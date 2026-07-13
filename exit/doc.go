// Package exit defines the exit statuses a command may terminate with, and the
// translation of errors into them.
//
// The exit status is a command's only machine-readable output. This package
// gives an application a vocabulary of statuses that follow the POSIX-style
// sysexits.h conventions, and a way to decide which one an error deserves.
//
// That decision is made by a [Classifier]. An application composes classifiers
// to layer its own errors on top of the ones the standard library produces, so
// that each failure mode of a command reaches a distinct status without every
// runner needing to know what a status is.
package exit
