package term_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	xterm "golang.org/x/term"
)

// mustDisableEcho disables echo on w, failing the test if the disable itself
// errors, and returns the restore func.
func mustDisableEcho(t *testing.T, c *term.Console, w io.Writer) term.RestoreFunc {
	t.Helper()
	restore, err := c.DisableEcho(w)
	if err != nil {
		t.Fatalf("DisableEcho() unexpected error: %v", err)
	}
	return restore
}

func TestConsole_DisableEcho(t *testing.T) {
	t.Parallel()

	testErr := errors.New("make raw failed")

	testCases := []struct {
		name       string
		writer     io.Writer
		makeRawErr error
		wantErr    error
	}{
		{
			name:       "MakeRawFails",
			writer:     &fdWriter{fd: 1},
			makeRawErr: testErr,
			wantErr:    testErr,
		}, {
			name:       "WriterWithoutDescriptor",
			writer:     &bytes.Buffer{},
			makeRawErr: nil,
			wantErr:    term.ErrNotDescriptor,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := &term.Console{
				MakeRaw: func(int) (*xterm.State, error) { return nil, tc.makeRawErr },
				Restore: func(int, *xterm.State) error { return nil },
			}

			// Act
			_, err := sut.DisableEcho(tc.writer)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Console.DisableEcho(...) error got %v, want %v", got, want)
			}
		})
	}
}

func TestConsole_DisableEcho_Restore(t *testing.T) {
	t.Parallel()

	testErr := errors.New("restore failed")

	testCases := []struct {
		name       string
		restoreErr error
		wantErr    error
	}{
		{
			name:       "RestoreSucceeds",
			restoreErr: nil,
			wantErr:    nil,
		}, {
			name:       "RestoreFails",
			restoreErr: testErr,
			wantErr:    testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := &term.Console{
				MakeRaw: func(int) (*xterm.State, error) { return nil, nil },
				Restore: func(int, *xterm.State) error { return tc.restoreErr },
			}
			restore := mustDisableEcho(t, sut, &fdWriter{fd: 1})

			// Act
			err := restore()

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("restore() error got %v, want %v", got, want)
			}
		})
	}
}
