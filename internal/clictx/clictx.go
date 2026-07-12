package clictx

import (
	"context"
	"io"
	"os"

	"github.com/bitwizeshift/go-cli/richtext"
)

type ctxKey int

const (
	ctxKeyIO ctxKey = iota
)

type writerContext struct {
	stdout io.Writer
	stderr io.Writer
}

// WithWriters returns a copy of ctx carrying stdout and stderr as the command's
// output and error streams, retrievable with [Writers].
func WithWriters(ctx context.Context, stdout, stderr io.Writer) context.Context {
	ctx = context.WithValue(ctx, ctxKeyIO, writerContext{
		stdout: stdout,
		stderr: stderr,
	})
	return ctx
}

// Writers returns the output and error streams stored on ctx by [WithWriters].
// It returns [os.Stdout] and [os.Stderr] when ctx carries no streams.
func Writers(ctx context.Context) (stdout, stderr io.Writer) {
	if writers := ctx.Value(ctxKeyIO); writers != nil {
		writers := writers.(writerContext)
		return writers.stdout, writers.stderr
	}
	outStream := richtext.NewWriter(os.Stdout, richtext.DefaultTheme)
	errStream := richtext.NewWriter(os.Stderr, richtext.DefaultTheme)
	return outStream, errStream
}
