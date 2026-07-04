package annotation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/internal/annotation"
)

// newFlagSet returns a flag set with three boolean flags "a", "b", and "c".
func newFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("a", false, "")
	fs.Bool("b", false, "")
	fs.Bool("c", false, "")
	return fs
}

// lookupAll returns the flags in fs identified by names.
func lookupAll(fs *pflag.FlagSet, names []string) []*pflag.Flag {
	flags := make([]*pflag.Flag, 0, len(names))
	for _, name := range names {
		flags = append(flags, fs.Lookup(name))
	}
	return flags
}

func TestIsRequired(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		requiredTogether []string
		required         []string
		flag             string
		want             bool
	}{
		{
			name:     "MarkedFlagIsRequired",
			required: []string{"a"},
			flag:     "a",
			want:     true,
		},
		{
			name:     "UnmarkedFlagIsNotRequired",
			required: []string{"a"},
			flag:     "b",
			want:     false,
		},
		{
			name:             "MarkedFlagAlreadyInGroupIsRequired",
			requiredTogether: []string{"a", "b"},
			required:         []string{"a"},
			flag:             "a",
			want:             true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := newFlagSet()
			annotation.MarkRequiredTogether(lookupAll(fs, tc.requiredTogether)...)
			annotation.MarkRequired(lookupAll(fs, tc.required)...)
			target := fs.Lookup(tc.flag)

			// Act
			required := annotation.IsRequired(target)

			// Assert
			if got, want := required, tc.want; !cmp.Equal(got, want) {
				t.Errorf("IsRequired(...) = %v, want %v", got, want)
			}
		})
	}
}

func TestRequiredTogether(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		groups [][]string
		flag   string
		want   []string
	}{
		{
			name:   "MemberReportsFullGroupIncludingSelf",
			groups: [][]string{{"a", "b"}},
			flag:   "a",
			want:   []string{"a", "b"},
		},
		{
			name:   "NonMemberReportsEmpty",
			groups: [][]string{{"a", "b"}},
			flag:   "c",
			want:   nil,
		},
		{
			name:   "MultipleGroupsAreUnioned",
			groups: [][]string{{"a", "b"}, {"a", "c"}},
			flag:   "a",
			want:   []string{"a", "b", "c"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := newFlagSet()
			for _, group := range tc.groups {
				annotation.MarkRequiredTogether(lookupAll(fs, group)...)
			}
			target := fs.Lookup(tc.flag)

			// Act
			requiredWith := annotation.RequiredTogether(target)

			// Assert
			if got, want := requiredWith, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("RequiredTogether(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
		})
	}
}

func TestMutuallyExclusive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		group []string
		flag  string
		want  []string
	}{
		{
			name:  "MemberReportsFullGroupIncludingSelf",
			group: []string{"a", "c"},
			flag:  "a",
			want:  []string{"a", "c"},
		},
		{
			name:  "NonMemberReportsEmpty",
			group: []string{"a", "c"},
			flag:  "b",
			want:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := newFlagSet()
			annotation.MarkMutuallyExclusive(lookupAll(fs, tc.group)...)
			target := fs.Lookup(tc.flag)

			// Act
			exclusiveWith := annotation.MutuallyExclusive(target)

			// Assert
			if got, want := exclusiveWith, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("MutuallyExclusive(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
		})
	}
}

func TestOneRequired(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		group []string
		flag  string
		want  []string
	}{
		{
			name:  "MemberReportsFullGroupIncludingSelf",
			group: []string{"a", "b", "c"},
			flag:  "b",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "NonMemberReportsEmpty",
			group: []string{"a", "b"},
			flag:  "c",
			want:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := newFlagSet()
			annotation.MarkOneRequired(lookupAll(fs, tc.group)...)
			target := fs.Lookup(tc.flag)

			// Act
			oneRequiredWith := annotation.OneRequired(target)

			// Assert
			if got, want := oneRequiredWith, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("OneRequired(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
		})
	}
}

