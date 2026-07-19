package arity

import (
	"errors"
	"fmt"
)

// ErrBadArity indicates a command received a number of positional arguments
// outside the range its [Arity] permits.
var ErrBadArity = errors.New("bad arity")

// Arity is an inclusive range of positional-argument counts a command permits.
// When unbounded is set the range has no upper limit and hi is ignored.
//
// The zero-value permits no arguments at all.
type Arity struct {
	lo        int
	hi        int
	unbounded bool
}

// Between returns an [Arity] permitting between lo and hi arguments inclusive.
// It panics if lo is negative or greater than hi.
func Between(lo, hi int) Arity {
	if lo < 0 {
		panic("arity: negative lower bound")
	}
	if hi < lo {
		panic("arity: upper bound below lower bound")
	}
	return Arity{lo: lo, hi: hi}
}

// AtLeast returns an [Arity] permitting lo or more arguments. It panics if lo
// is negative.
func AtLeast(lo int) Arity {
	if lo < 0 {
		panic("arity: negative lower bound")
	}
	return Arity{lo: lo, unbounded: true}
}

// Validate reports whether n is a permitted argument count, returning a wrapped
// [ErrBadArity] describing the permitted counts when it is not.
func (a Arity) Validate(n int) error {
	if a.contains(n) {
		return nil
	}
	return &badArityError{arity: a, got: n}
}

// String returns a human-readable description of the permitted argument counts,
// such as "at most 1 argument" or "between 1 and 3 arguments".
func (a Arity) String() string {
	switch {
	case a.unbounded && a.lo == 0:
		return "any number of arguments"
	case a.unbounded:
		return "at least " + a.arguments(a.lo)
	case a.lo == 0 && a.hi == 0:
		return "no arguments"
	case a.lo == a.hi:
		return "exactly " + a.arguments(a.lo)
	case a.lo == 0:
		return "at most " + a.arguments(a.hi)
	default:
		return fmt.Sprintf("between %d and %d arguments", a.lo, a.hi)
	}
}

var _ fmt.Stringer = Arity{}

// contains reports whether n falls within the permitted range.
func (a Arity) contains(n int) bool {
	if n < a.lo {
		return false
	}
	return a.unbounded || n <= a.hi
}

// arguments renders n with the correctly pluralized noun, such as "1 argument"
// or "3 arguments".
func (Arity) arguments(n int) string {
	if n == 1 {
		return "1 argument"
	}
	return fmt.Sprintf("%d arguments", n)
}

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
