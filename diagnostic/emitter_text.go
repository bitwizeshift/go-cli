package diagnostic

import "io"

type textEmitter struct {
	writer io.Writer
}

func (e *textEmitter) Emit(msg *Diagnostic) error {
	_, err := e.writer.Write([]byte(msg.Message))
	return err
}
