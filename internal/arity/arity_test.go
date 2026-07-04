package arity_test

import (
	"encoding"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/internal/arity"
)

// mustUnmarshal parses spec into u, failing the test on error.
func mustUnmarshal(t *testing.T, u encoding.TextUnmarshaler, spec string) {
	t.Helper()
	if err := u.UnmarshalText([]byte(spec)); err != nil {
		t.Fatalf("UnmarshalText(%q) = %v, want nil", spec, err)
	}
}

func TestArity_UnmarshalText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		spec    string
		wantErr error
	}{
		{
			name:    "exact",
			spec:    "3",
			wantErr: nil,
		},
		{
			name:    "greater than",
			spec:    ">3",
			wantErr: nil,
		},
		{
			name:    "greater or equal",
			spec:    ">=3",
			wantErr: nil,
		},
		{
			name:    "less than",
			spec:    "<3",
			wantErr: nil,
		},
		{
			name:    "less or equal",
			spec:    "<=3",
			wantErr: nil,
		},
		{
			name:    "exclusive range",
			spec:    "1..3",
			wantErr: nil,
		},
		{
			name:    "inclusive range",
			spec:    "1..=3",
			wantErr: nil,
		},
		{
			name:    "zero",
			spec:    "0",
			wantErr: nil,
		},
		{
			name:    "less or equal zero",
			spec:    "<=0",
			wantErr: nil,
		},
		{
			name:    "greater than zero",
			spec:    ">0",
			wantErr: nil,
		},
		{
			name:    "multiple ranges",
			spec:    "<2, >3",
			wantErr: nil,
		},
		{
			name:    "touching ranges",
			spec:    "<=2, >=3",
			wantErr: nil,
		},
		{
			name:    "ranges with gap",
			spec:    "<2, >=3",
			wantErr: nil,
		},
		{
			name:    "empty",
			spec:    "",
			wantErr: arity.ErrEmpty,
		},
		{
			name:    "whitespace only",
			spec:    "   ",
			wantErr: arity.ErrEmpty,
		},
		{
			name:    "empty term",
			spec:    "3,",
			wantErr: arity.ErrSyntax,
		},
		{
			name:    "bad greater or equal",
			spec:    ">=abc",
			wantErr: arity.ErrSyntax,
		},
		{
			name:    "bad less or equal",
			spec:    "<=abc",
			wantErr: arity.ErrSyntax,
		},
		{
			name:    "bad greater than",
			spec:    ">abc",
			wantErr: arity.ErrSyntax,
		},
		{
			name:    "bad less than",
			spec:    "<abc",
			wantErr: arity.ErrSyntax,
		},
		{
			name:    "bad exact",
			spec:    "abc",
			wantErr: arity.ErrSyntax,
		},
		{
			name:    "bad range lower",
			spec:    "abc..3",
			wantErr: arity.ErrSyntax,
		},
		{
			name:    "bad range upper",
			spec:    "1..abc",
			wantErr: arity.ErrSyntax,
		},
		{
			name:    "empty range lower",
			spec:    "..3",
			wantErr: arity.ErrSyntax,
		},
		{
			name:    "negative bound",
			spec:    ">=-1",
			wantErr: arity.ErrNegative,
		},
		{
			name:    "less than zero",
			spec:    "<0",
			wantErr: arity.ErrEmptyRange,
		},
		{
			name:    "reversed exclusive range",
			spec:    "3..1",
			wantErr: arity.ErrEmptyRange,
		},
		{
			name:    "reversed inclusive range",
			spec:    "3..=1",
			wantErr: arity.ErrEmptyRange,
		},
		{
			name:    "empty exclusive range",
			spec:    "1..1",
			wantErr: arity.ErrEmptyRange,
		},
		{
			name:    "overlapping bounded ranges",
			spec:    "<=3, >=2",
			wantErr: arity.ErrOverlap,
		},
		{
			name:    "overlapping unbounded range",
			spec:    ">2, 5",
			wantErr: arity.ErrOverlap,
		},
		{
			name:    "internal whitespace",
			spec:    "< 2",
			wantErr: arity.ErrSyntax,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sut arity.Arity

			// Act
			err := sut.UnmarshalText([]byte(tc.spec))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Arity.UnmarshalText(%q) = %v, want %v", tc.spec, got, want)
			}
		})
	}
}

