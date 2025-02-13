package diagnostic

import "log"

// logEmitter is an [Emitter] that always prints object in log format.
type logEmitter struct {
	logger *log.Logger
}

// Emit emits a diagnostic message in log format.
func (e *logEmitter) Emit(msg *Diagnostic) error {
	e.logger.Printf("%v\t%s", msg.Severity, msg.Message)
	return nil
}

var _ Emitter = (*logEmitter)(nil)
