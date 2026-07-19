package completion_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/internal/completion"
)

func TestAddFlag(t *testing.T) {
	t.Parallel()

	// Arrange
	target := newStringFlag("flag")

	// Act
	completion.AddFlag(target, func(string) ([]string, completion.Directive) {
		return nil, completion.Default
	})
	ids := len(target.Annotations[completion.Annotation])

	// Assert
	if got, want := ids, 1; !cmp.Equal(got, want) {
		t.Errorf("AddFlag(...) recorded %d ids, want %d", got, want)
	}
}

func TestFlagFunc_FlagWithCompletion_ReturnsCompletion(t *testing.T) {
	t.Parallel()

	// Arrange
	target := newStringFlag("flag")
	completion.AddFlag(target, func(toComplete string) ([]string, completion.Directive) {
		return []string{toComplete + "-value"}, completion.NoFileComp
	})

	// Act
	sut := completion.FlagFunc(target)
	values, directive := complete(t, sut, "flag")

	// Assert
	if got, want := values, []string{"flag-value"}; !cmp.Equal(got, want) {
		t.Errorf("FlagFunc(...)(...) values = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
	if got, want := directive, completion.NoFileComp; !cmp.Equal(got, want) {
		t.Errorf("FlagFunc(...)(...) directive = %v, want %v", got, want)
	}
}

func TestFlagFunc_FlagWithoutCompletion_ReturnsNil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		annotations map[string][]string
	}{
		{
			name:        "Unannotated",
			annotations: nil,
		},
		{
			name:        "UnknownCompletionID",
			annotations: map[string][]string{completion.Annotation: {"not-a-registered-id"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			target := newStringFlag("flag")
			target.Annotations = tc.annotations

			// Act
			sut := completion.FlagFunc(target)

			// Assert
			if got, want := sut == nil, true; !cmp.Equal(got, want) {
				t.Errorf("FlagFunc(...) = nil %t, want %t", got, want)
			}
		})
	}
}

func TestRegisterFlags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		values        []string
		directive     completion.Directive
		want          []string
		wantDirective cobra.ShellCompDirective
	}{
		{
			name:          "DefaultDefersToShell",
			values:        nil,
			directive:     completion.Default,
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveDefault,
		},
		{
			name:          "NoFileCompOffersCandidatesOnly",
			values:        []string{"json", "yaml"},
			directive:     completion.NoFileComp,
			want:          []string{"json", "yaml"},
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:          "FilterFileExtCompletesExtensions",
			values:        []string{"json"},
			directive:     completion.FilterFileExt,
			want:          []string{"json"},
			wantDirective: cobra.ShellCompDirectiveFilterFileExt,
		},
		{
			name:          "FilterDirsCompletesDirectories",
			values:        nil,
			directive:     completion.FilterDirs,
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveFilterDirs,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cmd := newStringFlagCommand("flag")
			completion.AddFlag(cmd.Flags().Lookup("flag"), func(string) ([]string, completion.Directive) {
				return tc.values, tc.directive
			})
			completion.RegisterFlags(cmd)

			// Act
			values, directive := invokeCompletion(t, cmd, "flag", "")

			// Assert
			if got, want := values, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("completion values = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
			if got, want := directive, tc.wantDirective; !cmp.Equal(got, want) {
				t.Errorf("completion directive = %v, want %v", got, want)
			}
		})
	}
}

func TestRegisterFlags_Unannotated_RegistersNothing(t *testing.T) {
	t.Parallel()

	// Arrange
	cmd := newStringFlagCommand("flag")

	// Act
	completion.RegisterFlags(cmd)

	// Assert
	if _, got := cmd.GetFlagCompletionFunc("flag"); got != false {
		t.Errorf("GetFlagCompletionFunc(...) registered = %t, want %t", got, false)
	}
}

// newStringFlag registers a string flag named name on a fresh flag set and
// returns it, for exercising helpers that operate on a single flag.
func newStringFlag(name string) *pflag.Flag {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String(name, "", "")
	return fs.Lookup(name)
}

// newStringFlagCommand returns a command carrying a single string flag named
// name, for exercising flag completion registration.
func newStringFlagCommand(name string) *cobra.Command {
	cmd := &cobra.Command{Use: "root"}
	cmd.Flags().String(name, "", "")
	return cmd
}

// complete invokes fn with toComplete, failing the test if no completion
// function was found.
func complete(t testing.TB, fn completion.Func, toComplete string) ([]string, completion.Directive) {
	t.Helper()
	if fn == nil {
		t.Fatalf("FlagFunc(...) = nil, want a completion function")
	}
	return fn(toComplete)
}

// invokeCompletion invokes the completion function registered on cmd for the
// named flag, failing the test if none was registered.
func invokeCompletion(t testing.TB, cmd *cobra.Command, name, toComplete string) ([]string, cobra.ShellCompDirective) {
	t.Helper()
	complete, ok := cmd.GetFlagCompletionFunc(name)
	if !ok {
		t.Fatalf("GetFlagCompletionFunc(%q) = _, false, want true", name)
	}
	return complete(cmd, nil, toComplete)
}
