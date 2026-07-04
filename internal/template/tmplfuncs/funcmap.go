package tmplfuncs

import (
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/template/palette"
)

// NewFunc returns the template function map shared by every renderer. Each entry
// is a function that yields an object whose methods templates call directly,
// such as {{ palette.Error … }}, {{ build.VCS }}, and {{ text.Wrap … }}. Only
// the palette varies with colour settings; build and text are stateless.
func NewFunc(p palette.Palette) template.FuncMap {
	return template.FuncMap{
		"build":   func() *Build { return &DefaultBuild },
		"palette": func() palette.Palette { return p },
		"text":    func() Text { return Text{} },
	}
}
