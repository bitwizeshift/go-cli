package flagtest_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/flag/flagtest"
	"github.com/bitwizeshift/go-cli/internal/annotation"
)

type FakeT struct {
	testing.TB
	IsHelper bool
	Fatals   []string
}

func (ft *FakeT) Helper() {
	ft.IsHelper = true
}

func (ft *FakeT) Fatalf(format string, args ...any) {
	ft.Fatals = append(ft.Fatals, fmt.Sprintf(format, args...))
}

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		args       []string
		wantFatals int
	}{
		{
			name:       "ValidArgsDoNotFail",
			args:       []string{"--flag"},
			wantFatals: 0,
		},
		{
			name:       "InvalidArgsFail",
			args:       []string{"--unknown"},
			wantFatals: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			registry := flagtest.NewRegistry()
			flag.Add(registry, "flag", new(bool))
			ft := &FakeT{}

			// Act
			flagtest.Parse(ft, registry, tc.args...)

			// Assert
			if got, want := len(ft.Fatals), tc.wantFatals; !cmp.Equal(got, want) {
				t.Errorf("Parse(...) fatals = %d, want %d", got, want)
			}
		})
	}
}

func TestCompleteFlag(t *testing.T) {
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
			name:           "Values",
			option:         flag.CompleteFrom("json", "yaml"),
			toComplete:     "j",
			want:           []string{"json"},
			wantExtensions: nil,
			wantValues:     true,
			wantFiles:      false,
			wantDirs:       false,
		},
		{
			name:           "Files",
			option:         flag.CompleteFiles(),
			toComplete:     "",
			want:           nil,
			wantExtensions: nil,
			wantValues:     false,
			wantFiles:      true,
			wantDirs:       false,
		},
		{
			name:           "FilesMatchingExtensions",
			option:         flag.CompleteFilesMatching("json"),
			toComplete:     "",
			want:           nil,
			wantExtensions: []string{"json"},
			wantValues:     false,
			wantFiles:      true,
			wantDirs:       false,
		},
		{
			name:           "Dirs",
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
			target := flag.Add(registry, "value", new(string), tc.option)

			// Act
			sut := flagtest.CompleteFlag(target, tc.toComplete)

			// Assert
			if got, want := sut.Candidates(), tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Completion.Candidates() = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
			if got, want := sut.FileExtensions(), tc.wantExtensions; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Completion.FileExtensions() = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
			if got, want := sut.CompletesValues(), tc.wantValues; got != want {
				t.Errorf("Completion.CompletesValues() = %t, want %t", got, want)
			}
			if got, want := sut.CompletesFiles(), tc.wantFiles; got != want {
				t.Errorf("Completion.CompletesFiles() = %t, want %t", got, want)
			}
			if got, want := sut.CompletesDirs(), tc.wantDirs; got != want {
				t.Errorf("Completion.CompletesDirs() = %t, want %t", got, want)
			}
		})
	}
}

func TestCompleteFlag_FlagWithoutCompletion_ReturnsNil(t *testing.T) {
	t.Parallel()

	// Arrange
	registry := flagtest.NewRegistry()
	target := flag.Add(registry, "value", new(string))

	// Act
	sut := flagtest.CompleteFlag(target, "")

	// Assert
	if got, want := sut, (*flagtest.Completion)(nil); got != want {
		t.Errorf("CompleteFlag(...) = %v, want %v", got, want)
	}
}

func TestCompletion_NilCompletion_CompletesNothing(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := (*flagtest.Completion)(nil)

	// Act
	candidates, extensions := sut.Candidates(), sut.FileExtensions()
	values, files, dirs := sut.CompletesValues(), sut.CompletesFiles(), sut.CompletesDirs()

	// Assert
	if got, want := candidates, []string(nil); !cmp.Equal(got, want) {
		t.Errorf("Completion.Candidates() = %v, want %v", got, want)
	}
	if got, want := extensions, []string(nil); !cmp.Equal(got, want) {
		t.Errorf("Completion.FileExtensions() = %v, want %v", got, want)
	}
	if got, want := values, false; got != want {
		t.Errorf("Completion.CompletesValues() = %t, want %t", got, want)
	}
	if got, want := files, false; got != want {
		t.Errorf("Completion.CompletesFiles() = %t, want %t", got, want)
	}
	if got, want := dirs, false; got != want {
		t.Errorf("Completion.CompletesDirs() = %t, want %t", got, want)
	}
}

func TestAllFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	registry := flagtest.NewRegistry()
	verbose := flag.Add(registry, "verbose", new(bool), flag.Shorthand("v"))
	name := flag.Add(registry, "name", new(string))
	flag.Add(registry, "region", new(string))
	novalue := flag.Add(registry, "novalue", new(int))
	annotation.MarkRequired(name)
	annotation.MarkRequiredTogether(verbose, name)
	annotation.MarkMutuallyExclusive(verbose, novalue)
	annotation.MarkOneRequired(name, novalue)
	annotation.AddToGroup("Location Flags", novalue)

	// Act
	flags := flagtest.AllFlags(registry)

	// Assert
	want := []*flagtest.Flag{
		{
			Long:            "name",
			Type:            "string",
			Required:        true,
			RequiredWith:    []string{"name", "verbose"},
			OneRequiredWith: []string{"name", "novalue"},
		},
		{
			Long:            "novalue",
			Type:            "int",
			Group:           "Location Flags",
			ExclusiveWith:   []string{"novalue", "verbose"},
			OneRequiredWith: []string{"name", "novalue"},
		},
		{
			Long: "region",
			Type: "string",
		},
		{
			Long:          "verbose",
			Short:         "v",
			Type:          "bool",
			RequiredWith:  []string{"name", "verbose"},
			ExclusiveWith: []string{"novalue", "verbose"},
		},
	}
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("AllFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestLongFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	var b bool
	registry := flagtest.NewRegistry()
	flag.Add(registry, "alpha", &b, flag.Shorthand("a"))
	flag.Add(registry, "beta", &b)

	// Act
	long := flagtest.LongFlags(registry)

	// Assert
	if got, want := long, []string{"alpha", "beta"}; !cmp.Equal(got, want) {
		t.Errorf("LongFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestShortFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	registry := flagtest.NewRegistry()
	flag.Add(registry, "alpha", new(bool), flag.Shorthand("a"))
	flag.Add(registry, "beta", new(bool))

	// Act
	short := flagtest.ShortFlags(registry)

	// Assert
	if got, want := short, []string{"a", ""}; !cmp.Equal(got, want) {
		t.Errorf("ShortFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}
