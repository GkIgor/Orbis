package lexer_test

import (
	"testing"

	"github.com/orbisui/orbis/compiler/internal/diagnostics"
	"github.com/orbisui/orbis/compiler/internal/lexer"
)

// helper creates a lexer, tokenizes, and returns tokens + diagnostics.
func tokenize(t *testing.T, input string) ([]lexer.Token, *diagnostics.Collector) {
	t.Helper()
	diag := diagnostics.NewCollector()
	l := lexer.New(input, "test.html", diag)
	tokens := l.Tokenize()
	return tokens, diag
}

// expectTokenTypes checks that the token types match the expected sequence.
func expectTokenTypes(t *testing.T, tokens []lexer.Token, expected []lexer.TokenType) {
	t.Helper()
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d\nTokens: %v", len(expected), len(tokens), tokenTypesStr(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token[%d]: expected %s, got %s (value=%q)", i, expected[i], tok.Type, tok.Value)
		}
	}
}

func tokenTypesStr(tokens []lexer.Token) []string {
	var result []string
	for _, t := range tokens {
		result = append(result, t.Type.String()+"("+t.Value+")")
	}
	return result
}

// --- Simple HTML Tags ---

func TestSimpleTag(t *testing.T) {
	tokens, diag := tokenize(t, "<div></div>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenTagOpen,      // <
		lexer.TokenIdentifier,   // div
		lexer.TokenTagClose,     // >
		lexer.TokenTagSlashOpen, // </
		lexer.TokenIdentifier,   // div
		lexer.TokenTagClose,     // >
		lexer.TokenEOF,
	})
}

func TestSelfClosingTag(t *testing.T) {
	tokens, diag := tokenize(t, "<br/>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenTagOpen,      // <
		lexer.TokenIdentifier,   // br
		lexer.TokenTagSelfClose, // />
		lexer.TokenEOF,
	})
}

func TestNestedTags(t *testing.T) {
	tokens, diag := tokenize(t, "<div><span></span></div>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenTagOpen, lexer.TokenIdentifier, lexer.TokenTagClose, // <div>
		lexer.TokenTagOpen, lexer.TokenIdentifier, lexer.TokenTagClose, // <span>
		lexer.TokenTagSlashOpen, lexer.TokenIdentifier, lexer.TokenTagClose, // </span>
		lexer.TokenTagSlashOpen, lexer.TokenIdentifier, lexer.TokenTagClose, // </div>
		lexer.TokenEOF,
	})
}

// --- Attributes ---

func TestTagWithAttribute(t *testing.T) {
	tokens, diag := tokenize(t, `<div class="container"></div>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenTagOpen,      // <
		lexer.TokenIdentifier,   // div
		lexer.TokenIdentifier,   // class
		lexer.TokenEquals,       // =
		lexer.TokenString,       // container
		lexer.TokenTagClose,     // >
		lexer.TokenTagSlashOpen, // </
		lexer.TokenIdentifier,   // div
		lexer.TokenTagClose,     // >
		lexer.TokenEOF,
	})
	// Verify attribute value
	if tokens[4].Value != "container" {
		t.Errorf("attribute value = %q, want %q", tokens[4].Value, "container")
	}
}

func TestMultipleAttributes(t *testing.T) {
	tokens, diag := tokenize(t, `<div class="a" id="b"></div>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	// <div class="a" id="b">
	// TagOpen Ident Ident Eq Str Ident Eq Str TagClose ...
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenTagOpen, lexer.TokenIdentifier,
		lexer.TokenIdentifier, lexer.TokenEquals, lexer.TokenString,
		lexer.TokenIdentifier, lexer.TokenEquals, lexer.TokenString,
		lexer.TokenTagClose,
		lexer.TokenTagSlashOpen, lexer.TokenIdentifier, lexer.TokenTagClose,
		lexer.TokenEOF,
	})
}

// --- Event Bindings ---

