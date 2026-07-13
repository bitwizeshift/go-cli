package cli

import (
	"fmt"

	"github.com/bitwizeshift/go-cli/exit"
	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/bitwizeshift/go-cli/richtext"
)

// Option configures how a [CLI] is constructed from a specification.
type Option interface {
	apply(*config)
}

// config holds the resolved options used to build a [CLI].
type config struct {
	runners    map[string]spec.Runner
	theme      *richtext.Theme
	colour     spec.ColourMode
	sizer      term.Sizer
	classifier exit.Classifier
}

// newConfig resolves options into a config, panicking if two options bind a
// runner to the same command id.
func newConfig(options ...Option) *config {
	cfg := &config{
		runners:    map[string]spec.Runner{},
		classifier: exit.POSIXClassifier,
	}
	for _, opt := range options {
		opt.apply(cfg)
	}
	return cfg
}

type option func(*config)

func (o option) apply(c *config) { o(c) }

// BindRunner binds runner to the command identified by id in the specification.
//
// It panics if id has already been bound by an earlier option; a bound id that
// matches no command is reported when the [CLI] is constructed.
func BindRunner(id string, runner Runner) Option {
	return option(func(c *config) {
		if _, ok := c.runners[id]; ok {
			panic(fmt.Sprintf("cli: duplicate runner bound to id %q", id))
		}
		c.runners[id] = runner
	})
}

// Theme sets the [richtext.Theme] used to resolve the styling tags in the CLI's
// output. When unset, [richtext.DefaultTheme] is used.
func Theme(theme *richtext.Theme) Option {
	return option(func(c *config) {
		c.theme = theme
	})
}

// DisableColour forces the CLI's output to be uncoloured, regardless of the
// destination.
//
// It is mutually exclusive with [ForceColour]: setting a colour mode more than
// once, or setting both, panics.
func DisableColour() Option {
	return option(func(c *config) {
		setColour(c, spec.ColourDisabled)
	})
}

// ForceColour forces the CLI's output to be coloured, regardless of the
// destination.
//
// It is mutually exclusive with [DisableColour]: setting a colour mode more than
// once, or setting both, panics.
func ForceColour() Option {
	return option(func(c *config) {
		setColour(c, spec.ColourEnabled)
	})
}

// TerminalWidth sets a fixed width for the terminal width, instead of being
// dynamic based on the terminal size. Setting columns less than 60 will panic.
func TerminalWidth(columns int) Option {
	return option(func(c *config) {
		if columns < 60 {
			panic(fmt.Sprintf("TerminalWidth set to %d, which is not enough for basic formatting", columns))
		}
		c.sizer = term.FixedSizer(columns)
	})
}

// ExitClassifier sets the [exit.Classifier] for this application to control
// what the underlying exit code is given application-state errors.
func ExitClassifier(classifier exit.Classifier) Option {
	return option(func(c *config) {
		if classifier == nil {
			panic("nil classifier provided to cli.ExitClassifier")
		}
		c.classifier = classifier
	})
}

// setColour transitions the config's colour mode, panicking on any transition
// away from the default: a mode may be selected at most once.
func setColour(c *config, mode spec.ColourMode) {
	if c.colour != spec.ColourAuto {
		panic("cli: colour mode already set")
	}
	c.colour = mode
}
