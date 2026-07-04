// Package spec turns a declarative YAML command specification into a runnable
// [github.com/spf13/cobra.Command] tree.
//
// A specification is decoded through [Build], which binds runners to commands by
// id and wires each command's flags, help, usage, and error rendering. The
// exported specification types exist so their decoding can be tested directly;
// the package is internal so callers depend only on the public facade that wraps
// it.
package spec