func TestArity_Contains(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		spec string
		n    int
		want bool
	}{
		{
			name: "exact match",
			spec: "3",
			n:    3,
			want: true,
		},
		{
			name: "exact below",
			spec: "3",
			n:    2,
			want: false,
		},
		{
			name: "exact above",
			spec: "3",
			n:    4,
			want: false,
		},
		{
			name: "greater than excludes bound",
			spec: ">3",
			n:    3,
			want: false,
		},
		{
			name: "greater than includes above",
			spec: ">3",
			n:    4,
			want: true,
		},
		{
			name: "greater or equal includes bound",
			spec: ">=3",
			n:    3,
			want: true,
		},
		{
			name: "less than excludes bound",
			spec: "<3",
			n:    3,
			want: false,
		},
		{
			name: "less than includes below",
			spec: "<3",
			n:    2,
			want: true,
		},
		{
			name: "less or equal includes bound",
			spec: "<=3",
			n:    3,
			want: true,
		},
		{
			name: "exclusive range excludes upper",
			spec: "1..3",
			n:    3,
			want: false,
		},
		{
			name: "exclusive range includes interior",
			spec: "1..3",
			n:    2,
			want: true,
		},
		{
			name: "exclusive range includes lower",
			spec: "1..3",
			n:    1,
			want: true,
		},
		{
			name: "exclusive range below lower",
			spec: "1..3",
			n:    0,
			want: false,
		},
		{
			name: "inclusive range includes upper",
			spec: "1..=3",
			n:    3,
			want: true,
		},
		{
			name: "first of multiple ranges",
			spec: "<2, >3",
			n:    1,
			want: true,
		},
		{
			name: "second of multiple ranges",
			spec: "<2, >3",
			n:    4,
			want: true,
		},
		{
			name: "between multiple ranges",
			spec: "<2, >3",
			n:    3,
			want: false,
		},
		{
			name: "negative count",
			spec: ">=0",
			n:    -1,
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sut arity.Arity
			mustUnmarshal(t, &sut, tc.spec)

			// Act
			ok := sut.Contains(tc.n)

			// Assert
			if got, want := ok, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Arity.Contains(%d) = %v, want %v", tc.n, got, want)
			}
		})
	}
}

func TestArity_Contains_ZeroValue_ReturnsFalse(t *testing.T) {
	t.Parallel()

	// Arrange
	var sut arity.Arity

	// Act
	ok := sut.Contains(0)

	// Assert
	if got, want := ok, false; !cmp.Equal(got, want) {
		t.Errorf("Arity.Contains(0) = %v, want %v", got, want)
	}
}

func TestArityFunc_UnmarshalText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		spec    string
		wantErr error
	}{
		{
			name:    "valid",
			spec:    "1..=3",
			wantErr: nil,
		},
		{
			name:    "invalid",
			spec:    "abc",
			wantErr: arity.ErrSyntax,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sut arity.ArityFunc

			// Act
			err := sut.UnmarshalText([]byte(tc.spec))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("ArityFunc.UnmarshalText(%q) = %v, want %v", tc.spec, got, want)
			}
		})
	}
}

func TestArityFunc(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		spec    string
		args    []string
		wantErr error
	}{
		{
			name:    "within range",
			spec:    "1..=3",
			args:    []string{"a", "b"},
			wantErr: nil,
		},
		{
			name:    "below range",
			spec:    "1..=3",
			args:    []string{},
			wantErr: arity.ErrBadArity,
		},
		{
			name:    "above range",
			spec:    "1..=3",
			args:    []string{"a", "b", "c", "d"},
			wantErr: arity.ErrBadArity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sut arity.ArityFunc
			mustUnmarshal(t, &sut, tc.spec)

			// Act
			err := sut(nil, tc.args)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("ArityFunc(%v) = %v, want %v", tc.args, got, want)
			}
		})
	}
}