func TestEventBinding(t *testing.T) {
	tokens, diag := tokenize(t, `<button (click)="handler()"></button>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenTagOpen,        // <
		lexer.TokenIdentifier,     // button
		lexer.TokenEventBindOpen,  // (
		lexer.TokenIdentifier,     // click
		lexer.TokenEventBindClose, // )
		lexer.TokenEquals,         // =
		lexer.TokenString,         // handler()
		lexer.TokenTagClose,       // >
		lexer.TokenTagSlashOpen,   // </
		lexer.TokenIdentifier,     // button
		lexer.TokenTagClose,       // >
		lexer.TokenEOF,
	})
	// Verify event name
	if tokens[3].Value != "click" {
		t.Errorf("event name = %q, want %q", tokens[3].Value, "click")
	}
	// Verify handler
	if tokens[6].Value != "handler()" {
		t.Errorf("handler = %q, want %q", tokens[6].Value, "handler()")
	}
}

// --- Interpolation ---

func TestInterpolation(t *testing.T) {
	tokens, diag := tokenize(t, "{{ name }}")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenInterpolationOpen,
		lexer.TokenInterpolationContent,
		lexer.TokenInterpolationClose,
		lexer.TokenEOF,
	})
	if tokens[1].Value != "name" {
		t.Errorf("interpolation content = %q, want %q", tokens[1].Value, "name")
	}
}

func TestInterpolationPropertyAccess(t *testing.T) {
	tokens, diag := tokenize(t, "{{ config.title }}")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if tokens[1].Value != "config.title" {
		t.Errorf("interpolation content = %q, want %q", tokens[1].Value, "config.title")
	}
}

// --- Text Content ---

func TestTextContent(t *testing.T) {
	tokens, diag := tokenize(t, "Hello World")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenText,
		lexer.TokenEOF,
	})
	if tokens[0].Value != "Hello World" {
		t.Errorf("text = %q, want %q", tokens[0].Value, "Hello World")
	}
}

func TestMixedTextAndInterpolation(t *testing.T) {
	tokens, diag := tokenize(t, "Hello {{ name }}!")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenText,                 // "Hello "
		lexer.TokenInterpolationOpen,    // {{
		lexer.TokenInterpolationContent, // name
		lexer.TokenInterpolationClose,   // }}
		lexer.TokenText,                 // "!"
		lexer.TokenEOF,
	})
}

// --- Loop and If Tags ---

func TestLoopTag(t *testing.T) {
	tokens, diag := tokenize(t, `<loop for="item in items" index="i"></loop>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	// <loop for="item in items" index="i">
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenTagOpen, lexer.TokenIdentifier, // < loop
		lexer.TokenIdentifier, lexer.TokenEquals, lexer.TokenString, // for="item in items"
		lexer.TokenIdentifier, lexer.TokenEquals, lexer.TokenString, // index="i"
		lexer.TokenTagClose,                                                 // >
		lexer.TokenTagSlashOpen, lexer.TokenIdentifier, lexer.TokenTagClose, // </loop>
		lexer.TokenEOF,
	})
	// Verify tag name is "loop"
	if tokens[1].Value != "loop" {
		t.Errorf("tag name = %q, want %q", tokens[1].Value, "loop")
	}
}

func TestIfTag(t *testing.T) {
	tokens, diag := tokenize(t, `<if condition="app.visible"></if>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenTagOpen, lexer.TokenIdentifier, // < if
		lexer.TokenIdentifier, lexer.TokenEquals, lexer.TokenString, // condition="app.visible"
		lexer.TokenTagClose,                                                 // >
		lexer.TokenTagSlashOpen, lexer.TokenIdentifier, lexer.TokenTagClose, // </if>
		lexer.TokenEOF,
	})
}

// --- Position Tracking ---

func TestPositionTracking(t *testing.T) {
	tokens, _ := tokenize(t, "<div>\n  {{ x }}\n</div>")

	// First token '<' should be at line 1, col 1
	if tokens[0].Pos.Line != 1 || tokens[0].Pos.Column != 1 {
		t.Errorf("first token pos = %d:%d, want 1:1", tokens[0].Pos.Line, tokens[0].Pos.Column)
	}

	// Find the interpolation open {{ — it should be on line 2
	for _, tok := range tokens {
		if tok.Type == lexer.TokenInterpolationOpen {
			if tok.Pos.Line != 2 {
				t.Errorf("interpolation line = %d, want 2", tok.Pos.Line)
			}
			break
		}
	}

	// Find the closing </div> — it should be on line 3
	for _, tok := range tokens {
		if tok.Type == lexer.TokenTagSlashOpen {
			if tok.Pos.Line != 3 {
				t.Errorf("closing tag line = %d, want 3", tok.Pos.Line)
			}
			break
		}
	}
}

func TestFileNameTracking(t *testing.T) {
	diag := diagnostics.NewCollector()
	l := lexer.New("<div>", "myfile.html", diag)
	tokens := l.Tokenize()

	for _, tok := range tokens {
		if tok.Pos.File != "myfile.html" {
			t.Errorf("token file = %q, want %q", tok.Pos.File, "myfile.html")
		}
	}
}

// --- Error Cases ---

func TestUnclosedInterpolation(t *testing.T) {
	_, diag := tokenize(t, "{{ hello")
	if !diag.HasErrors() {
		t.Error("expected error for unclosed interpolation")
	}
	errs := diag.Errors()
	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}
	if errs[0].Code != "E001" {
		t.Errorf("error code = %q, want %q", errs[0].Code, "E001")
	}
}

func TestUnclosedString(t *testing.T) {
	_, diag := tokenize(t, `<div class="unclosed>`)
	if !diag.HasErrors() {
		t.Error("expected error for unclosed string")
	}
	errs := diag.Errors()
	found := false
	for _, e := range errs {
		if e.Code == "E004" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error code E004 for unclosed string")
	}
}

// --- Component Tags ---

func TestComponentTag(t *testing.T) {
	tokens, diag := tokenize(t, "<CounterItem></CounterItem>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	// Component tags are tokenized the same as elements — differentiation happens in the parser.
	if tokens[1].Value != "CounterItem" {
		t.Errorf("tag name = %q, want %q", tokens[1].Value, "CounterItem")
	}
}

// --- Combined Template ---

func TestFullTemplate(t *testing.T) {
	input := `<div class="container">
  <h1>{{ config.title }}</h1>
  <button (click)="toggle()">Toggle</button>
  <if condition="app.visible">
    <loop for="item in app.items" index="i">
      <CounterItem/>
    </loop>
  </if>
</div>`

	tokens, diag := tokenize(t, input)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}

	// Just verify we get a reasonable number of tokens and no errors
	if len(tokens) < 20 {
		t.Errorf("expected at least 20 tokens, got %d", len(tokens))
	}

	// Last token should be EOF
	if tokens[len(tokens)-1].Type != lexer.TokenEOF {
		t.Error("last token should be EOF")
	}
}

// --- Empty Input ---

func TestEmptyInput(t *testing.T) {
	tokens, diag := tokenize(t, "")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenEOF,
	})
}

