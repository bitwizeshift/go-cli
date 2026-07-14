package annotation_test

import (
	"context"
	"errors"
	"strings"
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

func TestIssueURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		annotations map[string]string
		want        string
	}{
		{
			name:        "IssueURLSetReturnsURL",
			annotations: map[string]string{annotation.AnnotationIssueURL: "https://example.test/issues"},
			want:        "https://example.test/issues",
		},
		{
			name:        "NilAnnotationsReturnsEmpty",
			annotations: nil,
			want:        "",
		},
		{
			name:        "UnrelatedAnnotationReturnsEmpty",
			annotations: map[string]string{"other": "value"},
			want:        "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cmd := &cobra.Command{Use: "test", Annotations: tc.annotations}

			// Act
			url := annotation.IssueURL(cmd)

			// Assert
			if got, want := url, tc.want; !cmp.Equal(got, want) {
				t.Errorf("IssueURL(...) = %q, want %q", got, want)
			}
		})
	}
}

func TestAddIssueURL(t *testing.T) {
	t.Parallel()

	// Arrange
	const url = "https://example.test/issues"
	root := &cobra.Command{Use: "root"}
	left := &cobra.Command{Use: "left"}
	right := &cobra.Command{Use: "right"}
	shared := &cobra.Command{Use: "shared"}
	root.AddCommand(left, right)
	left.AddCommand(shared)
	right.AddCommand(shared)

	// Act
	annotation.AddIssueURL(root, url)

	// Assert
	urls := map[string]string{
		"root":   annotation.IssueURL(root),
		"left":   annotation.IssueURL(left),
		"right":  annotation.IssueURL(right),
		"shared": annotation.IssueURL(shared),
	}
	want := map[string]string{"root": url, "left": url, "right": url, "shared": url}
	if got, want := urls, want; !cmp.Equal(got, want) {
		t.Errorf("AddIssueURL(...) = %v, want %v", got, want)
	}
}

func TestAddENVFallback(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		envs []string
		want []string
	}{
		{
			name: "SingleEnvRecorded",
			envs: []string{"FOO"},
			want: []string{"FOO"},
		},
		{
			name: "MultipleEnvsAccumulate",
			envs: []string{"FOO", "BAR"},
			want: []string{"FOO", "BAR"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			target := newStringFlag("flag")

			// Act
			for _, env := range tc.envs {
				annotation.AddENVFallback(target, env)
			}

			// Assert
			if got, want := target.Annotations[annotation.AnnotationENVFallback], tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("AddENVFallback(...) = %v, want %v", got, want)
			}
		})
	}
}

func TestAddFuncFallback(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		count int
		want  int
	}{
		{
			name:  "SingleFuncRecorded",
			count: 1,
			want:  1,
		},
		{
			name:  "MultipleFuncsAccumulate",
			count: 2,
			want:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			target := newStringFlag("flag")

			// Act
			for range tc.count {
				annotation.AddFuncFallback(target, func(context.Context) (string, error) { return "", nil })
			}

			// Assert
			if got, want := len(target.Annotations[annotation.AnnotationFuncFallback]), tc.want; got != want {
				t.Errorf("AddFuncFallback(...) recorded %d ids, want %d", got, want)
			}
		})
	}
}

// funcResult declares the return values of a fallback function registered
// during a [annotation.SetFlagFallbacks] test case.
type funcResult struct {
	value string
	err   error
}

func TestSetFlagFallbacks(t *testing.T) {
	errCompute := errors.New("compute failed")

	testCases := []struct {
		name      string
		envs      []string
		setEnv    map[string]string
		funcs     []funcResult
		wantValue string
		wantErr   error
	}{
		{
			name:      "EnvPresentSetsValue",
			envs:      []string{"FLAG_ENV"},
			setEnv:    map[string]string{"FLAG_ENV": "from-env"},
			wantValue: "from-env",
		},
		{
			name:      "EnvTakesPrecedenceOverFunc",
			envs:      []string{"FLAG_ENV"},
			setEnv:    map[string]string{"FLAG_ENV": "from-env"},
			funcs:     []funcResult{{value: "from-func"}},
			wantValue: "from-env",
		},
		{
			name:      "EnvAbsentFallsToFunc",
			envs:      []string{"FLAG_ENV"},
			funcs:     []funcResult{{value: "from-func"}},
			wantValue: "from-func",
		},
		{
			name:      "EmptyFuncFallsThroughToNext",
			funcs:     []funcResult{{value: ""}, {value: "second"}},
			wantValue: "second",
		},
		{
			name:      "AllFuncsEmptyLeavesDefault",
			funcs:     []funcResult{{value: ""}},
			wantValue: "",
		},
		{
			name:      "FuncErrorWrapsComputeSentinel",
			funcs:     []funcResult{{err: errCompute}},
			wantValue: "",
			wantErr:   annotation.ErrComputingFuncFlag,
		},
		{
			name:      "NoFallbacksLeavesDefault",
			wantValue: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			fs.String("flag", "", "")
			target := fs.Lookup("flag")
			for _, env := range tc.envs {
				annotation.AddENVFallback(target, env)
			}
			for _, fn := range tc.funcs {
				annotation.AddFuncFallback(target, func(context.Context) (string, error) {
					return fn.value, fn.err
				})
			}
			for key, value := range tc.setEnv {
				t.Setenv(key, value)
			}
			ctx := context.Background()

			// Act
			err := annotation.SetFlagFallbacks(ctx, fs)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("SetFlagFallbacks(...) = %v, want %v", got, want)
			}
			if got, want := target.Value.String(), tc.wantValue; got != want {
				t.Errorf("flag value = %q, want %q", got, want)
			}
		})
	}
}

