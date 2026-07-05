package flag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/internal/annotation"
)

// completeFlag registers fs onto a cobra command, wires the flag completions,
// and invokes the completion function registered for the named flag. It fails
// the test if no completion was registered.
func completeFlag(t testing.TB, fs *pflag.FlagSet, name, toComplete string) ([]string, cobra.ShellCompDirective) {
	t.Helper()
	cmd := &cobra.Command{Use: "root"}
	cmd.Flags().AddFlagSet(fs)
	annotation.RegisterFlagCompletions(cmd)
	complete, ok := cmd.GetFlagCompletionFunc(name)
	if !ok {
		t.Fatalf("GetFlagCompletionFunc(%q) = _, false, want true", name)
	}
	return complete(cmd, nil, toComplete)
}

func TestCompletionOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		option        flag.Option
		toComplete    string
		want          []string
		wantDirective cobra.ShellCompDirective
	}{
		{
			name:          "CompleteFromMatchesPrefix",
			option:        flag.CompleteFrom("json", "yaml", "jsonl"),
			toComplete:    "js",
			want:          []string{"json", "jsonl"},
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:          "CompleteFromEmptyPrefixMatchesAll",
			option:        flag.CompleteFrom("json", "yaml"),
			toComplete:    "",
			want:          []string{"json", "yaml"},
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:          "CompleteFromNoMatchReturnsEmpty",
			option:        flag.CompleteFrom("json", "yaml"),
			toComplete:    "x",
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:          "CompleterFuncPassesCandidatesThrough",
			option:        flag.CompleterFunc(func(toComplete string) []string { return []string{toComplete + "-done"} }),
			toComplete:    "value",
			want:          []string{"value-done"},
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:          "CompleteFilesDefersToShell",
			option:        flag.CompleteFiles(),
			toComplete:    "",
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveDefault,
		},
		{
			name:          "CompleteFilesMatchingNormalizesExtensions",
			option:        flag.CompleteFilesMatching(".json", "yaml"),
			toComplete:    "",
			want:          []string{"json", "yaml"},
			wantDirective: cobra.ShellCompDirectiveFilterFileExt,
		},
		{
			name:          "CompleteDirsFiltersDirectories",
			option:        flag.CompleteDirs(),
			toComplete:    "",
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveFilterDirs,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			var dst string
			flag.Add(flag.NewRegistry(fs), "value", &dst, tc.option)

			// Act
			values, directive := completeFlag(t, fs, "value", tc.toComplete)

			// Assert
			if got, want := values, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("completeFlag(...) values = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
			if got, want := directive, tc.wantDirective; got != want {
				t.Errorf("completeFlag(...) directive = %v, want %v", got, want)
			}
		})
	}
}

func TestCompletionOptions_Conflict_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []flag.Option
	}{
		{
			name:    "TwoCompleteFrom",
			options: []flag.Option{flag.CompleteFrom("a"), flag.CompleteFrom("b")},
		},
		{
			name:    "CompleteFilesAndDirs",
			options: []flag.Option{flag.CompleteFiles(), flag.CompleteDirs()},
		},
		{
			name:    "CompleterFuncAndCompleteFrom",
			options: []flag.Option{flag.CompleterFunc(func(string) []string { return nil }), flag.CompleteFrom("a")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var dst string

			// Act & Assert
			requirePanic(t, func() { flag.Add(fs, "value", &dst, tc.options...) })
		})
	}
}
