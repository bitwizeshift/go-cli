package format_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/format"
)

func TestTruncate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		text  string
		width int
		want  string
	}{
		{
			name:  "ZeroWidthReturnsInputUnchanged",
			text:  "hello",
			width: 0,
			want:  "hello",
		}, {
			name:  "NegativeWidthReturnsInputUnchanged",
			text:  "hello",
			width: -3,
			want:  "hello",
		}, {
			name:  "FitsUnchanged",
			text:  "hello",
			width: 10,
			want:  "hello",
		}, {
			name:  "ExactWidthUnchanged",
			text:  "hello",
			width: 5,
			want:  "hello",
		}, {
			name:  "CutAppendsEllipsis",
			text:  "hello world",
			width: 8,
			want:  "hello w…",
		}, {
			name:  "WidthOneIsEllipsisOnly",
			text:  "hello",
			width: 1,
			want:  "…",
		}, {
			name:  "CountsRunesNotBytes",
			text:  "héllo wörld",
			width: 8,
			want:  "héllo w…",
		}, {
			name:  "MultibyteWithinLimitUnchanged",
			text:  "café",
			width: 4,
			want:  "café",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			text := tc.text
			width := tc.width

			// Act
			truncated := format.Truncate(text, width)

			// Assert
			if got, want := truncated, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Truncate(%q, %d) = %q, want %q", text, width, got, want)
			}
		})
	}
}
