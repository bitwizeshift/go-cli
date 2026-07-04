// Package progress renders live, updating command-line progress indicators.
//
// This provides both Bar and Progress forms of indicators, each with
// configurable color schemes. Progress appearances update incrementally based
// on manual update calls via [Render] operations. This enables the caller to
// manage the update period manually inline with other operations.
//
// To avoid manual updates, an [Animator]
package progress
