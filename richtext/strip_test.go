package richtext_test

import (
	"testing"

	"github.com/bitwizeshift/go-cli/richtext"
	"github.com/google/go-cmp/cmp"
)

func TestStrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "NoTags",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "KnownTagsRemoved",
			input: "[fg:red]a[attr:bold]b[/attr][/fg]",
			want:  "ab",
		},
		{
			name:  "UnknownNamespacePreserved",
			input: "[foo:bar]kept[/foo]",
			want:  "[foo:bar]kept[/foo]",
		},
		{
			name:  "IncompleteTrailingTagPreserved",
			input: "text[fg",
			want:  "text[fg",
		},
		{
			name:  "RawRegionContentsKeptMarkersRemoved",
			input: "[richtext:off]a[fg:red]b[/richtext]",
			want:  "a[fg:red]b",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			stripped := richtext.Strip(tc.input)

			// Assert
			if got, want := stripped, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Strip() = %q, want %q", got, want)
			}
		})
	}
}

func TestLen(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "PlainASCII",
			input: "hello",
			want:  5,
		},
		{
			name:  "TagsIgnored",
			input: "[fg:red]hello[/fg]",
			want:  5,
		},
		{
			name:  "MultibyteRunesCountedOnce",
			input: "[fg:red]héllo[/fg]",
			want:  5,
		},
		{
			name:  "RawRegionCountsVisibleContents",
			input: "[richtext:off][x][/richtext]",
			want:  3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			length := richtext.Len(tc.input)

			// Assert
			if got, want := length, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Len() = %d, want %d", got, want)
			}
		})
	}
}
