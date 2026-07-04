package cli

import (
	"fmt"

	"github.com/bitwizeshift/go-cli/internal/spec"
)

// Option configures how a [CLI] is constructed from a specification.
type Option interface {
	apply(*config)
}

// config holds the resolved options used to build a [CLI].
type config struct {
	runners map[string]spec.Runner
}

// newConfig resolves options into a config, panicking if two options bind a
// runner to the same command id.
func newConfig(options ...Option) *config {
	cfg := &config{runners: map[string]spec.Runner{}}
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
