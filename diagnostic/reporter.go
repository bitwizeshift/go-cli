package diagnostic

import (
	"io"
	"log"
	"os"

	"rodusek.dev/pkg/cli/formatting/wrap"
)

// Emitter represents a mechanism to emit [Diagnostic] messages.
type Emitter interface {
	// Emit emits a diagnostic message.
	Emit(msg *Diagnostic) error
}

// Metrics collects the number of diagnostics emitted by a [Reporter].
type Metrics struct {
	// Errors is the number of error diagnostics emitted.
	Errors int

	// Warnings is the number of warning diagnostics emitted.
	Warnings int

	// Notices is the number of notice diagnostics emitted.
	Notices int

	// Debugs is the number of debug diagnostics emitted.
	Debugs int
}

// Reporter reports [Diagnostic] messages to a given output.
type Reporter struct {
	// emitter is the underlying emitter that emits diagnostics.
	emitter Emitter

	// metrics are the metrics collected by the reporter.
	metrics Metrics

	// showDebug is a flag that determines whether debug messages should be
	// emitted. If false, debug messages won't be emitted but will still be
	// counted in the metrics.
	showDebug bool
}

// NewReporter creates a new [Reporter] that emits diagnostics to the given
// [Emitter]. If the [Emitter] is nil, the returned reporter will not emit
// any diagnostics.
func NewReporter(e Emitter) *Reporter {
	if e == nil {
		e = noopEmitter{}
	}
	return &Reporter{
		emitter: e,
	}
}

// NewGitHubReporter creates a new [Reporter] that emits diagnostics to the
// given writer in a format that is suitable for GitHub Actions. If the writer
// is nil, the returned reporter will not emit any diagnostics.
func NewGitHubReporter(w io.Writer) *Reporter {
	var emitter Emitter = noopEmitter{}
	if w != nil {
		emitter = &jsonEmitter{writer: w}
	}
	return &Reporter{
		emitter: emitter,
	}
}

// TerminalOptions represents options for terminal reporters.
type TerminalOptions struct {
	// BasePath is the path to show files relative to when emitting terminal
	// diagnostics. If unspecified, the current working directory is used.
	BasePath string
}

func (to *TerminalOptions) basePath() string {
	if to == nil || to.BasePath == "" {
		cwd, _ := os.Getwd()
		return cwd
	}
	return to.BasePath
}

// NewTerminalReporter creates a new [Reporter] that emits diagnostics to the
// given writer in a format that is suitable for ANSI terminals. If the writer
// is nil, the returned reporter will not emit any diagnostics.
func NewTerminalReporter(w io.Writer, opts *TerminalOptions) *Reporter {
	var emitter Emitter = noopEmitter{}
	if w != nil {
		wrapper, _ := wrap.FromTerminal(w)
		emitter = &terminalEmitter{
			writer:   w,
			wrapper:  wrapper,
			basePath: opts.basePath(),
		}
	}
	return &Reporter{
		emitter: emitter,
	}
}

// NewTextReporter creates a new [Reporter] that emits diagnostics to the
// given writer in a text format. If the writer is nil, the returned reporter
// will not emit any diagnostics.
func NewTextReporter(w io.Writer) *Reporter {
	var emitter Emitter = noopEmitter{}
	if w != nil {
		emitter = &textEmitter{writer: w}
	}
	return &Reporter{
		emitter: emitter,
	}
}

// NewJSONReporter creates a new [Reporter] that emits diagnostics to the
// given writer in a JSON format. If the writer is nil, the returned reporter
// will not emit any diagnostics.
func NewJSONReporter(w io.Writer) *Reporter {
	var emitter Emitter = noopEmitter{}
	if w != nil {
		emitter = &jsonEmitter{writer: w}
	}
	return &Reporter{
		emitter: emitter,
	}
}

// NewLogReporter creates a new [Reporter] that emits diagnostics to the given
// writer in a log format. If the writer is nil, the returned reporter will not
// emit any diagnostics.
func NewLogReporter(w io.Writer) *Reporter {
	var emitter Emitter = noopEmitter{}
	if w != nil {
		emitter = &logEmitter{logger: log.New(w, "", 0)}
	}
	return &Reporter{
		emitter: emitter,
	}
}

// NewNoopReporter creates a new [Reporter] that does not emit any diagnostics.
func NewNoopReporter() *Reporter {
	return &Reporter{
		emitter: noopEmitter{},
	}
}

// Report emits a diagnostic message to the underlying [Emitter].
// If the provided diagnostic message is invalid -- such as having an invalid
// severity, or missing a required field -- an error will be returned.
func (r *Reporter) Report(msg *Diagnostic) error {
	if err := msg.validate(); err != nil {
		return err
	}
	switch msg.Severity {
	case SeverityError:
		r.metrics.Errors++
	case SeverityWarning:
		r.metrics.Warnings++
	case SeverityNotice:
		r.metrics.Notices++
	case SeverityDebug:
		r.metrics.Debugs++
		if !r.showDebug {
			return nil
		}
	}
	if r.emitter == nil {
		return nil
	}
	return r.emitter.Emit(msg)
}

// ShowDebug enables or disables the emission of debug messages.
func (r *Reporter) ShowDebug(b bool) {
	r.showDebug = b
}

// Errorf emits a formatted error diagnostic message.
func (r *Reporter) Errorf(format string, args ...any) error {
	return r.Report(Errorf(format, args...))
}

// Warningf emits a formatted warning diagnostic message.
func (r *Reporter) Warningf(format string, args ...any) error {
	return r.Report(Warningf(format, args...))
}

// Noticef emits a formatted notice diagnostic message.
func (r *Reporter) Noticef(format string, args ...any) error {
	return r.Report(Noticef(format, args...))
}

// Debugf emits a formatted debug diagnostic message.
func (r *Reporter) Debugf(format string, args ...any) error {
	return r.Report(Debugf(format, args...))
}

// Fatalf emits a formatted error diagnostic message and exits the program with
// a non-zero exit code.
func (r *Reporter) Fatalf(format string, args ...any) {
	r.Report(Errorf(format, args...))
	os.Exit(1)
}

// Metrics returns the metrics collected by the reporter.
func (r *Reporter) Metrics() *Metrics {
	return &r.metrics
}
