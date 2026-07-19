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

func TestForArgs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		fns        map[int]completion.Func
		unmatched  completion.Func
		args       []string
		toComplete string
		want       offered
	}{
		{
			name: "CompletesRegisteredIndex",
			fns: map[int]completion.Func{
				0: completerOf([]string{"alpha", "beta"}, completion.NoFileComp),
			},
			unmatched:  nil,
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
			unmatched:  nil,
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
			unmatched:  nil,
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
			unmatched:  nil,
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
			unmatched:  nil,
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
			unmatched:  nil,
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
			unmatched:  nil,
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
			unmatched:  nil,
			args:       nil,
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  cobra.ShellCompDirectiveFilterDirs,
			},
		},
		{
			name:       "UnregisteredIndexDefersToUnmatched",
			fns:        map[int]completion.Func{},
			unmatched:  completerOf([]string{"rest"}, completion.NoFileComp),
			args:       []string{"alpha"},
			toComplete: "",
			want: offered{
				Candidates: []string{"rest"},
				Directive:  cobra.ShellCompDirectiveNoFileComp,
			},
		},
		{
			name: "PositionalWinsOverUnmatchedAtItsIndex",
			fns: map[int]completion.Func{
				0: completerOf([]string{"first"}, completion.NoFileComp),
			},
			unmatched:  completerOf([]string{"rest"}, completion.NoFileComp),
			args:       nil,
			toComplete: "",
			want: offered{
				Candidates: []string{"first"},
				Directive:  cobra.ShellCompDirectiveNoFileComp,
			},
		},
		{
			name: "UnmatchedClaimsIndexBeyondPositionals",
			fns: map[int]completion.Func{
				0: completerOf([]string{"first"}, completion.NoFileComp),
			},
			unmatched:  completerOf([]string{"rest"}, completion.NoFileComp),
			args:       []string{"alpha", "beta"},
			toComplete: "",
			want: offered{
				Candidates: []string{"rest"},
				Directive:  cobra.ShellCompDirectiveNoFileComp,
			},
		},
		{
			name:       "PassesToCompleteThroughToUnmatched",
			fns:        nil,
			unmatched:  echoCompleter,
			args:       nil,
			toComplete: "pre",
			want: offered{
				Candidates: []string{"pre-done"},
				Directive:  cobra.ShellCompDirectiveNoFileComp,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := completion.ForArgs(tc.fns, tc.unmatched)

			// Act
			candidates, directive := sut(nil, tc.args, tc.toComplete)

			// Assert
			offer := offered{
				Candidates: candidates,
				Directive:  directive,
			}
			if got, want := offer, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("ForArgs(...)(...) = %+v, want %+v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}

func TestForArgs_Nil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		fns       map[int]completion.Func
		unmatched completion.Func
		want      bool
	}{
		{
			name:      "NoCompleters",
			fns:       map[int]completion.Func{},
			unmatched: nil,
			want:      true,
		},
		{
			name:      "NilMap",
			fns:       nil,
			unmatched: nil,
			want:      true,
		},
		{
			name:      "UnmatchedOnly",
			fns:       nil,
			unmatched: completerOf([]string{"rest"}, completion.NoFileComp),
			want:      false,
		},
		{
			name: "PositionalsOnly",
			fns: map[int]completion.Func{
				0: completerOf([]string{"first"}, completion.NoFileComp),
			},
			unmatched: nil,
			want:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			sut := completion.ForArgs(tc.fns, tc.unmatched)

			// Assert
			if got, want := sut == nil, tc.want; !cmp.Equal(got, want) {
				t.Errorf("ForArgs(...) = nil %t, want %t", got, want)
			}
		})
	}
}
