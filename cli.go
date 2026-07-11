package cli

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/spf13/cobra"
)

// Runner is the generalized "run" behavior that each bound command executes.
//
// Implementations may additionally implement [flag.Registrar],
// or expose reachable fields that do, so that flags are registered automatically
// when the command tree is built.
type Runner interface {
	// Run executes the command with the resolved positional arguments.
	Run(ctx context.Context, args ...string) error
}

// PanicError is the error produced when a [Runner] terminates by panicking,
// carrying the recovered value and the captured stack trace.
type PanicError = spec.PanicError

var (
	// ErrUsage is a sentinel error that, when returned from a [Runner], prints
	// the command's usage message.
	ErrUsage = spec.ErrUsage

	// ErrPanic is the sentinel error a [PanicError] unwraps to.
	ErrPanic = spec.ErrPanic
)

// CLI is a fully assembled command-line application ready to be executed.
type CLI struct {
	cmd *cobra.Command
}

// FromReader builds a [CLI] from a YAML specification read from r, binding
// runners supplied via [BindRunner].
//
// It panics if the specification cannot be decoded or if a bound runner id
// matches no command, since the specification is expected to be embedded in the
// binary and therefore known-good at build time.
func FromReader(r io.Reader, options ...Option) *CLI {
	cfg := newConfig(options...)
	cmd, err := spec.Build(r, spec.Options{
		Runners: cfg.runners,
		Theme:   cfg.theme,
		Colour:  cfg.colour,
	})
	if err != nil {
		panic("cli: " + err.Error())
	}
	return &CLI{cmd: cmd}
}

// FromBytes builds a [CLI] from a YAML specification held in data. It is a
// convenience wrapper around [FromReader].
func FromBytes(data []byte, options ...Option) *CLI {
	return FromReader(bytes.NewReader(data), options...)
}

// CobraCommand returns the underlying [github.com/spf13/cobra.Command] for
// callers that need direct access to the command tree.
func (c *CLI) CobraCommand() *cobra.Command {
	return c.cmd
}

// Run executes the application against ctx and reports the resulting [ExitCode]
// without terminating the process. It returns [ExitSuccess] on success,
// [ExitPanic] for a recovered panic, and [ExitError] for any other error.
func (c *CLI) Run(ctx context.Context) ExitCode {
	switch err := spec.Execute(ctx, c.cmd); {
	case err == nil:
		return ExitSuccess
	case errors.Is(err, ErrPanic):
		return ExitPanic
	default:
		return ExitError
	}
}

// Execute runs the application against a background context and terminates the
// process with the resulting [ExitCode]. It does not return.
func (c *CLI) Execute() {
	ctx := context.Background()
	c.Run(ctx).Exit()
}
