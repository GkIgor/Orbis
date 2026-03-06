// Package lexer implements the DSL tokenizer for the Orbis compiler.
// It produces a stream of tokens from Orbis template source code.
//
// Supported constructs per framework_proposal_v_1.md §25:
//   - HTML tags (open, close, self-closing)
//   - <loop> and <if> structural tags
//   - {{ interpolation }}
//   - Attributes (name="value")
//   - Event bindings ((event)="handler()")
//
// The lexer is deterministic, tracks line/column positions, and emits
// diagnostics for malformed input without panicking.
package lexer

import (
	"github.com/orbisui/orbis/compiler/internal/diagnostics"
)

// TokenType identifies the kind of token.
type TokenType int

const (
	// TokenEOF signals the end of input.
	TokenEOF TokenType = iota
	// TokenText represents raw text content between tags/interpolations.
	TokenText
	// TokenTagOpen represents '<' when opening a tag.
	TokenTagOpen
	// TokenTagClose represents '>'.
	TokenTagClose
	// TokenTagSlashOpen represents '</' for closing tags.
	TokenTagSlashOpen
	// TokenTagSelfClose represents '/>' for self-closing tags.
	TokenTagSelfClose
	// TokenIdentifier represents a tag name or attribute name.
	TokenIdentifier
	// TokenEquals represents '='.
	TokenEquals
	// TokenString represents a quoted attribute value (includes quotes in raw value).
	TokenString
	// TokenInterpolationOpen represents '{{'.
	TokenInterpolationOpen
	// TokenInterpolationClose represents '}}'.
	TokenInterpolationClose
	// TokenInterpolationContent represents the expression inside {{ }}.
	TokenInterpolationContent
	// TokenEventBindOpen represents '(' in event bindings.
	TokenEventBindOpen
	// TokenEventBindClose represents ')' in event bindings.
	TokenEventBindClose
)

// String returns the string representation of a TokenType.
func (tt TokenType) String() string {
	switch tt {
	case TokenEOF:
		return "EOF"
	case TokenText:
		return "Text"
	case TokenTagOpen:
		return "TagOpen"
	case TokenTagClose:
		return "TagClose"
	case TokenTagSlashOpen:
		return "TagSlashOpen"
	case TokenTagSelfClose:
		return "TagSelfClose"
	case TokenIdentifier:
		return "Identifier"
	case TokenEquals:
		return "Equals"
	case TokenString:
		return "String"
	case TokenInterpolationOpen:
		return "InterpolationOpen"
	case TokenInterpolationClose:
		return "InterpolationClose"
	case TokenInterpolationContent:
		return "InterpolationContent"
	case TokenEventBindOpen:
		return "EventBindOpen"
	case TokenEventBindClose:
		return "EventBindClose"
	default:
		return "Unknown"
	}
}

// Token represents a single lexical token.
type Token struct {
	Type  TokenType                  `json:"type"`
	Value string                     `json:"value"`
	Pos   diagnostics.SourceLocation `json:"position"`
}

// Lexer performs tokenization of Orbis DSL template source.
// It is not safe for concurrent use.
type Lexer struct {
	input       []rune
	pos         int
	line        int
	col         int
	file        string
	diagnostics *diagnostics.Collector
	inTag       bool    // true when scanning inside a tag (between < and >)
	buffer      []Token // internal buffer for multi-token productions
}

// New creates a new Lexer for the given source input and file name.
func New(input string, file string, diag *diagnostics.Collector) *Lexer {
	return &Lexer{
		input:       []rune(input),
		pos:         0,
		line:        1,
		col:         1,
		file:        file,
		diagnostics: diag,
		inTag:       false,
	}
}

// Tokenize processes the entire input and returns all tokens.
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.nextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}

// Diagnostics returns the diagnostic collector used by this lexer.
func (l *Lexer) Diagnostics() *diagnostics.Collector {
	return l.diagnostics
}

// currentLoc returns the current source location.
func (l *Lexer) currentLoc() diagnostics.SourceLocation {
	return diagnostics.SourceLocation{
		File:   l.file,
		Line:   l.line,
		Column: l.col,
	}
}

// peek returns the current rune without advancing.
func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

