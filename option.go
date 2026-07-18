package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/bitwizeshift/go-cli/exit"
	"github.com/bitwizeshift/go-cli/internal/spec"
	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/bitwizeshift/go-cli/richtext"
	"github.com/bitwizeshift/go-cli/update"
)

// Option configures how a [CLI] is constructed from a specification.
type Option interface {
	apply(*config)
}

// config holds the resolved options used to build a [CLI].
type config struct {
	builders   map[string]spec.Builder
	theme      *richtext.Theme
	colour     spec.ColourMode
	sizer      term.Sizer
	classifier exit.Classifier

	buildVersion    string
	buildSource     string
	updateTTL       time.Duration
	updateProviders map[string]update.Provider
}

// newConfig resolves options into a config, panicking if two options bind a
// runner to the same command id.
func newConfig(options ...Option) *config {
	cfg := &config{
		builders:        map[string]spec.Builder{},
		classifier:      exit.POSIXClassifier,
		updateProviders: map[string]update.Provider{},
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
// Prefer [BindBuilder] for binding builders instead though.
func BindRunner(id string, runner Runner) Option {
	return BindBuilder(id, inlineRunner{Runner: runner})
}

// BindBuilder binds builder to the command identified by id in the
// specification.
//
// It panics if id has already been bound by an earlier option; a bound id that
// matches no command is reported whne the [CLI] is constructed.
// Binding a [Builder] is the preferred way of building CLI runners because this
// preserves a separation-of-concerns on responsibility (building vs executing).
func BindBuilder(id string, builder Builder) Option {
	return option(func(c *config) {
		if _, ok := c.builders[id]; ok {
			panic(fmt.Sprintf("cli: duplicate builder bound to id %q", id))
		}
		c.builders[id] = builderWrapper{Builder: builder}
	})
}

type builderWrapper struct {
	Builder Builder
}

func (w builderWrapper) Build(ctx context.Context) (spec.Runner, error) {
	runner, err := w.Builder.Build(ctx)
	return spec.Runner(runner), err
}

type inlineRunner struct {
	Runner Runner
}

func (r inlineRunner) Build(context.Context) (Runner, error) {
	return r.Runner, nil
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

// CurrentVersion sets the running build's version, typically injected at build
// time with -ldflags. It enables update checking together with [BuildSource] and
// at least one [UpdateProvider]; without all three, no update check is performed.
func CurrentVersion(version string) Option {
	return option(func(c *config) {
		c.buildVersion = version
	})
}

// BuildSource sets the distribution channel the running build was installed from,
// such as "github" or "brew". It selects which registered [UpdateProvider] is
// consulted for updates.
func BuildSource(source string) Option {
	return option(func(c *config) {
		c.buildSource = source
	})
}

// UpdateTTL sets how long a cached update check is reused before the source is
// queried again. When unset, a default of 24 hours is used.
func UpdateTTL(d time.Duration) Option {
	return option(func(c *config) {
		c.updateTTL = d
	})
}

// UpdateProvider registers provider as the source of update information under
// name. The name is matched against [BuildSource] and against keys of the
// "update-sources" specification block, whose values configure the provider.
//
// It panics if name has already been registered by an earlier option.
func UpdateProvider(name string, provider update.Provider) Option {
	return option(func(c *config) {
		if _, ok := c.updateProviders[name]; ok {
			panic(fmt.Sprintf("cli: duplicate update provider registered for %q", name))
		}
		c.updateProviders[name] = provider
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
