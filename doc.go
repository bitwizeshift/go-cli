// Package cli assembles a command-line application from a declarative YAML
// specification and a set of bound runners.
//
// An application is constructed with [FromBytes] or [FromReader], binding each
// command's behavior with [BindRunner]. The result is a [CLI] whose [CLI.Execute]
// runs the application and terminates the process with the appropriate
// [ExitCode]. Construction panics on a malformed or mis-bound specification,
// which is intended to be embedded in the binary and validated at build time.
package cli
