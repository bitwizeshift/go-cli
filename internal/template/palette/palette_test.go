package palette_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/ansi"
	"github.com/bitwizeshift/go-cli/internal/template/palette"
)

func TestColour(t *testing.T) {
	t.Parallel()

	sut := palette.Colour{
		TitleColour:    ansi.Green,
		HeadingColour:  ansi.Yellow,
		LabelColour:    ansi.Cyan,
		ValueColour:    ansi.White,
		EmphasisColour: ansi.Bold,
		ErrorColour:    ansi.Red,
		QuoteColour:    ansi.BrightBlack,
		GutterColour:   ansi.Blue,
		URLColour:      ansi.Magenta,
	}

	testCases := []struct {
		name   string
		method func(string) string
		colour ansi.Colour
	}{
		{
			name:   "Title",
			method: sut.Title,
			colour: ansi.Green,
		}, {
			name:   "Heading",
			method: sut.Heading,
			colour: ansi.Yellow,
		}, {
			name:   "Label",
			method: sut.Label,
			colour: ansi.Cyan,
		}, {
			name:   "Value",
			method: sut.Value,
			colour: ansi.White,
		}, {
			name:   "Emphasis",
			method: sut.Emphasis,
			colour: ansi.Bold,
		}, {
			name:   "Error",
			method: sut.Error,
			colour: ansi.Red,
		}, {
			name:   "Quote",
			method: sut.Quote,
			colour: ansi.BrightBlack,
		}, {
			name:   "Gutter",
			method: sut.Gutter,
			colour: ansi.Blue,
		}, {
			name:   "URL",
			method: sut.URL,
			colour: ansi.Magenta,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			want := string(tc.colour) + "sample" + string(ansi.Reset)

			// Act
			painted := tc.method("sample")

			// Assert
			if got, want := painted, want; !cmp.Equal(got, want) {
				t.Errorf("Colour.%s(sample) = %q, want %q", tc.name, got, want)
			}
		})
	}
}

func TestNoColour(t *testing.T) {
	t.Parallel()

	sut := palette.NoColour{}

	testCases := []struct {
		name   string
		method func(string) string
	}{
		{
			name:   "Title",
			method: sut.Title,
		}, {
			name:   "Heading",
			method: sut.Heading,
		}, {
			name:   "Label",
			method: sut.Label,
		}, {
			name:   "Value",
			method: sut.Value,
		}, {
			name:   "Emphasis",
			method: sut.Emphasis,
		}, {
			name:   "Error",
			method: sut.Error,
		}, {
			name:   "Quote",
			method: sut.Quote,
		}, {
			name:   "Gutter",
			method: sut.Gutter,
		}, {
			name:   "URL",
			method: sut.URL,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			input := "sample"

			// Act
			painted := tc.method(input)

			// Assert
			if got, want := painted, input; !cmp.Equal(got, want) {
				t.Errorf("NoColour.%s(sample) = %q, want %q", tc.name, got, want)
			}
		})
	}
}