// peekAt returns the rune at the given offset from current position.
func (l *Lexer) peekAt(offset int) rune {
	idx := l.pos + offset
	if idx >= len(l.input) || idx < 0 {
		return 0
	}
	return l.input[idx]
}

// advance moves forward one rune and updates line/column tracking.
func (l *Lexer) advance() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

// nextToken returns the next token from the input.
func (l *Lexer) nextToken() Token {
	// Drain the buffer first (used by multi-token productions).
	if len(l.buffer) > 0 {
		tok := l.buffer[0]
		l.buffer = l.buffer[1:]
		return tok
	}

	if l.pos >= len(l.input) {
		if l.inTag {
			l.diagnostics.AddError(l.currentLoc(), "E002", "unexpected end of input inside tag")
			l.inTag = false
		}
		return Token{Type: TokenEOF, Value: "", Pos: l.currentLoc()}
	}

	if l.inTag {
		return l.scanInsideTag()
	}
	return l.scanOutsideTag()
}

// scanOutsideTag scans tokens when we're outside any HTML tag.
func (l *Lexer) scanOutsideTag() Token {
	// Check for interpolation opening {{
	if l.peek() == '{' && l.peekAt(1) == '{' {
		return l.scanInterpolation()
	}

	// Check for tag opening
	if l.peek() == '<' {
		return l.scanTagStart()
	}

	// Otherwise, scan raw text
	return l.scanText()
}

// scanText scans raw text content until a tag or interpolation is encountered.
func (l *Lexer) scanText() Token {
	loc := l.currentLoc()
	var text []rune

	for l.pos < len(l.input) {
		// Stop at interpolation
		if l.peek() == '{' && l.peekAt(1) == '{' {
			break
		}
		// Stop at tag
		if l.peek() == '<' {
			break
		}
		text = append(text, l.advance())
	}

	return Token{Type: TokenText, Value: string(text), Pos: loc}
}

// scanInterpolation scans {{ expression }}.
func (l *Lexer) scanInterpolation() Token {
	loc := l.currentLoc()

	// Consume {{
	l.advance()
	l.advance()

	openTok := Token{Type: TokenInterpolationOpen, Value: "{{", Pos: loc}

	// We return the {{ token first — the caller will get content and }} on subsequent calls.
	// Instead, we'll return all three as part of the tokenize loop.
	// Actually, let's scan the full interpolation and produce multiple tokens by switching
	// to returning the open token and letting next calls handle content/close.
	// For simplicity, we just return the open token; the next call will scan content.

	// But we need a way to track we're in interpolation mode.
	// Let's use a simpler approach: scan the entire interpolation here and push tokens.
	// We'll refactor to use a token buffer.

	// Actually, the cleanest design: return open token, and have nextToken detect we just
	// opened an interpolation. We track this with a state flag.

	// Let's use the simplest correct approach: scan everything here, return as 3 tokens
	// via a buffer.
	_ = openTok

	// Scan content
	contentLoc := l.currentLoc()
	var content []rune
	for l.pos < len(l.input) {
		if l.peek() == '}' && l.peekAt(1) == '}' {
			break
		}
		content = append(content, l.advance())
	}

	contentStr := trimWhitespace(string(content))

	// Consume }}
	closeLoc := l.currentLoc()
	if l.peek() == '}' && l.peekAt(1) == '}' {
		l.advance()
		l.advance()
	} else {
		l.diagnostics.AddError(loc, "E001", "unclosed interpolation '{{', expected '}}'")
	}

	// Push content and close tokens into buffer, return open token.
	l.pushToken(Token{Type: TokenInterpolationContent, Value: contentStr, Pos: contentLoc})
	l.pushToken(Token{Type: TokenInterpolationClose, Value: "}}", Pos: closeLoc})

	return openTok
}

// pushToken adds a token to the internal buffer.
func (l *Lexer) pushToken(tok Token) {
	l.buffer = append(l.buffer, tok)
}

// scanTagStart scans the beginning of a tag (< or </).
func (l *Lexer) scanTagStart() Token {
	loc := l.currentLoc()

	l.advance() // consume '<'

	// Check for closing tag </
	if l.peek() == '/' {
		l.advance() // consume '/'
		l.inTag = true
		return Token{Type: TokenTagSlashOpen, Value: "</", Pos: loc}
	}

	l.inTag = true
	return Token{Type: TokenTagOpen, Value: "<", Pos: loc}
}

