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

// hiddenFlags registers a bool flag for each entry in hidden, marking the flag
// hidden when its entry is set.
func hiddenFlags(t testing.TB, hidden []bool) []*arg.FlagArg {
	t.Helper()

	cl := argtest.NewCommandLine()
	var flags []*arg.FlagArg
	for i, h := range hidden {
		var options []arg.Option
		if h {
			options = append(options, arg.Hidden())
		}
		flags = append(flags, addFlag(cl, fmt.Sprintf("flag%d", i), new(bool), options...))
	}
	return flags
}

func TestGroup(t *testing.T) {
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
				fg := addFlag(cl, f.name, new(bool), arg.Usage(f.usage))

				// Act
				arg.Group(f.group, fg)
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
