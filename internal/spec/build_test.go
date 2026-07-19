package spec_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/bitwizeshift/go-cli/internal/spec/spectest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
)

// flaggedRunner is a [spec.Runner] that registers flags, used to verify that
// [spec.Build] registers a bound runner's flags and their completions.
type flaggedRunner struct {
	verbose bool
	format  string
}

func (fr *flaggedRunner) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(
		arg.Flag("verbose", &fr.verbose),
		arg.Flag("format", &fr.format, arg.CompleteFrom("json", "yaml")),
	)
}

func (fr *flaggedRunner) Run(context.Context) error {
	return nil
}

var (
	_ spec.Runner   = (*flaggedRunner)(nil)
	_ arg.Registrar = (*flaggedRunner)(nil)
)

// positionalRunner is a [spec.Runner] that registers positional arguments, used
// to verify that [spec.Build] reconciles their completions onto the command.
type positionalRunner struct {
	input  string
	format string
}

func (pr *positionalRunner) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(
		arg.Positional("input", 0, &pr.input, arg.CompleteFiles()),
		arg.Positional("format", 1, &pr.format, arg.CompleteFrom("json", "yaml")),
	)
}

func (pr *positionalRunner) Run(context.Context) error {
	return nil
}

var (
	_ spec.Runner   = (*positionalRunner)(nil)
	_ arg.Registrar = (*positionalRunner)(nil)
)

// offered is what a command offers for a word being completed: the candidates it
// returns, and the cobra directive telling a shell how to treat them.
type offered struct {
	Candidates []string
	Directive  cobra.ShellCompDirective
}

const rootWithChild = `
id: root
use: root
commands:
  default:
    - id: child
      use: child
`

func toBuilders(runners map[string]spec.Runner) map[string]spec.Builder {
	result := map[string]spec.Builder{}
	for key, runner := range runners {
		result[key] = spectest.PassThroughBuilder(runner)
	}
	return result
}

func TestBuild(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		runners map[string]spec.Runner
		want    string
		wantErr error
	}{
		{
			name:    "builds command tree",
			input:   "id: root\nuse: root <command>\n",
			runners: map[string]spec.Runner{"root": spectest.NoOpRunner()},
			want:    "root <command>",
		},
		{
			name:    "reports unbound runner",
			input:   "id: root\nuse: root\n",
			runners: map[string]spec.Runner{"ghost": spectest.NoOpRunner()},
			wantErr: spec.ErrUnboundRunner,
		},
		{
			name:    "reports malformed specification",
			input:   "id: root\ncommands: not-a-mapping\n",
			runners: map[string]spec.Runner{},
			wantErr: cmpopts.AnyError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			reader := strings.NewReader(tc.input)

			// Act
			cmd, err := spec.Build(reader, spec.Options{Builders: toBuilders(tc.runners)})

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("spec.Build(...) = _, %v, want %v", got, want)
			}
			use := commandUse(cmd)
			if got, want := use, tc.want; got != want {
				t.Errorf("spec.Build(...).Use = %q, want %q", got, want)
			}
		})
	}
}
func commandUse(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	return cmd.Use
}

func TestBuild_Colour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		colour   spec.ColourMode
		wantANSI bool
	}{
		{
			name:     "Enabled",
			colour:   spec.ColourEnabled,
			wantANSI: true,
		},
		{
			name:     "Disabled",
			colour:   spec.ColourDisabled,
			wantANSI: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var out bytes.Buffer
			sut := build(t, "id: root\nuse: root\n", spec.Options{
				Builders: toBuilders(map[string]spec.Runner{
					"root": spectest.NoOpRunner(),
				}),
				Colour: tc.colour,
				Stdout: &out,
				Stderr: &out,
			})

			// Act
			err := sut.Help()

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Help() = %v, want nil", err)
			}
			if got, want := strings.Contains(out.String(), "\x1b"), tc.wantANSI; !cmp.Equal(got, want) {
				t.Errorf("Help() coloured = %t, want %t (output %q)", got, want, out.String())
			}
		})
	}
}

func TestBuild_DefaultGroup_LeavesChildUngrouped(t *testing.T) {
	t.Parallel()

	// Arrange
	reader := strings.NewReader(rootWithChild)

	// Act
	sut, err := spec.Build(reader, spec.Options{
		Builders: toBuilders(map[string]spec.Runner{
			"root": spectest.NoOpRunner(),
		}),
	})

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("spec.Build(...) = _, %v, want nil", err)
	}
	child := subcommand(t, sut, "child")
	if got, want := child.GroupID, ""; got != want {
		t.Errorf("child.GroupID = %q, want %q", got, want)
	}
}

func TestBuild_NamedGroup_AssignsGroupID(t *testing.T) {
	t.Parallel()

	// Arrange
	const input = `
id: root
use: root
commands:
  Named Commands:
    - id: child
      use: child
`
	reader := strings.NewReader(input)

	// Act
	sut, err := spec.Build(reader, spec.Options{
		Builders: toBuilders(map[string]spec.Runner{
			"root": spectest.NoOpRunner(),
		}),
	})

	// Assert
	if err != nil {
		t.Fatalf("spec.Build(...) = _, %v, want nil", err)
	}
	child := subcommand(t, sut, "child")
	if got, want := child.GroupID, "Named-Commands"; got != want {
		t.Errorf("child.GroupID = %q, want %q", got, want)
	}
	if got, want := groupTitles(sut), []string{"Named Commands"}; !cmp.Equal(got, want) {
		t.Errorf("groupTitles(sut) = %v, want %v", got, want)
	}
}

