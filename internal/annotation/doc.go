/*
Package annotation provides a thin internal wrapper around custom annotations
that the cli package and subpackages use. These annotations enable configuring
[pflag.Flag] objects _directly_ instead of requiring the companion [cobra.Command]
object where it's not needed. This helps to reduce the command-coupling with
flag objects.

Being an internal package also aids in visibility, since the shared logic can
be reused in the flagtest package without needing duplication or high-coupling.
*/
package annotation
