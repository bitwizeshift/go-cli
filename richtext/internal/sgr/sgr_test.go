package sgr_test

import (
	"testing"

	"github.com/bitwizeshift/go-cli/richtext/internal/sgr"
	"github.com/google/go-cmp/cmp"
)

func TestSequence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		params []int
		want   string
	}{
		{
			name:   "NoParameters",
			params: nil,
			want:   "",
		},
		{
			name:   "SingleParameter",
			params: []int{1},
			want:   "\x1b[1m",
		},
		{
			name:   "MultipleParameters",
			params: []int{1, 3, 31},
			want:   "\x1b[1;3;31m",
		},
		{
			name:   "TrueColourParameters",
			params: []int{38, 2, 255, 0, 128},
			want:   "\x1b[38;2;255;0;128m",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			seq := sgr.Sequence(tc.params...)

			// Assert
			if got, want := seq, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Sequence() = %q, want %q", got, want)
			}
		})
	}
}

func TestTrueColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		background bool
		r, g, b    uint8
		want       []int
	}{
		{
			name:       "Foreground",
			background: false,
			r:          10,
			g:          20,
			b:          30,
			want:       []int{38, 2, 10, 20, 30},
		},
		{
			name:       "Background",
			background: true,
			r:          10,
			g:          20,
			b:          30,
			want:       []int{48, 2, 10, 20, 30},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			params := sgr.TrueColour(tc.background, tc.r, tc.g, tc.b)

			// Assert
			if got, want := params, tc.want; !cmp.Equal(got, want) {
				t.Errorf("TrueColour() = %v, want %v", got, want)
			}
		})
	}
}

func TestBackground(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		foreground int
		want       int
	}{
		{
			name:       "Standard",
			foreground: 31,
			want:       41,
		},
		{
			name:       "Bright",
			foreground: 91,
			want:       101,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			code := sgr.Background(tc.foreground)

			// Assert
			if got, want := code, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Background() = %d, want %d", got, want)
			}
		})
	}
}
