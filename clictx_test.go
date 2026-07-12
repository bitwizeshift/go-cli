package cli_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/internal/clictx"
	"github.com/bitwizeshift/go-cli/internal/term"
)

// sameWriter compares [io.Writer] values by identity rather than by their
// unexported contents.
var sameWriter = cmp.Comparer(func(lhs, rhs io.Writer) bool {
	return lhs == rhs
})

func TestOutStream(t *testing.T) {
	t.Parallel()

	// Arrange
	stdout := &bytes.Buffer{}
	ctx := clictx.WithWriters(context.Background(), stdout, &bytes.Buffer{})

	// Act
	writer := cli.OutStream(ctx)

	// Assert
	if got, want := writer, io.Writer(stdout); !cmp.Equal(got, want, sameWriter) {
		t.Errorf("OutStream(ctx) = %v, want %v", got, want)
	}
}

func TestErrStream(t *testing.T) {
	t.Parallel()

	// Arrange
	stderr := &bytes.Buffer{}
	ctx := clictx.WithWriters(context.Background(), &bytes.Buffer{}, stderr)

	// Act
	writer := cli.ErrStream(ctx)

	// Assert
	if got, want := writer, io.Writer(stderr); !cmp.Equal(got, want, sameWriter) {
		t.Errorf("ErrStream(ctx) = %v, want %v", got, want)
	}
}

func TestStreamColumns(t *testing.T) {
	t.Parallel()

	// Arrange
	const fixedColumns = 42
	stderr := &bytes.Buffer{}
	ctx := clictx.WithSizer(context.Background(), term.FixedSizer(fixedColumns))

	// Act
	columns := cli.StreamColumns(ctx, stderr)

	// Assert
	if got, want := columns, fixedColumns; !cmp.Equal(got, want) {
		t.Errorf("StreamColumns(ctx) = %v, want %v", got, want)
	}
}
