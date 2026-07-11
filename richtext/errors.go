package richtext

import (
	"errors"
	"fmt"
)

var (
	// ErrNotStruct reports that the value passed to [NewTheme] is not a struct
	// or a pointer to one.
	ErrNotStruct = errors.New("not a struct")

	// ErrThemeFieldType reports that a theme-tagged field is not a style.Style.
	ErrThemeFieldType = errors.New("theme field is not a style")

	// ErrDuplicateTheme reports that a theme name is declared more than once.
	ErrDuplicateTheme = errors.New("duplicate theme name")

	// ErrUnbalancedTag reports a closing tag that does not match the most
	// recently opened tag.
	ErrUnbalancedTag = errors.New("unbalanced closing tag")

	// ErrUnclosedTag reports that input ended while tags were still open.
	ErrUnclosedTag = errors.New("unclosed tag")
)

// TagError describes a problem with a specific tag encountered while writing.
// It wraps one of the package's sentinel errors.
type TagError struct {
	Namespace string
	Err       error
}

// Error implements [error].
func (e *TagError) Error() string {
	return fmt.Sprintf("tag [%s]: %v", e.Namespace, e.Err)
}

// Unwrap returns the wrapped sentinel error.
func (e *TagError) Unwrap() error { return e.Err }

// ThemeFieldError describes a struct field that [NewTheme] could not accept. It
// wraps one of the package's sentinel errors.
type ThemeFieldError struct {
	Field string // name of the struct field
	Name  string // theme name from the field's tag
	Err   error
}

// Error implements [error].
func (e *ThemeFieldError) Error() string {
	return fmt.Sprintf("theme field %q (%q): %v", e.Field, e.Name, e.Err)
}

// Unwrap returns the wrapped sentinel error.
func (e *ThemeFieldError) Unwrap() error { return e.Err }
