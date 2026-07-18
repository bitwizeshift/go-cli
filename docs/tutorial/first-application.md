# Your First `go-cli` Application

This tutorial builds a small `greeter` application from scratch, then tests it.
By the end you will have a command tree, a flag, and two levels of test coverage.

It assumes you know Go and have used a CLI framework before. It does not assume
you know `go-cli`.

**Goal:** Create your first `go-cli` application.

## Install

```sh
go get github.com/bitwizeshift/go-cli
```

The module path ends in `go-cli`, but the package is named `cli`:

```go
import "github.com/bitwizeshift/go-cli"
```

## The split: YAML declares, Go implements

`go-cli` splits a command into two halves.

The *specification* -- names, summaries, descriptions, aliases, argument counts,
nesting -- lives in a YAML file that you embed into the binary. The *behavior*
lives in Go. The two are joined by an `id`.

This is the library's central opinion. It keeps long help text out of Go string
literals, and it leaves your Go code holding only the parts worth testing.

## Step 1: Write the specification

Create `app.yaml`:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/bitwizeshift/go-cli/refs/heads/master/command.schema.json
id: root
use: greeter <command>
version: 0.1.0
summary: Greets people
description: |
  greeter is a small demonstration application.

  It exists to show how a go-cli specification is bound to Go behavior.
issue-url: https://github.com/you/greeter/issues

commands:
  default:
    - id: greet
      use: greet <name>
      summary: Greets someone by name
      arity: 1
```

The first line points your editor at [`command.schema.json`][schema], which gets
you completion and validation while you write. A relative path works too, if you
vendor the schema alongside your source.

Two fields are required on every node: `id` and `use`. `id` is the key you bind
Go code to, and it never appears in the user interface. Everything else is
optional. The full field list is in the [command schema reference][schema-ref].

`arity: 1` declares that `greet` takes exactly one positional argument. The
framework enforces this before your code runs, which matters in step 2.

## Step 2: Implement a runner

A command's behavior is a `cli.Runner`:

```go
type Runner interface {
  Run(ctx context.Context, args ...string) error
}
```

That is the whole interface. Create `main.go`:

```go
package main

import (
  "context"
  _ "embed"
  "fmt"

  "github.com/bitwizeshift/go-cli"
)

//go:embed app.yaml
var configYAML []byte

func main() {
  cli.FromBytes(configYAML,
    cli.BindRunner("greet", &GreetRunner{}),
  ).Execute()
}

// GreetRunner backs "greet".
type GreetRunner struct{}

func (*GreetRunner) Run(ctx context.Context, args ...string) error {
  fmt.Fprintf(cli.OutStream(ctx), "Hello, %s!\n", args[0])
  return nil
}

var _ cli.Runner = (*GreetRunner)(nil)
```

Four things are worth pulling out of that.

**`args[0]` is not a bug.** `arity: 1` already rejected any other count, so by
the time `Run` executes, `len(args)` is exactly 1. Declare the arity in YAML and
you can index positionally without needing defensive checks.

**`FromBytes` panics on a bad specification.** A malformed YAML document, or an
`id` bound to no command, panics rather than returning an error. The
specification is embedded at build time, so a failure here is a programming
error, not a runtime condition. You will see it the first time you run the
binary.

**`FromBytes(...).Execute()` handles the exit code.** You don't need to manually
call `os.Exit` with a nonzero exit-code; the `CLI`'s `Execute()` function will
do this for you. It will return `1` for normal error cases, and `2` for panic
recoveries.

**Write to `cli.OutStream(ctx)`, not to `os.Stdout`.** See below.

## Step 3: Write to the context, not to stdout

`fmt.Println` writes to the process's stdout. Nothing can intercept that, which
means nothing can test it.

The framework puts its writers on the context:

```go
func OutStream(ctx context.Context) io.Writer
func ErrStream(ctx context.Context) io.Writer
```

These are the writers the CLI was built with. They handle colour and
[richtext markup][richtext] for you, and -- the point of this section -- a test
can swap them out. Making every runner take its writer from `ctx` is what makes
the rest of this tutorial possible.

If you need to lay text out to the terminal width, `cli.StreamColumns(ctx, w)`
reports the column count for a given writer, resolving to a sane default when
the destination is not a terminal.

## Step 4: Add a flag

Flags are declared in Go, not in YAML. A runner opts into flags by also
implementing `arg.Registrar`:

```go
type Registrar interface {
  RegisterArgs(fs *Registry)
}
```

Give `GreetRunner` a `--loud` flag:

```go
import "github.com/bitwizeshift/go-cli/arg"

