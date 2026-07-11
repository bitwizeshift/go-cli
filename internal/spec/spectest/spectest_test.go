package spectest_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/bitwizeshift/go-cli/internal/spec/spectest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRunners(t *testing.T) {
	t.Parallel()

	errSentinel := errors.New("sentinel")

	testCases := []struct {
		name    string
		sut     spec.Runner
		wantErr error
	}{
		{
			name:    "OK succeeds",
			sut:     spectest.NoOpRunner(),
			wantErr: nil,
		},
		{
			name:    "Err returns the error",
			sut:     spectest.Err(errSentinel),
			wantErr: errSentinel,
		},
		{
			name:    "Usage returns ErrUsage",
			sut:     spectest.UsageRunner(),
			wantErr: spec.ErrUsage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()

			// Act
			err := tc.sut.Run(ctx)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("sut.Run(ctx) = %v, want %v", got, want)
			}
		})
	}
}

func TestPanic_WhenRun_Panics(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := spectest.PanicRunner("boom")
	ctx := context.Background()

	// Act
	recovered := recoverRun(sut, ctx)

	// Assert
	if got, want := recovered, "boom"; !cmp.Equal(got, want) {
		t.Errorf("recovered panic = %v, want %v", got, want)
	}
}

// recoverRun runs sut and returns the value it panics with, or nil if it does
// not panic.
func recoverRun(sut spec.Runner, ctx context.Context) (recovered any) {
	defer func() { recovered = recover() }()
	_ = sut.Run(ctx)
	return nil
}
