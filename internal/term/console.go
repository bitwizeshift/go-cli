package term

import (
	"errors"
	"fmt"
	"io"

	"golang.org/x/term"
)

var ErrNotDescriptor = errors.New("writer is not a file descriptor")

type RestoreFunc = func() error

type EchoDisabler interface {
	DisableEcho(w io.Writer) (RestoreFunc, error)
}

type Console struct {
	MakeRaw func(fd int) (*term.State, error)
	Restore func(fd int, state *term.State) error
}

func (c *Console) DisableEcho(w io.Writer) (RestoreFunc, error) {
	file, ok := any(w).(interface{ Fd() uintptr })
	if !ok {
		return nil, fmt.Errorf("echo: %w", ErrNotDescriptor)
	}
	fd := file.Fd()
	state, err := c.MakeRaw(int(fd))
	if err != nil {
		return nil, err
	}
	restore := func() error {
		return c.Restore(int(fd), state)
	}
	return restore, nil
}

var DefaultConsole = &Console{
	MakeRaw: term.MakeRaw,
	Restore: term.Restore,
}
