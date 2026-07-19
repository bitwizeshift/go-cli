package arg_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
	"github.com/bitwizeshift/go-cli/internal/completion"
)

// offered is what an argument offers for a word being completed: the candidates it
// returns, and the directive telling a shell how to treat them.
type offered struct {
	Candidates []string
	Directive  completion.Directive
}

// completionOf registers a string flag carrying option, and completes it with
// the partial word toComplete. It fails the test if option registered no
// completion on the flag.
func completionOf(t testing.TB, option arg.Option, toComplete string) offered {
	t.Helper()

	f := addFlag(argtest.NewCommandLine(), "value", new(string), option)
	complete := completion.FlagFunc(f.Flag())
	if complete == nil {
		t.Fatalf("Add(...) registered no completion function, want one")
	}
	candidates, directive := complete(toComplete)
	return offered{
		Candidates: candidates,
		Directive:  directive,
	}
}

// suffixCompleter completes the word being completed by suffixing it, to observe
// that the word reaches a [arg.CompleterFunc] unaltered.
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
		option     arg.Option
		toComplete string
		want       offered
	}{
		{
			name:       "CompleteFromMatchesPrefix",
			option:     arg.CompleteFrom("json", "yaml", "jsonl"),
			toComplete: "js",
			want: offered{
				Candidates: []string{"json", "jsonl"},
				Directive:  completion.NoFileComp,
			},
		},
		{
			name:       "CompleteFromEmptyPrefixMatchesAll",
			option:     arg.CompleteFrom("json", "yaml"),
			toComplete: "",
			want: offered{
				Candidates: []string{"json", "yaml"},
				Directive:  completion.NoFileComp,
			},
		},
		{
			name:       "CompleteFromNoMatchOffersNothing",
			option:     arg.CompleteFrom("json", "yaml"),
			toComplete: "x",
			want: offered{
				Candidates: nil,
				Directive:  completion.NoFileComp,
			},
		},
		{
			name:       "CompleterFuncOffersItsCandidates",
			option:     arg.CompleterFunc(suffixCompleter),
			toComplete: "value",
			want: offered{
				Candidates: []string{"value-done"},
				Directive:  completion.NoFileComp,
			},
		},
		{
			name:       "CompleteFilesDefersToShell",
			option:     arg.CompleteFiles(),
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  completion.Default,
			},
		},
		{
			name:       "CompleteFilesMatchingNormalizesExtensions",
			option:     arg.CompleteFilesMatching(".json", "yaml"),
			toComplete: "",
			want: offered{
				Candidates: []string{"json", "yaml"},
				Directive:  completion.FilterFileExt,
			},
		},
		{
			name:       "CompleteDirsFiltersDirectories",
			option:     arg.CompleteDirs(),
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  completion.FilterDirs,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.option

			// Act
			offer := completionOf(t, sut, tc.toComplete)

			// Assert
			if got, want := offer, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Add(..., option) completion = %+v, want %+v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}

func TestCompletionOptions_Conflict_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []arg.FlagOption
	}{
		{
			name:    "TwoCompleteFrom",
			options: []arg.FlagOption{arg.CompleteFrom("a"), arg.CompleteFrom("b")},
		},
		{
			name:    "CompleteFilesAndDirs",
			options: []arg.FlagOption{arg.CompleteFiles(), arg.CompleteDirs()},
		},
		{
			name:    "CompleterFuncAndCompleteFrom",
			options: []arg.FlagOption{arg.CompleterFunc(noCompleter), arg.CompleteFrom("a")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			var dst string

			// Act & Assert
			requirePanic(t, func() { addFlag(cl, "value", &dst, tc.options...) })
		})
	}
}
