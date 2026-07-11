// Package panichandler renders a report describing an uncaught panic: the error
// message, the stack trace, and an optional invitation to file an issue.
//
// It exists to give the CLI a readable crash report in place of Go's default
// panic dump.
package panichandler

//go:generate go run golden_gen.go
