package codegen_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/orbisui/orbis/compiler/internal/ast"
	"github.com/orbisui/orbis/compiler/internal/codegen"
	"github.com/orbisui/orbis/compiler/internal/diagnostics"
)

// loc returns a test source location.
func loc() diagnostics.SourceLocation {
	return diagnostics.SourceLocation{File: "test.html", Line: 1, Column: 1}
}

// generate helper: creates generator, generates, returns string.
func generate(t *testing.T, nodes []ast.Node) string {
	t.Helper()
	gen := codegen.New()
	result, err := gen.Generate(nodes)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	return string(result)
}

// --- 1) Simple Element ---

func TestGenerateSimpleElement(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{Tag: "div", Loc: loc()},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "simple_element.js", output)
}

// --- 2) Nested Elements ---

func TestGenerateNestedElements(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag: "div",
			Loc: loc(),
			Children: []ast.Node{
				&ast.Element{Tag: "span", Loc: loc()},
				&ast.Element{Tag: "p", Loc: loc()},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "nested_elements.js", output)
}

// --- 3) Element with Attributes ---

func TestGenerateElementWithAttributes(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag: "div",
			Loc: loc(),
			Attributes: []ast.Attribute{
				{Name: "class", Value: "container p-6", Location: loc()},
				{Name: "id", Value: "main", Location: loc()},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "element_with_attributes.js", output)
}

// --- 4) If Block ---

