# Customizing Exit Codes

This tutorial takes an application whose errors all surface as exit code `1`,
and teaches it to report a distinct exit code per failure mode -- so that the
shell scripts and CI jobs calling your binary can branch on *why* it failed,
not just *that* it failed.

It assumes you have read [Your First `go-cli` Application][first-app].

**Goal:** Translate your application's own errors into your own exit codes, while
keeping the framework's translation of standard-library errors.

## How exit codes are produced

A runner reports failure by returning an `error`:

```go
func (r *DeployRunner) Run(ctx context.Context, _ ...string) error {
  cfg, err := config.FromFile(r.path)
  if err != nil {
    return err
  }
  // ...
}
```

The framework passes that error to an [`exit.Classifier`], which returns the
[`exit.Code`] the process exits with:

```go
type Classifier interface {
  ClassifyError(error) exit.Code
}
```

Without configuration, applications use [`exit.POSIXClassifier`]. It recognizes
the errors of the standard library -- `fs.ErrNotExist` becomes `exit.CodeNoInput`
(66), `fs.ErrPermission` becomes `exit.CodeNoPerm` (77), a `*net.DNSError`
becomes `exit.CodeNoHost` (68), and so on. It matches with `errors.Is` and
`errors.As`, so wrapping does not defeat it: an error returned as
`fmt.Errorf("read config: %w", err)` classifies as whatever it wraps.

Any error it does not recognize -- which includes every error your application
defines itself -- goes unclassified, and the command exits `1`. That is the gap
this tutorial closes.

## Part 1: Give the failures distinct types

A classifier can only distinguish failures the error values already distinguish.
If every failure is `errors.New("something went wrong")`, no classifier can help
you. So the first step is in your own packages, not in `go-cli`.

Use sentinels when the failure carries no data:

```go
package deploy

var (
  // ErrNoManifest is returned when the deployment manifest is absent.
  ErrNoManifest = errors.New("no deployment manifest")

  // ErrLocked is returned when another deployment holds the environment lock.
  ErrLocked = errors.New("environment is locked")
)
```

And an error type when the failure carries data worth reporting:

```go
// ValidationError reports a manifest field that failed validation.
type ValidationError struct {
  Field  string
  Reason string
}

func (e *ValidationError) Error() string {
  return fmt.Sprintf("manifest: %s: %s", e.Field, e.Reason)
}
```

Return these wrapped with context as usual (`fmt.Errorf("deploy: %w", ErrLocked)`).
The classifier will still see through the wrapping.

## Part 2: Choosing the codes

Exit codes are a public interface of your binary. Pick them deliberately:

* **0** is success (`exit.CodeSuccess`), and **1** is an unclassified failure.
* **2--63** carry no conventional meaning. This range is yours.
* **64--78** are the [sysexits.h] conventions, exposed one-for-one as `exit.Code`
  constants: `EX_USAGE` is `exit.CodeUsage`, `EX_CONFIG` is `exit.CodeConfig`,
  and so on. See the [`exit`][exit] package for all predefined exit codes.

* **126 and above** are reserved by the shell: 126 and 127 report that a command
  could not be executed or found, and 128+N reports termination by signal N. Do
  not produce them.
* **Negative** values are not valid exit code statuses.

Prefer the sysexits constant when one fits -- `exit.CodeConfig` for a bad config
file, `exit.CodeTempFail` for something the caller should retry -- and reach for
the 2-63 range only for failures that are genuinely specific to your domain.
For the running example: a missing manifest is a missing input, an invalid
manifest is bad input data, and a held lock is a retryable failure.

## Part 3: Write the classifier

[`exit.ClassifierFunc`] adapts a plain function to the `Classifier` interface.
If you need something more structured, you can also define a type that implements
`Classifier` directly -- but for this tutorial we will just focus on a `func`.

Match with `errors.Is` and `errors.As`:

```go
package deploy

import (
  "errors"

  "github.com/bitwizeshift/go-cli/exit"
)

var ExitClassifier = exit.ClassifierFunc(func(err error) exit.Code {
  var validationErr *deploy.ValidationError

  switch {
  case errors.Is(err, ErrNoManifest):
    return exit.CodeNoInput
  case errors.Is(err, ErrLocked):
    return exit.CodeTempFail
  case errors.As(err, &validationErr):
    return exit.CodeDataErr
  }
  return exit.CodeUnknown
})
```

