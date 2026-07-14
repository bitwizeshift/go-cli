package flag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/flag/flagtest"
)

func TestCompletionOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		option         flag.Option
		toComplete     string
		want           []string
		wantExtensions []string
		wantValues     bool
		wantFiles      bool
		wantDirs       bool
	}{
		{
			name:           "CompleteFromMatchesPrefix",
			option:         flag.CompleteFrom("json", "yaml", "jsonl"),
			toComplete:     "js",
			want:           []string{"json", "jsonl"},
			wantExtensions: nil,
			wantValues:     true,
			wantFiles:      false,
			wantDirs:       false,
		},
		{
			name:           "CompleteFromEmptyPrefixMatchesAll",
			option:         flag.CompleteFrom("json", "yaml"),
			toComplete:     "",
			want:           []string{"json", "yaml"},
			wantExtensions: nil,
			wantValues:     true,
			wantFiles:      false,
			wantDirs:       false,
		},
		{
			name:           "CompleteFromNoMatchReturnsEmpty",
			option:         flag.CompleteFrom("json", "yaml"),
			toComplete:     "x",
			want:           nil,
			wantExtensions: nil,
			wantValues:     true,
			wantFiles:      false,
			wantDirs:       false,
		},
		{
			name:           "CompleterFuncPassesCandidatesThrough",
			option:         flag.CompleterFunc(suffixCompleter),
			toComplete:     "value",
			want:           []string{"value-done"},
			wantExtensions: nil,
			wantValues:     true,
			wantFiles:      false,
			wantDirs:       false,
		},
		{
			name:           "CompleteFilesDefersToShell",
			option:         flag.CompleteFiles(),
			toComplete:     "",
			want:           nil,
			wantExtensions: nil,
			wantValues:     false,
			wantFiles:      true,
			wantDirs:       false,
		},
		{
			name:           "CompleteFilesMatchingNormalizesExtensions",
			option:         flag.CompleteFilesMatching(".json", "yaml"),
			toComplete:     "",
			want:           nil,
			wantExtensions: []string{"json", "yaml"},
			wantValues:     false,
			wantFiles:      true,
			wantDirs:       false,
		},
		{
			name:           "CompleteDirsFiltersDirectories",
			option:         flag.CompleteDirs(),
			toComplete:     "",
			want:           nil,
			wantExtensions: nil,
			wantValues:     false,
			wantFiles:      false,
			wantDirs:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			registry := flagtest.NewRegistry()
			var dst string
			sut := flag.Add(registry, "value", &dst, tc.option)

			// Act
			completion := flagtest.CompleteFlag(sut, tc.toComplete)

			// Assert
			if got, want := completion.Candidates(), tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("CompleteFlag(...).Candidates() = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
			if got, want := completion.FileExtensions(), tc.wantExtensions; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("CompleteFlag(...).FileExtensions() = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
			if got, want := completion.CompletesValues(), tc.wantValues; got != want {
				t.Errorf("CompleteFlag(...).CompletesValues() = %t, want %t", got, want)
			}
			if got, want := completion.CompletesFiles(), tc.wantFiles; got != want {
				t.Errorf("CompleteFlag(...).CompletesFiles() = %t, want %t", got, want)
			}
			if got, want := completion.CompletesDirs(), tc.wantDirs; got != want {
				t.Errorf("CompleteFlag(...).CompletesDirs() = %t, want %t", got, want)
			}
		})
	}
}

func TestCompletionOptions_NoCompletionOption_DoesNotComplete(t *testing.T) {
	t.Parallel()

	// Arrange
	registry := flagtest.NewRegistry()
	var dst string
	sut := flag.Add(registry, "value", &dst)

	// Act
	completion := flagtest.CompleteFlag(sut, "")

	// Assert
	if got, want := completion, (*flagtest.Completion)(nil); got != want {
		t.Errorf("CompleteFlag(...) = %v, want %v", got, want)
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

// suffixCompleter completes the word being completed by suffixing it, to observe
// that the word reaches a [flag.CompleterFunc] unaltered.
func suffixCompleter(toComplete string) []string {
	return []string{toComplete + "-done"}
}

// noCompleter is a completer that offers no candidates.
func noCompleter(string) []string {
	return nil
}
