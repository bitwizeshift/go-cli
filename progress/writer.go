package progress

import (
	"io"
	"sync"

	"github.com/bitwizeshift/go-cli/internal/redraw"
)

// Writer renders a [Renderer] in place on W, redrawing it on each
// [Writer.Update]. When animation is disabled — the [ansi.Enabler] reports
// false, as it does when W is not a terminal — intermediate frames are
// suppressed and only the final state is emitted by [Writer.Done].
type Writer struct {
	// W is the destination. It must be set.
	W io.Writer

	// DisableAnimation can be specified to disable animating over the same
	// line. This is generally necessary for
	DisableAnimation bool

	mu      sync.Mutex
	redraw  *redraw.Writer
	ready   bool
	enabled bool
	last    Renderer
}

// Update redraws r. When animation is disabled it records r without drawing, so
// that [Writer.Done] can emit the final frame. It returns any error from the
// underlying writer.
func (w *Writer) Update(r Renderer) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.init()

	w.last = r
	if !w.enabled {
		return nil
	}
	return w.redraw.Draw(r.Render())
}

// Done finishes the animation. When animation was disabled it draws the final
// recorded frame; either way it terminates the block with a newline. It returns
// any error from the underlying writer.
func (w *Writer) Done() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.init()

	if !w.enabled && w.last != nil {
		if err := w.redraw.Draw(w.last.Render()); err != nil {
			return err
		}
	}
	return w.redraw.Flush()
}

// init lazily resolves the redraw target and the animation decision, so a bare
// Writer{W: ...} literal is usable.
func (w *Writer) init() {
	if w.ready {
		return
	}
	w.redraw = redraw.NewWriter(w.W)
	w.enabled = !w.DisableAnimation
	w.ready = true
}
