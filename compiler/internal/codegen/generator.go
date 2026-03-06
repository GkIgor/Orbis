// Package codegen implements the JavaScript code generator for the Orbis compiler.
// It transforms AST nodes into deterministic JavaScript render functions.
//
// Generated output follows RFC 0001 — Render Engine Deterministic Model:
//  1. beforeRender()
//  2. Clear ShadowRoot
//  3. Execute compiled DOM instructions
//  4. Attach children
//  5. afterRender()
//
// All output is deterministic: identical AST input produces identical JS output.
// No randomness, no reflection, no dynamic property enumeration.
package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/orbisui/orbis/compiler/internal/ast"
)

// Generator defines the interface for code generation.
type Generator interface {
	// Generate transforms AST nodes into JavaScript code.
	Generate(nodes []ast.Node) ([]byte, error)
}

// JSGenerator produces deterministic JavaScript render functions from AST.
// It is not safe for concurrent use.
type JSGenerator struct {
	buf     bytes.Buffer
	indent  int
	counter int      // deterministic variable counter for unique DOM references
	locals  []string // stack of loop-scoped local variable names (iterators + indexes)
}

// New creates a new JSGenerator.
func New() *JSGenerator {
	return &JSGenerator{}
}

// Generate transforms AST nodes into a complete render function.
// The output is a self-contained JavaScript function that creates DOM nodes
// and appends them to the provided root element.
func (g *JSGenerator) Generate(nodes []ast.Node) ([]byte, error) {
	g.buf.Reset()
	g.indent = 0
	g.counter = 0
	g.locals = nil

	g.writeLine("function render(ctx, root) {")
	g.indent++

	// Step 1: beforeRender lifecycle hook
	g.writeLine("if (ctx.beforeRender) ctx.beforeRender();")

	// Step 2: Clear previous subtree
	g.writeLine("root.innerHTML = \"\";")

	// Step 3: DOM construction
	for _, node := range nodes {
		g.emitNode(node, "root")
	}

	// Step 5: afterRender lifecycle hook
	g.writeLine("if (ctx.afterRender) ctx.afterRender();")

	g.indent--
	g.writeLine("}")

	return g.buf.Bytes(), nil
}

// emitNode dispatches to the appropriate emitter based on node type.
func (g *JSGenerator) emitNode(node ast.Node, parentVar string) {
	switch n := node.(type) {
	case *ast.Element:
		g.emitElement(n, parentVar)
	case *ast.Component:
		g.emitComponent(n, parentVar)
	case *ast.Loop:
		g.emitLoop(n, parentVar)
	case *ast.If:
		g.emitIf(n, parentVar)
	case *ast.Text:
		g.emitText(n, parentVar)
	case *ast.Interpolation:
		g.emitInterpolation(n, parentVar)
	}
}

// emitElement generates DOM creation code for an HTML element.
// Output: const _elN = document.createElement("tag"); parent.appendChild(_elN);
func (g *JSGenerator) emitElement(el *ast.Element, parentVar string) {
	varName := g.nextVar()

	g.writeLine(fmt.Sprintf("const %s = document.createElement(\"%s\");", varName, el.Tag))

	// Emit attributes
	for _, attr := range el.Attributes {
		g.writeLine(fmt.Sprintf("%s.setAttribute(\"%s\", \"%s\");",
			varName, escapeJS(attr.Name), escapeJS(attr.Value)))
	}

	// Emit event bindings
	for _, ev := range el.EventBindings {
		g.emitEventBinding(varName, ev)
	}

	// Emit children
	for _, child := range el.Children {
		g.emitNode(child, varName)
	}

	// Append to parent
	g.writeLine(fmt.Sprintf("%s.appendChild(%s);", parentVar, varName))
}

// emitComponent generates instantiation code for a child component.
// Output: const _compN = new ComponentName(); _compN.render(ctx, wrapper);
func (g *JSGenerator) emitComponent(comp *ast.Component, parentVar string) {
	varName := g.nextVar()
	wrapperVar := g.nextVar()

	// Create a wrapper element for the component's ShadowRoot mount point
	g.writeLine(fmt.Sprintf("const %s = document.createElement(\"%s\");",
		wrapperVar, strings.ToLower(comp.Name)))

	// Emit attributes on the wrapper
	for _, attr := range comp.Attributes {
		g.writeLine(fmt.Sprintf("%s.setAttribute(\"%s\", \"%s\");",
			wrapperVar, escapeJS(attr.Name), escapeJS(attr.Value)))
	}

	// Emit event bindings on the wrapper
	for _, ev := range comp.EventBindings {
		g.emitEventBinding(wrapperVar, ev)
	}

	// Instantiate the component
	g.writeLine(fmt.Sprintf("const %s = new %s();", varName, comp.Name))

	// Mount: attach ShadowRoot and render
	g.writeLine(fmt.Sprintf("const %sShadow = %s.attachShadow({ mode: \"open\" });",
		varName, wrapperVar))
	g.writeLine(fmt.Sprintf("%s.render(%s, %sShadow);", varName, varName, varName))

	// Append wrapper to parent
	g.writeLine(fmt.Sprintf("%s.appendChild(%s);", parentVar, wrapperVar))
}

