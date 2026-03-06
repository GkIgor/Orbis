// Package diagnostics provides structured error reporting for the Orbis compiler.
// All diagnostics include file, line, column, and an actionable message.
// No panic is used for user errors.
package diagnostics

import (
	"fmt"
	"strings"
)

// Severity represents the severity level of a diagnostic.
type Severity int

const (
	// Error indicates a compilation error that prevents successful parsing.
	Error Severity = iota
	// Warning indicates a potential issue that does not prevent compilation.
	Warning
)

// String returns the string representation of a Severity.
func (s Severity) String() string {
	switch s {
	case Error:
		return "error"
	case Warning:
		return "warning"
	default:
		return "unknown"
	}
}

// SourceLocation represents a position in source code.
type SourceLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// String returns a human-readable representation of the source location.
func (loc SourceLocation) String() string {
	return fmt.Sprintf("%s:%d:%d", loc.File, loc.Line, loc.Column)
}

// Diagnostic represents a single compiler diagnostic message.
type Diagnostic struct {
	Location SourceLocation `json:"location"`
	Severity Severity       `json:"severity"`
	Code     string         `json:"code"`
	Message  string         `json:"message"`
}

// String returns a formatted diagnostic message.
func (d Diagnostic) String() string {
	return fmt.Sprintf("%s [%s] %s: %s", d.Location, d.Code, d.Severity, d.Message)
}

// Collector accumulates diagnostics during compilation.
// It is not safe for concurrent use.
type Collector struct {
	diagnostics []Diagnostic
}

// NewCollector creates a new empty Collector.
func NewCollector() *Collector {
	return &Collector{
		diagnostics: make([]Diagnostic, 0),
	}
}

// AddError adds an error diagnostic.
func (c *Collector) AddError(loc SourceLocation, code string, message string) {
	c.diagnostics = append(c.diagnostics, Diagnostic{
		Location: loc,
		Severity: Error,
		Code:     code,
		Message:  message,
	})
}

// AddWarning adds a warning diagnostic.
func (c *Collector) AddWarning(loc SourceLocation, code string, message string) {
	c.diagnostics = append(c.diagnostics, Diagnostic{
		Location: loc,
		Severity: Warning,
		Code:     code,
		Message:  message,
	})
}

// HasErrors returns true if any error diagnostics have been collected.
func (c *Collector) HasErrors() bool {
	for _, d := range c.diagnostics {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

// Errors returns only the error diagnostics.
func (c *Collector) Errors() []Diagnostic {
	var errs []Diagnostic
	for _, d := range c.diagnostics {
		if d.Severity == Error {
			errs = append(errs, d)
		}
	}
	return errs
}

// All returns all collected diagnostics.
func (c *Collector) All() []Diagnostic {
	result := make([]Diagnostic, len(c.diagnostics))
	copy(result, c.diagnostics)
	return result
}

// Count returns the total number of diagnostics.
func (c *Collector) Count() int {
	return len(c.diagnostics)
}

// ErrorCount returns the number of error diagnostics.
func (c *Collector) ErrorCount() int {
	count := 0
	for _, d := range c.diagnostics {
		if d.Severity == Error {
			count++
		}
	}
	return count
}

// FormatAll returns a human-readable string of all diagnostics.
func (c *Collector) FormatAll() string {
	if len(c.diagnostics) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, d := range c.diagnostics {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(d.String())
	}
	return sb.String()
}
