// Package usage renders the short advisory shown when a command is used
// incorrectly.
//
// It exists so that a usage error prints a brief, optionally coloured pointer to
// --help rather than cobra's full usage dump. Callers configure a [Renderer]
// with a palette and render the advisory for a command to an [io.Writer].
package usage

//go:generate go run golden_gen.go
