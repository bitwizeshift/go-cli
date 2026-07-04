// Package panichandler renders a report describing an uncaught panic: the error
// message, the stack trace, and an optional invitation to file an issue.
//
// It exists to give the CLI a readable, optionally coloured crash report in
// place of Go's default panic dump. Callers configure a [Renderer] with a
// palette and render a [PanicContext] to an [io.Writer].
//
// Colour is optional: a plain palette produces deterministic plain text, which
// is what the golden test compares against, while a colour palette applies an
// ANSI scheme.
package panichandler

//go:generate go run golden_gen.go
