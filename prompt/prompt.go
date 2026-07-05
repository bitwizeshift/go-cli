package prompt

import (
	"context"
	"io"
	"os"

	"github.com/bitwizeshift/go-cli/prompt/internal/ask"
)

// Prompter reads answers to interactive prompts, writing prompts and echoed
// characters to Out and reading replies from In. Secret answers are masked with
// HiddenChar.
type Prompter struct {
	Out io.Writer
	In  io.Reader

	HiddenChar rune
}

// Line writes prompt to Out and returns the entered line.
func (p *Prompter) Line(ctx context.Context, prompt string) (string, error) {
	return p.asker().Line(ctx, prompt)
}

// Confirm writes prompt and returns the parsed yes/no answer, re-prompting until
// the reply is recognised.
func (p *Prompter) Confirm(ctx context.Context, prompt string) (bool, error) {
	return p.asker().Confirm(ctx, prompt)
}

// Secret writes prompt and reads a reply without revealing it, masking each rune
// with HiddenChar. It errors when In/Out is not an interactive terminal.
func (p *Prompter) Secret(ctx context.Context, prompt string) (string, error) {
	return p.asker().Secret(ctx, prompt)
}

// Value writes prompt, reads one line, and unmarshals it into v.
func (p *Prompter) Value(ctx context.Context, prompt string, v any) error {
	return p.asker().Value(ctx, prompt, v)
}

func (p *Prompter) asker() *ask.Asker {
	return ask.DefaultAsker(p.Out, p.In, p.HiddenChar)
}

// DefaultPrompter reads from standard input and writes to standard output,
// masking secret input with '*'.
var DefaultPrompter = Prompter{
	Out:        os.Stdout,
	In:         os.Stdin,
	HiddenChar: '*',
}

// Line prompts on standard output and reads a line from standard input.
func Line(ctx context.Context, prompt string) (string, error) {
	return DefaultPrompter.Line(ctx, prompt)
}

// Confirm prompts for and reads a yes/no answer from standard input.
func Confirm(ctx context.Context, prompt string) (bool, error) {
	return DefaultPrompter.Confirm(ctx, prompt)
}

// Secret prompts for and reads a masked answer from the standard input terminal.
func Secret(ctx context.Context, prompt string) (string, error) {
	return DefaultPrompter.Secret(ctx, prompt)
}

// Value prompts for a line on standard input and unmarshals it into v.
func Value(ctx context.Context, prompt string, v any) error {
	return DefaultPrompter.Value(ctx, prompt, v)
}
