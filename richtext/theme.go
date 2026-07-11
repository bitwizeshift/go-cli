package richtext

import (
	"reflect"

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
		"quote":    {Foreground: style.BrightBlack},
		"gutter":   {Foreground: style.White},
		"url":      {Foreground: style.BrightWhite, Attributes: style.Underline},
	},
}

// NewTheme builds a Theme from the exported fields of v that carry a
// `theme:"name"` struct tag. v may be a struct or a pointer to one; each tagged
// field must be of type [style.Style]. Fields without the tag are ignored.
//
// It returns [ErrNotStruct] when v is not a struct, and a [*ThemeFieldError]
// (wrapping [ErrThemeFieldType] or [ErrDuplicateTheme]) for a tagged field of
// the wrong type or a repeated name.
func NewTheme(v any) (*Theme, error) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	styleType := reflect.TypeFor[style.Style]()
	rt := rv.Type()
	styles := make(map[string]style.Style)
	for i := range rt.NumField() {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}
		name, ok := field.Tag.Lookup("theme")
		if !ok {
			continue
		}
		if field.Type != styleType {
			return nil, &ThemeFieldError{Field: field.Name, Name: name, Err: ErrThemeFieldType}
		}
		if _, dup := styles[name]; dup {
			return nil, &ThemeFieldError{Field: field.Name, Name: name, Err: ErrDuplicateTheme}
		}
		styles[name] = rv.Field(i).Interface().(style.Style)
	}
	return &Theme{styles: styles}, nil
}

// New derives a Theme from v that overrides the receiver: names declared in v
// take precedence, and any other name falls back to the receiver and its
// ancestors. It follows the same rules and errors as [NewTheme].
func (t *Theme) New(v any) (*Theme, error) {
	derived, err := NewTheme(v)
	if err != nil {
		return nil, err
	}
	derived.parent = t
	return derived, nil
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
