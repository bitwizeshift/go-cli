package token_test

import (
	"testing"

	"github.com/bitwizeshift/go-cli/richtext/internal/token"
	"github.com/google/go-cmp/cmp"
)

func TestScanner_Scan(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		chunks []string
		want   []token.Token
	}{
		{
			name:   "PlainText",
			chunks: []string{"hello world"},
			want: []token.Token{
				{Kind: token.Text, Raw: "hello world"},
			},
		},
		{
			name:   "OpenTextClose",
			chunks: []string{"[fg:red]hi[/fg]"},
			want: []token.Token{
				{Kind: token.Open, Raw: "[fg:red]", Namespace: "fg", Field: "red"},
				{Kind: token.Text, Raw: "hi"},
				{Kind: token.Close, Raw: "[/fg]", Namespace: "fg"},
			},
		},
		{
			name:   "UnknownNamespaceTagShape",
			chunks: []string{"[foo:bar]x[/foo]"},
			want: []token.Token{
				{Kind: token.Open, Raw: "[foo:bar]", Namespace: "foo", Field: "bar"},
				{Kind: token.Text, Raw: "x"},
				{Kind: token.Close, Raw: "[/foo]", Namespace: "foo"},
			},
		},
		{
			name:   "FieldRetainsInnerPunctuation",
			chunks: []string{"[fg:rgb(1,2,3)]"},
			want: []token.Token{
				{Kind: token.Open, Raw: "[fg:rgb(1,2,3)]", Namespace: "fg", Field: "rgb(1,2,3)"},
			},
		},
		{
			name:   "BracketWithoutColon",
			chunks: []string{"[not a tag]"},
			want: []token.Token{
				{Kind: token.Text, Raw: "[not a tag]"},
			},
		},
		{
			name:   "EmptyNamespace",
			chunks: []string{"[:red]"},
			want: []token.Token{
				{Kind: token.Text, Raw: "[:red]"},
			},
		},
		{
			name:   "NonLetterCloseNamespace",
			chunks: []string{"[/1]"},
			want: []token.Token{
				{Kind: token.Text, Raw: "[/1]"},
			},
		},
		{
			name:   "TagSplitAcrossChunks",
			chunks: []string{"[fg:", "red]hi"},
			want: []token.Token{
				{Kind: token.Open, Raw: "[fg:red]", Namespace: "fg", Field: "red"},
				{Kind: token.Text, Raw: "hi"},
			},
		},
		{
			name:   "TextThenBracketSplitAcrossChunks",
			chunks: []string{"ab[", "fg:red]"},
			want: []token.Token{
				{Kind: token.Text, Raw: "ab"},
				{Kind: token.Open, Raw: "[fg:red]", Namespace: "fg", Field: "red"},
			},
		},
		{
			name:   "IncompleteTagFlushedAsText",
			chunks: []string{"abc[fg"},
			want: []token.Token{
				{Kind: token.Text, Raw: "abc"},
				{Kind: token.Text, Raw: "[fg"},
			},
		},
		{
			name:   "RawRegionPassesInnerTagsThrough",
			chunks: []string{"[richtext:off]a[fg:red]b[/richtext]"},
			want: []token.Token{
				{Kind: token.Open, Raw: "[richtext:off]", Namespace: "richtext", Field: "off"},
				{Kind: token.Text, Raw: "a[fg:red]b"},
				{Kind: token.Close, Raw: "[/richtext]", Namespace: "richtext"},
			},
		},
		{
			name:   "RawRegionCloseSplitAcrossChunks",
			chunks: []string{"[richtext:off]hi[/richt", "ext]bye"},
			want: []token.Token{
				{Kind: token.Open, Raw: "[richtext:off]", Namespace: "richtext", Field: "off"},
				{Kind: token.Text, Raw: "hi"},
				{Kind: token.Close, Raw: "[/richtext]", Namespace: "richtext"},
				{Kind: token.Text, Raw: "bye"},
			},
		},
		{
			name:   "RawRegionWithNestedTags",
			chunks: []string{"[richtext:off][fg:red]hi[/fg][/richt", "ext]bye"},
			want: []token.Token{
				{Kind: token.Open, Raw: "[richtext:off]", Namespace: "richtext", Field: "off"},
				{Kind: token.Text, Raw: "[fg:red]hi[/fg]"},
				{Kind: token.Close, Raw: "[/richtext]", Namespace: "richtext"},
				{Kind: token.Text, Raw: "bye"},
			},
		},
		{
			name:   "RawRegionUnclosedFlushesText",
			chunks: []string{"[richtext:off]abc"},
			want: []token.Token{
				{Kind: token.Open, Raw: "[richtext:off]", Namespace: "richtext", Field: "off"},
				{Kind: token.Text, Raw: "abc"},
			},
		},
		{
			name:   "RichtextNamespaceOtherFieldParsesAsTag",
			chunks: []string{"[richtext:on]x[/richtext]"},
			want: []token.Token{
				{Kind: token.Open, Raw: "[richtext:on]", Namespace: "richtext", Field: "on"},
				{Kind: token.Text, Raw: "x"},
				{Kind: token.Close, Raw: "[/richtext]", Namespace: "richtext"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sut token.Scanner
			var tokens []token.Token

			// Act
			for _, chunk := range tc.chunks {
				tokens = append(tokens, sut.Scan([]byte(chunk))...)
			}
			if flushed, ok := sut.Flush(); ok {
				tokens = append(tokens, flushed)
			}

			// Assert
			if got, want := tokens, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Scan() = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
		})
	}
}
