// Package userdir resolves the per-user base directories for application storage
// that the os package does not provide, following each operating system's
// conventions.
//
// The operating system and its environment are supplied to each resolver rather
// than read from the process, so the resolution for every platform can be
// exercised from any host. The os-backed values are wired in by the caller.
package userdir
