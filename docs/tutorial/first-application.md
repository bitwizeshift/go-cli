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

The *specification* of commands -- names, summaries, descriptions, aliases,
nesting -- lives in a YAML file that you embed into the binary. The *behavior*,
such as the arguments used, execution logic, etc, lives in Go. The two parts
are bound by an ID.

This is the library's central opinion. It keeps long help text out of Go string
literals, and it leaves your Go code holding only the parts worth testing.

## Step 1: Write the specification

Create `app.yaml`:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/bitwizeshift/go-cli/refs/heads/master/command.schema.json
name: greeter
version: 0.1.0
summary: Greets people
description: |
  greeter is a small demonstration application.

  It exists to show how a go-cli specification is bound to Go behavior.
issue-url: https://github.com/you/greeter/issues

commands:
  default:
    - name: greet
      summary: Greets someone by name
```

The first line points your editor at [`command.schema.json`][schema], which gets
you completion and validation while you write. A relative path works too, if you
vendor the schema alongside your source.

One field is required on every node: `name`, the name the command is invoked by.
Everything else is optional. The full field list is in the
[command schema reference][schema-ref].

## Step 2: Implement a runner

A command's behavior is a `cli.Runner`:

```go
type Runner interface {
  Run(ctx context.Context) error
}
```

That is the whole interface -- it takes no arguments, because a command's
arguments are bound to fields before it runs. A runner declares those bindings
by also implementing `arg.Registrar`:

```go
type Registrar interface {
  RegisterArgs(cl *arg.CommandLine)
}
```

Create `main.go`:

```go
package main

import (
  "context"
  _ "embed"
  "fmt"

  "github.com/bitwizeshift/go-cli"
  "github.com/bitwizeshift/go-cli/arg"
)

//go:embed app.yaml
var configYAML []byte

func main() {
  cli.FromBytes(configYAML,
    cli.BindRunner("greeter.greet", &GreetRunner{}),
  ).Execute()
}

// GreetRunner backs "greet".
type GreetRunner struct {
  name string
}

func (g *GreetRunner) RegisterArgs(cl *arg.CommandLine) {
  cl.Add(arg.Positional("name", 0, &g.name,
    arg.Usage("who to greet"),
    arg.Required(),
  ))
}

func (g *GreetRunner) Run(ctx context.Context) error {
  fmt.Fprintf(cli.OutStream(ctx), "Hello, %s!\n", g.name)
  return nil
}

var (
  _ cli.Runner    = (*GreetRunner)(nil)
  _ arg.Registrar = (*GreetRunner)(nil)
)
```

Four things are worth pulling out of that.

**`g.name` is never empty.** `arg.Required()` tells the framework that the
command line must reach argument 0, so `greeter greet` with no argument is
rejected before `Run` executes. Registering one positional also caps the command
at one argument, so `greeter greet a b` is rejected too. No defensive checks
needed.

**`FromBytes` panics on a bad specification.** A malformed YAML document, or an
id path bound to no command, panics rather than returning an error. The
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

Flags are declared in Go, not in YAML, and they go through the same
`RegisterArgs` method the positional argument used. Give `GreetRunner` a
`--loud` flag:

```go
type GreetRunner struct {
  name string
  loud bool
}

func (g *GreetRunner) RegisterArgs(cl *arg.CommandLine) {
  cl.Add(
    arg.Positional("name", 0, &g.name,
      arg.Usage("who to greet"),
      arg.Required(),
    ),
    arg.Flag("loud", &g.loud,
      arg.Shorthand("l"),
      arg.Usage("shout the greeting"),
    ),
  )
}

func (g *GreetRunner) Run(ctx context.Context) error {
  greeting := fmt.Sprintf("Hello, %s!", g.name)
  if g.loud {
    greeting = strings.ToUpper(greeting)
  }
  fmt.Fprintln(cli.OutStream(ctx), greeting)
  return nil
}
```

`arg.Required()` reads the same on both: on a flag it demands `--loud` be
given, on a positional it demands the argument be supplied.

`arg.Flag` is generic over the destination pointer, so the type of `&g.loud`
determines how the flag parses. Because `loud` is a `bool`, `--loud` works as a
bare flag with no value. `arg.Flag` only constructs the flag; `cl.Add` registers
it on the command line.

Registration happens automatically: when the framework binds a runner, it checks
whether that runner implements `Registrar` and calls `RegisterArgs` if so. You
never touch a `pflag.FlagSet` directly, and you never declare argument counts by
hand -- the registered arguments are what the framework validates against.

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
      sut := &GreetRunner{name: "world", loud: tc.loud}

      // Act
      err := sut.Run(ctx)

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

Both the flag and the positional argument are set by assigning struct fields.
There is no parsing involved, and no dependency on the flag's name, its
shorthand, or the argument's position. Testing what the command *does* is
decoupled from how it is *spelled*.

## Where to go next

* [`examples/simple`](../../examples/simple) is a fuller application using this
  same shape: grouped commands, nested subcommands, aliases, flag constraints,
  and a deliberate panic.
* [Command Schema reference][schema-ref] documents every YAML field, including
  command grouping and how positional arguments determine the counts a command
  accepts.
* [Creating custom reusable flag types][custom-flags] builds flags that are
  themselves testable, injectable components.
* [Formatted output with richtext][richtext] covers styling what you print.

[schema]: ../../command.schema.json
[schema-ref]: ../reference/command-schema.md
[richtext]: ./richtext-output.md
[custom-flags]: ./custom-flags.md
