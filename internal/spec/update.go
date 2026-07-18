package spec

import (
	"context"
	"time"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/template"
	"github.com/bitwizeshift/go-cli/internal/template/help"
	"github.com/bitwizeshift/go-cli/update"
	"github.com/spf13/cobra"
)

// defaultUpdateTTL is how long a cached update check is reused when no TTL is
// configured.
const defaultUpdateTTL = 24 * time.Hour

// updateTimeout bounds the update check performed while rendering help.
const updateTimeout = 3 * time.Second

// UpdateOptions configures update-availability checking for a command tree.
type UpdateOptions struct {
	// Version is the running build's version.
	Version string

	// Source is the distribution channel the build was installed from.
	Source string

	// TTL is how long a cached check is reused. A non-positive value uses a
	// 24-hour default.
	TTL time.Duration

	// Providers maps a source name to the provider that reports its latest
	// version. Each provider is configured from the matching "update-sources"
	// entry.
	Providers map[string]update.Provider
}

// enabled reports whether update checking has the version, source, and provider
// needed to run.
func (o UpdateOptions) enabled() bool {
	return o.Version != "" && o.Source != "" && len(o.Providers) > 0
}

// checker builds the update [update.Checker] for app, memoizing lookups through
// cache, or nil when checking is not enabled. It returns a decoding error when an
// update-source configuration cannot be decoded into its provider.
func (o UpdateOptions) checker(app *Application, cache update.Cache) (*update.Checker, error) {
	if !o.enabled() {
		return nil, nil
	}
	ttl := o.TTL
	if ttl <= 0 {
		ttl = defaultUpdateTTL
	}
	registry := update.ProviderRegistry{}
	for name, provider := range o.Providers {
		if node, ok := app.UpdateSources[name]; ok {
			if err := node.Decode(provider); err != nil {
				return nil, err
			}
		}
		registry[name] = &update.CacheProvider{
			Provider: provider,
			Source:   name,
			TTL:      ttl,
			Cache:    cache,
			Now:      time.Now,
		}
	}
	build := update.BuildInfo{Version: o.Version, Source: o.Source}
	return update.NewChecker(build, &registry), nil
}

// installUpdateHelp replaces cmd's help function with one that appends an update
// advisory when a newer version is available. cl supplies cmd's positional
// arguments for the rendered help. The check runs under a short timeout and
// contributes nothing on error.
func installUpdateHelp(cmd *cobra.Command, checker *update.Checker, cl *arg.CommandLine) {
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		stdout := cmd.OutOrStdout()
		renderer := template.DefaultRenderEngine.HelpRenderer(stdout)
		renderer.Notice = updateNotice(cmd.Context(), checker)
		_ = renderer.Render(stdout, cmd, cl)
	})
}

// updateNotice runs checker under a short timeout and returns a help notice when
// a newer version is available, or nil on no update or any error.
func updateNotice(ctx context.Context, checker *update.Checker) *help.Notice {
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	result, err := checker.Check(ctx)
	if err != nil || !result.Available {
		return nil
	}
	return &help.Notice{
		Current: result.Current,
		Latest:  result.Latest,
	}
}
