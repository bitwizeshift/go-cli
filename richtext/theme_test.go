package richtext_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/richtext"
	"github.com/bitwizeshift/go-cli/richtext/style"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const (
	greenFg = "\x1b[32m"
	blueFg  = "\x1b[34m"
)

// derivedTheme returns a child theme that overrides "title" and adds "note",
// deriving a parent that defines "title" and "heading".
func derivedTheme(t *testing.T) *richtext.Theme {
	t.Helper()

	parent := richtext.NewTheme(map[string]style.Style{
		"title":   {Foreground: style.Cyan, Attributes: style.Bold},
		"heading": {Foreground: style.Green},
	})
	return parent.New(map[string]style.Style{
		"title": {Foreground: style.Red},
		"note":  {Foreground: style.Blue},
	})
}

func TestTheme_New(t *testing.T) {
	t.Parallel()

	child := derivedTheme(t)

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "OverridesParent",
			input: "[theme:title]x[/theme]",
			want:  reset + red + "x" + reset,
		},
		{
			name:  "FallsBackToParent",
			input: "[theme:heading]x[/theme]",
			want:  reset + greenFg + "x" + reset,
		},
		{
			name:  "UsesOwnName",
			input: "[theme:note]x[/theme]",
			want:  reset + blueFg + "x" + reset,
		},
		{
			name:  "UnknownNameFallsThroughToReset",
			input: "[theme:missing]x[/theme]",
			want:  "x",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf strings.Builder
			sut := richtext.NewWriter(&buf, child)
			sut.ForceColour()

			// Act
			_, err := sut.Write([]byte(tc.input))

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Write() = %v, want nil", got)
			}
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Write() output = %q, want %q", got, want)
			}
		})
	}
}

func TestNewTheme_InvalidName_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		styles    map[string]style.Style
		themeName string
	}{
		{
			name:      "EmptyName",
			styles:    map[string]style.Style{"": {Foreground: style.Red}},
			themeName: "",
		},
		{
			name:      "NameContainsBracket",
			styles:    map[string]style.Style{"a]b": {Foreground: style.Red}},
			themeName: "a]b",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var recovered any

			// Act
			func() {
				defer func() { recovered = recover() }()
				_ = richtext.NewTheme(tc.styles)
			}()

			// Assert
			message, ok := recovered.(string)
			want := fmt.Sprintf("%q", tc.themeName)
			if !ok || !strings.Contains(message, want) {
				t.Fatalf("NewTheme() panic = %v, want a message containing %q", recovered, want)
			}
		})
	}
}

func TestTheme_New_InvalidName_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		styles    map[string]style.Style
		themeName string
	}{
		{
			name:      "EmptyName",
			styles:    map[string]style.Style{"": {Foreground: style.Red}},
			themeName: "",
		},
		{
			name:      "NameContainsBracket",
			styles:    map[string]style.Style{"a]b": {Foreground: style.Red}},
			themeName: "a]b",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			parent := richtext.NewTheme(map[string]style.Style{
				"title": {Foreground: style.Green},
			})
			var recovered any

			// Act
			func() {
				defer func() { recovered = recover() }()
				_ = parent.New(tc.styles)
			}()

			// Assert
			message, ok := recovered.(string)
			want := fmt.Sprintf("%q", tc.themeName)
			if !ok || !strings.Contains(message, want) {
				t.Fatalf("New() panic = %v, want a message containing %q", recovered, want)
			}
		})
	}
}

func TestTagError_Error_ContainsNamespace(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := &richtext.TagError{Namespace: "fg", Err: richtext.ErrUnclosedTag}

	// Act
	msg := sut.Error()

	// Assert
	if got, want := strings.Contains(msg, "fg"), true; !cmp.Equal(got, want) {
		t.Errorf("Error() = %q, want it to contain %q", msg, "fg")
	}
}
