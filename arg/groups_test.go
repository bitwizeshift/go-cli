package arg_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
)

// groupFlag is a declarative specification of a flag to register, along with
// the group it belongs to. An empty group leaves the flag ungrouped.
type groupFlag struct {
	name  string
	usage string
	group string
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

// hiddenFlags registers a bool flag for each entry in hidden, marking the flag
// hidden when its entry is set.
func hiddenFlags(t testing.TB, hidden []bool) []*arg.Flag {
	t.Helper()

	cl := argtest.NewCommandLine()
	var flags []*arg.Flag
	for i, h := range hidden {
		var options []arg.Option
		if h {
			options = append(options, arg.Hidden())
		}
		flags = append(flags, arg.AddFlag(cl, fmt.Sprintf("flag%d", i), new(bool), options...))
	}
	return flags
}

func TestGroups(t *testing.T) {
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
				added := arg.AddFlag(cl, f.name, new(bool), arg.Usage(f.usage))
				arg.AddToGroup(f.group, added)
			}

			// Act
			groups := groupInfosOf(arg.Groups(cl))

			// Assert
			if got, want := groups, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Groups(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}

func TestAddToGroup(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		flags []groupFlag
		want  []*argtest.Flag
	}{
		{
			name: "GroupedFlagReportsItsGroup",
			flags: []groupFlag{
				{name: "region", group: "Location Flags"},
			},
			want: []*argtest.Flag{
				{Long: "region", Type: "bool", Group: "Location Flags"},
			},
		},
		{
			name: "UngroupedFlagHasNoGroup",
			flags: []groupFlag{
				{name: "verbose"},
			},
			want: []*argtest.Flag{
				{Long: "verbose", Type: "bool"},
			},
		},
		{
			name: "SharedGroupAppliesToEveryFlag",
			flags: []groupFlag{
				{name: "host", group: "Connection Flags"},
				{name: "port", group: "Connection Flags"},
			},
			want: []*argtest.Flag{
				{Long: "host", Type: "bool", Group: "Connection Flags"},
				{Long: "port", Type: "bool", Group: "Connection Flags"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			for _, f := range tc.flags {
				fg := arg.AddFlag(cl, f.name, new(bool), arg.Usage(f.usage))

				// Act
				arg.AddToGroup(f.group, fg)
			}

			// Assert
			names := argtest.AllFlags(cl)
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
			sut := &arg.FlagGroup{Flags: hiddenFlags(t, tc.hidden)}

			// Act
			hidden := sut.Hidden()

			// Assert
			if got, want := hidden, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Group.Hidden() = %v, want %v", got, want)
			}
		})
	}
}
