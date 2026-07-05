// Package term provides utilities for inspecting and controlling the terminal
// a command's streams are attached to.
//
// It resolves terminal properties — such as the display width and whether
// character echo can be toggled — from the writer or reader in use, degrading
// gracefully when a stream is not a real terminal. Callers compose the small
// interfaces here to decide how to render and read interactive output.
package term
