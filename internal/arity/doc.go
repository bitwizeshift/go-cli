// Package arity parses a small specification language describing how many
// positional arguments a command may accept.
//
// Specifications are read from textual configuration through
// [Arity.UnmarshalText] or [ArityFunc.UnmarshalText]. See those methods for the
// accepted grammar.
package arity
