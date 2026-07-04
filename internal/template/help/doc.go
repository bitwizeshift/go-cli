// Package help renders coloured, group-organised help output for a
// [github.com/spf13/cobra.Command].
//
// It exists because cobra's built-in help cannot colourise output or present
// flags and subcommands by group. Callers configure a [Renderer] with a column
// width and a palette, then render the help for a command to an [io.Writer].
package help

//go:generate go run golden_gen.go
