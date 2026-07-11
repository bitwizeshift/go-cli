package style_test

import (
	"testing"

	"github.com/bitwizeshift/go-cli/richtext/style"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestStyle_Merge(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		base    style.Style
		overlay style.Style
		want    style.Style
	}{
		{
			name:    "OverlayColourReplacesBase",
			base:    style.Style{Foreground: style.Red},
			overlay: style.Style{Foreground: style.Blue},
			want:    style.Style{Foreground: style.Blue},
		},
		{
			name:    "UnsetOverlayColourKeepsBase",
			base:    style.Style{Foreground: style.Red},
			overlay: style.Style{Background: style.White},
			want:    style.Style{Foreground: style.Red, Background: style.White},
		},
		{
			name:    "AttributesCombine",
			base:    style.Style{Attributes: style.Bold},
			overlay: style.Style{Attributes: style.Italic},
			want:    style.Style{Attributes: style.Bold | style.Italic},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			merged := tc.base.Merge(tc.overlay)

			// Assert
			if got, want := merged, tc.want; !cmp.Equal(got, want, cmpopts.EquateComparable(style.Style{})) {
				t.Errorf("Merge() = %v, want %v", got, want)
			}
		})
	}
}

func TestStyle_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		style style.Style
		want  string
	}{
		{
			name:  "ZeroStyle",
			style: style.Style{},
			want:  "",
		},
		{
			name:  "ForegroundOnly",
			style: style.Style{Foreground: style.Red},
			want:  "\x1b[31m",
		},
		{
			name:  "BackgroundOnly",
			style: style.Style{Background: style.Red},
			want:  "\x1b[41m",
		},
		{
			name:  "AttributesOnly",
			style: style.Style{Attributes: style.Bold | style.Italic},
			want:  "\x1b[1;3m",
		},
		{
			name:  "AttributesBeforeColours",
			style: style.Style{Foreground: style.Green, Background: style.Black, Attributes: style.Bold},
			want:  "\x1b[1;32;40m",
		},
		{
			name:  "TrueColourForeground",
			style: style.Style{Foreground: style.RGB(255, 0, 128)},
			want:  "\x1b[38;2;255;0;128m",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			seq := tc.style.String()

			// Assert
			if got, want := seq, tc.want; !cmp.Equal(got, want) {
				t.Errorf("String() = %q, want %q", got, want)
			}
		})
	}
}
