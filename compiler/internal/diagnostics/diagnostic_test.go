package diagnostics_test

import (
	"strings"
	"testing"

	"github.com/orbisui/orbis/compiler/internal/diagnostics"
)

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity diagnostics.Severity
		expected string
	}{
		{diagnostics.Error, "error"},
		{diagnostics.Warning, "warning"},
		{diagnostics.Severity(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.severity.String(); got != tt.expected {
			t.Errorf("Severity(%d).String() = %q, want %q", tt.severity, got, tt.expected)
		}
	}
}

func TestSourceLocationString(t *testing.T) {
	loc := diagnostics.SourceLocation{File: "test.html", Line: 10, Column: 5}
	expected := "test.html:10:5"
	if got := loc.String(); got != expected {
		t.Errorf("SourceLocation.String() = %q, want %q", got, expected)
	}
}

func TestDiagnosticString(t *testing.T) {
	d := diagnostics.Diagnostic{
		Location: diagnostics.SourceLocation{File: "app.html", Line: 3, Column: 12},
		Severity: diagnostics.Error,
		Code:     "E001",
		Message:  "unclosed tag",
	}
	got := d.String()
	if !strings.Contains(got, "app.html:3:12") {
		t.Errorf("Diagnostic.String() missing location, got: %q", got)
	}
	if !strings.Contains(got, "E001") {
		t.Errorf("Diagnostic.String() missing error code, got: %q", got)
	}
	if !strings.Contains(got, "error") {
		t.Errorf("Diagnostic.String() missing severity, got: %q", got)
	}
	if !strings.Contains(got, "unclosed tag") {
		t.Errorf("Diagnostic.String() missing message, got: %q", got)
	}
}

func TestCollectorEmpty(t *testing.T) {
	c := diagnostics.NewCollector()
	if c.HasErrors() {
		t.Error("empty collector should not have errors")
	}
	if c.Count() != 0 {
		t.Errorf("empty collector count = %d, want 0", c.Count())
	}
	if c.ErrorCount() != 0 {
		t.Errorf("empty collector error count = %d, want 0", c.ErrorCount())
	}
	if got := c.FormatAll(); got != "" {
		t.Errorf("empty collector FormatAll() = %q, want empty", got)
	}
}

func TestCollectorAddError(t *testing.T) {
	c := diagnostics.NewCollector()
	loc := diagnostics.SourceLocation{File: "test.html", Line: 1, Column: 1}
	c.AddError(loc, "E001", "test error")

	if !c.HasErrors() {
		t.Error("collector should have errors")
	}
	if c.Count() != 1 {
		t.Errorf("count = %d, want 1", c.Count())
	}
	if c.ErrorCount() != 1 {
		t.Errorf("error count = %d, want 1", c.ErrorCount())
	}

	errs := c.Errors()
	if len(errs) != 1 {
		t.Fatalf("Errors() returned %d items, want 1", len(errs))
	}
	if errs[0].Code != "E001" {
		t.Errorf("error code = %q, want %q", errs[0].Code, "E001")
	}
}

func TestCollectorAddWarning(t *testing.T) {
	c := diagnostics.NewCollector()
	loc := diagnostics.SourceLocation{File: "test.html", Line: 2, Column: 3}
	c.AddWarning(loc, "W001", "test warning")

	if c.HasErrors() {
		t.Error("collector should not have errors (only warnings)")
	}
	if c.Count() != 1 {
		t.Errorf("count = %d, want 1", c.Count())
	}
	if c.ErrorCount() != 0 {
		t.Errorf("error count = %d, want 0", c.ErrorCount())
	}
}

func TestCollectorMixed(t *testing.T) {
	c := diagnostics.NewCollector()
	loc1 := diagnostics.SourceLocation{File: "a.html", Line: 1, Column: 1}
	loc2 := diagnostics.SourceLocation{File: "a.html", Line: 5, Column: 10}
	c.AddError(loc1, "E001", "error one")
	c.AddWarning(loc2, "W001", "warning one")
	c.AddError(loc1, "E002", "error two")

	if c.Count() != 3 {
		t.Errorf("count = %d, want 3", c.Count())
	}
	if c.ErrorCount() != 2 {
		t.Errorf("error count = %d, want 2", c.ErrorCount())
	}
	if !c.HasErrors() {
		t.Error("collector should have errors")
	}

	all := c.All()
	if len(all) != 3 {
		t.Errorf("All() returned %d items, want 3", len(all))
	}

	errs := c.Errors()
	if len(errs) != 2 {
		t.Errorf("Errors() returned %d items, want 2", len(errs))
	}
}

func TestCollectorFormatAll(t *testing.T) {
	c := diagnostics.NewCollector()
	loc := diagnostics.SourceLocation{File: "test.html", Line: 1, Column: 1}
	c.AddError(loc, "E001", "first error")
	c.AddError(loc, "E002", "second error")

	formatted := c.FormatAll()
	lines := strings.Split(formatted, "\n")
	if len(lines) != 2 {
		t.Errorf("FormatAll() produced %d lines, want 2", len(lines))
	}
}

func TestCollectorAllReturnsDefensiveCopy(t *testing.T) {
	c := diagnostics.NewCollector()
	loc := diagnostics.SourceLocation{File: "test.html", Line: 1, Column: 1}
	c.AddError(loc, "E001", "original")

	all := c.All()
	all[0].Message = "modified"

	// Original should be unchanged
	original := c.All()
	if original[0].Message != "original" {
		t.Error("All() should return a defensive copy")
	}
}
