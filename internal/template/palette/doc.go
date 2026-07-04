// Package palette defines the colour roles used to style rendered CLI output.
//
// It exists so that every renderer shares one styling vocabulary instead of
// each defining its own. A [Palette] assigns a styling to each role; [Colour]
// applies an ANSI scheme and [NoColour] leaves text plain. Renderers consume a
// [Palette] through the template function map and never implement one
// themselves.
package palette
