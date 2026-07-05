package termtest_test

import (
	"errors"
	"io"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/bitwizeshift/go-cli/internal/term/termtest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// mustDisableEcho disables echo, failing the test if the disable itself errors,
// and returns the restore func.
func mustDisableEcho(t *testing.T, d term.EchoDisabler) term.RestoreFunc {
	t.Helper()
	restore, err := d.DisableEcho(io.Discard)
	if err != nil {
		t.Fatalf("DisableEcho() unexpected error: %v", err)
	}
	return restore
}

func TestEchoDisabler_DisableEcho(t *testing.T) {
	t.Parallel()

	testErr := errors.New("echo unavailable")

	testCases := []struct {
		name     string
		disabler term.EchoDisabler
		wantErr  error
	}{
		{
			name:     "NoOp",
			disabler: termtest.NoOpEchoDisabler(),
			wantErr:  nil,
		}, {
			name:     "ErrEchoDisabler",
			disabler: termtest.ErrEchoDisabler(testErr),
			wantErr:  testErr,
		}, {
			name:     "ErrRestoreDisabler",
			disabler: termtest.ErrRestoreDisabler(testErr),
			wantErr:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.disabler

			// Act
			_, err := sut.DisableEcho(io.Discard)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("DisableEcho(...) error got %v, want %v", got, want)
			}
		})
	}
}

func TestRecorder_DisableEcho(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := &termtest.Recorder{}

	// Act
	restore := mustDisableEcho(t, sut)
	disabledDuring := sut.Disabled
	restoreErr := restore()

	// Assert
	if got, want := disabledDuring, true; !cmp.Equal(got, want) {
		t.Errorf("Recorder.Disabled during read got %v, want %v", got, want)
	}
	if got, want := restoreErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("restore() error got %v, want %v", got, want)
	}
	if got, want := sut.Disabled, false; !cmp.Equal(got, want) {
		t.Errorf("Recorder.Disabled after restore got %v, want %v", got, want)
	}
}

func TestEchoDisabler_Restore(t *testing.T) {
	t.Parallel()

	testErr := errors.New("restore unavailable")

	testCases := []struct {
		name     string
		disabler term.EchoDisabler
		wantErr  error
	}{
		{
			name:     "NoOp",
			disabler: termtest.NoOpEchoDisabler(),
			wantErr:  nil,
		}, {
			name:     "ErrRestoreDisabler",
			disabler: termtest.ErrRestoreDisabler(testErr),
			wantErr:  testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			restore := mustDisableEcho(t, tc.disabler)

			// Act
			err := restore()

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("restore() error got %v, want %v", got, want)
			}
		})
	}
}
