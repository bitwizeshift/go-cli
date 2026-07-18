# Creating Custom Reusable Flag Types

This tutorial builds a reusable GitHub flag component: one struct that owns
`--github-token`, `--github-owner`, `--github-repo`, and `--github-pr`, hands
back a configured API client, and can be dropped into any command in any binary.
Then it tests that component without running a CLI.

It assumes you have read [Your First `go-cli` Application][first-app].

**Goal:** Create your test first composable flag object.

## What "custom flag type" means here

If you have used [`pflag`] directly, you will expect to implement [`pflag.Value`]:
`Set`, `String`, `Type`. **Do not do that.** `go-cli` does not expose a `Value`
interface, and `arg.AddFlag` constructs the `pflag.Value` internally. Implementing
one gains you nothing.

The library instead splits the problem in two, and they are independently useful:

1. A custom flag **value**: a Go type that knows how to decode itself from a
   flag's raw bytes. This is one method.
2. A custom flag **component**: a struct that registers a coherent set of flags
   and exposes a high-level object built from them. This is the reusable unit,
   and it is what most of this tutorial is about.

[`pflag`]: https://pkg.go.dev/github.com/spf13/pflag
[`pflag.Value`]: https://pkg.go.dev/github.com/spf13/pflag#Value

## Part 1: A custom flag value

To make a type usable as a flag destination, a type must either implement
[`encoding.TextUnmarshaler`] or optionally `arg.Unmarshaler`:

```go
type Unmarshaler interface {
  UnmarshalArg(data []byte) error
}
```

Typically `encoding.TextUnmarshaler` is the more idiomatic thing to follow here,
but `arg.Unmarshaler` exists as an alternative to distinguish a custom encoding
that may only be used for flags. For this tutorial, we will implement a
`arg.Unmarshaler`.

Here is a validated enum:

```go
// Visibility is the visibility of a created repository.
type Visibility string

const (
  VisibilityPublic   Visibility = "public"
  VisibilityPrivate  Visibility = "private"
  VisibilityInternal Visibility = "internal"
)

func (v *Visibility) UnmarshalArg(data []byte) error {
  switch Visibility(data) {
  case VisibilityPublic, VisibilityPrivate, VisibilityInternal:
    *v = Visibility(data)
    return nil
  default:
    return fmt.Errorf("invalid visibility: %s", data)
  }
}
```

Returning an error fails parsing, and the message reaches the user. That is the
entire contract.

You do not need `UnmarshalArg` if your type already implements
`encoding.TextUnmarshaler` -- the decoder tries `arg.Unmarshaler` first, then
`encoding.TextUnmarshaler`, then falls back to built-in parsing for strings,
bools, ints, floats, `time.Duration`, and slices of those. Implement
`UnmarshalArg` when you want flag decoding to differ from text decoding, and
`UnmarshalText` when you want one rule everywhere.

### The type name in help output

By default the help output names the flag's type by kebab-casing the Go type,
package included: `github.Visibility` renders as `github-visibility`. If this
default is not preferable, you can override it:

```go
arg.AddFlag(registry, "visibility", &f.visibility,
  arg.Type("visibility"),
  arg.Usage("visibility of the new repository (public,private,internal)"),
)
```

`arg.Type` exists precisely so the type name is a property of the *flag*, not
baked into the type. That is also what keeps the type testable: there is no
`Type() string` method whose return value a test has to assert against.

## Part 2: The reusable component

A flag component is any type implementing [`arg.Registrar`]:

```go
type Registrar interface {
  RegisterArgs(cl *arg.CommandLine)
}
```

The shape that makes it reusable, rather than just a bag of flags:

* Parsed state lives in **unexported** fields. Nothing outside can read a
  half-configured token.
* Registration options -- flag name, shorthand, usage -- are **exported** fields,
  so a consumer can rename a flag that collides with one of theirs.
* The component exposes **behavior**, not data: a method returning a constructed,
  ready-to-use object following a **Factory** pattern.

```go
package githubflags

import (
  "context"
  "os"
  "strings"
  "time"

  "github.com/bitwizeshift/go-cli/arg"
  "github.com/google/go-github/v88/github"
)

// ClientFlags is a factory for constructing [github.Client] objects from
// command-line flags.
type ClientFlags struct {
  token   string
  baseURL string
}

func (f *ClientFlags) RegisterArgs(cl *arg.CommandLine) {
  arg.AddFlagToGroup("GitHub Flags",
    arg.AddFlag(cl, "github-token", &f.token,
      arg.Shorthand("T"),
      arg.Type("api-token"),
      arg.Usage("the GitHub API token to use for communication"),
      arg.DefaultFromEnv("GITHUB_TOKEN"),
      arg.DefaultFromEnv("GH_TOKEN"),
    ),
    arg.AddFlag(cl, "github-api-url", &f.baseURL,
      arg.Type("url"),
      arg.Usage("the base URL for the GitHub API"),
      arg.DefaultFromEnv("GITHUB_API_URL"),
      arg.Hidden(),
    ),
  )
}

var _ arg.Registrar = (*ClientFlags)(nil)

// Client returns a GitHub client configured from the parsed flags.
func (f *ClientFlags) Client() *github.Client {
  opts := []github.ClientOptionsFunc{github.WithTimeout(5 * time.Second)}
  if f.token != "" {
    opts = append(opts, github.WithAuthToken(f.token))
  }
  if f.baseURL != "" {
    opts = append(opts, github.WithEnterpriseURLs(f.baseURL, f.baseURL))
  }
  client, _ := github.NewClient(opts...)
  return client
}
```

Three details in `RegisterArgs` are doing real work.

