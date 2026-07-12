package clitest

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bitwizeshift/go-cli/internal/clictx"
	"github.com/bitwizeshift/go-cli/richtext"
)

// Output is an object that provides captured stdout, stderr, or combined output
// from a CLI runner. This type is configured and provided by [WithCaptureWriters]
// as part of a context construction.
type Output struct {
	// Stdout is a Stringer representing the current stdout output at the
	// time it is queried.
	Stdout fmt.Stringer

	// Stderr is a Stringer representing the current stderr output at the
	// time it is queried.
	Stderr fmt.Stringer

	// Combined is a Stringer representing the combined stdout and stderr at the
	// time it is queried.
	Combined fmt.Stringer
}

// WithCaptureWriters returns a context that configures an [Output] variable
// that allows for retrieving captured stdout, stderr, or combined stdout/stderr
// output as a string.
func WithCaptureWriters(ctx context.Context) (context.Context, *Output) {
	var outStream, errStream, mergedStream strings.Builder
	ctx = WithWriters(ctx,
		io.MultiWriter(&outStream, &mergedStream),
		io.MultiWriter(&errStream, &mergedStream),
	)
	return ctx, &Output{
		Stdout:   &outStream,
		Stderr:   &errStream,
		Combined: &mergedStream,
	}
}

// WithWriters assigns richtext writers to the context for the purposes of
// testing.
//
// If stdout or stderr does not refer to a *[richtext.Writer] object, the writer
// will be wrapped in one first that disables colours.
func WithWriters(ctx context.Context, stdout, stderr io.Writer) context.Context {
	stdout = defaultRichTextWriter(stdout)
	stderr = defaultRichTextWriter(stderr)
	return clictx.WithWriters(ctx, stdout, stderr)
}

func defaultRichTextWriter(w io.Writer) io.Writer {
	if w, ok := w.(*richtext.Writer); ok {
		return w
	}
	writer := richtext.NewWriter(w, richtext.DefaultTheme)
	writer.EnableColour(false)
	return writer
}