// emitLoop generates a for-loop that iterates over a collection.
// Output: for (let i = 0; i < ctx.collection.length; i++) { const item = ctx.collection[i]; ... }
func (g *JSGenerator) emitLoop(loop *ast.Loop, parentVar string) {
	indexVar := loop.Index
	iteratorVar := loop.Iterator
	collectionExpr := g.resolveExpr(loop.Collection)

	g.writeLine(fmt.Sprintf("for (let %s = 0; %s < %s.length; %s++) {",
		indexVar, indexVar, collectionExpr, indexVar))
	g.indent++

	// Bind the iterator variable to the current element
	g.writeLine(fmt.Sprintf("const %s = %s[%s];", iteratorVar, collectionExpr, indexVar))

	// Push loop locals onto the scope stack
	g.pushLocals(iteratorVar, indexVar)

	// Emit children inside the loop body
	for _, child := range loop.Children {
		g.emitNode(child, parentVar)
	}

	// Pop loop locals
	g.popLocals(2)

	g.indent--
	g.writeLine("}")
}

// emitIf generates a conditional block.
// Output: if (ctx.condition) { ... }
func (g *JSGenerator) emitIf(ifNode *ast.If, parentVar string) {
	condExpr := g.resolveExpr(ifNode.Condition)

	g.writeLine(fmt.Sprintf("if (%s) {", condExpr))
	g.indent++

	for _, child := range ifNode.Children {
		g.emitNode(child, parentVar)
	}

	g.indent--
	g.writeLine("}")
}

// emitText generates a text node.
// Output: const _txtN = document.createTextNode("content"); parent.appendChild(_txtN);
func (g *JSGenerator) emitText(text *ast.Text, parentVar string) {
	content := strings.TrimSpace(text.Content)
	if content == "" {
		return
	}
	varName := g.nextVar()
	g.writeLine(fmt.Sprintf("const %s = document.createTextNode(\"%s\");", varName, escapeJS(content)))
	g.writeLine(fmt.Sprintf("%s.appendChild(%s);", parentVar, varName))
}

// emitInterpolation generates a text node whose content comes from the component context.
// Uses textContent semantics (no innerHTML) per security model §1.1.
// Output: const _txtN = document.createTextNode(String(expr)); parent.appendChild(_txtN);
func (g *JSGenerator) emitInterpolation(interp *ast.Interpolation, parentVar string) {
	varName := g.nextVar()
	expr := g.resolveExpr(interp.Expression)
	g.writeLine(fmt.Sprintf("const %s = document.createTextNode(String(%s));", varName, expr))
	g.writeLine(fmt.Sprintf("%s.appendChild(%s);", parentVar, varName))
}

// emitEventBinding generates an addEventListener call.
// Uses direct function reference — no string-based handler evaluation.
// Output: el.addEventListener("event", function() { ctx.handler(); });
func (g *JSGenerator) emitEventBinding(elVar string, ev ast.EventBinding) {
	g.writeLine(fmt.Sprintf("%s.addEventListener(\"%s\", function() { ctx.%s; });",
		elVar, escapeJS(ev.Event), ev.Handler))
}

// resolveExpr resolves an expression against the current scope.
// If the expression root matches a loop-local variable (iterator or index),
// it is used directly. Otherwise, it is prefixed with "ctx.".
func (g *JSGenerator) resolveExpr(expr string) string {
	// Extract the root identifier (before any '.' access)
	root := expr
	dotIdx := strings.Index(expr, ".")
	if dotIdx != -1 {
		root = expr[:dotIdx]
	}

	// Check if root is a loop-local variable
	for _, local := range g.locals {
		if root == local {
			return expr // use directly
		}
	}

	return "ctx." + expr
}

// pushLocals adds variable names to the scope stack.
func (g *JSGenerator) pushLocals(names ...string) {
	g.locals = append(g.locals, names...)
}

// popLocals removes N entries from the scope stack.
func (g *JSGenerator) popLocals(n int) {
	if n > len(g.locals) {
		g.locals = nil
		return
	}
	g.locals = g.locals[:len(g.locals)-n]
}

// nextVar returns the next deterministic variable name.
func (g *JSGenerator) nextVar() string {
	name := fmt.Sprintf("_el%d", g.counter)
	g.counter++
	return name
}

// writeLine writes an indented line to the buffer.
func (g *JSGenerator) writeLine(line string) {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteString("  ")
	}
	g.buf.WriteString(line)
	g.buf.WriteString("\n")
}

// escapeJS escapes a string for safe inclusion in JavaScript string literals.
func escapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