**`arg.AddFlag` returns the `*pflag.Flag` it created**, which is why the whole body
can nest inside a single `arg.AddFlagToGroup("GitHub Flags", ...)` call. The group
is a help-output heading; ungrouped flags fall under "General Flags".

**Defaults are layered.** `arg.DefaultFromEnv` can be applied more than once --
above, `GITHUB_TOKEN` is consulted, then `GH_TOKEN`.

Fallbacks only run when the flag was not given on the command line.

**`arg.Hidden()`** keeps `--github-api-url` functional but out of the help
output. Use it for escape hatches you support but do not advertise.

[`arg.Registrar`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/arg#Registrar

## Part 3: Wiring it into a command

Here is the magical part: The runner does not have to implement
`arg.Registrar`, and it does not have to know that flags exist.

See below:

```go
type ClientProvider interface {
  Client() *github.Client
}

type HelloRunner struct {
  ClientProvider ClientProvider
}

func (r *HelloRunner) Run(ctx context.Context, _ ...string) error {
  client := r.ClientProvider.Client()

  user, _, err := client.Users.Get(ctx, "")
  if err != nil {
    return err
  }
  fmt.Fprintf(cli.OutStream(ctx), "Hello %s\n", pull.GetLogin())
  return nil
}

func main() {
  flags := &githubflags.ClientFlags{}
  cli.FromBytes(configYAML,
    cli.BindRunner("hello", &HelloRunner{
      ClientProvider: flags,
    }),
  ).Execute()
}
```

When the framework binds a runner, it calls `arg.Register` on it. If the runner
implements `Registrar`, its `RegisterArgs` runs. If it does not, `Register`
walks the runner's exported fields -- through structs, pointers, interfaces,
slices, and maps -- and registers every `Registrar` it finds. `HelloRunner` gets
its flags because its fields point at something that has them.

That walk has two properties worth knowing:

**Registration is deduplicated by pointer identity.** Above, one `*ClientFlags`
is reachable through three separate fields, and its flags are registered exactly
once.

**A `Registrar` stops the recursion.** If your component has fields that are
themselves `Registrar`s, `Register` will not descend into them -- your
`RegisterArgs` owns that, and must call `arg.Register(cl, f.child)`
itself. Exclude a field from the walk entirely with a `flag:"ignore"` or
`flag:"-"` struct tag.

The payoff is the runner's testability. `HelloRunner` depends on three interfaces,
not on a flag parser, so a test substitutes fakes and never touches a
`FlagSet`.

## Part 4: Testing the component

Split this the same way the component is split: assert the flags' *shape*, then
assert the component's *behavior*. `arg/argtest` covers both.

Build the `arg.CommandLine` using `argtest` and then call `arg.Register`:

```go
cl := argtest.NewCommandLine()
arg.Register(cl, sut)
```

### Shape

`flagtest.LongFlags` reports every registered flag as its long-flag form sorted
by long name. Compare the whole slice at once, and a regression in any flag's
spelling:

```go
package githubflags_test

import (
  "testing"

  "github.com/google/go-cmp/cmp"
  "github.com/google/go-cmp/cmp/cmpopts"
  "github.com/spf13/pflag"

  "github.com/bitwizeshift/go-cli/arg"
  "github.com/bitwizeshift/go-cli/arg/argtest"
)

func TestClientFlags_RegisterArgs(t *testing.T) {
  t.Parallel()

  // Arrange
  cl := argtest.NewCommandLine()
  sut := &githubflags.ClientFlags{}

  // Act
  arg.Register(cl, sut)

  // Assert
  wantFlags := []string{"github-api-url", "api-token"}
  if got, want := flagtest.LongFlags(pfs), wantFlags; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
    t.Errorf("ClientFlags.Register(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
  }
}
```

`flagtest.LongFlags` and `flagtest.ShortFlags` return just the names, for when
you only care that a flag exists.

You can also test full properties with `flagtest.AllFlags` if you want to
evaluate short flags, usage, grouping, constraints, or general annotations.
Just be careful not to devolve into a change-detector test by relying on exact
strings.

### Behavior

`flagtest.Parse` parses arguments into the flag set and fails the test via
`Fatalf` if parsing errors. Drive the real flags, then assert on the component's
*accessor* -- not on its private fields, which is the whole point of hiding them:

```go
func TestClientFlags_Client(t *testing.T) {
  t.Parallel()

  // Arrange
  pfs := pflag.NewFlagSet("test", pflag.ContinueOnError)
  sut := &githubflags.ClientFlags{}
  arg.Register(arg.NewRegistry(pfs), sut)

  // Act
  client := sut.Client()

  // Assert
  defaultURL := "https://api.github.com/"
  if got, want := client.BasURL, defaultURL; !cmp.Equal(got, want) {
    t.Errorf("ClientFlags.Client() BaseURL = %q, want %q", got, want)
  }
  // ...
}
```

Use the same structure to test a rejected value: parse an invalid `--visibility`
and assert the error, which is where a custom `UnmarshalArg` earns its keep.

## Shell completion

Completion is a flag property, registered alongside the flag:

```go
arg.AddFlag(registry, "visibility", &f.visibility,
  arg.CompleteFrom("public", "private", "internal"),
)
```

`arg.CompleterFunc` computes candidates dynamically, and `arg.CompleteFiles`,
`arg.CompleteFilesMatching`, and `arg.CompleteDirs` defer to the shell.
Applying two completion options to one flag panics.

## Where to go next

* [`examples/custom-flags`](../../examples/custom-flags) is this pattern as a
  running program.
* [Formatted output with richtext][richtext] covers styling the output your
  runner produces.

[first-app]: ./first-application.md
[richtext]: ./richtext-output.md
[`encoding.TextUnmarshaler`]: https://pkg.go.dev/encoding#TextUnmarshaler
