package spec

import (
	"errors"
	"fmt"
)

var (
	// ErrUsage is a sentinel error that, when returned from a [Runner], causes
	// the command's usage message to be printed.
	ErrUsage = errors.New("bad usage")

	// ErrPanic is the sentinel error that a [PanicError] unwraps to, identifying
	// a run that terminated by recovering from a panic.
	ErrPanic = errors.New("panic")

	// ErrUnboundRunner indicates a runner was bound to a command id that does not
	// exist in the specification.
	ErrUnboundRunner = errors.New("no command for bound runner")

	// ErrNotMapping indicates the commands node was not a YAML mapping of group
	// name to command list.
	ErrNotMapping = errors.New("commands must be a mapping of group name to commands")

	// ErrInvalidAppID indicates the app-id node was neither a string nor a
	// mapping of host operating system to identifier.
	ErrInvalidAppID = errors.New("app-id must be a string or a mapping of host os to id")

	// ErrUnknownHostOS indicates an app-id mapping was keyed by a host operating
	// system that is not recognized.
	ErrUnknownHostOS = errors.New("unknown host os")
)

// PanicError is the error produced when a [Runner] terminates by panicking. It
// carries the recovered value and the stack trace captured at recovery.
type PanicError struct {
	// Err is the value recovered from the panic.
	Err any

	// Stack is the stack trace captured at the point of recovery.
	Stack []byte
}

// Error returns the string form of the recovered panic value.
func (pe PanicError) Error() string {
	return fmt.Sprintf("%v", pe.Err)
}

// Unwrap returns [ErrPanic], so that a [PanicError] satisfies
// errors.Is(err, ErrPanic).
func (pe PanicError) Unwrap() error {
	return ErrPanic
}

var _ error = (*PanicError)(nil)
