package richtext

import (
	"fmt"
	"strings"

	"github.com/bitwizeshift/go-cli/richtext/style"
)

// Theme maps theme names to their styles. It is consulted by a [Writer] to
// resolve [theme:name] tags and is safe for concurrent reads. A Theme may
// derive from a parent (see [Theme.New]), falling back to it for names it does
// not define itself.
type Theme struct {
	parent *Theme
	styles map[string]style.Style
}

// DefaultTheme is the standard colour scheme used by the go-cli project.
// Users can derive themes from this to override basic stylings by using
// [Theme.New].
var DefaultTheme = &Theme{
	styles: map[string]style.Style{
		"title":    {Foreground: style.Green},
		"heading":  {Foreground: style.Yellow},
		"label":    {Foreground: style.Cyan},
		"value":    {Foreground: style.White},
		"emphasis": {Foreground: style.White, Attributes: style.Bold},
		"error":    {Foreground: style.Red},
		"warning":  {Foreground: style.Yellow},
		"info":     {Foreground: style.Green},
		"debug":    {Foreground: style.Cyan},
		"quote":    {Foreground: style.BrightBlack},
		"gutter":   {Foreground: style.White},
		"url":      {Foreground: style.BrightWhite, Attributes: style.Underline},
	},
}

// NewTheme builds a Theme that maps each name in styles to its style. A name is
// referenced by a [theme:name] tag, so it must be non-empty and must not contain
// a ']'; NewTheme panics if any name violates this.
func NewTheme(styles map[string]style.Style) *Theme {
	return &Theme{styles: copyStyles(styles)}
}

// New derives a Theme from styles that overrides the receiver: names declared in
// styles take precedence, and any other name falls back to the receiver and its
// ancestors. It follows the same naming rules and panic behaviour as [NewTheme].
func (t *Theme) New(styles map[string]style.Style) *Theme {
	return &Theme{parent: t, styles: copyStyles(styles)}
}

// copyStyles returns a private copy of styles, panicking on any name that could
// never be referenced by a [theme:name] tag.
func copyStyles(styles map[string]style.Style) map[string]style.Style {
	out := make(map[string]style.Style, len(styles))
	for name, s := range styles {
		if !validThemeName(name) {
			panic(fmt.Sprintf("richtext: invalid theme name %q", name))
		}
		out[name] = s
	}
	return out
}

// validThemeName reports whether name can be referenced by a [theme:name] tag: a
// non-empty string that does not contain the tag terminator ']'.
func validThemeName(name string) bool {
	return name != "" && !strings.ContainsRune(name, ']')
}

// lookup resolves a theme by name, searching this theme before its ancestors.
func (t *Theme) lookup(name string) (style.Style, bool) {
	if s, ok := t.styles[name]; ok {
		return s, true
	}
	if t.parent != nil {
		return t.parent.lookup(name)
	}
	return style.Style{}, false
}