func TestConfigureFlags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		required          []string
		requiredTogether  []string
		mutuallyExclusive []string
		oneRequired       []string
		args              []string
		wantErr           error
	}{
		{
			name:    "NoConstraintsSucceeds",
			args:    nil,
			wantErr: nil,
		},
		{
			name:     "RequiredMissingFails",
			required: []string{"a"},
			args:     nil,
			wantErr:  cmpopts.AnyError,
		},
		{
			name:     "RequiredSatisfiedSucceeds",
			required: []string{"a"},
			args:     []string{"--a"},
			wantErr:  nil,
		},
		{
			name:             "RequiredTogetherPartialFails",
			requiredTogether: []string{"a", "b"},
			args:             []string{"--a"},
			wantErr:          cmpopts.AnyError,
		},
		{
			name:             "RequiredTogetherAllSucceeds",
			requiredTogether: []string{"a", "b"},
			args:             []string{"--a", "--b"},
			wantErr:          nil,
		},
		{
			name:             "RequiredTogetherNoneSucceeds",
			requiredTogether: []string{"a", "b"},
			args:             nil,
			wantErr:          nil,
		},
		{
			name:              "MutuallyExclusiveConflictFails",
			mutuallyExclusive: []string{"a", "b"},
			args:              []string{"--a", "--b"},
			wantErr:           cmpopts.AnyError,
		},
		{
			name:              "MutuallyExclusiveSingleSucceeds",
			mutuallyExclusive: []string{"a", "b"},
			args:              []string{"--a"},
			wantErr:           nil,
		},
		{
			name:        "OneRequiredNoneFails",
			oneRequired: []string{"a", "b"},
			args:        nil,
			wantErr:     cmpopts.AnyError,
		},
		{
			name:        "OneRequiredOneSucceeds",
			oneRequired: []string{"a", "b"},
			args:        []string{"--a"},
			wantErr:     nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cmd := newCommand()
			annotation.MarkRequired(lookupAll(cmd.Flags(), tc.required)...)
			annotation.MarkRequiredTogether(lookupAll(cmd.Flags(), tc.requiredTogether)...)
			annotation.MarkMutuallyExclusive(lookupAll(cmd.Flags(), tc.mutuallyExclusive)...)
			annotation.MarkOneRequired(lookupAll(cmd.Flags(), tc.oneRequired)...)
			annotation.ConfigureFlags(cmd)
			cmd.SetArgs(tc.args)

			// Act
			err := cmd.Execute()

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Execute() = %v, want %v", got, want)
			}
		})
	}
}

// groupAssignment records a single [annotation.AddToGroup] call to perform
// during arrangement.
type groupAssignment struct {
	group string
	flags []string
}

func TestGroup(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		assignments []groupAssignment
		flag        string
		want        string
	}{
		{
			name:        "UngroupedReturnsEmpty",
			assignments: nil,
			flag:        "a",
			want:        "",
		},
		{
			name:        "GroupedReturnsName",
			assignments: []groupAssignment{{group: "Files", flags: []string{"a"}}},
			flag:        "a",
			want:        "Files",
		},
		{
			name:        "MultipleFlagsShareGroup",
			assignments: []groupAssignment{{group: "Files", flags: []string{"a", "b"}}},
			flag:        "b",
			want:        "Files",
		},
		{
			name: "ReassignOverwritesPreviousGroup",
			assignments: []groupAssignment{
				{group: "Old", flags: []string{"a"}},
				{group: "New", flags: []string{"a"}},
			},
			flag: "a",
			want: "New",
		},
		{
			name:        "GroupNameWithSpacesPreserved",
			assignments: []groupAssignment{{group: "General Flags", flags: []string{"a"}}},
			flag:        "a",
			want:        "General Flags",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := newFlagSet()
			for _, a := range tc.assignments {
				annotation.AddToGroup(a.group, lookupAll(fs, a.flags)...)
			}
			target := fs.Lookup(tc.flag)

			// Act
			group := annotation.Group(target)

			// Assert
			if got, want := group, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Group(...) = %q, want %q", got, want)
			}
		})
	}
}

// newCommand returns a silent no-op command with three boolean flags "a", "b",
// and "c".
func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "test",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return nil
		},
	}
	cmd.Flags().Bool("a", false, "")
	cmd.Flags().Bool("b", false, "")
	cmd.Flags().Bool("c", false, "")
	return cmd
}
