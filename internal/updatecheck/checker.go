package updatecheck

import "context"

// Result is the outcome of a single update check.
type Result struct {
	// Available reports whether Latest is a strictly newer release than the
	// running version.
	Available bool

	// Current is the running version, as canonical v-prefixed semver.
	Current string

	// Latest is the newest version reported by the source, as canonical
	// v-prefixed semver. It is empty when no update information was available.
	Latest string
}

// Checker decides whether the running build has a newer release available on the
// channel it was distributed through.
type Checker struct {
	build    BuildInfo
	registry *ProviderRegistry
}

// NewChecker returns a [Checker] that compares build against the latest version
// reported by the provider registered for build.Source in registry.
func NewChecker(build BuildInfo, registry *ProviderRegistry) *Checker {
	return &Checker{
		build:    build,
		registry: registry,
	}
}

// Check queries the provider registered for the build's source and compares its
// latest version against the running version.
//
// The returned [Result] reports Available false, without error, when no provider
// is registered for the source, when the source reports no version, or when the
// running version is not valid semver. It returns the provider's error when the
// lookup fails.
func (c *Checker) Check(ctx context.Context) (Result, error) {
	latest, err := c.registry.LatestVersion(ctx, c.build.Source)
	if err != nil {
		return Result{}, err
	}
	current := c.build.Version
	if canonical, err := CanonicalVersion(current); err == nil {
		current = canonical
	}
	return Result{
		Available: latest != "" && IsNewer(latest, current),
		Current:   current,
		Latest:    latest,
	}, nil
}
