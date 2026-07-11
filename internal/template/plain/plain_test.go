package plain_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/internal/template/plain"
	"github.com/bitwizeshift/go-cli/richtext"
)

func TestRender(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		markup  string
		want    string
		wantErr error
	}{
		{
			name:    "StripsThemeTags",
			markup:  "[theme:title]hi[/theme]",
			want:    "hi",
			wantErr: nil,
		},
		{
			name:    "KeepsRawRegionContents",
			markup:  "[richtext:off][x][/richtext]",
			want:    "[x]",
			wantErr: nil,
		},
		{
			name:    "UnbalancedMarkup",
			markup:  "[theme:title]hi",
			want:    "",
			wantErr: richtext.ErrUnclosedTag,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			rendered, err := plain.Render(tc.markup)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Render(%q) = %v, want %v", tc.markup, got, want)
			}
			if got, want := rendered, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Render(%q) = %q, want %q", tc.markup, got, want)
			}
		})
	}
}
