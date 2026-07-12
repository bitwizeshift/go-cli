package cli

import (
	"context"
	"io"

	"github.com/bitwizeshift/go-cli/internal/clictx"
)

// OutStream returns the output writer from the application's context.
func OutStream(ctx context.Context) io.Writer {
	stdout, _ := clictx.Writers(ctx)
	return stdout
}

// ErrStream returns the error writer from the application's context.
func ErrStream(ctx context.Context) io.Writer {
	_, stderr := clictx.Writers(ctx)
	return stderr
}

// StreamColumns reports the column width to render w at. Width is resolved per
// writer, which can determine whether it's a terminal stream or some other
// non-TTY stream.
func StreamColumns(ctx context.Context, w io.Writer) int {
	return clictx.Columns(ctx, w)
}
