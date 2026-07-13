// Package flag simplifies command-line flag registration by wrapping
// [github.com/spf13/pflag] behind a small, type-driven surface.
//
// Rather than exposing a distinct registration call per value type, the package
// decodes flag values through [Unmarshal] so that any supported type -- and any
// type that decodes itself -- can be bound to a flag with a single entrypoint.
// It additionally layers on grouping, requirement constraints, shell completion,
// and recursive registration of flag-bearing values.
package flag
