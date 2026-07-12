package cli_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/internal/spec/spectest"
	"github.com/bitwizeshift/go-cli/richtext"
	"github.com/google/go-cmp/cmp"
)

const rootWithChild = `
id: root
use: root
commands:
  default:
    - id: child
      use: child
`

func TestFromReader_BuildsCommandTree(t *testing.T) {
	t.Parallel()

	// Arrange
	reader := strings.NewReader(rootWithChild)

	// Act
	sut := cli.FromReader(reader, cli.BindRunner("root", spectest.NoOpRunner()))

	// Assert
	if got, want := sut.CobraCommand().Use, "root"; got != want {
		t.Errorf("cli.FromReader(...).CobraCommand().Use = %q, want %q", got, want)
	}
	if got, want := childNames(sut), []string{"child"}; !cmp.Equal(got, want) {
		t.Errorf("child names = %v, want %v", got, want)
	}
}

func TestFromBytes_BuildsCommandTree(t *testing.T) {
	t.Parallel()

	// Arrange
	data := []byte(rootWithChild)

	// Act
	sut := cli.FromBytes(data, cli.BindRunner("root", spectest.NoOpRunner()))

	// Assert
	if got, want := sut.CobraCommand().Use, "root"; got != want {
		t.Errorf("cli.FromBytes(...).CobraCommand().Use = %q, want %q", got, want)
	}
	if got, want := childNames(sut), []string{"child"}; !cmp.Equal(got, want) {
		t.Errorf("child names = %v, want %v", got, want)
	}
}

func TestFromBytes_InvalidSpecification_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		input        string
		options      []cli.Option
		wantContains string
	}{
		{
			name:         "invalid yaml",
			input:        "use: root\ncommands: [unclosed",
			wantContains: "cli:",
		},
		{
			name:         "unbound runner",
			input:        "id: root\nuse: root\n",
			options:      []cli.Option{cli.BindRunner("ghost", spectest.NoOpRunner())},
			wantContains: "no command",
		},
		{
			name:         "duplicate binding",
			input:        "id: root\nuse: root\n",
			options:      []cli.Option{cli.BindRunner("root", spectest.NoOpRunner()), cli.BindRunner("root", spectest.NoOpRunner())},
			wantContains: "duplicate",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			data := []byte(tc.input)

			// Act
			recovered := recoverPanic(func() {
				cli.FromBytes(data, tc.options...)
			})

			// Assert
			message, _ := recovered.(string)
			if got, want := strings.Contains(message, tc.wantContains), true; got != want {
				t.Fatalf("recovered panic = %q, want to contain %q", message, tc.wantContains)
			}
		})
	}
}

func TestFromReader_StyleOptions_Build(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []cli.Option
	}{
		{
			name:    "Theme",
			options: []cli.Option{cli.Theme(richtext.DefaultTheme)},
		},
		{
			name:    "DisableColour",
			options: []cli.Option{cli.DisableColour()},
		},
		{
			name:    "ForceColour",
			options: []cli.Option{cli.ForceColour()},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			options := append(tc.options, cli.BindRunner("root", spectest.NoOpRunner()))

			// Act
			sut := cli.FromBytes([]byte("id: root\nuse: root\n"), options...)

			// Assert
			if got, want := sut.CobraCommand().Use, "root"; !cmp.Equal(got, want) {
				t.Errorf("cli.FromBytes(...).CobraCommand().Use = %q, want %q", got, want)
			}
		})
	}
}

func TestFromReader_ConflictingColourOptions_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []cli.Option
	}{
		{
			name:    "DisableThenForce",
			options: []cli.Option{cli.DisableColour(), cli.ForceColour()},
		},
		{
			name:    "ForceThenDisable",
			options: []cli.Option{cli.ForceColour(), cli.DisableColour()},
		},
		{
			name:    "DisableTwice",
			options: []cli.Option{cli.DisableColour(), cli.DisableColour()},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			recovered := recoverPanic(func() {
				cli.FromBytes([]byte("id: root\nuse: root\n"), tc.options...)
			})

			// Assert
			const substr = "colour mode already set"
			message, _ := recovered.(string)
			if got, want := strings.Contains(message, substr), true; got != want {
				t.Fatalf("recovered panic = %q, want to contain %q", message, substr)
			}
		})
	}
}

func TestFromReader_InvalidSizeOptions_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []cli.Option
	}{
		{
			name:    "Negative",
			options: []cli.Option{cli.TerminalWidth(-20)},
		},
		{
			name:    "LessThan60",
			options: []cli.Option{cli.TerminalWidth(50)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			recovered := recoverPanic(func() {
				cli.FromBytes([]byte("id: root\nuse: root\n"), tc.options...)
			})

			// Assert
			const substr = "not enough"
			message, _ := recovered.(string)
			if got, want := strings.Contains(message, substr), true; got != want {
				t.Fatalf("recovered panic = %q, want to contain %q", message, substr)
			}
		})
	}
}

func TestCLI_Run(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")
	testCases := []struct {
		name   string
		runner cli.Runner
		want   cli.ExitCode
	}{
		{
			name:   "Success",
			runner: spectest.NoOpRunner(),
			want:   cli.ExitSuccess,
		},
		{
			name:   "RunnerError",
			runner: spectest.Err(testErr),
			want:   cli.ExitError,
		},
		{
			name:   "RecoveredPanic",
			runner: spectest.PanicRunner("kaboom"),
			want:   cli.ExitPanic,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := cli.FromBytes([]byte("id: root\nuse: root\n"), cli.BindRunner("root", tc.runner))
			var stderr strings.Builder
			sut.CobraCommand().SetOut(&stderr)
			sut.CobraCommand().SetErr(&stderr)
			sut.CobraCommand().SetArgs(nil)
			ctx := context.Background()

			// Act
			code := sut.Run(ctx)

			// Assert
			if got, want := code, tc.want; !cmp.Equal(got, want) {
				t.Errorf("sut.Run(ctx) = %d, want %d", got, want)
			}
		})
	}
}

// childNames returns the names of the direct subcommands of the CLI's command,
// excluding cobra's built-in commands.
func childNames(sut *cli.CLI) []string {
	var names []string
	for _, c := range sut.CobraCommand().Commands() {
		if c.Name() == "completion" || c.Name() == "help" {
			continue
		}
		names = append(names, c.Name())
	}
	return names
}

// recoverPanic runs fn and returns the value it panics with, or nil if it does
// not panic.
func recoverPanic(fn func()) (recovered any) {
	defer func() { recovered = recover() }()
	fn()
	return nil
}
