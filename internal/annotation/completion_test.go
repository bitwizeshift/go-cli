package annotation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"

	"github.com/bitwizeshift/go-cli/internal/annotation"
)

func TestAddCompletion(t *testing.T) {
	t.Parallel()

	// Arrange
	target := newStringFlag("flag")

	// Act
	annotation.AddCompletion(target, func(string) ([]string, annotation.CompletionDirective) {
		return nil, annotation.CompletionDefault
	})
	ids := len(target.Annotations[annotation.AnnotationCompletion])

	// Assert
	if got, want := ids, 1; !cmp.Equal(got, want) {
		t.Errorf("AddCompletion(...) recorded %d ids, want %d", got, want)
	}
}

func TestGetCompletionFunc_FlagWithCompletion_ReturnsCompletion(t *testing.T) {
	t.Parallel()

	// Arrange
	target := newStringFlag("flag")
	annotation.AddCompletion(target, func(toComplete string) ([]string, annotation.CompletionDirective) {
		return []string{toComplete + "-value"}, annotation.CompletionNoFileComp
	})

	// Act
	sut := annotation.GetCompletionFunc(target)
	values, directive := complete(t, sut, "flag")

	// Assert
	if got, want := values, []string{"flag-value"}; !cmp.Equal(got, want) {
		t.Errorf("GetCompletionFunc(...)(...) values = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
	if got, want := directive, annotation.CompletionNoFileComp; !cmp.Equal(got, want) {
		t.Errorf("GetCompletionFunc(...)(...) directive = %v, want %v", got, want)
	}
}

func TestGetCompletionFunc_FlagWithoutCompletion_ReturnsNil(t *testing.T) {
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
			annotations: map[string][]string{annotation.AnnotationCompletion: {"not-a-registered-id"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			target := newStringFlag("flag")
			target.Annotations = tc.annotations

			// Act
			sut := annotation.GetCompletionFunc(target)

			// Assert
			if got, want := sut == nil, true; !cmp.Equal(got, want) {
				t.Errorf("GetCompletionFunc(...) = nil %t, want %t", got, want)
			}
		})
	}
}

func TestRegisterFlagCompletions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		values        []string
		directive     annotation.CompletionDirective
		want          []string
		wantDirective cobra.ShellCompDirective
	}{
		{
			name:          "DefaultDefersToShell",
			values:        nil,
			directive:     annotation.CompletionDefault,
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveDefault,
		},
		{
			name:          "NoFileCompOffersCandidatesOnly",
			values:        []string{"json", "yaml"},
			directive:     annotation.CompletionNoFileComp,
			want:          []string{"json", "yaml"},
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:          "FilterFileExtCompletesExtensions",
			values:        []string{"json"},
			directive:     annotation.CompletionFilterFileExt,
			want:          []string{"json"},
			wantDirective: cobra.ShellCompDirectiveFilterFileExt,
		},
		{
			name:          "FilterDirsCompletesDirectories",
			values:        nil,
			directive:     annotation.CompletionFilterDirs,
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveFilterDirs,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cmd := newStringFlagCommand("flag")
			annotation.AddCompletion(cmd.Flags().Lookup("flag"), func(string) ([]string, annotation.CompletionDirective) {
				return tc.values, tc.directive
			})
			annotation.RegisterFlagCompletions(cmd)

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

func TestRegisterFlagCompletions_Unannotated_RegistersNothing(t *testing.T) {
	t.Parallel()

	// Arrange
	cmd := newStringFlagCommand("flag")

	// Act
	annotation.RegisterFlagCompletions(cmd)

	// Assert
	if _, got := cmd.GetFlagCompletionFunc("flag"); got != false {
		t.Errorf("GetFlagCompletionFunc(...) registered = %t, want %t", got, false)
	}
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
func complete(t testing.TB, fn annotation.CompletionFunc, toComplete string) ([]string, annotation.CompletionDirective) {
	t.Helper()
	if fn == nil {
		t.Fatalf("GetCompletionFunc(...) = nil, want a completion function")
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
