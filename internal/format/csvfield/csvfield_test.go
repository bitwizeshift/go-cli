package csvfield_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/internal/format/csvfield"
)

func TestSplit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		value   string
		want    []string
		wantErr error
	}{
		{
			name:    "EmptyValueYieldsNoFields",
			value:   "",
			want:    nil,
			wantErr: nil,
		}, {
			name:    "SingleField",
			value:   "alpha",
			want:    []string{"alpha"},
			wantErr: nil,
		}, {
			name:    "SeparatesOnComma",
			value:   "alpha,beta,gamma",
			want:    []string{"alpha", "beta", "gamma"},
			wantErr: nil,
		}, {
			name:    "QuotedFieldRetainsComma",
			value:   `"alpha,beta",gamma`,
			want:    []string{"alpha,beta", "gamma"},
			wantErr: nil,
		}, {
			name:    "EmptyFieldsPreserved",
			value:   "alpha,,gamma",
			want:    []string{"alpha", "", "gamma"},
			wantErr: nil,
		}, {
			name:    "UnterminatedQuoteReportsError",
			value:   `"alpha`,
			want:    nil,
			wantErr: cmpopts.AnyError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			fields, err := csvfield.Split(tc.value)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Split(...) = %v, want %v", got, want)
			}
			if got, want := fields, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Split(...) = %v, want %v", got, want)
			}
		})
	}
}
