// Package arg manages command-line argument registration.
//
// Arguments come in three different and distinct forms:
//
//   - Flags: these are posix-style long and short flags (e.g. --foo, -f, etc)
//     and are backed by the [github.com/spf13/pflag] library. These are
//     constructed with [Flag].
//
//   - Positional: these are fixed-position arguments (e.g. the nth index of a
//     non-flag argument). These are constructed with [Positional].
//
//   - Unmatched: this is always a []string, and contains every argument that
//     does nat satisfy one of the above two category. These are constructed
//     with [Unmatched].
//
// All forms of arguments are strongly-typed, are automatically populated as
// part of command invocations, and offer a variety of settings, including:
//
//   - Callbacks on invocation/setting
//
//   - Cosmetic configurations like what the "type" appears as
//
//   - Auto-completion settings
//
//   - etc
//
// Arguments are conventionally registered through types that implement the
// [Registrar] interface, which is visited by [CommandLine] objects during
// command construction at CLI runtime.
package arg
