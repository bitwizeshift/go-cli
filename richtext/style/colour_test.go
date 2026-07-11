package style_test

import (
	"testing"

	"github.com/bitwizeshift/go-cli/richtext/style"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestColourByName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		input  string
		want   style.Colour
		wantOK bool
	}{
		{
			name:   "StandardColour",
			input:  "red",
			want:   style.Red,
			wantOK: true,
		},
		{
			name:   "BrightColour",
			input:  "brightred",
			want:   style.BrightRed,
			wantOK: true,
		},
		{
			name:   "UnknownColour",
			input:  "chartreuse",
			want:   style.Colour{},
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			colour, ok := style.ColourByName(tc.input)

			// Assert
			if got, want := ok, tc.wantOK; !cmp.Equal(got, want) {
				t.Fatalf("ColourByName() ok = %t, want %t", got, want)
			}
			if got, want := colour, tc.want; !cmp.Equal(got, want, cmpopts.EquateComparable(style.Colour{})) {
				t.Errorf("ColourByName() = %v, want %v", got, want)
			}
		})
	}
}

func TestColour_UnmarshalText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    style.Colour
		wantErr error
	}{
		{
			name:    "NamedColour",
			input:   "green",
			want:    style.Green,
			wantErr: nil,
		},
		{
			name:    "TrueColour",
			input:   "rgb(1,2,3)",
			want:    style.RGB(1, 2, 3),
			wantErr: nil,
		},
		{
			name:    "TrueColourWithSpaces",
			input:   "rgb(10, 20, 30)",
			want:    style.RGB(10, 20, 30),
			wantErr: nil,
		},
		{
			name:    "ComponentOutOfRange",
			input:   "rgb(256,0,0)",
			want:    style.Colour{},
			wantErr: style.ErrUnknownColour,
		},
		{
			name:    "WrongComponentCount",
			input:   "rgb(1,2)",
			want:    style.Colour{},
			wantErr: style.ErrUnknownColour,
		},
		{
			name:    "MissingClosingParen",
			input:   "rgb(1,2,3",
			want:    style.Colour{},
			wantErr: style.ErrUnknownColour,
		},
		{
			name:    "NonNumericComponent",
			input:   "rgb(a,b,c)",
			want:    style.Colour{},
			wantErr: style.ErrUnknownColour,
		},
		{
			name:    "UnknownName",
			input:   "octarine",
			want:    style.Colour{},
			wantErr: style.ErrUnknownColour,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var colour style.Colour

			// Act
			err := colour.UnmarshalText([]byte(tc.input))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("UnmarshalText() = %v, want %v", got, want)
			}
			if got, want := colour, tc.want; !cmp.Equal(got, want, cmpopts.EquateComparable(style.Colour{})) {
				t.Errorf("UnmarshalText() colour = %v, want %v", got, want)
			}
		})
	}
}
