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
