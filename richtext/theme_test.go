package richtext_test

import (
	"errors"
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

	parent, err := richtext.NewTheme(struct {
		Title   style.Style `theme:"title"`
		Heading style.Style `theme:"heading"`
	}{
		Title:   style.Style{Foreground: style.Cyan, Attributes: style.Bold},
		Heading: style.Style{Foreground: style.Green},
	})
	if err != nil {
		t.Fatalf("NewTheme() = %v, want nil", err)
	}

	child, err := parent.New(struct {
		Title style.Style `theme:"title"`
		Note  style.Style `theme:"note"`
	}{
		Title: style.Style{Foreground: style.Red},
		Note:  style.Style{Foreground: style.Blue},
	})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}
	return child
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

func TestTheme_New_InvalidInput_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	parent, err := richtext.NewTheme(struct {
		Title style.Style `theme:"title"`
	}{})
	if err != nil {
		t.Fatalf("NewTheme() = %v, want nil", err)
	}

	// Act
	derived, err := parent.New(42)

	// Assert
	if got, want := err, richtext.ErrNotStruct; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("New() = %v, want %v", got, want)
	}
	if got, want := derived == nil, true; !cmp.Equal(got, want) {
		t.Errorf("New() theme = %v, want nil", derived)
	}
}

func TestNewTheme(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   any
		wantErr error
	}{
		{
			name:    "NotAStruct",
			input:   42,
			wantErr: richtext.ErrNotStruct,
		},
		{
			name:    "NilPointer",
			input:   (*int)(nil),
			wantErr: richtext.ErrNotStruct,
		},
		{
			name: "WrongFieldType",
			input: struct {
				Title int `theme:"title"`
			}{Title: 1},
			wantErr: richtext.ErrThemeFieldType,
		},
		{
			name: "DuplicateName",
			input: struct {
				A style.Style `theme:"dup"`
				B style.Style `theme:"dup"`
			}{},
			wantErr: richtext.ErrDuplicateTheme,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			theme, err := richtext.NewTheme(tc.input)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("NewTheme() = %v, want %v", got, want)
			}
			if got, want := theme == nil, true; !cmp.Equal(got, want) {
				t.Errorf("NewTheme() theme = %v, want nil", theme)
			}
		})
	}
}

func TestNewTheme_WrongFieldType_ReturnsFieldError(t *testing.T) {
	t.Parallel()

	// Arrange
	input := struct {
		Heading int `theme:"heading"`
	}{Heading: 1}

	// Act
	_, err := richtext.NewTheme(input)

	// Assert
	var fieldErr *richtext.ThemeFieldError
	if !errors.As(err, &fieldErr) {
		t.Fatalf("NewTheme() = %v, want *richtext.ThemeFieldError", err)
	}
	if got, want := fieldErr.Field, "Heading"; !cmp.Equal(got, want) {
		t.Errorf("Field = %q, want %q", got, want)
	}
	if got, want := fieldErr.Name, "heading"; !cmp.Equal(got, want) {
		t.Errorf("Name = %q, want %q", got, want)
	}
}

func TestThemeFieldError_Error_ContainsFieldAndName(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := &richtext.ThemeFieldError{Field: "Heading", Name: "heading", Err: richtext.ErrThemeFieldType}

	// Act
	msg := sut.Error()

	// Assert
	if got, want := strings.Contains(msg, "Heading"), true; !cmp.Equal(got, want) {
		t.Errorf("Error() = %q, want it to contain %q", msg, "Heading")
	}
	if got, want := strings.Contains(msg, "heading"), true; !cmp.Equal(got, want) {
		t.Errorf("Error() = %q, want it to contain %q", msg, "heading")
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
