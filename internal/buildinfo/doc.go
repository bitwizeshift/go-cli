// Package buildinfo reads facts about the running binary that the Go toolchain
// embeds at link time.
//
// The information is read through injectable hooks rather than the runtime
// functions directly, so callers can present deterministic values in tests.
package buildinfo
