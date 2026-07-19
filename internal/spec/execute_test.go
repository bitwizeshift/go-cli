package spec_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/bitwizeshift/go-cli/internal/clictx"
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
			var stderr strings.Builder
			sut := newRootCommand(t, tc.runner, &stderr)
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
	var stderr strings.Builder
	sut := newRootCommand(t, spectest.Err(testErr), &stderr)

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
	var stderr strings.Builder
	sut := newRootCommand(t, spectest.NoOpRunner(), &stderr)
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
	var stderr strings.Builder
	sut := newRootCommand(t, spectest.NoOpRunner(), &stderr)
	sut.Flags().String("token", "", "")
	argdef.AddFuncFallback(sut.Flags().Lookup("token"), func(context.Context) (string, error) {
		return "", fallbackErr
	})
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

// storageCapture is a [spec.Runner] that records whether application storage
// was present on the context it ran with.
type storageCapture struct {
	injected bool
}

func (sc *storageCapture) Run(ctx context.Context) error {
	sc.injected = clictx.Storage(ctx) != nil
	return nil
}

func TestExecute_InjectsStorage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		spec    string
		binding string
		want    bool
	}{
		{
			name:    "ExplicitAppID",
			spec:    "name: root\napp-id: com.example.app\n",
			binding: "root",
			want:    true,
		},
		{
			name:    "AppIDFromName",
			spec:    "name: mytool\n",
			binding: "mytool",
			want:    true,
		},
		{
			name:    "AppIDFromBinary",
			spec:    "name: \"\"\n",
			binding: "",
			want:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			runner := &storageCapture{}
			sut := build(t, tc.spec, spec.Options{
				Builders: toBuilders(map[string]spec.Runner{tc.binding: runner}),
				Stdout:   io.Discard,
				Stderr:   io.Discard,
			})
			ctx := context.Background()

			// Act
			err := spec.Execute(ctx, sut)

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("spec.Execute(...) = %v, want %v", got, want)
			}
			if got, want := runner.injected, tc.want; got != want {
				t.Errorf("storage injected = %t, want %t", got, want)
			}
		})
	}
}

// positionalCapture is a [spec.Runner] and [arg.Registrar] that binds a leading
// positional argument and collects the remaining unmatched arguments.
type positionalCapture struct {
	first string
	rest  []string
}

func (pc *positionalCapture) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(
		arg.Positional("first", 0, &pc.first),
		arg.Unmatched("rest", &pc.rest),
	)
}

func (pc *positionalCapture) Run(context.Context) error {
	return nil
}

func TestExecute_BindsPositionalArguments(t *testing.T) {
	t.Parallel()

	// Arrange
	runner := &positionalCapture{}
	sut := build(t, "name: root\n", spec.Options{
		Builders: toBuilders(map[string]spec.Runner{"root": runner}),
		Stdout:   io.Discard,
		Stderr:   io.Discard,
	})
	sut.SetArgs([]string{"alpha", "beta", "gamma"})
	ctx := context.Background()

	// Act
	err := spec.Execute(ctx, sut)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("spec.Execute(...) = %v, want %v", got, want)
	}
	if got, want := runner.first, "alpha"; !cmp.Equal(got, want) {
		t.Errorf("positional first = %q, want %q", got, want)
	}
	if got, want := runner.rest, []string{"beta", "gamma"}; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("unmatched arguments = %v, want %v", got, want)
	}
}

// positionalFailure is a [spec.Runner] and [arg.Registrar] whose positional
// argument fails to decode a non-numeric value.
type positionalFailure struct {
	count int
}

func (pf *positionalFailure) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(arg.Positional("count", 0, &pf.count))
}

func (pf *positionalFailure) Run(context.Context) error {
	return nil
}

func TestExecute_PositionalBindError_ShowsUsage(t *testing.T) {
	t.Parallel()

	// Arrange
	var stderr strings.Builder
	sut := build(t, "name: root\n", spec.Options{
		Builders: toBuilders(map[string]spec.Runner{"root": &positionalFailure{}}),
		Stdout:   io.Discard,
		Stderr:   &stderr,
	})
	sut.SetArgs([]string{"not-a-number"})
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

// newRootCommand builds a single root command bound to runner, routing both of
// its output streams to w.
func newRootCommand(t testing.TB, runner spec.Runner, w io.Writer) *cobra.Command {
	t.Helper()
	cmd, err := spec.Build(strings.NewReader("name: root\n"), spec.Options{
		Builders: toBuilders(map[string]spec.Runner{"root": runner}),
		Stdout:   w,
		Stderr:   w,
	})
	if err != nil {
		t.Fatalf("spec.Build(...) = _, %v, want nil", err)
	}
	return cmd
}
