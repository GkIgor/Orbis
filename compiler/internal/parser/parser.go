// Package parser implements a recursive descent parser for the Orbis DSL.
// It consumes tokens from the lexer and produces AST nodes.
//
// The parser follows the structural grammar defined in framework_proposal_v_1.md §25.2:
//   - <loop> requires for and index attributes
//   - <if> requires condition attribute
//   - Component tags start with uppercase letter
//   - Element tags start with lowercase letter
//   - Nested structures are unlimited
//
// The parser is deterministic, produces no side effects, and reports
// errors through the diagnostics collector without panicking.
package parser

import (
	"strings"
	"unicode"

	"github.com/orbisui/orbis/compiler/internal/ast"
	"github.com/orbisui/orbis/compiler/internal/diagnostics"
	"github.com/orbisui/orbis/compiler/internal/lexer"
)

// Parser produces an AST from a token stream.
// It is not safe for concurrent use.
type Parser struct {
	tokens      []lexer.Token
	pos         int
	file        string
	diagnostics *diagnostics.Collector
}

// New creates a new Parser for the given token stream.
func New(tokens []lexer.Token, file string, diag *diagnostics.Collector) *Parser {
	return &Parser{
		tokens:      tokens,
		pos:         0,
		file:        file,
		diagnostics: diag,
	}
}

