package flag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/flag/flagtest"
	"github.com/bitwizeshift/go-cli/internal/annotation"
)

// completion is what a flag offers for a word being completed: the candidates it
// returns, and the directive telling a shell how to treat them.
type completion struct {
	Candidates []string
	Directive  annotation.CompletionDirective
}

// completionOf registers a string flag carrying option, and completes it with
// the partial word toComplete. It fails the test if option registered no
// completion on the flag.
func completionOf(t testing.TB, option flag.Option, toComplete string) completion {
	t.Helper()

	f := flag.Add(flagtest.NewRegistry(), "value", new(string), option)
	complete := annotation.GetCompletionFunc(f.Flag())
	if complete == nil {
		t.Fatalf("Add(...) registered no completion function, want one")
	}
	candidates, directive := complete(toComplete)
	return completion{
		Candidates: candidates,
		Directive:  directive,
	}
}

// suffixCompleter completes the word being completed by suffixing it, to observe
// that the word reaches a [flag.CompleterFunc] unaltered.
func suffixCompleter(toComplete string) []string {
	return []string{toComplete + "-done"}
}

// noCompleter is a completer that offers no candidates.
func noCompleter(string) []string {
	return nil
}

func TestCompletionOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		option     flag.Option
		toComplete string
		want       completion
	}{
		{
			name:       "CompleteFromMatchesPrefix",
			option:     flag.CompleteFrom("json", "yaml", "jsonl"),
			toComplete: "js",
			want: completion{
				Candidates: []string{"json", "jsonl"},
				Directive:  annotation.CompletionNoFileComp,
			},
		},
		{
			name:       "CompleteFromEmptyPrefixMatchesAll",
			option:     flag.CompleteFrom("json", "yaml"),
			toComplete: "",
			want: completion{
				Candidates: []string{"json", "yaml"},
				Directive:  annotation.CompletionNoFileComp,
			},
		},
		{
			name:       "CompleteFromNoMatchOffersNothing",
			option:     flag.CompleteFrom("json", "yaml"),
			toComplete: "x",
			want: completion{
				Candidates: nil,
				Directive:  annotation.CompletionNoFileComp,
			},
		},
		{
			name:       "CompleterFuncOffersItsCandidates",
			option:     flag.CompleterFunc(suffixCompleter),
			toComplete: "value",
			want: completion{
				Candidates: []string{"value-done"},
				Directive:  annotation.CompletionNoFileComp,
			},
		},
		{
			name:       "CompleteFilesDefersToShell",
			option:     flag.CompleteFiles(),
			toComplete: "",
			want: completion{
				Candidates: nil,
				Directive:  annotation.CompletionDefault,
			},
		},
		{
			name:       "CompleteFilesMatchingNormalizesExtensions",
			option:     flag.CompleteFilesMatching(".json", "yaml"),
			toComplete: "",
			want: completion{
				Candidates: []string{"json", "yaml"},
				Directive:  annotation.CompletionFilterFileExt,
			},
		},
		{
			name:       "CompleteDirsFiltersDirectories",
			option:     flag.CompleteDirs(),
			toComplete: "",
			want: completion{
				Candidates: nil,
				Directive:  annotation.CompletionFilterDirs,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.option

			// Act
			offered := completionOf(t, sut, tc.toComplete)

			// Assert
			if got, want := offered, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Add(..., option) completion = %+v, want %+v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
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
			options: []flag.Option{flag.CompleterFunc(noCompleter), flag.CompleteFrom("a")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			registry := flagtest.NewRegistry()
			var dst string

			// Act & Assert
			requirePanic(t, func() { flag.Add(registry, "value", &dst, tc.options...) })
		})
	}
}
