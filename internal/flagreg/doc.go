/*
Package flagreg provides the base definition of the flag registry objects.

This package mostly exists to provide a level of indirection between the "real"
implementation, while still sharing elements strictly internally to other
packages. This separation is what makes it possible for the [flag.Registry] to
be constructed internally to the CLI construction, and how the [flagtest] package
can enable creating custom registries.
*/
package flagreg
