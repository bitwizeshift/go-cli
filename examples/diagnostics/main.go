// Command diagnostics is a self-contained showcase of the go-cli diagnostic
// logger.
//
// It reports a representative set of diagnostics (an error, a warning, and an
// info, some carrying source locations) through a [diagnostic.LoggerFlag]. The
// --output-format flag selects between the human-readable text rendering, GitHub
// Actions workflow commands, and newline-delimited JSON.
package main

import (
	"context"
	_ "embed"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/diagnostic"
)

// loggerProvider yields a [diagnostic.Logger] bound to the CLI's error stream
// and the format selected on the command line.
type loggerProvider interface {
	Logger(ctx context.Context) *diagnostic.Logger
}

// rootRunner backs the top-level command and emits a fixed set of diagnostics
// so each output format can be observed by changing --output-format.
type rootRunner struct {
	LoggerProvider loggerProvider
}

func (r *rootRunner) Run(ctx context.Context, _ ...string) error {
	logger := r.LoggerProvider.Logger(ctx)

	logger.Error(ctx, &diagnostic.Diagnostic{
		ID:      "E1001",
		Title:   "unresolved symbol",
		Message: "cannot find function `render` in this scope",
		Location: &diagnostic.Location{
			File:        "src/main.psx",
			LineStart:   42,
			ColumnStart: 5,
			ColumnEnd:   11,
		},
	})
	logger.Warn(ctx, &diagnostic.Diagnostic{
		ID:      "W2002",
		Title:   "deprecated call",
		Message: "`legacyInit` is deprecated; use `init` instead",
		Location: &diagnostic.Location{
			File:      "src/boot.psx",
			LineStart: 7,
		},
	})
	logger.Info(ctx, &diagnostic.Diagnostic{
		ID:      "I3003",
		Title:   "build complete",
		Message: "recompiled 3 modules in 1.2s",
	})
	return nil
}

var _ cli.Runner = (*rootRunner)(nil)

//go:embed app.yaml
var configYAML []byte

func main() {
	cli.FromBytes(configYAML,
		cli.BindRunner("root", &rootRunner{
			LoggerProvider: &diagnostic.LoggerFlag{},
		}),
	).Execute()
}
