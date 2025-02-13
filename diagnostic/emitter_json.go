package diagnostic

import (
	"encoding/json"
	"io"
	"path/filepath"
)

// jsonEmitter is an [Emitter] that always prints object in JSON format.
type jsonEmitter struct {
	writer io.Writer
}

// Emit emits a diagnostic message in JSON format.
func (e *jsonEmitter) Emit(msg *Diagnostic) error {
	if msg.File != "" {
		if abs, err := filepath.Abs(msg.File); err == nil {
			msg.File = abs
		}
	}
	return json.NewEncoder(e.writer).Encode(msg)
}

var _ Emitter = (*jsonEmitter)(nil)
