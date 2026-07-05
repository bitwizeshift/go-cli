package spec_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/flag/flagtest"
	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/bitwizeshift/go-cli/internal/spec/spectest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
)

// flaggedRunner is a [spec.Runner] that registers a single flag, used to verify
// that [spec.Build] registers a bound runner's flags.
type flaggedRunner struct {
	verbose bool
}

func (fr *flaggedRunner) RegisterFlags(registry *flag.Registry) {
	flag.Add(registry, "verbose", &fr.verbose)
}

func (fr *flaggedRunner) Run(context.Context, ...string) error {
	return nil
}

var (
	_ spec.Runner    = (*flaggedRunner)(nil)
	_ flag.Registrar = (*flaggedRunner)(nil)
)

const rootWithChild = `
id: root
use: root
commands:
  default:
    - id: child
      use: child
`

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
			runners: map[string]spec.Runner{"root": spectest.OK()},
			want:    "root <command>",
		},
		{
			name:    "reports unbound runner",
			input:   "id: root\nuse: root\n",
			runners: map[string]spec.Runner{"ghost": spectest.OK()},
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
			cmd, err := spec.Build(reader, tc.runners)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("spec.Build(...) = _, %v, want %v", got, want)
			}
			use := ""
			if cmd != nil {
				use = cmd.Use
			}
			if got, want := use, tc.want; got != want {
				t.Errorf("spec.Build(...).Use = %q, want %q", got, want)
			}
		})
	}
}

func TestBuild_DefaultGroup_LeavesChildUngrouped(t *testing.T) {
	t.Parallel()

	// Arrange
	reader := strings.NewReader(rootWithChild)

	// Act
	sut, err := spec.Build(reader, map[string]spec.Runner{"root": spectest.OK()})

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
	sut, err := spec.Build(reader, map[string]spec.Runner{"root": spectest.OK()})

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
	sut, err := spec.Build(reader, map[string]spec.Runner{"root": &flaggedRunner{}})

	// Assert
	if err != nil {
		t.Fatalf("spec.Build(...) = _, %v, want nil", err)
	}
	if got, want := flagtest.LongFlags(sut.Flags()), []string{"verbose"}; !cmp.Equal(got, want) {
		t.Errorf("flags = %v, want %v", got, want)
	}
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