// Parse parses the token stream and returns a list of top-level AST nodes.
func (p *Parser) Parse() []ast.Node {
	var nodes []ast.Node
	for !p.isEOF() {
		node := p.parseNode()
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// Diagnostics returns the diagnostic collector used by this parser.
func (p *Parser) Diagnostics() *diagnostics.Collector {
	return p.diagnostics
}

// current returns the current token.
func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

// peek returns the current token without advancing.
func (p *Parser) peek() lexer.Token {
	return p.current()
}

// advance moves to the next token and returns the previous one.
func (p *Parser) advance() lexer.Token {
	tok := p.current()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

// expect consumes the current token if it matches the expected type.
// Returns the token if matched, or emits a diagnostic and returns an empty token.
func (p *Parser) expect(tt lexer.TokenType) (lexer.Token, bool) {
	tok := p.current()
	if tok.Type == tt {
		p.advance()
		return tok, true
	}
	p.diagnostics.AddError(
		tok.Pos,
		"E100",
		"expected "+tt.String()+", got "+tok.Type.String()+" '"+tok.Value+"'",
	)
	return tok, false
}

// isEOF returns true if the current token is EOF.
func (p *Parser) isEOF() bool {
	return p.current().Type == lexer.TokenEOF
}

// parseNode dispatches to the appropriate parsing function based on the current token.
func (p *Parser) parseNode() ast.Node {
	tok := p.current()

	switch tok.Type {
	case lexer.TokenInterpolationOpen:
		return p.parseInterpolation()
	case lexer.TokenText:
		return p.parseText()
	case lexer.TokenTagOpen:
		return p.parseOpenTag()
	case lexer.TokenTagSlashOpen:
		// Unexpected closing tag at top level — skip and report
		p.diagnostics.AddError(tok.Pos, "E101", "unexpected closing tag")
		p.skipUntilTagClose()
		return nil
	default:
		// Skip unexpected tokens
		p.diagnostics.AddError(tok.Pos, "E102", "unexpected token: "+tok.Type.String()+" '"+tok.Value+"'")
		p.advance()
		return nil
	}
}

// parseText consumes a text token and returns a Text node.
func (p *Parser) parseText() ast.Node {
	tok := p.advance()
	// Skip whitespace-only text nodes
	if strings.TrimSpace(tok.Value) == "" {
		return nil
	}
	return &ast.Text{
		Content: tok.Value,
		Loc:     tok.Pos,
	}
}

// parseInterpolation parses {{ expression }}.
func (p *Parser) parseInterpolation() ast.Node {
	openTok := p.advance() // consume InterpolationOpen

	contentTok := p.current()
	expression := ""
	if contentTok.Type == lexer.TokenInterpolationContent {
		expression = contentTok.Value
		p.advance()
	}

	if p.current().Type == lexer.TokenInterpolationClose {
		p.advance()
	} else {
		p.diagnostics.AddError(openTok.Pos, "E103", "unclosed interpolation")
	}

	return &ast.Interpolation{
		Expression: expression,
		Loc:        openTok.Pos,
	}
}

// parseOpenTag parses an opening tag and its children.
func (p *Parser) parseOpenTag() ast.Node {
	openTok := p.advance() // consume TagOpen '<'

	// Expect tag name
	nameTok, ok := p.expect(lexer.TokenIdentifier)
	if !ok {
		p.skipUntilTagClose()
		return nil
	}

	tagName := nameTok.Value

	// Dispatch based on tag name
	switch tagName {
	case "loop":
		return p.parseLoop(openTok.Pos)
	case "if":
		return p.parseIf(openTok.Pos)
	default:
		return p.parseElementOrComponent(tagName, openTok.Pos)
	}
}

// parseElementOrComponent parses an HTML element or component tag.
func (p *Parser) parseElementOrComponent(tagName string, loc diagnostics.SourceLocation) ast.Node {
	attrs, events := p.parseAttributes()

	// Check for self-closing
	if p.current().Type == lexer.TokenTagSelfClose {
		p.advance()
		if isComponent(tagName) {
			return &ast.Component{
				Name:          tagName,
				Attributes:    attrs,
				EventBindings: events,
				Loc:           loc,
				SelfClosing:   true,
			}
		}
		return &ast.Element{
			Tag:           tagName,
			Attributes:    attrs,
			EventBindings: events,
			Loc:           loc,
			SelfClosing:   true,
		}
	}

	// Expect >
	if _, ok := p.expect(lexer.TokenTagClose); !ok {
		p.skipUntilTagClose()
	}

	// Parse children until closing tag
	children := p.parseChildren(tagName)

	// Expect closing tag </tagName>
	p.expectClosingTag(tagName)

	if isComponent(tagName) {
		return &ast.Component{
			Name:          tagName,
			Attributes:    attrs,
			EventBindings: events,
			Children:      children,
			Loc:           loc,
		}
	}
	return &ast.Element{
		Tag:           tagName,
		Attributes:    attrs,
		EventBindings: events,
		Children:      children,
		Loc:           loc,
	}
}

// parseLoop parses a <loop for="item in items" index="i"> ... </loop> construct.
func (p *Parser) parseLoop(loc diagnostics.SourceLocation) ast.Node {
	attrs, _ := p.parseAttributes()

	// Extract required attributes
	forAttr := findAttr(attrs, "for")
	indexAttr := findAttr(attrs, "index")

	if forAttr == "" {
		p.diagnostics.AddError(loc, "E104", "<loop> requires 'for' attribute (e.g., for=\"item in items\")")
	}
	if indexAttr == "" {
		p.diagnostics.AddError(loc, "E105", "<loop> requires 'index' attribute (e.g., index=\"i\")")
	}

	// Parse the 'for' expression: "item in collection"
	iterator, collection := parseForExpression(forAttr)

	// Expect >
	if _, ok := p.expect(lexer.TokenTagClose); !ok {
		p.skipUntilTagClose()
	}

	// Parse children
	children := p.parseChildren("loop")

	// Expect </loop>
	p.expectClosingTag("loop")

	return &ast.Loop{
		Iterator:   iterator,
		Collection: collection,
		Index:      indexAttr,
		Children:   children,
		Loc:        loc,
	}
}

// parseIf parses an <if condition="expression"> ... </if> construct.
func (p *Parser) parseIf(loc diagnostics.SourceLocation) ast.Node {
	attrs, _ := p.parseAttributes()

	// Extract required attribute
	condition := findAttr(attrs, "condition")
	if condition == "" {
		p.diagnostics.AddError(loc, "E106", "<if> requires 'condition' attribute")
	}

	// Expect >
	if _, ok := p.expect(lexer.TokenTagClose); !ok {
		p.skipUntilTagClose()
	}

	// Parse children
	children := p.parseChildren("if")

	// Expect </if>
	p.expectClosingTag("if")

	return &ast.If{
		Condition: condition,
		Children:  children,
		Loc:       loc,
	}
}

// parseAttributes parses attributes and event bindings within a tag.
func (p *Parser) parseAttributes() ([]ast.Attribute, []ast.EventBinding) {
	var attrs []ast.Attribute
	var events []ast.EventBinding

	for {
		tok := p.current()

		// Stop at tag close tokens
		if tok.Type == lexer.TokenTagClose || tok.Type == lexer.TokenTagSelfClose || tok.Type == lexer.TokenEOF {
			break
		}

		// Event binding: (event)="handler"
		if tok.Type == lexer.TokenEventBindOpen {
			event := p.parseEventBinding()
			if event != nil {
				events = append(events, *event)
			}
			continue
		}

		// Normal attribute: name="value"
		if tok.Type == lexer.TokenIdentifier {
			attr := p.parseAttribute()
			if attr != nil {
				attrs = append(attrs, *attr)
			}
			continue
		}

		// Skip unexpected tokens inside attribute list
		p.diagnostics.AddError(tok.Pos, "E107", "unexpected token in attribute list: "+tok.Type.String())
		p.advance()
	}

	return attrs, events
}

// parseAttribute parses a single attribute: name="value".
func (p *Parser) parseAttribute() *ast.Attribute {
	nameTok := p.advance() // consume identifier

	// Check for = sign
	if p.current().Type != lexer.TokenEquals {
		// Attribute without value (boolean attribute)
		return &ast.Attribute{
			Name:     nameTok.Value,
			Value:    "",
			Location: nameTok.Pos,
		}
	}
	p.advance() // consume '='

	// Expect string value
	if p.current().Type != lexer.TokenString {
		p.diagnostics.AddError(nameTok.Pos, "E108", "expected attribute value after '='")
		return nil
	}
	valueTok := p.advance()

	return &ast.Attribute{
		Name:     nameTok.Value,
		Value:    valueTok.Value,
		Location: nameTok.Pos,
	}
}

// parseEventBinding parses an event binding: (event)="handler()".
func (p *Parser) parseEventBinding() *ast.EventBinding {
	openTok := p.advance() // consume '('

	// Expect event name
	if p.current().Type != lexer.TokenIdentifier {
		p.diagnostics.AddError(openTok.Pos, "E109", "expected event name after '('")
		return nil
	}
	nameTok := p.advance()

	// Expect ')'
	if p.current().Type != lexer.TokenEventBindClose {
		p.diagnostics.AddError(openTok.Pos, "E110", "expected ')' after event name")
		return nil
	}
	p.advance() // consume ')'

	// Expect '='
	if p.current().Type != lexer.TokenEquals {
		p.diagnostics.AddError(openTok.Pos, "E111", "expected '=' after event binding")
		return nil
	}
	p.advance() // consume '='

	// Expect handler string
	if p.current().Type != lexer.TokenString {
		p.diagnostics.AddError(openTok.Pos, "E112", "expected handler string after '='")
		return nil
	}
	handlerTok := p.advance()

	return &ast.EventBinding{
		Event:    nameTok.Value,
		Handler:  handlerTok.Value,
		Location: openTok.Pos,
	}
}

// parseChildren parses child nodes until a closing tag for the given parent is found.
func (p *Parser) parseChildren(parentTag string) []ast.Node {
	var children []ast.Node

	for !p.isEOF() {
		// Check for closing tag
		if p.current().Type == lexer.TokenTagSlashOpen {
			// Peek ahead to see if this is our closing tag
			if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.TokenIdentifier &&
				p.tokens[p.pos+1].Value == parentTag {
				break
			}
		}

		node := p.parseNode()
		if node != nil {
			children = append(children, node)
		}
	}

	return children
}

// expectClosingTag expects and consumes a closing tag </tagName>.
func (p *Parser) expectClosingTag(tagName string) {
	if p.isEOF() {
		p.diagnostics.AddError(
			p.current().Pos,
			"E113",
			"expected closing tag </"+tagName+">, but reached end of input",
		)
		return
	}

	if _, ok := p.expect(lexer.TokenTagSlashOpen); !ok {
		return
	}

	if p.current().Type == lexer.TokenIdentifier && p.current().Value == tagName {
		p.advance()
	} else {
		p.diagnostics.AddError(
			p.current().Pos,
			"E114",
			"expected </"+tagName+">, got </"+p.current().Value+">",
		)
	}

	if p.current().Type == lexer.TokenTagClose {
		p.advance()
	}
}

// skipUntilTagClose advances until a TagClose or TagSelfClose token is found.
func (p *Parser) skipUntilTagClose() {
	for !p.isEOF() {
		tok := p.current()
		if tok.Type == lexer.TokenTagClose || tok.Type == lexer.TokenTagSelfClose {
			p.advance()
			return
		}
		p.advance()
	}
}

// isComponent returns true if the tag name starts with an uppercase letter.
func isComponent(tagName string) bool {
	if len(tagName) == 0 {
		return false
	}
	return unicode.IsUpper(rune(tagName[0]))
}

// findAttr searches for an attribute by name and returns its value.
func findAttr(attrs []ast.Attribute, name string) string {
	for _, a := range attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

// parseForExpression parses "item in collection" into iterator and collection.
func parseForExpression(expr string) (string, string) {
	parts := strings.SplitN(expr, " in ", 2)
	if len(parts) != 2 {
		return expr, ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
