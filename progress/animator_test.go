package progress_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/progress"
	"github.com/bitwizeshift/go-cli/progress/progresstest"
)

func TestAnimator_Tick(t *testing.T) {
	t.Parallel()

	testErr := errors.New("write refused")
	testCases := []struct {
		name      string
		sends     int
		ok        int
		wantFrame int
		want      string
		wantErr   error
	}{
		{
			name:      "NoTicks",
			sends:     0,
			ok:        alwaysOK,
			wantFrame: 0,
			want:      "",
			wantErr:   nil,
		}, {
			name:      "OneTick",
			sends:     1,
			ok:        alwaysOK,
			wantFrame: 1,
			want:      "\\\n",
			wantErr:   nil,
		}, {
			name:      "TwoTicks",
			sends:     2,
			ok:        alwaysOK,
			wantFrame: 2,
			want:      "\\\r\x1b[0J|\n",
			wantErr:   nil,
		}, {
			name:      "RedrawFailureIsReturnedByStop",
			sends:     1,
			ok:        0,
			wantFrame: 1,
			want:      "",
			wantErr:   testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ticker := progresstest.NewTicker()
			fw := &flakyWriter{ok: tc.ok, err: testErr}
			spinner := progress.Spinner{Frames: progress.LineFrames}
			writer := &progress.Writer{W: fw}
			sut := &progress.Animator{
				Writer:    writer,
				Target:    &spinner,
				NewTicker: func() progress.Ticker { return ticker },
			}
			ctx := context.Background()

			// Act
			sut.Start(ctx)
			for range tc.sends {
				ticker.Send()
			}
			err := sut.Stop()

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Animator.Stop() = %v, want %v", got, want)
			}
			if got, want := spinner.Frame, tc.wantFrame; !cmp.Equal(got, want) {
				t.Errorf("Animator advanced Frame = %d, want %d", got, want)
			}
			if got, want := fw.buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Animator output = %q, want %q", got, want)
			}
		})
	}
}

func TestAnimator_ContextCancellationStopsLoop(t *testing.T) {
	t.Parallel()

	// Arrange
	ticker := progresstest.NewTicker()
	spinner := progress.Spinner{Frames: progress.LineFrames}
	var fw flakyWriter
	fw.ok = alwaysOK
	writer := &progress.Writer{W: &fw}
	sut := &progress.Animator{
		Writer:    writer,
		Target:    &spinner,
		NewTicker: func() progress.Ticker { return ticker },
	}
	ctx, cancel := context.WithCancel(context.Background())

	// Act
	sut.Start(ctx)
	cancel()
	err := sut.Stop()

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Animator.Stop() = %v, want %v", got, want)
	}
	if got, want := spinner.Frame, 0; !cmp.Equal(got, want) {
		t.Errorf("Animator advanced Frame = %d, want %d", got, want)
	}
}

func TestAnimator_DefaultTickerStartsAndStopsCleanly(t *testing.T) {
	t.Parallel()

	// Arrange
	spinner := progress.LineSpinner
	var fw flakyWriter
	fw.ok = alwaysOK
	writer := &progress.Writer{W: &fw, DisableAnimation: true}
	sut := &progress.Animator{Writer: writer, Target: &spinner}
	ctx := context.Background()

	// Act
	sut.Start(ctx)
	err := sut.Stop()

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Animator.Stop() = %v, want %v", got, want)
	}
}
