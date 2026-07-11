// Package help renders group-organised help output for a
// [github.com/spf13/cobra.Command].
//
// It exists because cobra's built-in help cannot style output or present flags
// and subcommands by group.
package help

//go:generate go run golden_gen.go