The last line is a contract: `exit.CodeUnknown` means "I don't know what this
error is", and defers the decision to a future classifier, if one is available.
It is a classification sentinel rather than a status. This makes it possible to
support and ship `exit.Classifier` objects as a library that callers can easily
compound together.

Installing this classifier on its own would discard the standard-library
translation: a `fs.ErrNotExist` raised deep inside your deploy would no longer
reach `exit.CodeNoInput`, because your classifier does not recognize it and
nothing else is consulted. Part 4 will teach you how to fix this.

## Part 4: Compose it with the POSIX classifier

[`exit.FallbackClassifier`] is a classifier primitive which enables chaining of
classifiers, based on the `exit.CodeUnknown` value. The first one to return
something other than `exit.CodeUnknown` decides the exit code.

That is precisely what lets you keep the standard-library translation while
adding your own:

```go
classifier := exit.FallbackClassifier{
  deploy.ExitClassifier,
  // optionally, other classifiers ...
  exit.POSIXClassifier,
}
```

Read it as a chain of responsibility: your classifier gets first refusal on every
error, and anything it declines falls through to `exit.POSIXClassifier`, which
still maps `fs.ErrNotExist`, `context.DeadlineExceeded`, `syscall.Errno`, and the
rest.

**Order matters.** Both classifiers defer with `exit.CodeUnknown`, so the earlier
one wins whenever both recognize an error. Yours goes first: if your `deploy`
package defines an error that also wraps `fs.ErrNotExist`, you want your code for
it, not the generic one.

If no classifier in the chain recognizes the error, the chain returns
`exit.CodeUnknown` and the command exits `1` -- the same outcome as an
application with no classifier configured at all. An unclassified error never
exits `0`.

## Part 5: Wire it into the application

Install the classifier with the [`cli.ExitClassifier`] option:

```go
func main() {
  cli.FromBytes(configYAML,
    cli.BindRunner("deploy", &DeployRunner{}),
    cli.ExitClassifier(exit.FallbackClassifier{
      deploy.ExitClassifier,
      exit.POSIXClassifier,
    }),
  ).Execute()
}
```

Verify it end to end from the shell, which is where these codes are consumed:

```sh
$ myapp deploy --manifest missing.yaml
$ echo $?
66
```

## Testing the classifier

A classifier is a logical unit with a well-defined responsibility, which makes
testing it easy:

```go
package deploy_test

// ...

func TestExitClassifier(t *testing.T) {
  t.Parallel()
  testCases := []struct{
    name string
    err  error
    want exit.Code
  }{
    {
      name: "NoManifest",
      err:  fmt.Errorf("deploy: %w", deploy.ErrNoManifest),
      want: exit.CodeNoInput,
    }, {
      name: "ValidationFailure",
      err:  &deploy.ValidationError{Field: "region", Reason: "unknown"},
      want: exit.CodeDataErr,
    }, {
      name: "UnrecognizedError",
      err:  errors.New("something went wrong"),
      want: exit.CodeUnknown,
    },
  }

  for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
      t.Parallel()

      // Arrange
      sut := deploy.ExitClassifier

      // Act
      code := sut.ClassifyError(tc.err)

      // Assert
      if got, want := code, tc.want; !cmp.Equal(got, want) {
        t.Errorf("ExitClassifier.ClassifyError(...) = %v, want %v", got, want)
      }
    })
  }
}
```

## Where to go next

* [Your First `go-cli` Application][first-app] covers the runner and binding model
  these classifiers sit behind.
* The [`exit`][exit] package reference lists every predefined code.

[first-app]: ./first-application.md
[sysexits.h]: https://man.freebsd.org/cgi/man.cgi?query=sysexits
[exit]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/exit
[`exit.Classifier`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/exit#Classifier
[`exit.ClassifierFunc`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/exit#ClassifierFunc
[`exit.Code`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/exit#Code
[`exit.POSIXClassifier`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/exit#POSIXClassifier
[`exit.FallbackClassifier`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/exit#FallbackClassifier
[`cli.ExitClassifier`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli#ExitClassifier
