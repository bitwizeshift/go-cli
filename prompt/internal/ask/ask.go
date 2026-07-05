package ask

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/bitwizeshift/go-cli/internal/term"
)

// ErrInterrupted is returned by [Asker.Secret] when the user aborts entry with
// Ctrl-C or Ctrl-D.
var ErrInterrupted = errors.New("input interrupted")

// Asker reads answers to prompts from In, writing the prompts and any echoed
// characters to Out. Secret answers are masked with HiddenChar after EchoDisabler
// suppresses the terminal's own echo.
type Asker struct {
	Out io.Writer
	In  io.Reader

	EchoDisabler term.EchoDisabler

	HiddenChar rune
}

// DefaultAsker returns an [Asker] writing to out, reading from in, masking secret
// input with hidden, and using the real terminal for echo control.
func DefaultAsker(out io.Writer, in io.Reader, hidden rune) *Asker {
	return &Asker{
		Out:          out,
		In:           in,
		HiddenChar:   hidden,
		EchoDisabler: term.DefaultConsole,
	}
}

// Line writes prompt to Out and returns the next line of input with its trailing
// newline removed. It returns ctx.Err() if ctx is cancelled before a line is
// read, or io.EOF if input ends without one.
func (a *Asker) Line(ctx context.Context, prompt string) (string, error) {
	if _, err := io.WriteString(a.Out, prompt); err != nil {
		return "", err
	}
	return a.readLine(ctx, bufio.NewReader(a.In))
}

// Confirm writes prompt followed by a "[y/n]" hint and returns the parsed
// answer, treating "y"/"yes" as true and "n"/"no" as false (case-insensitive).
// It re-prompts on any other input, and returns ctx.Err() on cancellation.
func (a *Asker) Confirm(ctx context.Context, prompt string) (bool, error) {
	reader := bufio.NewReader(a.In)
	for {
		if _, err := io.WriteString(a.Out, prompt+" [y/n]: "); err != nil {
			return false, err
		}
		line, err := a.readLine(ctx, reader)
		if err != nil {
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}
	}
}

// Value writes prompt, reads one line, and unmarshals it into v via [Unmarshal].
// It returns ctx.Err() on cancellation or a parse error for invalid input, and
// panics if v is not a supported target type.
func (a *Asker) Value(ctx context.Context, prompt string, v any) error {
	line, err := a.Line(ctx, prompt)
	if err != nil {
		return err
	}
	return Unmarshal(line, v)
}

// Secret writes prompt, disables the terminal's echo, and reads one line while
// echoing HiddenChar for each rune. Echo is restored on every return path. It
// returns the entered text, the [term.EchoDisabler]'s error when the stream is
// not an interactive terminal, [ErrInterrupted] on Ctrl-C/Ctrl-D, or ctx.Err().
func (a *Asker) Secret(ctx context.Context, prompt string) (_ string, err error) {
	if _, err := io.WriteString(a.Out, prompt); err != nil {
		return "", err
	}
	restore, err := a.EchoDisabler.DisableEcho(a.Out)
	if err != nil {
		return "", err
	}
	defer func() {
		if rerr := restore(); rerr != nil && err == nil {
			err = rerr
		}
	}()
	return a.readSecret(ctx)
}

// readLine returns the next line from r with its trailing newline removed,
// honouring ctx while the blocking read runs in a goroutine.
func (a *Asker) readLine(ctx context.Context, r *bufio.Reader) (string, error) {
	type result struct {
		line string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		line, err := r.ReadString('\n')
		ch <- result{line: line, err: err}
	}()

	select {
	case <-ctx.Done():
		// No newline was echoed, so move off the prompt line before returning.
		io.WriteString(a.Out, "\n")
		return "", ctx.Err()
	case res := <-ch:
		if res.err != nil && !(errors.Is(res.err, io.EOF) && res.line != "") {
			return "", res.err
		}
		return strings.TrimRight(res.line, "\r\n"), nil
	}
}

// readSecret reads runes until Enter, masking each with HiddenChar. It honours
// ctx, handles backspace, and returns [ErrInterrupted] on Ctrl-C/Ctrl-D.
func (a *Asker) readSecret(ctx context.Context) (string, error) {
	type result struct {
		r   rune
		err error
	}
	reader := bufio.NewReader(a.In)
	ch := make(chan result)
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			r, _, err := reader.ReadRune()
			select {
			case ch <- result{r: r, err: err}:
			case <-done:
				return
			}
			if err != nil {
				return
			}
		}
	}()

	var buf []rune
	for {
		select {
		case <-ctx.Done():
			// Raw mode: "\r\n" both returns to the first column and moves down.
			io.WriteString(a.Out, "\r\n")
			return "", ctx.Err()
		case res := <-ch:
			switch {
			case res.err != nil:
				if errors.Is(res.err, io.EOF) && len(buf) > 0 {
					return string(buf), nil
				}
				return "", res.err
			case res.r == '\r' || res.r == '\n':
				// The terminal is in raw mode, so a bare "\n" would only move
				// the cursor down; "\r\n" also returns it to the first column.
				io.WriteString(a.Out, "\r\n")
				return string(buf), nil
			case res.r == 0x7f || res.r == 0x08:
				if len(buf) > 0 {
					buf = buf[:len(buf)-1]
					io.WriteString(a.Out, "\b \b")
				}
			case res.r == 0x03 || res.r == 0x04:
				return "", ErrInterrupted
			default:
				buf = append(buf, res.r)
				io.WriteString(a.Out, string(a.HiddenChar))
			}
		}
	}
}
