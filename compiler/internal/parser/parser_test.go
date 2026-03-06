package parser_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/orbisui/orbis/compiler/internal/ast"
	"github.com/orbisui/orbis/compiler/internal/diagnostics"
	"github.com/orbisui/orbis/compiler/internal/lexer"
	"github.com/orbisui/orbis/compiler/internal/parser"
)

// helper: lex + parse
func parse(t *testing.T, input string) ([]ast.Node, *diagnostics.Collector) {
	t.Helper()
	diag := diagnostics.NewCollector()
	l := lexer.New(input, "test.html", diag)
	tokens := l.Tokenize()
	p := parser.New(tokens, "test.html", diag)
	nodes := p.Parse()
	return nodes, diag
}

// helper: serialize nodes to JSON for comparison
func nodesJSON(t *testing.T, nodes []ast.Node) string {
	t.Helper()
	data, err := ast.MarshalNodesJSON(nodes)
	if err != nil {
		t.Fatalf("failed to marshal AST: %v", err)
	}
	return string(data)
}

// --- Simple Element ---

func TestParseSimpleElement(t *testing.T) {
	nodes, diag := parse(t, "<div></div>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	elem, ok := nodes[0].(*ast.Element)
	if !ok {
		t.Fatalf("expected *ast.Element, got %T", nodes[0])
	}
	if elem.Tag != "div" {
		t.Errorf("tag = %q, want %q", elem.Tag, "div")
	}
}

func TestParseSelfClosingElement(t *testing.T) {
	nodes, diag := parse(t, "<br/>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	elem, ok := nodes[0].(*ast.Element)
	if !ok {
		t.Fatalf("expected *ast.Element, got %T", nodes[0])
	}
	if !elem.SelfClosing {
		t.Error("expected self-closing element")
	}
}

// --- Nested Elements ---

func TestParseNestedElements(t *testing.T) {
	nodes, diag := parse(t, "<div><span></span></div>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	elem := nodes[0].(*ast.Element)
	if len(elem.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(elem.Children))
	}
	child, ok := elem.Children[0].(*ast.Element)
	if !ok {
		t.Fatalf("expected child *ast.Element, got %T", elem.Children[0])
	}
	if child.Tag != "span" {
		t.Errorf("child tag = %q, want %q", child.Tag, "span")
	}
}

// --- Attributes ---

func TestParseWithAttributes(t *testing.T) {
	nodes, diag := parse(t, `<div class="container" id="main"></div>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	elem := nodes[0].(*ast.Element)
	if len(elem.Attributes) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(elem.Attributes))
	}
	if elem.Attributes[0].Name != "class" || elem.Attributes[0].Value != "container" {
		t.Errorf("attr[0] = %q=%q, want class=container", elem.Attributes[0].Name, elem.Attributes[0].Value)
	}
	if elem.Attributes[1].Name != "id" || elem.Attributes[1].Value != "main" {
		t.Errorf("attr[1] = %q=%q, want id=main", elem.Attributes[1].Name, elem.Attributes[1].Value)
	}
}

// --- Event Bindings ---

func TestParseEventBinding(t *testing.T) {
	nodes, diag := parse(t, `<button (click)="handler()"></button>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	elem := nodes[0].(*ast.Element)
	if len(elem.EventBindings) != 1 {
		t.Fatalf("expected 1 event binding, got %d", len(elem.EventBindings))
	}
	if elem.EventBindings[0].Event != "click" {
		t.Errorf("event = %q, want %q", elem.EventBindings[0].Event, "click")
	}
	if elem.EventBindings[0].Handler != "handler()" {
		t.Errorf("handler = %q, want %q", elem.EventBindings[0].Handler, "handler()")
	}
}

// --- Interpolation ---

func TestParseInterpolation(t *testing.T) {
	nodes, diag := parse(t, "{{ config.title }}")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	interp, ok := nodes[0].(*ast.Interpolation)
	if !ok {
		t.Fatalf("expected *ast.Interpolation, got %T", nodes[0])
	}
	if interp.Expression != "config.title" {
		t.Errorf("expression = %q, want %q", interp.Expression, "config.title")
	}
}

func TestParseInterpolationInsideElement(t *testing.T) {
	nodes, diag := parse(t, "<h1>{{ title }}</h1>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	elem := nodes[0].(*ast.Element)
	if len(elem.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(elem.Children))
	}
	interp, ok := elem.Children[0].(*ast.Interpolation)
	if !ok {
		t.Fatalf("expected *ast.Interpolation, got %T", elem.Children[0])
	}
	if interp.Expression != "title" {
		t.Errorf("expression = %q, want %q", interp.Expression, "title")
	}
}

// --- Component Tags ---

func TestParseComponentTag(t *testing.T) {
	nodes, diag := parse(t, "<CounterItem></CounterItem>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	comp, ok := nodes[0].(*ast.Component)
	if !ok {
		t.Fatalf("expected *ast.Component, got %T", nodes[0])
	}
	if comp.Name != "CounterItem" {
		t.Errorf("name = %q, want %q", comp.Name, "CounterItem")
	}
}

func TestParseSelfClosingComponent(t *testing.T) {
	nodes, diag := parse(t, "<CounterItem/>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	comp, ok := nodes[0].(*ast.Component)
	if !ok {
		t.Fatalf("expected *ast.Component, got %T", nodes[0])
	}
	if !comp.SelfClosing {
		t.Error("expected self-closing component")
	}
}

func TestParseComponentWithEventBinding(t *testing.T) {
	nodes, diag := parse(t, `<CounterItem (click)="select(i)"></CounterItem>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	comp := nodes[0].(*ast.Component)
	if len(comp.EventBindings) != 1 {
		t.Fatalf("expected 1 event binding, got %d", len(comp.EventBindings))
	}
	if comp.EventBindings[0].Handler != "select(i)" {
		t.Errorf("handler = %q, want %q", comp.EventBindings[0].Handler, "select(i)")
	}
}

// --- Loop ---

func TestParseLoop(t *testing.T) {
	nodes, diag := parse(t, `<loop for="item in app.items" index="i"><div></div></loop>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	loop, ok := nodes[0].(*ast.Loop)
	if !ok {
		t.Fatalf("expected *ast.Loop, got %T", nodes[0])
	}
	if loop.Iterator != "item" {
		t.Errorf("iterator = %q, want %q", loop.Iterator, "item")
	}
	if loop.Collection != "app.items" {
		t.Errorf("collection = %q, want %q", loop.Collection, "app.items")
	}
	if loop.Index != "i" {
		t.Errorf("index = %q, want %q", loop.Index, "i")
	}
	if len(loop.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(loop.Children))
	}
}

func TestParseLoopMissingFor(t *testing.T) {
	_, diag := parse(t, `<loop index="i"></loop>`)
	if !diag.HasErrors() {
		t.Error("expected error for missing 'for' attribute on <loop>")
	}
	found := false
	for _, e := range diag.Errors() {
		if e.Code == "E104" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error code E104")
	}
}

func TestParseLoopMissingIndex(t *testing.T) {
	_, diag := parse(t, `<loop for="item in items"></loop>`)
	if !diag.HasErrors() {
		t.Error("expected error for missing 'index' attribute on <loop>")
	}
	found := false
	for _, e := range diag.Errors() {
		if e.Code == "E105" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error code E105")
	}
}

// --- If ---

func TestParseIf(t *testing.T) {
	nodes, diag := parse(t, `<if condition="app.visible"><div></div></if>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	ifNode, ok := nodes[0].(*ast.If)
	if !ok {
		t.Fatalf("expected *ast.If, got %T", nodes[0])
	}
	if ifNode.Condition != "app.visible" {
		t.Errorf("condition = %q, want %q", ifNode.Condition, "app.visible")
	}
	if len(ifNode.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(ifNode.Children))
	}
}

func TestParseIfMissingCondition(t *testing.T) {
	_, diag := parse(t, `<if><div></div></if>`)
	if !diag.HasErrors() {
		t.Error("expected error for missing 'condition' attribute on <if>")
	}
	found := false
	for _, e := range diag.Errors() {
		if e.Code == "E106" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error code E106")
	}
}

// --- Text ---

func TestParseTextNode(t *testing.T) {
	nodes, diag := parse(t, "<div>Hello World</div>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	elem := nodes[0].(*ast.Element)
	if len(elem.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(elem.Children))
	}
	text, ok := elem.Children[0].(*ast.Text)
	if !ok {
		t.Fatalf("expected *ast.Text, got %T", elem.Children[0])
	}
	if text.Content != "Hello World" {
		t.Errorf("content = %q, want %q", text.Content, "Hello World")
	}
}

// --- Mixed Content ---

func TestParseMixedContent(t *testing.T) {
	nodes, diag := parse(t, "<div>Hello {{ name }}!</div>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	elem := nodes[0].(*ast.Element)
	// Should have: Text("Hello ") + Interpolation("name") + Text("!")
	if len(elem.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(elem.Children))
	}
	if elem.Children[0].Type() != ast.TextNodeType {
		t.Errorf("child[0] type = %s, want Text", elem.Children[0].Type())
	}
	if elem.Children[1].Type() != ast.InterpolationNodeType {
		t.Errorf("child[1] type = %s, want Interpolation", elem.Children[1].Type())
	}
	if elem.Children[2].Type() != ast.TextNodeType {
		t.Errorf("child[2] type = %s, want Text", elem.Children[2].Type())
	}
}

// --- AST Snapshot Tests ---

func TestSnapshotSimpleElement(t *testing.T) {
	nodes, diag := parse(t, `<div class="container"></div>`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	compareSnapshot(t, "simple_element.json", nodes)
}

func TestSnapshotLoopWithChildren(t *testing.T) {
	input := `<loop for="item in items" index="i"><p>{{ item }}</p></loop>`
	nodes, diag := parse(t, input)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	compareSnapshot(t, "loop_with_children.json", nodes)
}

func TestSnapshotIfWithComponent(t *testing.T) {
	input := `<if condition="app.visible"><CounterItem (click)="select(i)"/></if>`
	nodes, diag := parse(t, input)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	compareSnapshot(t, "if_with_component.json", nodes)
}

func TestSnapshotFullTemplate(t *testing.T) {
	input := `<div class="p-6 bg-gray-100 container">
  <h1 class="title">{{ config.title }}</h1>
  <button class="btn" (click)="toggle()">Toggle List</button>
  <if condition="app.visible">
    <div class="mt-4">
      <loop for="item in app.items" index="i">
        <CounterItem (click)="select(i)"/>
      </loop>
    </div>
  </if>
</div>`
	nodes, diag := parse(t, input)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	compareSnapshot(t, "full_template.json", nodes)
}

// compareSnapshot compares AST output against a golden file.
// If the golden file doesn't exist, it creates it (update mode).
// Set UPDATE_SNAPSHOTS=1 to regenerate golden files.
func compareSnapshot(t *testing.T, name string, nodes []ast.Node) {
	t.Helper()

	actual := nodesJSON(t, nodes)
	goldenPath := filepath.Join("testdata", name)

	if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
		if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("updated golden file: %s", goldenPath)
		return
	}

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		// Golden file doesn't exist — create it
		if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("created golden file: %s", goldenPath)
		return
	}

	// Normalize JSON for comparison
	var expectedObj, actualObj interface{}
	if err := json.Unmarshal(golden, &expectedObj); err != nil {
		t.Fatalf("failed to parse golden file %s: %v", name, err)
	}
	if err := json.Unmarshal([]byte(actual), &actualObj); err != nil {
		t.Fatalf("failed to parse actual AST: %v", err)
	}

	expectedJSON, _ := json.MarshalIndent(expectedObj, "", "  ")
	actualJSON, _ := json.MarshalIndent(actualObj, "", "  ")

	if string(expectedJSON) != string(actualJSON) {
		t.Errorf("snapshot mismatch for %s\n--- expected ---\n%s\n--- actual ---\n%s",
			name, string(expectedJSON), string(actualJSON))
	}
}

// --- Parser Edge Cases for Coverage ---

func TestParseUnclosedTag(t *testing.T) {
	_, diag := parse(t, "<div>")
	if !diag.HasErrors() {
		t.Error("expected error for unclosed tag")
	}
	found := false
	for _, e := range diag.Errors() {
		if e.Code == "E113" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error code E113 for missing closing tag")
	}
}

func TestParseUnexpectedClosingTag(t *testing.T) {
	_, diag := parse(t, "</div>")
	if !diag.HasErrors() {
		t.Error("expected error for unexpected closing tag")
	}
	found := false
	for _, e := range diag.Errors() {
		if e.Code == "E101" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error code E101 for unexpected closing tag")
	}
}

func TestParseBooleanAttribute(t *testing.T) {
	nodes, diag := parse(t, "<div disabled></div>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	elem := nodes[0].(*ast.Element)
	if len(elem.Attributes) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(elem.Attributes))
	}
	if elem.Attributes[0].Name != "disabled" {
		t.Errorf("attr name = %q, want %q", elem.Attributes[0].Name, "disabled")
	}
	if elem.Attributes[0].Value != "" {
		t.Errorf("boolean attr should have empty value, got %q", elem.Attributes[0].Value)
	}
}

func TestParserDiagnosticsAccessor(t *testing.T) {
	diag := diagnostics.NewCollector()
	l := lexer.New("<div></div>", "test.html", diag)
	tokens := l.Tokenize()
	p := parser.New(tokens, "test.html", diag)
	if p.Diagnostics() != diag {
		t.Error("Diagnostics() should return the same collector")
	}
}

func TestParseEmptyTemplate(t *testing.T) {
	nodes, diag := parse(t, "")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes for empty template, got %d", len(nodes))
	}
}

func TestParseComponentWithChildren(t *testing.T) {
	nodes, diag := parse(t, "<MyComponent><span></span></MyComponent>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	comp := nodes[0].(*ast.Component)
	if comp.Name != "MyComponent" {
		t.Errorf("name = %q, want %q", comp.Name, "MyComponent")
	}
	if len(comp.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(comp.Children))
	}
}

func TestParseDeepNesting(t *testing.T) {
	input := `<div><span><p><em><strong>text</strong></em></p></span></div>`
	nodes, diag := parse(t, input)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	// Verify deep nesting works
	elem := nodes[0].(*ast.Element)
	if elem.Tag != "div" {
		t.Errorf("root tag = %q, want div", elem.Tag)
	}
	// Navigate down
	span := elem.Children[0].(*ast.Element)
	p := span.Children[0].(*ast.Element)
	em := p.Children[0].(*ast.Element)
	strong := em.Children[0].(*ast.Element)
	if strong.Tag != "strong" {
		t.Errorf("deepest tag = %q, want strong", strong.Tag)
	}
}

func TestParseMultipleTopLevelNodes(t *testing.T) {
	nodes, diag := parse(t, "<div></div><span></span>")
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.FormatAll())
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestParseLoopMalformedForExpression(t *testing.T) {
	// No "in" keyword — parseForExpression returns the whole string as iterator
	nodes, diag := parse(t, `<loop for="items" index="i"></loop>`)
	// Should still parse (errors for missing "in" are semantic, not structural)
	_ = diag
	loop := nodes[0].(*ast.Loop)
	if loop.Iterator != "items" {
		t.Errorf("iterator = %q, want %q", loop.Iterator, "items")
	}
}
