package spec_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/bitwizeshift/go-cli/internal/spec/spectest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
)

func TestExecute(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")
	testCases := []struct {
		name    string
		runner  spec.Runner
		args    []string
		wantErr error
	}{
		{
			name:    "success",
			runner:  spectest.NoOpRunner(),
			wantErr: nil,
		},
		{
			name:    "runner error",
			runner:  spectest.Err(testErr),
			wantErr: testErr,
		},
		{
			name:    "usage error",
			runner:  spectest.UsageRunner(),
			wantErr: spec.ErrUsage,
		},
		{
			name:    "recovered panic",
			runner:  spectest.PanicRunner("kaboom"),
			wantErr: spec.ErrPanic,
		},
		{
			name:    "argument parsing failure",
			runner:  spectest.NoOpRunner(),
			args:    []string{"--nope"},
			wantErr: cmpopts.AnyError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := newRootCommand(t, tc.runner)
			var stderr strings.Builder
			sut.SetOut(&stderr)
			sut.SetErr(&stderr)
			sut.SetArgs(tc.args)

			// Act
			err := spec.Execute(context.Background(), sut)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("spec.Execute(...) = %v, want %v", got, want)
			}
		})
	}
}

func TestExecute_RunnerError_OmitsUsageAdvisory(t *testing.T) {
	t.Parallel()

	// Arrange
	testErr := errors.New("test error")
	sut := newRootCommand(t, spectest.Err(testErr))
	var stderr strings.Builder
	sut.SetOut(&stderr)
	sut.SetErr(&stderr)

	// Act
	err := spec.Execute(context.Background(), sut)

	// Assert
	if got, want := err, testErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("spec.Execute(...) = %v, want %v", got, want)
	}
	if got, want := strings.Contains(stderr.String(), "test error"), true; got != want {
		t.Errorf("stderr contains error message = %t, want %t", got, want)
	}
}

func TestExecute_ParseError_ShowsUsageAdvisory(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newRootCommand(t, spectest.NoOpRunner())
	var stderr strings.Builder
	sut.SetOut(&stderr)
	sut.SetErr(&stderr)
	sut.SetArgs([]string{"--nope"})

	// Act
	err := spec.Execute(context.Background(), sut)

	// Assert
	if got, want := err, cmpopts.AnyError; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("spec.Execute(...) = %v, want %v", got, want)
	}
	if got, want := strings.Contains(stderr.String(), "unknown flag"), true; got != want {
		t.Errorf("stderr contains flag error = %t, want %t", got, want)
	}
}

func TestExecute_FallbackError_ShowsUsage(t *testing.T) {
	t.Parallel()

	// Arrange
	fallbackErr := errors.New("fallback failed")
	sut := newRootCommand(t, spectest.NoOpRunner())
	sut.Flags().String("token", "", "")
	annotation.AddFuncFallback(sut.Flags().Lookup("token"), func(context.Context) (string, error) {
		return "", fallbackErr
	})
	var stderr strings.Builder
	sut.SetOut(&stderr)
	sut.SetErr(&stderr)
	ctx := context.Background()

	// Act
	err := spec.Execute(ctx, sut)

	// Assert
	if got, want := err, spec.ErrUsage; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("spec.Execute(...) = %v, want %v", got, want)
	}
	if got, want := strings.Contains(stderr.String(), "--help"), true; got != want {
		t.Errorf("stderr shows usage = %t, want %t", got, want)
	}
}

// newRootCommand builds a single root command bound to runner.
func newRootCommand(t testing.TB, runner spec.Runner) *cobra.Command {
	t.Helper()
	cmd, err := spec.Build(strings.NewReader("id: root\nuse: root\n"), map[string]spec.Runner{"root": runner})
	if err != nil {
		t.Fatalf("spec.Build(...) = _, %v, want nil", err)
	}
	return cmd
}
