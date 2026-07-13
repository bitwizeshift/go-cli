package diagslog

import (
	"io"
	"log/slog"
)

// NewJSONHandler returns a [slog.Handler] that emits each Diagnostic as one JSON line
// to w.
func NewJSONHandler(w io.Writer) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})
}
