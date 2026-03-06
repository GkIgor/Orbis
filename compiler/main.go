// Package main provides the Orbis compiler entry point.
// The compiler reads an Orbis DSL template file, tokenizes it,
// parses it into an AST, and optionally generates JavaScript output.
//
// Usage:
//
//	orbis-compiler <file.html>          — output AST as JSON
//	orbis-compiler <file.html> --emit-js — output generated JavaScript render function
//
// This is the Compiler layer (Go) per framework_proposal_v_1.md §27.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orbisui/orbis/compiler/internal/ast"
	"github.com/orbisui/orbis/compiler/internal/codegen"
	"github.com/orbisui/orbis/compiler/internal/diagnostics"
	"github.com/orbisui/orbis/compiler/internal/lexer"
	"github.com/orbisui/orbis/compiler/internal/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: orbis-compiler <file.html> [--emit-js]")
		os.Exit(1)
	}

	filePath := os.Args[1]
	emitJS := false
	for _, arg := range os.Args[2:] {
		if arg == "--emit-js" {
			emitJS = true
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
		os.Exit(1)
	}

	result, exitCode := compile(string(data), filePath, emitJS)
	fmt.Print(result)
	os.Exit(exitCode)
}

// compile runs the lexer, parser, and optionally the code generator.
// Returns string output and an exit code (0 for success, 1 for errors).
func compile(source string, filePath string, emitJS bool) (string, int) {
	diag := diagnostics.NewCollector()

	// Lexer phase
	lex := lexer.New(source, filePath, diag)
	tokens := lex.Tokenize()

	// Parser phase
	p := parser.New(tokens, filePath, diag)
	nodes := p.Parse()

	// Check for errors before code generation
	if diag.HasErrors() {
		return formatDiagnostics(diag), 1
	}

	// Code generation phase (optional)
	if emitJS {
		gen := codegen.New()
		jsCode, err := gen.Generate(nodes)
		if err != nil {
			return fmt.Sprintf("error: code generation failed: %s\n", err), 1
		}
		return string(jsCode), 0
	}

	// Default: output AST as JSON
	return formatAST(nodes, diag)
}

// formatAST serializes AST nodes and diagnostics as JSON.
func formatAST(nodes []ast.Node, diag *diagnostics.Collector) (string, int) {
	type nodeWrapper struct {
		Type string      `json:"type"`
		Node interface{} `json:"node"`
	}
	type compilerOutput struct {
		Nodes       interface{}              `json:"nodes"`
		Diagnostics []diagnostics.Diagnostic `json:"diagnostics,omitempty"`
	}

	var nodeData interface{}
	if len(nodes) > 0 {
		wrappers := make([]nodeWrapper, len(nodes))
		for i, n := range nodes {
			wrappers[i] = nodeWrapper{
				Type: n.Type().String(),
				Node: n,
			}
		}
		nodeData = wrappers
	} else {
		nodeData = []interface{}{}
	}

	output := compilerOutput{
		Nodes:       nodeData,
		Diagnostics: diag.All(),
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal output: %s"}`, err), 1
	}

	exitCode := 0
	if diag.HasErrors() {
		exitCode = 1
	}

	return string(jsonBytes) + "\n", exitCode
}

// formatDiagnostics outputs diagnostics as JSON when compilation has errors.
func formatDiagnostics(diag *diagnostics.Collector) string {
	type errOutput struct {
		Diagnostics []diagnostics.Diagnostic `json:"diagnostics"`
	}
	output := errOutput{Diagnostics: diag.All()}
	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonBytes) + "\n"
}
