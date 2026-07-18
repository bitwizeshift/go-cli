/*
Package argreg provides the base definition of the command-line registry object.

This package exists to provide a level of indirection between the "real"
[arg.CommandLine] implementation and the internal state it must share with other
packages. This separation is what makes it possible for the [arg.CommandLine] to
be constructed internally to the CLI construction, and how the [argtest] package
can enable creating custom registries for tests.
*/
package argreg
