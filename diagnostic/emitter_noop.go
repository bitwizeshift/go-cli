package diagnostic

type noopEmitter struct{}

func (e noopEmitter) Emit(*Diagnostic) error {
	return nil
}

var _ Emitter = (*noopEmitter)(nil)