func TestGenerateIfBlock(t *testing.T) {
	nodes := []ast.Node{
		&ast.If{
			Condition: "visible",
			Loc:       loc(),
			Children: []ast.Node{
				&ast.Element{Tag: "div", Loc: loc()},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "if_block.js", output)
}

// --- 5) Loop Block ---

func TestGenerateLoopBlock(t *testing.T) {
	nodes := []ast.Node{
		&ast.Loop{
			Iterator:   "item",
			Collection: "items",
			Index:      "i",
			Loc:        loc(),
			Children: []ast.Node{
				&ast.Element{
					Tag: "p",
					Loc: loc(),
					Children: []ast.Node{
						&ast.Interpolation{Expression: "item", Loc: loc()},
					},
				},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "loop_block.js", output)
}

// --- 6) Nested If inside Loop ---

func TestGenerateNestedIfInLoop(t *testing.T) {
	nodes := []ast.Node{
		&ast.Loop{
			Iterator:   "item",
			Collection: "items",
			Index:      "i",
			Loc:        loc(),
			Children: []ast.Node{
				&ast.If{
					Condition: "item.active",
					Loc:       loc(),
					Children: []ast.Node{
						&ast.Element{
							Tag: "span",
							Loc: loc(),
							Children: []ast.Node{
								&ast.Interpolation{Expression: "item.name", Loc: loc()},
							},
						},
					},
				},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "nested_if_in_loop.js", output)
}

// --- 7) Component with Event Binding ---

func TestGenerateComponentWithEventBinding(t *testing.T) {
	nodes := []ast.Node{
		&ast.Component{
			Name: "CounterItem",
			Loc:  loc(),
			EventBindings: []ast.EventBinding{
				{Event: "click", Handler: "select(i)", Location: loc()},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "component_event_binding.js", output)
}

// --- 8) Interpolation ---

func TestGenerateInterpolation(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag: "h1",
			Loc: loc(),
			Children: []ast.Node{
				&ast.Interpolation{Expression: "config.title", Loc: loc()},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "interpolation.js", output)
}

// --- Text Node ---

func TestGenerateTextNode(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag: "p",
			Loc: loc(),
			Children: []ast.Node{
				&ast.Text{Content: "Hello World", Loc: loc()},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "text_node.js", output)
}

// --- Event Binding on Element ---

func TestGenerateElementWithEventBinding(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag: "button",
			Loc: loc(),
			Attributes: []ast.Attribute{
				{Name: "class", Value: "btn", Location: loc()},
			},
			EventBindings: []ast.EventBinding{
				{Event: "click", Handler: "toggle()", Location: loc()},
			},
			Children: []ast.Node{
				&ast.Text{Content: "Toggle", Loc: loc()},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "element_event_binding.js", output)
}

// --- Empty Template ---

func TestGenerateEmptyTemplate(t *testing.T) {
	nodes := []ast.Node{}
	output := generate(t, nodes)
	compareSnapshot(t, "empty_template.js", output)
}

// --- Full Pilot App Template ---

func TestGenerateFullTemplate(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag: "div",
			Loc: loc(),
			Attributes: []ast.Attribute{
				{Name: "class", Value: "p-6 bg-gray-100 container", Location: loc()},
			},
			Children: []ast.Node{
				&ast.Element{
					Tag: "h1",
					Loc: loc(),
					Attributes: []ast.Attribute{
						{Name: "class", Value: "title", Location: loc()},
					},
					Children: []ast.Node{
						&ast.Interpolation{Expression: "config.title", Loc: loc()},
					},
				},
				&ast.Element{
					Tag: "button",
					Loc: loc(),
					Attributes: []ast.Attribute{
						{Name: "class", Value: "btn", Location: loc()},
					},
					EventBindings: []ast.EventBinding{
						{Event: "click", Handler: "toggle()", Location: loc()},
					},
					Children: []ast.Node{
						&ast.Text{Content: "Toggle List", Loc: loc()},
					},
				},
				&ast.If{
					Condition: "app.visible",
					Loc:       loc(),
					Children: []ast.Node{
						&ast.Element{
							Tag: "div",
							Loc: loc(),
							Attributes: []ast.Attribute{
								{Name: "class", Value: "mt-4", Location: loc()},
							},
							Children: []ast.Node{
								&ast.Loop{
									Iterator:   "item",
									Collection: "app.items",
									Index:      "i",
									Loc:        loc(),
									Children: []ast.Node{
										&ast.Component{
											Name: "CounterItem",
											Loc:  loc(),
											EventBindings: []ast.EventBinding{
												{Event: "click", Handler: "select(i)", Location: loc()},
											},
											SelfClosing: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	output := generate(t, nodes)
	compareSnapshot(t, "full_template.js", output)
}

// --- Determinism Test ---

func TestGenerateDeterministic(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag: "div",
			Loc: loc(),
			Children: []ast.Node{
				&ast.Element{Tag: "span", Loc: loc()},
				&ast.Interpolation{Expression: "x", Loc: loc()},
			},
		},
	}

	// Generate twice and compare — must be identical
	gen := codegen.New()
	output1, err := gen.Generate(nodes)
	if err != nil {
		t.Fatalf("first Generate() error: %v", err)
	}

	gen2 := codegen.New()
	output2, err := gen2.Generate(nodes)
	if err != nil {
		t.Fatalf("second Generate() error: %v", err)
	}

	if string(output1) != string(output2) {
		t.Errorf("output not deterministic:\n--- run 1 ---\n%s\n--- run 2 ---\n%s",
			string(output1), string(output2))
	}
}

// --- Self-Closing Element ---

func TestGenerateSelfClosingElement(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag:         "br",
			Loc:         loc(),
			SelfClosing: true,
		},
	}
	output := generate(t, nodes)
	// Self-closing elements produce the same output — createElement + appendChild
	if output == "" {
		t.Error("expected non-empty output for self-closing element")
	}
}

// --- Whitespace-only Text ---

func TestGenerateWhitespaceTextSkipped(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag: "div",
			Loc: loc(),
			Children: []ast.Node{
				&ast.Text{Content: "   \n  \t  ", Loc: loc()},
			},
		},
	}
	output := generate(t, nodes)
	// Whitespace-only text nodes should be skipped
	if contains(output, "createTextNode") {
		t.Error("whitespace-only text should not produce createTextNode")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstr(s, substr)
}

func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- JS Escaping ---

func TestGenerateEscapedAttribute(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			Tag: "div",
			Loc: loc(),
			Attributes: []ast.Attribute{
				{Name: "data-value", Value: "he said \"hello\"", Location: loc()},
			},
		},
	}
	output := generate(t, nodes)
	// Quotes in attribute values must be escaped
	if !contains(output, `he said \"hello\"`) {
		t.Errorf("expected escaped quotes in output, got:\n%s", output)
	}
}

// compareSnapshot compares generated output against a golden file.
func compareSnapshot(t *testing.T, name string, actual string) {
	t.Helper()

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

	if string(golden) != actual {
		// For JS files, we just compare as strings directly
		t.Errorf("snapshot mismatch for %s\n--- expected ---\n%s\n--- actual ---\n%s",
			name, string(golden), actual)
	}

	_ = json.Unmarshal // import used for type compatibility
}
