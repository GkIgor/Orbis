// Package ast defines the Abstract Syntax Tree structures for the Orbis DSL.
// AST nodes are deterministic, serializable, and testable per framework_proposal_v_1.md §26.
package ast

import (
	"encoding/json"

	"github.com/orbisui/orbis/compiler/internal/diagnostics"
)

// NodeType identifies the kind of AST node.
type NodeType int

const (
	// ElementNodeType represents a standard HTML element (<div>, <span>, etc.).
	ElementNodeType NodeType = iota
	// ComponentNodeType represents a component reference (capitalized tag name).
	ComponentNodeType
	// LoopNodeType represents a <loop for="item in items" index="i"> construct.
	LoopNodeType
	// IfNodeType represents an <if condition="expression"> construct.
	IfNodeType
	// TextNodeType represents raw text content.
	TextNodeType
	// InterpolationNodeType represents a {{ expression }} interpolation.
	InterpolationNodeType
)

// String returns the string representation of a NodeType.
func (nt NodeType) String() string {
	switch nt {
	case ElementNodeType:
		return "Element"
	case ComponentNodeType:
		return "Component"
	case LoopNodeType:
		return "Loop"
	case IfNodeType:
		return "If"
	case TextNodeType:
		return "Text"
	case InterpolationNodeType:
		return "Interpolation"
	default:
		return "Unknown"
	}
}

// MarshalJSON implements custom JSON marshalling for NodeType.
func (nt NodeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(nt.String())
}

// Node is the interface that all AST nodes implement.
type Node interface {
	// Type returns the NodeType of this node.
	Type() NodeType
	// Location returns the source location of this node.
	Location() diagnostics.SourceLocation
}

// Attribute represents an HTML attribute (name="value").
type Attribute struct {
	Name     string                    `json:"name"`
	Value    string                    `json:"value"`
	Location diagnostics.SourceLocation `json:"location"`
}

// EventBinding represents an event binding ((event)="handler()").
type EventBinding struct {
	Event    string                    `json:"event"`
	Handler  string                    `json:"handler"`
	Location diagnostics.SourceLocation `json:"location"`
}

// Element represents a standard HTML element node.
type Element struct {
	Tag           string                    `json:"tag"`
	Attributes    []Attribute               `json:"attributes,omitempty"`
	EventBindings []EventBinding            `json:"eventBindings,omitempty"`
	Children      []Node                    `json:"children,omitempty"`
	Loc           diagnostics.SourceLocation `json:"location"`
	SelfClosing   bool                      `json:"selfClosing,omitempty"`
}

func (e *Element) Type() NodeType                    { return ElementNodeType }
func (e *Element) Location() diagnostics.SourceLocation { return e.Loc }

// Component represents a component reference node (capitalized tag name).
type Component struct {
	Name          string                    `json:"name"`
	Attributes    []Attribute               `json:"attributes,omitempty"`
	EventBindings []EventBinding            `json:"eventBindings,omitempty"`
	Children      []Node                    `json:"children,omitempty"`
	Loc           diagnostics.SourceLocation `json:"location"`
	SelfClosing   bool                      `json:"selfClosing,omitempty"`
}

func (c *Component) Type() NodeType                    { return ComponentNodeType }
func (c *Component) Location() diagnostics.SourceLocation { return c.Loc }

// Loop represents a <loop for="item in items" index="i"> node.
type Loop struct {
	Iterator   string                    `json:"iterator"`
	Collection string                    `json:"collection"`
	Index      string                    `json:"index"`
	Children   []Node                    `json:"children,omitempty"`
	Loc        diagnostics.SourceLocation `json:"location"`
}

func (l *Loop) Type() NodeType                    { return LoopNodeType }
func (l *Loop) Location() diagnostics.SourceLocation { return l.Loc }

// If represents an <if condition="expression"> node.
type If struct {
	Condition string                    `json:"condition"`
	Children  []Node                    `json:"children,omitempty"`
	Loc       diagnostics.SourceLocation `json:"location"`
}

func (i *If) Type() NodeType                    { return IfNodeType }
func (i *If) Location() diagnostics.SourceLocation { return i.Loc }

// Text represents raw text content in the template.
type Text struct {
	Content string                    `json:"content"`
	Loc     diagnostics.SourceLocation `json:"location"`
}

func (t *Text) Type() NodeType                    { return TextNodeType }
func (t *Text) Location() diagnostics.SourceLocation { return t.Loc }

// Interpolation represents a {{ expression }} interpolation node.
type Interpolation struct {
	Expression string                    `json:"expression"`
	Loc        diagnostics.SourceLocation `json:"location"`
}

func (ip *Interpolation) Type() NodeType                    { return InterpolationNodeType }
func (ip *Interpolation) Location() diagnostics.SourceLocation { return ip.Loc }

// Template represents the root of a parsed template (a list of top-level nodes).
type Template struct {
	Nodes []Node `json:"nodes"`
}

// MarshalJSON implements custom JSON marshalling for Node slices to include type discriminators.
// This is necessary because Node is an interface and built-in JSON marshalling
// does not include type information.
func MarshalNodeJSON(n Node) ([]byte, error) {
	type nodeWrapper struct {
		NodeType NodeType    `json:"type"`
		Node     interface{} `json:"node"`
	}
	return json.Marshal(nodeWrapper{
		NodeType: n.Type(),
		Node:     n,
	})
}

// MarshalNodesJSON marshals a slice of Nodes to JSON with type discriminators.
func MarshalNodesJSON(nodes []Node) ([]byte, error) {
	type nodeWrapper struct {
		NodeType NodeType    `json:"type"`
		Node     interface{} `json:"node"`
	}
	wrappers := make([]nodeWrapper, len(nodes))
	for i, n := range nodes {
		wrappers[i] = nodeWrapper{
			NodeType: n.Type(),
			Node:     n,
		}
	}
	return json.MarshalIndent(wrappers, "", "  ")
}
