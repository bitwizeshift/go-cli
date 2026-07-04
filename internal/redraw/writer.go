package redraw

import (
	"io"
	"strings"

	"github.com/bitwizeshift/go-cli/internal/ansi"
)

// Writer redraws a block of text in place on an underlying [io.Writer]. Each
// [Writer.Draw] erases the lines written by the previous Draw and replaces
// them, enabling live-updating output.
type Writer struct {
	w     io.Writer
	lines int
}

// NewWriter returns a [Writer] that draws to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Draw erases the block written by the previous Draw and writes s in its place.
// s may span multiple lines and should not end in a newline. It returns any
// error from the underlying writer.
func (w *Writer) Draw(s string) error {
	if err := w.eraseCurrent(); err != nil {
		return err
	}
	if err := w.write(s); err != nil {
		return err
	}
	w.lines = strings.Count(s, "\n") + 1
	return nil
}

// Flush terminates the current block with a newline and forgets it, leaving the
// block on screen so the next Draw starts fresh. It returns any error from the
// underlying writer.
func (w *Writer) Flush() error {
	if w.lines == 0 {
		return nil
	}
	w.lines = 0
	return w.write("\n")
}

// Clear erases the current block and forgets it. It returns any error from the
// underlying writer.
func (w *Writer) Clear() error {
	return w.eraseCurrent()
}

// eraseCurrent moves the cursor to the start of the current block and erases
// it, resetting the tracked line count.
func (w *Writer) eraseCurrent() error {
	if w.lines == 0 {
		return nil
	}
	up := ansi.CursorUp(w.lines - 1)
	w.lines = 0
	return w.write("\r" + up + string(ansi.ClearDown))
}

func (w *Writer) write(s string) error {
	_, err := io.WriteString(w.w, s)
	return err
}
