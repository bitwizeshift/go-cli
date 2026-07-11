package tmplfuncs

import (
	"text/template"
)

// NewFunc returns the template function map shared by every renderer. Each entry
// is a function that yields an object whose methods templates call directly,
// such as {{ build.VCS }} and {{ text.Wrap … }}.
func NewFunc() template.FuncMap {
	return template.FuncMap{
		"build": func() *Build { return &DefaultBuild },
		"text":  func() Text { return Text{} },
	}
}
