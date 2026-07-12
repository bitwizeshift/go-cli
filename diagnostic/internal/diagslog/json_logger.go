package diagslog

import (
	"io"
	"log/slog"
)

// NewJSONLogger returns a Logger that emits each Diagnostic as one JSON line
// to w.
func NewJSONHandler(w io.Writer) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})
}
