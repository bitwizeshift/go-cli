package clictx

import (
	"context"
	"io"
	"os"

	"github.com/bitwizeshift/go-cli/internal/storage"
	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/bitwizeshift/go-cli/richtext"
)

type ctxKey int

const (
	ctxKeyIO ctxKey = iota
	ctxKeySizer
	ctxKeyStorage
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

// WithSizer returns a copy of ctx carrying sizer as the policy for resolving
// terminal width, retrievable through [Columns].
func WithSizer(ctx context.Context, sizer term.Sizer) context.Context {
	return context.WithValue(ctx, ctxKeySizer, sizer)
}

// Columns reports the column width to render w at, using the [term.Sizer] stored
// on ctx by [WithSizer]. It falls back to [term.DefaultSizer] when ctx carries no
// sizer.
func Columns(ctx context.Context, w io.Writer) int {
	sizer := term.Sizer(term.DefaultSizer)
	if v := ctx.Value(ctxKeySizer); v != nil {
		sizer = v.(term.Sizer)
	}
	return sizer.Columns(underlying(w))
}

// WithStorage returns a copy of ctx carrying app as the application's storage
// roots, retrievable with [Storage].
func WithStorage(ctx context.Context, app *storage.AppStorage) context.Context {
	return context.WithValue(ctx, ctxKeyStorage, app)
}

// Storage returns the [storage.AppStorage] stored on ctx by [WithStorage], or
// nil when ctx carries none.
func Storage(ctx context.Context) *storage.AppStorage {
	if app, ok := ctx.Value(ctxKeyStorage).(*storage.AppStorage); ok {
		return app
	}
	return nil
}

// underlying returns the writer beneath w, following any writer that exposes a
// Writer() io.Writer method, so sizing can reach the file descriptor of the real
// terminal rather than a markup writer wrapped around it.
func underlying(w io.Writer) io.Writer {
	for {
		next, ok := w.(interface{ Writer() io.Writer })
		if !ok {
			return w
		}
		w = next.Writer()
	}
}
