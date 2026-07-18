// Package arg simplifies command-line registration by wrapping
// [github.com/spf13/pflag] behind a small, type-driven surface.
//
// Rather than exposing a distinct registration call per value type, the package
// decodes argument values through [Unmarshal] so that any supported type -- and
// any type that decodes itself -- can be bound with a single entrypoint. Flags
// are registered with [AddFlag]; positional and unmatched arguments with
// [Positional] and [Unmatched]. On top of registration the package layers
// grouping, requirement constraints, shell completion, and recursive
// registration of argument-bearing values.
package arg