func TestBuild_BoundRunner_RegistersFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	reader := strings.NewReader("id: root\nuse: root\n")

	// Act
	sut, err := spec.Build(reader, spec.Options{
		Builders: toBuilders(map[string]spec.Runner{
			"root": &flaggedRunner{},
		}),
	})
	cl := (*arg.CommandLine)(argdef.FromFlagSet(sut.Flags()))

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("spec.Build(...) = _, %v, want nil", err)
	}
	flags := argtest.LongFlags(cl)
	if got, want := flags, []string{"format", "verbose"}; !cmp.Equal(got, want) {
		t.Errorf("flags = %v, want %v", got, want)
	}
}

func TestBuild_BoundRunner_RegistersCompletions(t *testing.T) {
	t.Parallel()

	// Arrange
	reader := strings.NewReader("id: root\nuse: root\n")

	// Act
	sut, err := spec.Build(reader, spec.Options{
		Builders: toBuilders(map[string]spec.Runner{
			"root": &flaggedRunner{},
		}),
	})

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("spec.Build(...) got err %v, want %v", got, want)
	}
	_, registered := sut.GetFlagCompletionFunc("format")
	if got, want := registered, true; got != want {
		t.Errorf("GetFlagCompletionFunc(\"format\") registered = %t, want %t", got, want)
	}
}

func TestBuild_PositionalCompletions_CompletesByIndex(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		args       []string
		toComplete string
		want       offered
	}{
		{
			name:       "FirstPositionalDefersToShell",
			args:       nil,
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  cobra.ShellCompDirectiveDefault,
			},
		},
		{
			name:       "SecondPositionalOffersItsCandidates",
			args:       []string{"in.psx"},
			toComplete: "j",
			want: offered{
				Candidates: []string{"json"},
				Directive:  cobra.ShellCompDirectiveNoFileComp,
			},
		},
		{
			name:       "UnregisteredIndexDefersToShell",
			args:       []string{"in.psx", "json"},
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  cobra.ShellCompDirectiveDefault,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := buildRoot(t, &positionalRunner{})

			// Act
			candidates, directive := sut.ValidArgsFunction(sut, tc.args, tc.toComplete)

			// Assert
			offer := offered{
				Candidates: candidates,
				Directive:  directive,
			}
			if got, want := offer, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("ValidArgsFunction(...) = %+v, want %+v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}

func TestBuild_RunnerWithoutPositionals_RegistersNoArgsFunction(t *testing.T) {
	t.Parallel()

	// Arrange
	runner := &flaggedRunner{}

	// Act
	sut := buildRoot(t, runner)

	// Assert
	if got, want := sut.ValidArgsFunction == nil, true; !cmp.Equal(got, want) {
		t.Errorf("ValidArgsFunction = nil %t, want %t", got, want)
	}
}

func TestBuild_UnboundCommand_RegistersNoArgsFunction(t *testing.T) {
	t.Parallel()

	// Arrange
	reader := strings.NewReader(rootWithChild)
	root, err := spec.Build(reader, spec.Options{
		Builders: toBuilders(map[string]spec.Runner{"root": spectest.NoOpRunner()}),
	})
	if err != nil {
		t.Fatalf("spec.Build(...) = %v, want %v", err, error(nil))
	}

	// Act
	sut := root.Commands()[0]

	// Assert
	if got, want := sut.ValidArgsFunction == nil, true; !cmp.Equal(got, want) {
		t.Errorf("ValidArgsFunction = nil %t, want %t", got, want)
	}
}

// buildRoot builds a single-command tree binding runner to its root, failing the
// test if the specification cannot be built.
func buildRoot(t testing.TB, runner spec.Runner) *cobra.Command {
	t.Helper()

	reader := strings.NewReader("id: root\nuse: root\n")
	cmd, err := spec.Build(reader, spec.Options{
		Builders: toBuilders(map[string]spec.Runner{"root": runner}),
	})
	if err != nil {
		t.Fatalf("spec.Build(...) = %v, want %v", err, error(nil))
	}
	return cmd
}

func TestBuild_UnboundCommand_RunsHelp(t *testing.T) {
	t.Parallel()

	// Arrange
	const input = `
id: root
use: root
commands:
  default:
    - id: child
      use: child
      summary: a child command
`
	var out bytes.Buffer
	sut := build(t, input, spec.Options{
		Builders: toBuilders(map[string]spec.Runner{
			"root": spectest.NoOpRunner(),
		}),
		Stdout: &out,
		Stderr: &out,
	})
	sut.SetArgs([]string{"child"})

	// Act
	err := sut.Execute()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Execute(child) = %v, want nil", err)
	}
	if got, want := strings.Contains(out.String(), "root child"), true; !cmp.Equal(got, want) {
		t.Errorf("Execute(child) output = %q, want help containing %q", out.String(), "root child")
	}
}

// build decodes input into a command tree, failing the test on error.
func build(t testing.TB, input string, opts spec.Options) *cobra.Command {
	t.Helper()
	cmd, err := spec.Build(strings.NewReader(input), opts)
	if err != nil {
		t.Fatalf("spec.Build(...) = _, %v, want nil", err)
	}
	return cmd
}

// subcommand returns the child of parent named name, failing the test if absent.
func subcommand(t testing.TB, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	t.Fatalf("subcommand(%q): not found", name)
	return nil
}

// groupTitles returns the titles of the groups registered on cmd, in order.
func groupTitles(cmd *cobra.Command) []string {
	titles := make([]string, 0, len(cmd.Groups()))
	for _, g := range cmd.Groups() {
		titles = append(titles, g.Title)
	}
	return titles
}