// scanInsideTag scans tokens within a tag (attributes, >, />, identifiers, etc.).
func (l *Lexer) scanInsideTag() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		l.diagnostics.AddError(l.currentLoc(), "E002", "unexpected end of input inside tag")
		l.inTag = false
		return Token{Type: TokenEOF, Value: "", Pos: l.currentLoc()}
	}

	ch := l.peek()

	// Self-closing />
	if ch == '/' && l.peekAt(1) == '>' {
		loc := l.currentLoc()
		l.advance()
		l.advance()
		l.inTag = false
		return Token{Type: TokenTagSelfClose, Value: "/>", Pos: loc}
	}

	// Close tag >
	if ch == '>' {
		loc := l.currentLoc()
		l.advance()
		l.inTag = false
		return Token{Type: TokenTagClose, Value: ">", Pos: loc}
	}

	// Equals
	if ch == '=' {
		loc := l.currentLoc()
		l.advance()
		return Token{Type: TokenEquals, Value: "=", Pos: loc}
	}

	// Quoted string
	if ch == '"' || ch == '\'' {
		return l.scanString()
	}

	// Event binding (event)
	if ch == '(' {
		return l.scanEventBinding()
	}

	// Identifier (tag name or attribute name)
	if isIdentStart(ch) {
		return l.scanIdentifier()
	}

	// Unexpected character
	loc := l.currentLoc()
	l.diagnostics.AddError(loc, "E003", "unexpected character '"+string(ch)+"' inside tag")
	l.advance()
	return l.scanInsideTag()
}

// scanIdentifier scans an alphanumeric identifier.
func (l *Lexer) scanIdentifier() Token {
	loc := l.currentLoc()
	var ident []rune

	for l.pos < len(l.input) && isIdentChar(l.peek()) {
		ident = append(ident, l.advance())
	}

	return Token{Type: TokenIdentifier, Value: string(ident), Pos: loc}
}

// scanString scans a quoted string value.
func (l *Lexer) scanString() Token {
	loc := l.currentLoc()
	quote := l.advance() // consume opening quote
	var content []rune

	for l.pos < len(l.input) && l.peek() != quote {
		content = append(content, l.advance())
	}

	if l.pos < len(l.input) {
		l.advance() // consume closing quote
	} else {
		l.diagnostics.AddError(loc, "E004", "unclosed string literal")
	}

	return Token{Type: TokenString, Value: string(content), Pos: loc}
}

// scanEventBinding scans an event binding like (click).
func (l *Lexer) scanEventBinding() Token {
	loc := l.currentLoc()
	l.advance() // consume '('

	openTok := Token{Type: TokenEventBindOpen, Value: "(", Pos: loc}

	// Scan event name
	nameLoc := l.currentLoc()
	var name []rune
	for l.pos < len(l.input) && l.peek() != ')' && isIdentChar(l.peek()) {
		name = append(name, l.advance())
	}

	// Consume ')'
	closeLoc := l.currentLoc()
	if l.pos < len(l.input) && l.peek() == ')' {
		l.advance()
	} else {
		l.diagnostics.AddError(loc, "E005", "unclosed event binding '(', expected ')'")
	}

	l.pushToken(Token{Type: TokenIdentifier, Value: string(name), Pos: nameLoc})
	l.pushToken(Token{Type: TokenEventBindClose, Value: ")", Pos: closeLoc})

	return openTok
}

// skipWhitespace advances past any whitespace characters.
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

// isIdentStart returns true if the rune can start an identifier.
func isIdentStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// isIdentChar returns true if the rune can appear in an identifier.
func isIdentChar(ch rune) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9') || ch == '-'
}

// trimWhitespace trims leading and trailing whitespace from a string.
func trimWhitespace(s string) string {
	runes := []rune(s)
	start := 0
	end := len(runes)
	for start < end && isWhitespace(runes[start]) {
		start++
	}
	for end > start && isWhitespace(runes[end-1]) {
		end--
	}
	return string(runes[start:end])
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
