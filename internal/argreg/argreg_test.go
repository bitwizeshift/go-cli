package argreg_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/internal/argreg"
)

func TestBind(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		indices       []int
		unmatched     bool
		args          []string
		wantBound     map[int]string
		wantUnmatched []string
	}{
		{
			name:      "BindsPositionalByIndex",
			indices:   []int{0, 1},
			args:      []string{"alpha", "beta"},
			wantBound: map[int]string{0: "alpha", 1: "beta"},
		}, {
			name:      "SkipsOutOfRangeIndex",
			indices:   []int{0, 2},
			args:      []string{"alpha"},
			wantBound: map[int]string{0: "alpha"},
		}, {
			name:          "UnmatchedCollectsAllWhenNoPositionals",
			unmatched:     true,
			args:          []string{"a", "b", "c"},
			wantUnmatched: []string{"a", "b", "c"},
		}, {
			name:          "UnmatchedExcludesClaimedPreservingOrder",
			indices:       []int{0, 2},
			unmatched:     true,
			args:          []string{"a", "b", "c", "d"},
			wantBound:     map[int]string{0: "a", 2: "c"},
			wantUnmatched: []string{"b", "d"},
		}, {
			name:          "UnmatchedEmptyWhenAllClaimed",
			indices:       []int{0},
			unmatched:     true,
			args:          []string{"a"},
			wantBound:     map[int]string{0: "a"},
			wantUnmatched: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argreg.New()
			bound := map[int]string{}
			for _, index := range tc.indices {
				argreg.AddPositional(cl, &argreg.Positional{
					Index: index,
					Set:   func(value string) error { bound[index] = value; return nil },
				})
			}
			var rest []string
			if tc.unmatched {
				argreg.SetUnmatched(cl, &argreg.Unmatched{
					Set: func(values []string) error { rest = values; return nil },
				})
			}

			// Act
			err := argreg.Bind(cl, tc.args)

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Bind(...) = %v, want %v", got, want)
			}
			if got, want := bound, tc.wantBound; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Bind(...) bound = %v, want %v", got, want)
			}
			if got, want := rest, tc.wantUnmatched; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Bind(...) unmatched = %v, want %v", got, want)
			}
		})
	}
}

func TestBind_PositionalError(t *testing.T) {
	t.Parallel()

	// Arrange
	testErr := errors.New("boom")
	cl := argreg.New()
	argreg.AddPositional(cl, &argreg.Positional{
		Index: 0,
		Set:   func(string) error { return testErr },
	})

	// Act
	err := argreg.Bind(cl, []string{"x"})

	// Assert
	if got, want := err, testErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
}

func TestBind_UnmatchedError(t *testing.T) {
	t.Parallel()

	// Arrange
	testErr := errors.New("boom")
	cl := argreg.New()
	argreg.SetUnmatched(cl, &argreg.Unmatched{
		Set: func([]string) error { return testErr },
	})

	// Act
	err := argreg.Bind(cl, []string{"x"})

	// Assert
	if got, want := err, testErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
}
