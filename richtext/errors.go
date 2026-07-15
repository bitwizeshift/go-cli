package richtext

import (
	"errors"
	"fmt"
)

var (
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
