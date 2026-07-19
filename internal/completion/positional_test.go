package completion_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"

	"github.com/bitwizeshift/go-cli/internal/completion"
)

// offered is what a command offers for a word being completed: the candidates it
// returns, and the cobra directive telling a shell how to treat them.
type offered struct {
	Candidates []string
	Directive  cobra.ShellCompDirective
}

// completerOf returns a completion function offering candidates with directive,
// disregarding the word being completed.
func completerOf(candidates []string, directive completion.Directive) completion.Func {
	return func(string) ([]string, completion.Directive) {
		return candidates, directive
	}
}

// echoCompleter suffixes the word being completed, to observe that the word
// reaches the registered function unaltered.
func echoCompleter(toComplete string) ([]string, completion.Directive) {
	return []string{toComplete + "-done"}, completion.NoFileComp
}

func TestForPositionals(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		fns        map[int]completion.Func
		args       []string
		toComplete string
		want       offered
	}{
		{
			name: "CompletesRegisteredIndex",
			fns: map[int]completion.Func{
				0: completerOf([]string{"alpha", "beta"}, completion.NoFileComp),
			},
			args:       nil,
			toComplete: "",
			want: offered{
				Candidates: []string{"alpha", "beta"},
				Directive:  cobra.ShellCompDirectiveNoFileComp,
			},
		},
		{
			name: "SelectsIndexFromArgCount",
			fns: map[int]completion.Func{
				0: completerOf([]string{"first"}, completion.NoFileComp),
				1: completerOf([]string{"second"}, completion.NoFileComp),
			},
			args:       []string{"alpha"},
			toComplete: "",
			want: offered{
				Candidates: []string{"second"},
				Directive:  cobra.ShellCompDirectiveNoFileComp,
			},
		},
		{
			name: "UnregisteredIndexDefersToShell",
			fns: map[int]completion.Func{
				0: completerOf([]string{"first"}, completion.NoFileComp),
			},
			args:       []string{"alpha"},
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  cobra.ShellCompDirectiveDefault,
			},
		},
		{
			name: "SparseIndicesSkipGaps",
			fns: map[int]completion.Func{
				2: completerOf([]string{"third"}, completion.NoFileComp),
			},
			args:       []string{"alpha", "beta"},
			toComplete: "",
			want: offered{
				Candidates: []string{"third"},
				Directive:  cobra.ShellCompDirectiveNoFileComp,
			},
		},
		{
			name: "PassesToCompleteThrough",
			fns: map[int]completion.Func{
				0: echoCompleter,
			},
			args:       nil,
			toComplete: "pre",
			want: offered{
				Candidates: []string{"pre-done"},
				Directive:  cobra.ShellCompDirectiveNoFileComp,
			},
		},
		{
			name: "TranslatesDefault",
			fns: map[int]completion.Func{
				0: completerOf(nil, completion.Default),
			},
			args:       nil,
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  cobra.ShellCompDirectiveDefault,
			},
		},
		{
			name: "TranslatesFilterFileExt",
			fns: map[int]completion.Func{
				0: completerOf([]string{"json"}, completion.FilterFileExt),
			},
			args:       nil,
			toComplete: "",
			want: offered{
				Candidates: []string{"json"},
				Directive:  cobra.ShellCompDirectiveFilterFileExt,
			},
		},
		{
			name: "TranslatesFilterDirs",
			fns: map[int]completion.Func{
				0: completerOf(nil, completion.FilterDirs),
			},
			args:       nil,
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  cobra.ShellCompDirectiveFilterDirs,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := completion.ForPositionals(tc.fns)

			// Act
			candidates, directive := sut(nil, tc.args, tc.toComplete)

			// Assert
			offer := offered{
				Candidates: candidates,
				Directive:  directive,
			}
			if got, want := offer, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("ForPositionals(...)(...) = %+v, want %+v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}

func TestForPositionals_NoCompleters_ReturnsNil(t *testing.T) {
	t.Parallel()

	// Arrange
	fns := map[int]completion.Func{}

	// Act
	sut := completion.ForPositionals(fns)

	// Assert
	if got, want := sut == nil, true; !cmp.Equal(got, want) {
		t.Errorf("ForPositionals(...) = nil %t, want %t", got, want)
	}
}

func TestForPositionals_NilMap_ReturnsNil(t *testing.T) {
	t.Parallel()

	// Act
	sut := completion.ForPositionals(nil)

	// Assert
	if got, want := sut == nil, true; !cmp.Equal(got, want) {
		t.Errorf("ForPositionals(...) = nil %t, want %t", got, want)
	}
}
