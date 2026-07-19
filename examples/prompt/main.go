// Command prompt-demo is a self-contained showcase of the go-cli prompt package.
//
// It binds a runner per input kind to an embedded YAML specification, reading a
// plain line, a yes/no confirmation, a masked secret, and a typed value, each
// honouring the cancellation context the runtime provides.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/prompt"
)

//go:embed app.yaml
var configYAML []byte

func main() {
	cli.FromBytes(configYAML,
		cli.BindRunner("prompt-demo", &rootRunner{}),
		cli.BindRunner("prompt-demo.line", &lineRunner{}),
		cli.BindRunner("prompt-demo.confirm", &confirmRunner{}),
		cli.BindRunner("prompt-demo.secret", &secretRunner{}),
		cli.BindRunner("prompt-demo.value.string", &valueRunner[string]{label: "Enter a string: "}),
		cli.BindRunner("prompt-demo.value.int", &valueRunner[int]{label: "Enter an integer: "}),
		cli.BindRunner("prompt-demo.value.float", &valueRunner[float64]{label: "Enter a float: "}),
		cli.BindRunner("prompt-demo.value.bool", &valueRunner[bool]{label: "Enter a bool (true/false): "}),
		cli.BindRunner("prompt-demo.value.duration", &valueRunner[time.Duration]{label: "Enter a duration (e.g. 1m30s): "}),
	).Execute()
}

// rootRunner backs the top-level command and shows usage when invoked without a
// subcommand.
type rootRunner struct{}

func (*rootRunner) Run(context.Context) error {
	return cli.ErrUsage
}

// lineRunner backs "line" and reads a plain, echoed line.
type lineRunner struct{}

func (*lineRunner) Run(ctx context.Context) error {
	name, err := prompt.Line(ctx, "What is your name? ")
	if err != nil {
		return err
	}
	fmt.Printf("Hello, %s!\n", name)
	return nil
}

// confirmRunner backs "confirm" and reads a yes/no answer.
type confirmRunner struct{}

func (*confirmRunner) Run(ctx context.Context) error {
	ok, err := prompt.Confirm(ctx, "Do you want to continue?")
	if err != nil {
		return err
	}
	fmt.Printf("continue = %t\n", ok)
	return nil
}

// secretRunner backs "secret" and reads a masked value without revealing it.
type secretRunner struct{}

func (*secretRunner) Run(ctx context.Context) error {
	secret, err := prompt.Secret(ctx, "Enter a password: ")
	if err != nil {
		return err
	}
	fmt.Printf("captured a %d-character secret\n", len(secret))
	return nil
}

// valueRunner backs each "value <type>" subcommand, reading a value of type T
// via prompt.Value and printing the parsed result.
type valueRunner[T any] struct {
	label string
}

func (r *valueRunner[T]) Run(ctx context.Context) error {
	var value T
	if err := prompt.Value(ctx, r.label, &value); err != nil {
		return err
	}
	fmt.Printf("value = %v\n", value)
	return nil
}

var (
	_ cli.Runner = (*rootRunner)(nil)
	_ cli.Runner = (*lineRunner)(nil)
	_ cli.Runner = (*confirmRunner)(nil)
	_ cli.Runner = (*secretRunner)(nil)
	_ cli.Runner = (*valueRunner[int])(nil)
)
