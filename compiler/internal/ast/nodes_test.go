package ast_test

import (
	"encoding/json"
	"testing"

	"github.com/orbisui/orbis/compiler/internal/ast"
	"github.com/orbisui/orbis/compiler/internal/diagnostics"
)

func TestNodeTypeString(t *testing.T) {
	tests := []struct {
		nt       ast.NodeType
		expected string
	}{
		{ast.ElementNodeType, "Element"},
		{ast.ComponentNodeType, "Component"},
		{ast.LoopNodeType, "Loop"},
		{ast.IfNodeType, "If"},
		{ast.TextNodeType, "Text"},
		{ast.InterpolationNodeType, "Interpolation"},
		{ast.NodeType(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.nt.String(); got != tt.expected {
			t.Errorf("NodeType(%d).String() = %q, want %q", tt.nt, got, tt.expected)
		}
	}
}

func TestNodeTypeMarshalJSON(t *testing.T) {
	nt := ast.ElementNodeType
	data, err := json.Marshal(nt)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	if string(data) != `"Element"` {
		t.Errorf("MarshalJSON = %s, want %q", string(data), "Element")
	}
}

func TestElementNodeInterface(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 1, Column: 1}
	elem := &ast.Element{Tag: "div", Loc: loc}

	if elem.Type() != ast.ElementNodeType {
		t.Errorf("type = %v, want ElementNodeType", elem.Type())
	}
	if elem.Location() != loc {
		t.Errorf("location mismatch")
	}
}

func TestComponentNodeInterface(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 5, Column: 3}
	comp := &ast.Component{Name: "MyComp", Loc: loc}

	if comp.Type() != ast.ComponentNodeType {
		t.Errorf("type = %v, want ComponentNodeType", comp.Type())
	}
	if comp.Location() != loc {
		t.Errorf("location mismatch")
	}
}

func TestLoopNodeInterface(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 10, Column: 1}
	loop := &ast.Loop{Iterator: "item", Collection: "items", Index: "i", Loc: loc}

	if loop.Type() != ast.LoopNodeType {
		t.Errorf("type = %v, want LoopNodeType", loop.Type())
	}
	if loop.Location() != loc {
		t.Errorf("location mismatch")
	}
}

func TestIfNodeInterface(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 15, Column: 1}
	ifNode := &ast.If{Condition: "x > 0", Loc: loc}

	if ifNode.Type() != ast.IfNodeType {
		t.Errorf("type = %v, want IfNodeType", ifNode.Type())
	}
	if ifNode.Location() != loc {
		t.Errorf("location mismatch")
	}
}

func TestTextNodeInterface(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 20, Column: 5}
	text := &ast.Text{Content: "Hello", Loc: loc}

	if text.Type() != ast.TextNodeType {
		t.Errorf("type = %v, want TextNodeType", text.Type())
	}
	if text.Location() != loc {
		t.Errorf("location mismatch")
	}
}

func TestInterpolationNodeInterface(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 25, Column: 10}
	interp := &ast.Interpolation{Expression: "name", Loc: loc}

	if interp.Type() != ast.InterpolationNodeType {
		t.Errorf("type = %v, want InterpolationNodeType", interp.Type())
	}
	if interp.Location() != loc {
		t.Errorf("location mismatch")
	}
}

func TestMarshalNodeJSON(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 1, Column: 1}
	elem := &ast.Element{Tag: "div", Loc: loc}

	data, err := ast.MarshalNodeJSON(elem)
	if err != nil {
		t.Fatalf("MarshalNodeJSON error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result["type"] != "Element" {
		t.Errorf("type = %v, want Element", result["type"])
	}
}

func TestMarshalNodesJSON(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 1, Column: 1}
	nodes := []ast.Node{
		&ast.Element{Tag: "div", Loc: loc},
		&ast.Text{Content: "Hello", Loc: loc},
	}

	data, err := ast.MarshalNodesJSON(nodes)
	if err != nil {
		t.Fatalf("MarshalNodesJSON error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result))
	}
	if result[0]["type"] != "Element" {
		t.Errorf("node[0] type = %v, want Element", result[0]["type"])
	}
	if result[1]["type"] != "Text" {
		t.Errorf("node[1] type = %v, want Text", result[1]["type"])
	}
}

func TestElementJSONSerialization(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 1, Column: 1}
	elem := &ast.Element{
		Tag: "div",
		Attributes: []ast.Attribute{
			{Name: "class", Value: "container", Location: loc},
		},
		EventBindings: []ast.EventBinding{
			{Event: "click", Handler: "handle()", Location: loc},
		},
		Loc:         loc,
		SelfClosing: true,
	}

	data, err := json.Marshal(elem)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result["tag"] != "div" {
		t.Errorf("tag = %v, want div", result["tag"])
	}
	if result["selfClosing"] != true {
		t.Errorf("selfClosing = %v, want true", result["selfClosing"])
	}
}