// --- Edge Cases for Coverage ---

func TestUnexpectedEOFInsideTag(t *testing.T) {
	// Tag with attribute but no closing >, so scanInsideTag hits EOF
	_, diag := tokenize(t, `<div class="x"`)
	if !diag.HasErrors() {
		t.Error("expected error for EOF inside tag")
	}
	found := false
	for _, e := range diag.Errors() {
		if e.Code == "E002" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error code E002 for unexpected EOF inside tag")
	}
}

func TestUnexpectedCharInsideTag(t *testing.T) {
	// '@' is not a valid character inside a tag
	_, diag := tokenize(t, "<div @>")
	if !diag.HasErrors() {
		t.Error("expected error for unexpected character inside tag")
	}
	found := false
	for _, e := range diag.Errors() {
		if e.Code == "E003" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error code E003 for unexpected character")
	}
}

func TestUnclosedEventBinding(t *testing.T) {
	_, diag := tokenize(t, "<div (click>")
	if !diag.HasErrors() {
		t.Error("expected error for unclosed event binding")
	}
	found := false
	for _, e := range diag.Errors() {
		if e.Code == "E005" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error code E005 for unclosed event binding")
	}
}

func TestSingleQuoteString(t *testing.T) {
	tokens, diag := tokenize(t, "<div class='container'></div>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	// Find the string token and verify the value
	for _, tok := range tokens {
		if tok.Type == lexer.TokenString {
			if tok.Value != "container" {
				t.Errorf("single-quote string value = %q, want %q", tok.Value, "container")
			}
			return
		}
	}
	t.Error("expected a String token for single-quoted attribute")
}

func TestSingleOpenBraceNotInterpolation(t *testing.T) {
	tokens, diag := tokenize(t, "Hello { World")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	// Single '{' should be treated as text, not interpolation
	expectTokenTypes(t, tokens, []lexer.TokenType{
		lexer.TokenText,
		lexer.TokenEOF,
	})
	if tokens[0].Value != "Hello { World" {
		t.Errorf("text = %q, want %q", tokens[0].Value, "Hello { World")
	}
}

func TestTokenTypeUnknownString(t *testing.T) {
	tt := lexer.TokenType(999)
	if tt.String() != "Unknown" {
		t.Errorf("unknown TokenType.String() = %q, want %q", tt.String(), "Unknown")
	}
}

func TestDiagnosticsAccessor(t *testing.T) {
	diag := diagnostics.NewCollector()
	l := lexer.New("test", "file.html", diag)
	if l.Diagnostics() != diag {
		t.Error("Diagnostics() should return the same collector")
	}
}

func TestWhitespaceWithCarriageReturn(t *testing.T) {
	tokens, diag := tokenize(t, "<div\r\n  class=\"x\"\r\n></div>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	// Should successfully parse despite \r\n line endings
	foundDiv := false
	for _, tok := range tokens {
		if tok.Type == lexer.TokenIdentifier && tok.Value == "div" {
			foundDiv = true
			break
		}
	}
	if !foundDiv {
		t.Error("expected to find 'div' identifier")
	}
}