type GreetRunner struct {
  loud bool
}

func (g *GreetRunner) RegisterArgs(cl *arg.CommandLine) {
  cl.Add(arg.Flag("loud", &g.loud,
    arg.Shorthand("l"),
    arg.Usage("shout the greeting"),
  ))
}

func (g *GreetRunner) Run(ctx context.Context, args ...string) error {
  greeting := fmt.Sprintf("Hello, %s!", args[0])
  if g.loud {
    greeting = strings.ToUpper(greeting)
  }
  fmt.Fprintln(cli.OutStream(ctx), greeting)
  return nil
}

var _ arg.Registrar = (*GreetRunner)(nil)
```

`arg.Flag` is generic over the destination pointer, so the type of `&g.loud`
determines how the flag parses. Because `loud` is a `bool`, `--loud` works as a
bare flag with no value. `arg.Flag` only constructs the flag; `cl.Add` registers
it on the command line.

Registration happens automatically: when the framework binds a runner, it checks
whether that runner implements `Registrar` and calls `RegisterArgs` if so. You
never touch a `pflag.FlagSet` directly.

Flags can be grouped in the help output with `arg.AddToGroup`, and constrained
against each other with `arg.MarkRequired`, `arg.MarkMutuallyExclusive`,
`arg.MarkRequiredTogether`, and `arg.MarkOneRequired`. Building genuinely
reusable flag components is the subject of the [next tutorial][custom-flags].

## Step 5: Test the runner

Because the runner writes to the context, testing it needs no subprocess, no
temporary files, and no command tree. `clitest.WithCaptureWriters` returns a
context carrying capture writers plus the handle to read them back:

```go
func WithCaptureWriters(ctx context.Context) (context.Context, *Output)
```

`Output` exposes `Stdout`, `Stderr`, and `Combined`, each a `fmt.Stringer`
evaluated at the moment you call it. Colour is disabled on these writers, so
assertions compare against plain text.

```go
package main

import (
  "context"
  "testing"

  "github.com/bitwizeshift/go-cli/clitest"
  "github.com/google/go-cmp/cmp"
  "github.com/google/go-cmp/cmp/cmpopts"
)

func TestGreetRunner_Run(t *testing.T) {
  t.Parallel()

  testCases := []struct {
    name string
    loud bool
    want string
  }{
    {
      name: "Default",
      loud: false,
      want: "Hello, world!\n",
    },
    {
      name: "Loud",
      loud: true,
      want: "HELLO, WORLD!\n",
    },
  }

  for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
      t.Parallel()

      // Arrange
      ctx, output := clitest.WithCaptureWriters(context.Background())
      sut := &GreetRunner{loud: tc.loud}

      // Act
      err := sut.Run(ctx, "world")

      // Assert
      if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
        t.Fatalf("sut.Run(...) = %v, want nil", err)
      }
      if got, want := output.Stdout.String(), tc.want; !cmp.Equal(got, want) {
        t.Errorf("sut.Run(...) stdout = %q, want %q", got, want)
      }
    })
  }
}
```

The flag is set by assigning the struct field. There is no parsing involved, and
no dependency on the flag's name or shorthand. Testing what the command *does*
is decoupled from how it is *spelled*.

## Where to go next

* [`examples/simple`](../../examples/simple) is a fuller application using this
  same shape: grouped commands, nested subcommands, aliases, flag constraints,
  and a deliberate panic.
* [Command Schema reference][schema-ref] documents every YAML field, including
  the `arity` expression grammar and command grouping.
* [Creating custom reusable flag types][custom-flags] builds flags that are
  themselves testable, injectable components.
* [Formatted output with richtext][richtext] covers styling what you print.

[schema]: ../../command.schema.json
[schema-ref]: ../reference/command-schema.md
[richtext]: ./richtext-output.md
[custom-flags]: ./custom-flags.md
