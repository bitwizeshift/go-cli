// Package helptest provides a fixed command hierarchy used to exercise the
// help renderer.
//
// The same tree backs both the golden-file generator and the golden test, so
// the checked-in output and the freshly rendered output are guaranteed to
// describe the identical commands and flags. The tree deliberately mixes short
// and long descriptions, several flag types, and multiple groups so that
// wrapping and alignment are exercised.
package helptest
