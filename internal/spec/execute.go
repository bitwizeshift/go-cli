package spec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/bitwizeshift/go-cli/internal/clictx"
	"github.com/bitwizeshift/go-cli/internal/storage"
	"github.com/bitwizeshift/go-cli/internal/template"
	"github.com/bitwizeshift/go-cli/internal/template/panichandler"
	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/spf13/cobra"
)

// Execute runs cmd against ctx, rendering any resulting error, usage advisory,
// or panic report to the failing command's error stream.
//
// It returns nil on success, [ErrPanic] for a recovered panic, [ErrUsage] for
// an explicit usage error, or the runner's error otherwise. The returned error
// has already been reported to the user and is intended only for exit-status
// classification.
func Execute(ctx context.Context, cmd *cobra.Command) error {
	stdout := cmd.OutOrStdout()
	stderr := cmd.ErrOrStderr()
	defer closeStream(stdout)
	defer closeStream(stderr)

	target, err := cmd.ExecuteContextC(ctx)
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrPanic):
		// The panic report was already rendered while unwinding the runner.
	case errors.Is(err, ErrUsage):
		_ = target.Usage()
	case fromRunner(err):
		renderError(stderr, err)
	default:
		renderError(stderr, err)
		_ = target.Usage()
	}
	return err
}

// renderError writes a styled, newline-terminated error message to w.
func renderError(w io.Writer, err error) {
	_ = template.DefaultRenderEngine.Errorf(w, "%v", err)
	_, _ = fmt.Fprintln(w)
}

// closeStream flushes w when it is an [io.Closer] that implements Flush() error
func closeStream(w io.Writer) {
	if rw, ok := w.(io.Closer); ok {
		_ = rw.Close()
	}
}

// fromRunner reports whether err originated from a [Runner], as opposed to an
// argument-parsing failure produced before the runner was reached.
func fromRunner(err error) bool {
	var re runnerError
	return errors.As(err, &re)
}

// run adapts a [Runner] into a cobra RunE, installing signal-cancellation and
// panic recovery. A recovered panic is rendered as a crash report and returned
// as a [PanicError]; any other error is wrapped so that [Execute] can tell it
// apart from an argument-parsing failure. store is placed on the context so the
// runner can reach the application's storage roots.
func (i *CommandInfo) run(builder Builder, store *storage.AppStorage, cl *arg.CommandLine) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer cancel()
		stdout := cmd.OutOrStdout()
		stderr := cmd.ErrOrStderr()
		ctx = clictx.WithWriters(ctx, stdout, stderr)
		ctx = clictx.WithSizer(ctx, term.DefaultSizer)
		ctx = clictx.WithStorage(ctx, store)

		defer func() {
			if e := recover(); e != nil {
				stack := debug.Stack()
				err = PanicError{Err: e, Stack: stack}
				stderr := cmd.ErrOrStderr()
				pctx := panichandler.PanicContext{
					Err:      e,
					Stack:    stack,
					IssueURL: argdef.IssueURL(cmd),
				}
				_ = template.DefaultRenderEngine.PanicRenderer().Render(stderr, pctx)
			}
		}()

		// Compute fallback defaults for flags that aren't set, and assign the
		// values.
		if e := argdef.SetFlagFallbacks(ctx, cmd.Flags()); e != nil {
			return fmt.Errorf("%w: %w", ErrUsage, e)
		}

		// Bind positional and unmatched arguments into the runner before it runs.
		if e := argdef.Bind(ctx, (*argdef.CommandLine)(cl), args); e != nil {
			return fmt.Errorf("%w: %w", ErrUsage, e)
		}

		runner, e := builder.Build(ctx)
		if e != nil {
			return e
		}

		if e := runner.Run(ctx); e != nil {
			return runnerError{err: e}
		}
		return nil
	}
}

// runnerError marks an error as originating from a [Runner], distinguishing a
// runtime failure from a usage error produced by argument parsing.
type runnerError struct {
	err error
}

// Error returns the wrapped error's message.
func (re runnerError) Error() string { return re.err.Error() }

// Unwrap returns the wrapped error.
func (re runnerError) Unwrap() error { return re.err }
