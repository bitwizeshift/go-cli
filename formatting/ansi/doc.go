/*
Package ansi provides a light abstraction around ANSI escape codes.

The primary mechanism for using this package is to create an [SGRControlSequence]
object, and then use the `fmt` package to format the object an ANSI-formatted
string.
*/
package ansi
