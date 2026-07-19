package diagnostic

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/diagnostic/internal/diagslog"
	"github.com/bitwizeshift/go-cli/internal/term"
)

// FormatType selects one of the diagnostic output renderings.
type FormatType string

// Recognised values for FormatType.
const (
	FormatText   FormatType = "text"
	FormatGitHub FormatType = "github"
	FormatJSON   FormatType = "json"
)

// UnmarshalText implements encoding.TextUnmarshaler.
func (f *FormatType) UnmarshalText(text []byte) error {
	switch FormatType(text) {
	case FormatText, FormatGitHub, FormatJSON:
		*f = FormatType(text)
		return nil
	default:
		return fmt.Errorf("invalid format type: %s", text)
	}
}

// LoggerFlag represents a flag that can be used to control the behavior of a
// logger.
type LoggerFlag struct {
	LongFlag  string
	ShortFlag string
	Usage     string

	format FormatType
}

// RegisterArgs registers the LoggerFlag's flags with the provided FlagSet.
func (lf *LoggerFlag) RegisterArgs(cl *arg.CommandLine) {
	flagName := lf.LongFlag
	if flagName == "" {
		flagName = "output-format"
	}
	usage := lf.Usage
	if usage == "" {
		usage = "The output format for diagnostics to be printed in"
	}
	opts := []arg.FlagOption{
		arg.Usage(usage),
		arg.ValueLabel("format"),
		arg.CompleteFrom(string(FormatText), string(FormatGitHub), string(FormatJSON)),
	}
	if lf.ShortFlag != "" {
		opts = append(opts, arg.Shorthand(lf.ShortFlag))
	}
	cl.Add(arg.Flag(flagName, &lf.format, opts...))
}

var _ arg.Registrar = (*LoggerFlag)(nil)

// LoggerFor returns a [Logger] for the given writer.
func (lf *LoggerFlag) LoggerFor(w io.Writer) *Logger {
	var handler slog.Handler = diagslog.NewTextHandler(w, term.DefaultSizer)
	switch lf.format {
	case FormatJSON:
		handler = diagslog.NewJSONHandler(w)
	case FormatText:
		handler = diagslog.NewTextHandler(w, term.DefaultSizer)
	case FormatGitHub:
		handler = diagslog.NewGitHubHandler(w)
	}
	return NewLogger(handler)
}

// Logger returns a new [Logger] from the error stream stored within the
// CLI's context.
func (lf *LoggerFlag) Logger(ctx context.Context) *Logger {
	w := cli.ErrStream(ctx)
	return lf.LoggerFor(w)
}
