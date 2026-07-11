package tag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/template/tag"
)

func TestThemed(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		role string
		text string
		want string
	}{
		{
			name: "WrapsText",
			role: "title",
			text: "hi",
			want: "[theme:title]hi[/theme]",
		},
		{
			name: "EmptyText",
			role: "error",
			text: "",
			want: "[theme:error][/theme]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			wrapped := tag.Themed(tc.role, tc.text)

			// Assert
			if got, want := wrapped, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Themed(%q, %q) = %q, want %q", tc.role, tc.text, got, want)
			}
		})
	}
}

func TestRaw(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		text string
		want string
	}{
		{
			name: "PlainText",
			text: "hi",
			want: "[richtext:off]hi[/richtext]",
		},
		{
			name: "TextWithBrackets",
			text: "[fg:red]",
			want: "[richtext:off][fg:red][/richtext]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			wrapped := tag.Raw(tc.text)

			// Assert
			if got, want := wrapped, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Raw(%q) = %q, want %q", tc.text, got, want)
			}
		})
	}
}
