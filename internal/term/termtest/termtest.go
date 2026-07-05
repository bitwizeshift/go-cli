package termtest

import (
	"io"

	"github.com/bitwizeshift/go-cli/internal/term"
)

type echoDisablerFunc func(w io.Writer) (term.RestoreFunc, error)

func (f echoDisablerFunc) DisableEcho(w io.Writer) (term.RestoreFunc, error) {
	return f(w)
}

var _ term.EchoDisabler = (*echoDisablerFunc)(nil)

// Recorder is an [term.EchoDisabler] that tracks whether echo is currently
// disabled, modelling the echo state of a real terminal. Disabled is true
// between a successful DisableEcho and its restore.
type Recorder struct {
	Disabled bool
}

// DisableEcho implements [term.EchoDisabler], recording that echo is disabled
// until the returned restore func re-enables it.
func (r *Recorder) DisableEcho(io.Writer) (term.RestoreFunc, error) {
	r.Disabled = true
	return func() error {
		r.Disabled = false
		return nil
	}, nil
}

var _ term.EchoDisabler = (*Recorder)(nil)

func ErrEchoDisabler(err error) term.EchoDisabler {
	return echoDisablerFunc(func(w io.Writer) (term.RestoreFunc, error) {
		return nil, err
	})
}

func ErrRestoreDisabler(err error) term.EchoDisabler {
	return echoDisablerFunc(func(w io.Writer) (term.RestoreFunc, error) {
		restore := func() error {
			return err
		}
		return restore, nil
	})
}

func NoOpEchoDisabler() term.EchoDisabler {
	return echoDisablerFunc(func(w io.Writer) (term.RestoreFunc, error) {
		restore := func() error { return nil }
		return restore, nil
	})
}
