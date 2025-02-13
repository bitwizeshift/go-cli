package diagnostic

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
	"golang.org/x/term"
)

// Flags is a simple embeddable struct that provides a common set of flags
// for configuring diagnostic output
type Flags struct {
	outputFormat outputFormatFlag
	debug        bool
}

// RegisterFlags registers the diagnostic flags with the given flagset
func (f *Flags) RegisterFlags(fs *pflag.FlagSet) {
	fs.Var(&f.outputFormat, "output-format", "The format to output diagnostics in")
	fs.BoolVar(&f.debug, "debug", false, "enable debug output")
}

// Reporter creates a new reporter based on the flags provided.
func (f *Flags) Reporter(w io.Writer) *Reporter {
	reporter := NewNoopReporter()
	switch f.outputFormat {
	case "text":
		reporter = NewTextReporter(w)
	case "terminal":
		reporter = NewTerminalReporter(w, nil)
	case "json":
		reporter = NewJSONReporter(w)
	case "github":
		reporter = NewGitHubReporter(w)
	case "log":
		reporter = NewLogReporter(w)
	default:
		if fd, ok := w.(interface{ Fd() uintptr }); ok && term.IsTerminal(int(fd.Fd())) {
			reporter = NewTerminalReporter(w, nil)
		} else {
			reporter = NewTextReporter(w)
		}
	}
	reporter.ShowDebug(f.debug)
	return reporter
}

type outputFormatFlag string

func (f *outputFormatFlag) UnmarshalText(b []byte) error {
	switch string(b) {
	case "text", "terminal", "json", "github", "log", "none":
		*f = outputFormatFlag(b)
	default:
		return fmt.Errorf("output-format: invalid value '%s'", b)
	}
	return nil
}

func (f *outputFormatFlag) Set(s string) error {
	return f.UnmarshalText([]byte(s))
}

func (f *outputFormatFlag) Type() string {
	return "format"
}

func (f *outputFormatFlag) String() string {
	return string(*f)
}
