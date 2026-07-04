// Package redraw rewrites a block of terminal text in place, replacing what it
// last drew with the next block.
//
// It is the mechanism behind live-updating output such as progress bars and
// spinners. It operates purely on strings and is unaware of what they depict,
// so it composes with any renderer that produces one.
package redraw
