package tmplfuncs_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
)

func TestText_Wrap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		columns int
		input   string
		want    string
	}{
		{
			name:    "wraps at word boundary",
			columns: 7,
			input:   "one two three",
			want:    "one two\nthree",
		}, {
			name:    "non-positive width unchanged",
			columns: 0,
			input:   "one two three",
			want:    "one two three",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tmplfuncs.Text{}

			// Act
			wrapped := sut.Wrap(tc.columns, tc.input)

			// Assert
			if got, want := wrapped, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Text.Wrap(%d, %q) = %q, want %q", tc.columns, tc.input, got, want)
			}
		})
	}
}

func TestText_Indent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		n     int
		input string
		want  string
	}{
		{
			name:  "indents every line",
			n:     2,
			input: "first\nsecond",
			want:  "  first\n  second",
		}, {
			name:  "zero is unchanged",
			n:     0,
			input: "first\nsecond",
			want:  "first\nsecond",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tmplfuncs.Text{}

			// Act
			indented := sut.Indent(tc.n, tc.input)

			// Assert
			if got, want := indented, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Text.Indent(%d, %q) = %q, want %q", tc.n, tc.input, got, want)
			}
		})
	}
}

func TestText_IndentLines(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := tmplfuncs.Text{}
	lines := []string{"first", "second"}

	// Act
	indented := sut.IndentLines(2, lines)

	// Assert
	if got, want := indented, "  first\n  second"; !cmp.Equal(got, want) {
		t.Errorf("Text.IndentLines(2, %v) = %q, want %q", lines, got, want)
	}
}

func TestText_Pad(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		width int
		input string
		want  string
	}{
		{
			name:  "pads to width",
			width: 5,
			input: "ab",
			want:  "ab   ",
		}, {
			name:  "wider than width unchanged",
			width: 2,
			input: "abc",
			want:  "abc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tmplfuncs.Text{}

			// Act
			padded := sut.Pad(tc.width, tc.input)

			// Assert
			if got, want := padded, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Text.Pad(%d, %q) = %q, want %q", tc.width, tc.input, got, want)
			}
		})
	}
}

func TestText_Upper(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := tmplfuncs.Text{}

	// Act
	upper := sut.Upper("Build Details")

	// Assert
	if got, want := upper, "BUILD DETAILS"; !cmp.Equal(got, want) {
		t.Errorf("Text.Upper(%q) = %q, want %q", "Build Details", got, want)
	}
}

func TestText_MaxWidth(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		values []string
		want   int
	}{
		{
			name:   "widest string",
			values: []string{"VCS", "VCS Revision", "Target"},
			want:   len("VCS Revision"),
		}, {
			name:   "empty is zero",
			values: nil,
			want:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tmplfuncs.Text{}

			// Act
			width := sut.MaxWidth(tc.values)

			// Assert
			if got, want := width, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Text.MaxWidth(%v) = %d, want %d", tc.values, got, want)
			}
		})
	}
}
