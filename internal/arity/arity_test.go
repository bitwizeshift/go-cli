package arity_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/internal/arity"
)

func TestArity_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		arity   arity.Arity
		count   int
		wantErr error
	}{
		{
			name:    "ZeroValuePermitsNoArguments",
			arity:   arity.Arity{},
			count:   0,
			wantErr: nil,
		}, {
			name:    "ZeroValueRejectsArgument",
			arity:   arity.Arity{},
			count:   1,
			wantErr: arity.ErrBadArity,
		}, {
			name:    "ExactCountPermitted",
			arity:   arity.Between(2, 2),
			count:   2,
			wantErr: nil,
		}, {
			name:    "BelowExactCountRejected",
			arity:   arity.Between(2, 2),
			count:   1,
			wantErr: arity.ErrBadArity,
		}, {
			name:    "AboveExactCountRejected",
			arity:   arity.Between(2, 2),
			count:   3,
			wantErr: arity.ErrBadArity,
		}, {
			name:    "RangeLowerBoundPermitted",
			arity:   arity.Between(1, 3),
			count:   1,
			wantErr: nil,
		}, {
			name:    "RangeUpperBoundPermitted",
			arity:   arity.Between(1, 3),
			count:   3,
			wantErr: nil,
		}, {
			name:    "BelowRangeRejected",
			arity:   arity.Between(1, 3),
			count:   0,
			wantErr: arity.ErrBadArity,
		}, {
			name:    "AboveRangeRejected",
			arity:   arity.Between(1, 3),
			count:   4,
			wantErr: arity.ErrBadArity,
		}, {
			name:    "UnboundedPermitsFloor",
			arity:   arity.AtLeast(2),
			count:   2,
			wantErr: nil,
		}, {
			name:    "UnboundedPermitsLargeCount",
			arity:   arity.AtLeast(2),
			count:   1000,
			wantErr: nil,
		}, {
			name:    "UnboundedRejectsBelowFloor",
			arity:   arity.AtLeast(2),
			count:   1,
			wantErr: arity.ErrBadArity,
		}, {
			name:    "UnboundedFromZeroPermitsNone",
			arity:   arity.AtLeast(0),
			count:   0,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.arity

			// Act
			err := sut.Validate(tc.count)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("sut.Validate(%d) = %v, want %v", tc.count, got, want)
			}
		})
	}
}

func TestArity_Validate_Error(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := arity.Between(1, 3)

	// Act
	err := sut.Validate(4)

	// Assert
	if err == nil {
		t.Fatalf("sut.Validate(4) = nil, want %v", arity.ErrBadArity)
	}
	message := err.Error()
	if got, want := strings.Contains(message, sut.String()), true; !cmp.Equal(got, want) {
		t.Errorf("sut.Validate(4) message = %q, want it to describe %q", message, sut)
	}
	if got, want := strings.Contains(message, "4"), true; !cmp.Equal(got, want) {
		t.Errorf("sut.Validate(4) message = %q, want it to report the received count", message)
	}
}

func TestArity_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		arity arity.Arity
		want  string
	}{
		{
			name:  "NoArguments",
			arity: arity.Between(0, 0),
			want:  "no arguments",
		}, {
			name:  "ExactlyOne",
			arity: arity.Between(1, 1),
			want:  "exactly 1 argument",
		}, {
			name:  "ExactlyMany",
			arity: arity.Between(2, 2),
			want:  "exactly 2 arguments",
		}, {
			name:  "AtMost",
			arity: arity.Between(0, 1),
			want:  "at most 1 argument",
		}, {
			name:  "Between",
			arity: arity.Between(1, 3),
			want:  "between 1 and 3 arguments",
		}, {
			name:  "AtLeastOne",
			arity: arity.AtLeast(1),
			want:  "at least 1 argument",
		}, {
			name:  "AtLeastMany",
			arity: arity.AtLeast(2),
			want:  "at least 2 arguments",
		}, {
			name:  "AnyNumber",
			arity: arity.AtLeast(0),
			want:  "any number of arguments",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.arity

			// Act
			description := sut.String()

			// Assert
			if got, want := description, tc.want; !cmp.Equal(got, want) {
				t.Errorf("sut.String() = %q, want %q", got, want)
			}
		})
	}
}

func TestBetween_BadBounds_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		lo   int
		hi   int
		want bool
	}{
		{
			name: "NegativeLowerBound",
			lo:   -1,
			hi:   3,
			want: true,
		}, {
			name: "UpperBoundBelowLower",
			lo:   3,
			hi:   1,
			want: true,
		}, {
			name: "ValidBounds",
			lo:   1,
			hi:   3,
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			call := func() { arity.Between(tc.lo, tc.hi) }

			// Act
			panicked := recovered(call)

			// Assert
			if got, want := panicked, tc.want; !cmp.Equal(got, want) {
				t.Errorf("arity.Between(%d, %d) panicked = %t, want %t", tc.lo, tc.hi, got, want)
			}
		})
	}
}

func TestAtLeast_NegativeBound_Panics(t *testing.T) {
	t.Parallel()

	// Arrange
	call := func() { arity.AtLeast(-1) }

	// Act
	panicked := recovered(call)

	// Assert
	if got, want := panicked, true; !cmp.Equal(got, want) {
		t.Errorf("arity.AtLeast(-1) panicked = %t, want %t", got, want)
	}
}

// recovered reports whether fn panicked.
func recovered(fn func()) (panicked bool) {
	defer func() { panicked = recover() != nil }()
	fn()
	return
}
