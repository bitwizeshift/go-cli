package arity

import (
	"cmp"
	"encoding"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// ErrBadArity indicates a command received a number of positional arguments
	// outside the range permitted by an [ArityFunc].
	ErrBadArity = errors.New("bad arity")

	// ErrEmpty indicates the specification contained no terms.
	ErrEmpty = errors.New("empty arity specification")

	// ErrSyntax indicates a term was not a recognized arity expression.
	ErrSyntax = errors.New("invalid arity syntax")

	// ErrNegative indicates a term contained a negative bound, which cannot be
	// an argument count.
	ErrNegative = errors.New("negative arity bound")

	// ErrEmptyRange indicates a term described a range containing no values.
	ErrEmptyRange = errors.New("empty arity range")

	// ErrOverlap indicates two terms described overlapping ranges.
	ErrOverlap = errors.New("overlapping arity ranges")
)

// interval is an inclusive range of permitted argument counts. When unbounded
// is set the range has no upper limit and hi is ignored.
type interval struct {
	lo        int
	hi        int
	unbounded bool
}

// contains reports whether n falls within the interval.
func (i interval) contains(n int) bool {
	if n < i.lo {
		return false
	}
	return i.unbounded || n <= i.hi
}

// describe returns a human-readable phrase for the counts the interval permits,
// such as "at least 2 arguments" or "between 1 and 3 arguments".
func (i interval) describe() string {
	switch {
	case i.unbounded && i.lo == 0:
		return "any number of arguments"
	case i.unbounded:
		return "at least " + i.arguments(i.lo)
	case i.lo == 0 && i.hi == 0:
		return "no arguments"
	case i.lo == i.hi:
		return "exactly " + i.arguments(i.lo)
	case i.lo == 0:
		return "at most " + i.arguments(i.hi)
	default:
		return fmt.Sprintf("between %d and %d arguments", i.lo, i.hi)
	}
}

// arguments renders n with the correctly pluralized noun, such as "1 argument"
// or "3 arguments".
func (interval) arguments(n int) string {
	if n == 1 {
		return "1 argument"
	}
	return fmt.Sprintf("%d arguments", n)
}

// Arity describes the set of permitted argument counts parsed from a
// specification.
type Arity struct {
	intervals []interval
}

// Contains reports whether n is a permitted argument count. The zero-value
// [Arity] permits no counts.
func (a Arity) Contains(n int) bool {
	return slices.ContainsFunc(a.intervals, func(i interval) bool {
		return i.contains(n)
	})
}

// String returns a human-readable description of the permitted argument counts,
// such as "at most 1 argument" or, for disjoint ranges, "exactly 1 argument, or
// at least 3 arguments". The zero-value [Arity] returns an empty string.
func (a Arity) String() string {
	phrases := make([]string, len(a.intervals))
	for i := range a.intervals {
		phrases[i] = a.intervals[i].describe()
	}
	return a.joinAlternatives(phrases)
}

var _ fmt.Stringer = Arity{}

// joinAlternatives renders phrases as a natural-language list, separating the
// final phrase with "or".
func (Arity) joinAlternatives(phrases []string) string {
	switch len(phrases) {
	case 0:
		return ""
	case 1:
		return phrases[0]
	default:
		return strings.Join(phrases[:len(phrases)-1], ", ") + ", or " + phrases[len(phrases)-1]
	}
}

// UnmarshalText parses an arity specification and configures the [Arity] to
// permit the described argument counts.
//
// A specification is a comma-separated list of terms. Each term is an exact
// count ("3"), a comparison (">3", ">=3", "<3", "<=3"), or a range with an
// exclusive ("1..3") or inclusive ("1..=3") upper bound. Terms may touch but
// must not overlap.
//
// It returns [ErrEmpty], [ErrSyntax], [ErrNegative], [ErrEmptyRange], or
// [ErrOverlap] when the specification is invalid.
func (a *Arity) UnmarshalText(text []byte) error {
	intervals, err := parse(string(text))
	if err != nil {
		return err
	}
	a.intervals = intervals
	return nil
}

var _ encoding.TextUnmarshaler = (*Arity)(nil)

// ArityFunc validates that a command received a permitted number of positional
// arguments. It satisfies cobra's positional-argument validator signature.
type ArityFunc func(*cobra.Command, []string) error

