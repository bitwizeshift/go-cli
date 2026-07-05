# `go-cli`

[![Continuous Integration][ci-badge]][ci-link]

This library is an **opinionated** framework built around the [cobra] and [pflag]
libraries to provide a simple and easy way to build **testable** command-line
applications.

[cobra]: https://github.com/spf13/cobra
[pflag]: https://github.com/spf13/pflag
[ci-badge]: https://github.com/bitwizeshift/go-cli/actions/workflows/continuous-integration.yaml/badge.svg
[ci-link]: https://github.com/bitwizeshift/go-cli/actions/workflows/continuous-integration.yaml

## Quick Links

* [💼 Features](#features)
* [🚀 Examples](./examples)
* [🤨 Why `go-cli`?](#why-go-cli)

## Features

* 🚦 **Highly testable surface-area**: Your command execution becomes a simple
  [`cli.Runner`] that can use dependency-injected factories.

* 📄 **Configurable commands using YAML**: No more inline `string`s for `Short`
  or `Long`; write it in a `.yaml` file so that it's easy to read and maintain.

* 💻 **Automatic text resizing in terminals**: `--help` messages will
  automatically size-to-fit terminals of different sizes so that output remains
  readable.

* ⏩ **Simplified flag surface area**: No more dozens of `pflag.FlagSet` values;
  just use `flag.Add`, and it idiomatically uses Go to understand
  either [`encoding.TextUnmarshaler`] or custom [`flag.Unmarshaler`].

* 🛟 **Support for flag fallback defaults**: Flags can now support falling back
  to either environment variables, or custom functions to derive a default value.

* 📦 **Support for grouping flags**: A highly-requested but absent feature of
  both [`cobra`] and [`pflag`], available at last.

* 🎨 **Visually clear defaults**: This puts a fresh coat of paint on the default
  CLI experience of Cobra out of the box by leveraging conditional colours.

Overall, this library sacrifices Cobra's verbosity and flexibility for some
opinionated and visually clear defaults.

[`encoding.TextUnmarshaler`]: https://pkg.go.dev/encoding@go1.26.4#TextUnmarshaler
[`flag.Unmarshaler`]: https://pkg.go.dev/github.com/bitwizeshift/go-cli/flag#Unmarshaler

## Why `go-cli`?

Both `cobra` and `pflag` are libraries that _get the job done_, but are not
well-suited towards building testable or high-quality/scalable abstractions.
There are a number of sharp-edges that this library tackles:

* The `cobra.Command` and `pflag.FlagSet` share a number of responsibilities
  between them -- which means writing code following best practices like
  [Single-Responsibility-Principle][srp]. For example, flag names are dictated
  by the flag definitions, but -_flag completion_, _flag constraints_ (requires,
  mutually exclusive, etc.), etc are all part of the `cobra.Command` object.
  This leads code to need to know about both, instead of just knowing about flags.

* The `pflag.Value` abstraction is a poor abstraction for testability since it
  enables `Type()` and `String()` to vary, when the behavior should be
  consistent. In practice, testing `Type()` leads to high-coupling in unit
  tests, or missed coverage.

* There is no mechanism for grouping flags in any easy way, it needs to be
  hand-rolled.

* [pflag.FlagSet] offers entirely too many receivers for setting up flags.

* etc. The list goes on and on.

The primary motivation is tackling these sharp edges, while improving ergonomics,
at the sacrificed of overwhelming number of combinations users will never use.

[srp]: https://en.wikipedia.org/wiki/Single-responsibility_principle

## Disclaimer

I wrote this library for me, to satisfy my own personal needs -- and am sharing
this out in case anyone else finds value in it. I make my own opinionated
defaults that others may not agree with; and that's _fine_.
