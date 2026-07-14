package flag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/flag/flagtest"
)

// flagName reduces a [pflag.Flag] to its long name so that [flag.Group] values
// may be compared without depending on unexported flag state.
var flagName = cmp.Transformer("flagName", func(f *pflag.Flag) string {
	return f.Name
})

// groupFlag is a declarative specification of a flag to register, along with
// the group it belongs to. An empty group leaves the flag ungrouped.
type groupFlag struct {
	name  string
	usage string
	group string
}

func TestGroups(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		flags []groupFlag
		want  []*flag.Group
	}{
		{
			name:  "EmptyFlagSetHasNoGroups",
			flags: nil,
			want:  nil,
		},
		{
			name: "UngroupedFlagsFallInGeneralGroupSortedByName",
			flags: []groupFlag{
				{name: "b"},
				{name: "a"},
			},
			want: []*flag.Group{
				{
					Name:  "General Flags",
					Flags: []*pflag.Flag{{Name: "a"}, {Name: "b"}},
				},
			},
		},
		{
			name: "NamedGroupsSortedAlphabetically",
			flags: []groupFlag{
				{name: "a", group: "Zeta"},
				{name: "b", group: "Alpha"},
			},
			want: []*flag.Group{
				{
					Name:  "Alpha",
					Flags: []*pflag.Flag{{Name: "b"}},
				},
				{
					Name:  "Zeta",
					Flags: []*pflag.Flag{{Name: "a"}},
				},
			},
		},
		{
			name: "GeneralGroupSortsLastWhenSeenFirst",
			flags: []groupFlag{
				{name: "a"},
				{name: "z", group: "Named"},
			},
			want: []*flag.Group{
				{
					Name:  "Named",
					Flags: []*pflag.Flag{{Name: "z"}},
				},
				{
					Name:  "General Flags",
					Flags: []*pflag.Flag{{Name: "a"}},
				},
			},
		},
		{
			name: "GeneralGroupSortsLastWhenSeenLast",
			flags: []groupFlag{
				{name: "a", group: "Named"},
				{name: "z"},
			},
			want: []*flag.Group{
				{
					Name:  "Named",
					Flags: []*pflag.Flag{{Name: "a"}},
				},
				{
					Name:  "General Flags",
					Flags: []*pflag.Flag{{Name: "z"}},
				},
			},
		},
		{
			name: "FlagsInSameGroupSortedByName",
			flags: []groupFlag{
				{name: "b", group: "Shared"},
				{name: "a", group: "Shared"},
			},
			want: []*flag.Group{
				{
					Name:  "Shared",
					Flags: []*pflag.Flag{{Name: "a"}, {Name: "b"}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			for _, f := range tc.flags {
				fs.Bool(f.name, false, f.usage)
			}
			for _, f := range tc.flags {
				flag.AddToGroup(f.group, fs.Lookup(f.name))
			}

			// Act
			groups := flag.Groups(fs)

			// Assert
			if got, want := groups, tc.want; !cmp.Equal(got, want, flagName, cmpopts.EquateEmpty()) {
				t.Errorf("Groups(...) mismatch (-want +got):\n%s", cmp.Diff(want, got, flagName, cmpopts.EquateEmpty()))
			}
		})
	}
}

func TestAddToGroup(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		flags []groupFlag
		want  []*flagtest.Flag
	}{
		{
			name: "GroupedFlagReportsItsGroup",
			flags: []groupFlag{
				{name: "region", group: "Location Flags"},
			},
			want: []*flagtest.Flag{
				{Long: "region", Type: "bool", Group: "Location Flags"},
			},
		},
		{
			name: "UngroupedFlagHasNoGroup",
			flags: []groupFlag{
				{name: "verbose"},
			},
			want: []*flagtest.Flag{
				{Long: "verbose", Type: "bool"},
			},
		},
		{
			name: "SharedGroupAppliesToEveryFlag",
			flags: []groupFlag{
				{name: "host", group: "Connection Flags"},
				{name: "port", group: "Connection Flags"},
			},
			want: []*flagtest.Flag{
				{Long: "host", Type: "bool", Group: "Connection Flags"},
				{Long: "port", Type: "bool", Group: "Connection Flags"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			registry := flagtest.NewRegistry()
			for _, f := range tc.flags {
				fg := flag.Add(registry, f.name, new(bool), flag.Usage(f.usage))

				// Act
				flag.AddToGroup(f.group, fg)
			}

			// Assert
			names := flagtest.AllFlags(registry)
			if got, want := names, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("AddToGroup(...) mismatch (-want +got):\n%s", cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}

func TestGroup_Hidden(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		hidden []bool
		want   bool
	}{
		{
			name:   "NoFlagsIsHidden",
			hidden: nil,
			want:   true,
		},
		{
			name:   "SingleHiddenFlagIsHidden",
			hidden: []bool{true},
			want:   true,
		},
		{
			name:   "SingleVisibleFlagIsNotHidden",
			hidden: []bool{false},
			want:   false,
		},
		{
			name:   "AllHiddenFlagsIsHidden",
			hidden: []bool{true, true},
			want:   true,
		},
		{
			name:   "AnyVisibleFlagIsNotHidden",
			hidden: []bool{true, false},
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			group := &flag.Group{}
			for _, h := range tc.hidden {
				group.Flags = append(group.Flags, &pflag.Flag{Hidden: h})
			}

			// Act
			hidden := group.Hidden()

			// Assert
			if got, want := hidden, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Group.Hidden() = %v, want %v", got, want)
			}
		})
	}
}