func TestSetFlagFallbacks_SetError(t *testing.T) {
	testCases := []struct {
		name    string
		envs    []string
		setEnv  map[string]string
		funcs   []funcResult
		wantErr error
	}{
		{
			name:    "EnvValueRejectedWrapsEnvSentinel",
			envs:    []string{"FLAG_ENV"},
			setEnv:  map[string]string{"FLAG_ENV": "not-an-int"},
			wantErr: annotation.ErrSettingEnvFlag,
		},
		{
			name:    "FuncValueRejectedWrapsFuncSentinel",
			funcs:   []funcResult{{value: "not-an-int"}},
			wantErr: annotation.ErrSettingFuncFlag,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			fs.Int("flag", 0, "")
			target := fs.Lookup("flag")
			for _, env := range tc.envs {
				annotation.AddENVFallback(target, env)
			}
			for _, fn := range tc.funcs {
				annotation.AddFuncFallback(target, func(context.Context) (string, error) {
					return fn.value, fn.err
				})
			}
			for key, value := range tc.setEnv {
				t.Setenv(key, value)
			}
			ctx := context.Background()

			// Act
			err := annotation.SetFlagFallbacks(ctx, fs)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("SetFlagFallbacks(...) = %v, want %v", got, want)
			}
			if got, want := target.Value.String(), "0"; got != want {
				t.Errorf("flag value = %q, want %q", got, want)
			}
		})
	}
}

func TestSetFlagFallbacks_SkipsChangedFlag(t *testing.T) {
	// Arrange
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("flag", "", "")
	target := fs.Lookup("flag")
	annotation.AddENVFallback(target, "FLAG_ENV")
	annotation.AddFuncFallback(target, func(context.Context) (string, error) { return "from-func", nil })
	t.Setenv("FLAG_ENV", "from-env")
	if err := fs.Set("flag", "from-user"); err != nil {
		t.Fatalf("fs.Set(...) = %v, want nil", err)
	}
	ctx := context.Background()

	// Act
	err := annotation.SetFlagFallbacks(ctx, fs)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("SetFlagFallbacks(...) = %v, want %v", got, want)
	}
	if got, want := target.Value.String(), "from-user"; got != want {
		t.Errorf("flag value = %q, want %q", got, want)
	}
}

func TestSetFlagFallbacks_UnregisteredFuncFallback_SkipsFlag(t *testing.T) {
	// Arrange
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("flag", "", "")
	target := fs.Lookup("flag")
	target.Annotations = map[string][]string{
		annotation.AnnotationFuncFallback: {"not-a-registered-id"},
	}
	ctx := context.Background()

	// Act
	err := annotation.SetFlagFallbacks(ctx, fs)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("SetFlagFallbacks(...) = %v, want %v", got, want)
	}
	if got, want := target.Value.String(), ""; got != want {
		t.Errorf("flag value = %q, want %q", got, want)
	}
}

func TestSetFlagFallbacks_JoinsErrors(t *testing.T) {
	// Arrange
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Int("first", 0, "")
	fs.Int("second", 0, "")
	annotation.AddFuncFallback(fs.Lookup("first"), func(context.Context) (string, error) {
		return "", errors.New("first failed")
	})
	annotation.AddFuncFallback(fs.Lookup("second"), func(context.Context) (string, error) {
		return "not-an-int", nil
	})
	ctx := context.Background()

	// Act
	err := annotation.SetFlagFallbacks(ctx, fs)

	// Assert
	if got, want := errors.Is(err, annotation.ErrComputingFuncFlag), true; got != want {
		t.Errorf("errors.Is(err, ErrComputingFuncFlag) = %t, want %t", got, want)
	}
	if got, want := errors.Is(err, annotation.ErrSettingFuncFlag), true; got != want {
		t.Errorf("errors.Is(err, ErrSettingFuncFlag) = %t, want %t", got, want)
	}
}

func TestSetFlagFallbacks_EnvErrorNamesKey(t *testing.T) {
	// Arrange
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Int("flag", 0, "")
	annotation.AddENVFallback(fs.Lookup("flag"), "FLAG_ENV")
	t.Setenv("FLAG_ENV", "not-an-int")
	ctx := context.Background()

	// Act
	err := annotation.SetFlagFallbacks(ctx, fs)

	// Assert
	if got, want := err, annotation.ErrSettingEnvFlag; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("SetFlagFallbacks(...) = %v, want %v", got, want)
	}
	if got, want := strings.Contains(err.Error(), "FLAG_ENV"), true; got != want {
		t.Errorf("error names env key = %t, want %t", got, want)
	}
}

// newStringFlag registers a string flag named name on a fresh flag set and
// returns it, for exercising annotation helpers that operate on a single flag.
func newStringFlag(name string) *pflag.Flag {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String(name, "", "")
	return fs.Lookup(name)
}
