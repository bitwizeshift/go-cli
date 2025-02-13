/*
Package diagnostictest provides mechanisms for testing diagnostics emitted by
diagnostic reporters.
*/
package diagnostictest

import (
	"strings"

	"rodusek.dev/pkg/cli/diagnostic"
)

// Recorder is a simple structure that records diagnostics for testing purposes.
type Recorder struct {
	// Diagnostics is a slice of diagnostics that have been recorded.
	Diagnostics []*diagnostic.Diagnostic
}

// NewReporter creates a new [diagnostic.Reporter] that records diagnostics to
// the given [Recorder]. If the [Recorder] is nil, the returned reporter will
// not emit any diagnostics.
func NewReporter(recorder *Recorder) *diagnostic.Reporter {
	if recorder == nil {
		return diagnostic.NewNoopReporter()
	}
	reporter := diagnostic.NewReporter(recorder)
	reporter.ShowDebug(true)
	return reporter
}

// Filter returns a slice of diagnostics that match the given conditions.
func (r *Recorder) Filter(conditions ...Condition) []*diagnostic.Diagnostic {
	var diagnostics []*diagnostic.Diagnostic
	for _, d := range r.Diagnostics {
		if matchesAll(conditions, d) {
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics
}

// Count counts the number of diagnostics that match the given
// conditions.
func (r *Recorder) Count(conditions ...Condition) int {
	return len(r.Filter(conditions...))
}

// Contains returns true if there is at least one diagnostic that matches the
// given conditions.
func (r *Recorder) Contains(c ...Condition) bool {
	return r.Count(c...) > 0
}

// Emit records a diagnostic message.
// This is to satisfy the [diagnostic.Emitter] interface.
func (r *Recorder) Emit(msg *diagnostic.Diagnostic) error {
	r.Diagnostics = append(r.Diagnostics, msg)
	return nil
}

var _ diagnostic.Emitter = (*Recorder)(nil)

// Condition is a function that determines whether a diagnostic matches a
// certain condition.
type Condition func(*diagnostic.Diagnostic) bool

// HasMessage returns a [Condition] that checks if a diagnostic has a specific
// message.
func HasMessage(message string) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		return d.Message == message
	}
}

// ContainsMessage returns a [Condition] that checks if a diagnostic contains a
// substring of a message.
func ContainsMessage(substr string) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		return strings.Contains(d.Message, substr)
	}
}

// HasTitle returns a [Condition] that checks if a diagnostic has a specific
// title.
func HasTitle(s string) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		return d.Title == s
	}
}

// ContainsTitle returns a [Condition] that checks if a diagnostic contains a
// substring of a title.
func ContainsTitle(s string) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		return strings.Contains(d.Title, s)
	}
}

// HasSeverity returns a [Condition] that checks if a diagnostic has a specific
// severity.
func HasSeverity(severity ...diagnostic.Severity) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		for _, s := range severity {
			if d.Severity == s {
				return true
			}
		}
		return false
	}
}

// HasCode returns a [Condition] that checks if a diagnostic has a specific code.
func HasCode(s string) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		return d.Code == s
	}
}

// HasFile returns a [Condition] that checks if a diagnostic has a specific file.
func HasFile(s string) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		return d.File == s
	}
}

// HasStart returns a [Condition] that checks if a diagnostic has a specific
// start position.
func HasStart(line, column int) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		return d.Start.Line == line && d.Start.Column == column
	}
}

// HasEnd returns a [Condition] that checks if a diagnostic has a specific end
// position.
func HasEnd(line, column int) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		return d.End.Line == line && d.End.Column == column
	}
}

// HasRange returns a [Condition] that checks if a diagnostic has a specific
// range.
func HasRange(startLine, startColumn, endLine, endColumn int) Condition {
	return func(d *diagnostic.Diagnostic) bool {
		return d.Start.Line == startLine && d.Start.Column == startColumn &&
			d.End.Line == endLine && d.End.Column == endColumn
	}
}

func matchesAll(c []Condition, d *diagnostic.Diagnostic) bool {
	for _, condition := range c {
		if !condition(d) {
			return false
		}
	}
	return true
}
