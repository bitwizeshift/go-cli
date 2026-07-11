package style_test

import (
	"testing"

	"github.com/bitwizeshift/go-cli/richtext/style"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestAttributeByName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		input  string
		want   style.Attribute
		wantOK bool
	}{
		{
			name:   "Bold",
			input:  "bold",
			want:   style.Bold,
			wantOK: true,
		},
		{
			name:   "Reverse",
			input:  "reverse",
			want:   style.Reverse,
			wantOK: true,
		},
		{
			name:   "Unknown",
			input:  "sparkle",
			want:   0,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			attr, ok := style.AttributeByName(tc.input)

			// Assert
			if got, want := ok, tc.wantOK; !cmp.Equal(got, want) {
				t.Fatalf("AttributeByName() ok = %t, want %t", got, want)
			}
			if got, want := attr, tc.want; !cmp.Equal(got, want) {
				t.Errorf("AttributeByName() = %d, want %d", got, want)
			}
		})
	}
}

func TestAttribute_UnmarshalText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    style.Attribute
		wantErr error
	}{
		{
			name:    "KnownAttribute",
			input:   "italic",
			want:    style.Italic,
			wantErr: nil,
		},
		{
			name:    "UnknownAttribute",
			input:   "sparkle",
			want:    0,
			wantErr: style.ErrUnknownAttribute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var attr style.Attribute

			// Act
			err := attr.UnmarshalText([]byte(tc.input))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("UnmarshalText() = %v, want %v", got, want)
			}
			if got, want := attr, tc.want; !cmp.Equal(got, want) {
				t.Errorf("UnmarshalText() attr = %d, want %d", got, want)
			}
		})
	}
}
