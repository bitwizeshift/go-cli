package arg_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
)

func TestCommandLine_Flags(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	cl.Add(
		arg.Flag("verbose", new(bool), arg.Shorthand("v")),
		arg.Flag("name", new(string)),
	)
	want := []*arg.FlagArg{
		newFlag("name"),
		arg.Flag("verbose", new(bool), arg.Shorthand("v")),
	}

	// Act
	flags := cl.Flags()

	// Assert
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("CommandLine.Flags() = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestCommandLine_Add(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var (
		verbose bool
		name    string
		rest    []string
	)
	cl.Add(
		arg.Flag("verbose", &verbose, arg.Shorthand("v")),
		arg.Positional("name", 0, &name),
		arg.Unmatched(&rest),
	)

	// Act
	argtest.Parse(t, cl, "--verbose", "alpha", "beta")

	// Assert
	if got, want := verbose, true; !cmp.Equal(got, want) {
		t.Errorf("flag verbose = %t, want %t", got, want)
	}
	if got, want := name, "alpha"; !cmp.Equal(got, want) {
		t.Errorf("positional name = %q, want %q", got, want)
	}
	if got, want := rest, []string{"beta"}; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("unmatched arguments = %v, want %v", got, want)
	}
}

func TestCommandLine_AddFlagSet(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	fs := pflag.NewFlagSet("external", pflag.ContinueOnError)
	fs.String("alpha", "", "an external flag")
	fs.BoolP("beta", "b", false, "another external flag")

	// Act
	cl.AddFlagSet(fs)
	longFlags := argtest.LongFlags(cl)

	// Assert
	if got, want := longFlags, []string{"alpha", "beta"}; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("CommandLine.AddFlagSet(...) = %v, want %v", got, want)
	}
}

// groupInfo captures a [arg.FlagGroup] as its name and the long names of its
// flags, so that groups may be compared as plain data.
type groupInfo struct {
	Name  string
	Flags []string
}

// groupInfosOf reduces groups to their comparable form.
func groupInfosOf(groups []*arg.FlagGroup) []groupInfo {
	var result []groupInfo
	for _, g := range groups {
		info := groupInfo{Name: g.Name}
		for _, f := range g.Flags {
			info.Flags = append(info.Flags, f.Name())
		}
		result = append(result, info)
	}
	return result
}

func TestCommandLine_Groups(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		flags []groupFlag
		want  []groupInfo
	}{
		{
			name:  "EmptyRegistryHasNoGroups",
			flags: nil,
			want:  nil,
		},
		{
			name: "UngroupedFlagsFallInGeneralGroupSortedByName",
			flags: []groupFlag{
				{name: "b"},
				{name: "a"},
			},
			want: []groupInfo{
				{
					Name:  "General Flags",
					Flags: []string{"a", "b"},
				},
			},
		},
		{
			name: "NamedGroupsSortedAlphabetically",
			flags: []groupFlag{
				{name: "a", group: "Zeta"},
				{name: "b", group: "Alpha"},
			},
			want: []groupInfo{
				{
					Name:  "Alpha",
					Flags: []string{"b"},
				},
				{
					Name:  "Zeta",
					Flags: []string{"a"},
				},
			},
		},
		{
			name: "GeneralGroupSortsLastWhenSeenFirst",
			flags: []groupFlag{
				{name: "a"},
				{name: "z", group: "Named"},
			},
			want: []groupInfo{
				{
					Name:  "Named",
					Flags: []string{"z"},
				},
				{
					Name:  "General Flags",
					Flags: []string{"a"},
				},
			},
		},
		{
			name: "GeneralGroupSortsLastWhenSeenLast",
			flags: []groupFlag{
				{name: "a", group: "Named"},
				{name: "z"},
			},
			want: []groupInfo{
				{
					Name:  "Named",
					Flags: []string{"a"},
				},
				{
					Name:  "General Flags",
					Flags: []string{"z"},
				},
			},
		},
		{
			name: "FlagsInSameGroupSortedByName",
			flags: []groupFlag{
				{name: "b", group: "Shared"},
				{name: "a", group: "Shared"},
			},
			want: []groupInfo{
				{
					Name:  "Shared",
					Flags: []string{"a", "b"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			for _, f := range tc.flags {
				added := addFlag(cl, f.name, new(bool), arg.Usage(f.usage))
				arg.Group(f.group, added)
			}

			// Act
			groups := groupInfosOf(cl.Groups())

			// Assert
			if got, want := groups, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("CommandLine.Groups() = %v, want %v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}