// UnmarshalText parses an arity specification, using the same grammar and
// returning the same errors as [Arity.UnmarshalText]. On success it configures
// the [ArityFunc] to report a wrapped [ErrBadArity] when a command receives an
// argument count outside the permitted range.
func (af *ArityFunc) UnmarshalText(text []byte) error {
	var a Arity
	if err := a.UnmarshalText(text); err != nil {
		return err
	}
	*af = func(_ *cobra.Command, args []string) error {
		if !a.Contains(len(args)) {
			return &badArityError{arity: a, got: len(args)}
		}
		return nil
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*ArityFunc)(nil)

// badArityError reports that a command received a number of positional
// arguments outside the range permitted by its [Arity].
type badArityError struct {
	arity Arity
	got   int
}

// Error describes the permitted argument counts alongside the number received.
func (e *badArityError) Error() string {
	return fmt.Sprintf("accepts %s, but received %d", e.arity, e.got)
}

// Unwrap returns [ErrBadArity].
func (e *badArityError) Unwrap() error {
	return ErrBadArity
}

var _ error = (*badArityError)(nil)

// parse converts a specification into the sorted, non-overlapping intervals it
// describes.
func parse(spec string) ([]interval, error) {
	if strings.TrimSpace(spec) == "" {
		return nil, ErrEmpty
	}
	terms := strings.Split(spec, ",")
	intervals := make([]interval, 0, len(terms))
	for _, term := range terms {
		i, err := parseTerm(strings.TrimSpace(term))
		if err != nil {
			return nil, err
		}
		intervals = append(intervals, i)
	}
	slices.SortFunc(intervals, func(lhs, rhs interval) int {
		return cmp.Compare(lhs.lo, rhs.lo)
	})
	if err := checkOverlap(intervals); err != nil {
		return nil, err
	}
	return intervals, nil
}

// parseTerm converts a single trimmed term into its interval.
func parseTerm(term string) (interval, error) {
	switch {
	case term == "":
		return interval{}, fmt.Errorf("%w: empty term", ErrSyntax)
	case strings.Contains(term, ".."):
		return parseRange(term)
	case strings.HasPrefix(term, ">="):
		n, err := parseInt(term[2:])
		if err != nil {
			return interval{}, err
		}
		return interval{lo: n, unbounded: true}, nil
	case strings.HasPrefix(term, "<="):
		n, err := parseInt(term[2:])
		if err != nil {
			return interval{}, err
		}
		return interval{lo: 0, hi: n}, nil
	case strings.HasPrefix(term, ">"):
		n, err := parseInt(term[1:])
		if err != nil {
			return interval{}, err
		}
		return interval{lo: n + 1, unbounded: true}, nil
	case strings.HasPrefix(term, "<"):
		n, err := parseInt(term[1:])
		if err != nil {
			return interval{}, err
		}
		if n == 0 {
			return interval{}, fmt.Errorf("%w: %q", ErrEmptyRange, term)
		}
		return interval{lo: 0, hi: n - 1}, nil
	default:
		n, err := parseInt(term)
		if err != nil {
			return interval{}, err
		}
		return interval{lo: n, hi: n}, nil
	}
}

// parseRange converts a "lo..hi" or "lo..=hi" term into its interval.
func parseRange(term string) (interval, error) {
	inclusive := false
	lhs, rhs, ok := strings.Cut(term, "..=")
	if ok {
		inclusive = true
	} else {
		lhs, rhs, _ = strings.Cut(term, "..")
	}
	lo, err := parseInt(lhs)
	if err != nil {
		return interval{}, err
	}
	hi, err := parseInt(rhs)
	if err != nil {
		return interval{}, err
	}
	if !inclusive {
		hi--
	}
	if hi < lo {
		return interval{}, fmt.Errorf("%w: %q", ErrEmptyRange, term)
	}
	return interval{lo: lo, hi: hi}, nil
}

// parseInt parses a non-negative integer, rejecting negatives and any
// surrounding whitespace.
func parseInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("%w: empty number", ErrSyntax)
	}
	if strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("%w: %q", ErrNegative, s)
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%w: %q", ErrSyntax, s)
	}
	return n, nil
}

// checkOverlap reports an [ErrOverlap] when any two of the lo-sorted intervals
// share a value.
func checkOverlap(intervals []interval) error {
	for i := 1; i < len(intervals); i++ {
		prev, cur := intervals[i-1], intervals[i]
		if prev.unbounded || cur.lo <= prev.hi {
			return fmt.Errorf("%w: ranges starting at %d and %d", ErrOverlap, prev.lo, cur.lo)
		}
	}
	return nil
}
