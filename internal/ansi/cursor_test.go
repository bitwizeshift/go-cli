package ansi_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/ansi"
)

func TestCursorUp(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		n    int
		want string
	}{
		{
			name: "NegativeReturnsEmpty",
			n:    -3,
			want: "",
		}, {
			name: "ZeroReturnsEmpty",
			n:    0,
			want: "",
		}, {
			name: "OneLine",
			n:    1,
			want: "\x1b[1A",
		}, {
			name: "ManyLines",
			n:    5,
			want: "\x1b[5A",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			n := tc.n

			// Act
			sequence := ansi.CursorUp(n)

			// Assert
			if got, want := sequence, tc.want; !cmp.Equal(got, want) {
				t.Errorf("CursorUp(%d) = %q, want %q", n, got, want)
			}
		})
	}
}
