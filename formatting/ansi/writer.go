package ansi

import (
	"io"
	"regexp"
	"strconv"

	"golang.org/x/term"
)

// CanFormat returns true if the writer is a terminal that supports ANSI escape
// codes.
//
// If the environment variable NO_COLOR or NOCOLOR is set to a boolean-truthy
// value (e.g. a value for which [strconv.ParseBool] returns true), then this
// function will return false.
//
// To force this function to return true for a non-terminal writer, or when
// NO_COLOR/NOCOLOR is set, use [EnableFormat].
func CanFormat(w io.Writer) bool {
	if _, ok := w.(alwaysFormatter); ok {
		return true
	}
	if envIsTrue("NO_COLOR") || envIsTrue("NOCOLOR") {
		return false
	}
	fd, ok := w.(interface{ Fd() uintptr })
	if !ok {
		return false
	}
	return term.IsTerminal(int(fd.Fd()))
}

func envIsTrue(name string) bool {
	// b is false if parsing fails
	b, _ := strconv.ParseBool(name)
	return b
}

// NewWriter returns a writer that formats the output with ANSI escape codes
// if the underlying writer is a terminal that supports them; otherwise, it
// returns a writer that strips the codes.
//
// This is commonly used to enable common codepaths to be ansi-marked-up, but
// still work correctly when the output is redirected to a file or another
// non-terminal destination.
func NewWriter(w io.Writer) io.Writer {
	if CanFormat(w) {
		return w
	}
	// prevent double-wrapping of noFormatWriters.
	if _, ok := w.(noFormatter); ok {
		return w
	}
	return DisableFormat(w)
}

// EnableFormat returns a writer that formats the output with ANSI escape codes,
// even if the underlying writer is not a terminal.
//
// This will ensure that the writer, when used with [NewWriter] or [CanFormat],
// will always be treated as a terminal that supports ANSI escape codes.
func EnableFormat(w io.Writer) io.Writer {
	return &alwaysFormatWriter{W: w}
}

// DisableFormat returns a writer that strips ANSI escape codes from the output.
func DisableFormat(w io.Writer) io.Writer {
	return &noFormatWriter{W: w}
}

type alwaysFormatter interface {
	alwaysFormat()
}

type noFormatter interface {
	noFormat()
}

type alwaysFormatWriter struct {
	W io.Writer
}

func (w *alwaysFormatWriter) Write(p []byte) (n int, err error) {
	return w.W.Write(p)
}

func (w *alwaysFormatWriter) alwaysFormat() {}

var _ alwaysFormatter = (*alwaysFormatWriter)(nil)

// noFormatWriter is a writer that strips ANSI escape codes from the output.
type noFormatWriter struct {
	W io.Writer
}

// Write writes p to the underlying writer, stripping ANSI escape sequences.
//
// Note: If the sequence is not written in one call, the output may be
// corrupted; though this is unlikely, since ANSI escape sequences aren't
// generally split across multiple writes.
func (w *noFormatWriter) Write(p []byte) (n int, err error) {
	return w.W.Write(stripFormat.ReplaceAll(p, nil))
}

func (w *noFormatWriter) noFormat() {}

var _ noFormatter = (*noFormatWriter)(nil)

// stripFormat is a regular expression that matches ANSI escape codes.
var stripFormat = regexp.MustCompile(`\033\[[0-9;]*m`)
