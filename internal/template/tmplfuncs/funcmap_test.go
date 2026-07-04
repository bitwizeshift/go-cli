package tmplfuncs_test

import (
	"maps"
	"slices"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/template/palette"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
)

func TestNewFunc_Keys(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := tmplfuncs.NewFunc(palette.NoColour{})

	// Act
	keys := slices.Collect(maps.Keys(sut))
	sort.Strings(keys)

	// Assert
	if got, want := keys, []string{"build", "palette", "text"}; !cmp.Equal(got, want) {
		t.Errorf("NewFunc keys = %v, want %v", got, want)
	}
}

func TestNewFunc_Build_ReturnsDefaultBuild(t *testing.T) {
	t.Parallel()

	// Arrange
	funcs := tmplfuncs.NewFunc(palette.NoColour{})
	provider := funcs["build"].(func() *tmplfuncs.Build)

	// Act
	build := provider()

	// Assert
	if got, want := build, &tmplfuncs.DefaultBuild; got != want {
		t.Errorf("build() = %p, want %p", got, want)
	}
}

func TestNewFunc_Palette_ReturnsGivenPalette(t *testing.T) {
	t.Parallel()

	// Arrange
	funcs := tmplfuncs.NewFunc(palette.DefaultColour)
	provider := funcs["palette"].(func() palette.Palette)

	// Act
	selected := provider()

	// Assert
	if got, want := selected, palette.Palette(palette.DefaultColour); !cmp.Equal(got, want) {
		t.Errorf("palette() = %v, want %v", got, want)
	}
}

func TestNewFunc_Text_ReturnsText(t *testing.T) {
	t.Parallel()

	// Arrange
	funcs := tmplfuncs.NewFunc(palette.NoColour{})
	provider := funcs["text"].(func() tmplfuncs.Text)

	// Act
	text := provider()

	// Assert
	if got, want := text, (tmplfuncs.Text{}); !cmp.Equal(got, want) {
		t.Errorf("text() = %v, want %v", got, want)
	}
}
