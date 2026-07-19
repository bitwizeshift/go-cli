package updatecheck

import "context"

// Provider reports the latest released version available from a single
// distribution channel.
type Provider interface {
	// LatestVersion returns the latest available version as canonical, v-prefixed
	// semver (such as "v1.4.0"), or an error when the channel cannot be queried.
	LatestVersion(ctx context.Context) (string, error)
}

// ProviderRegistry maps a distribution source name to the [Provider] that looks
// up the latest version for that source.
type ProviderRegistry map[string]Provider

// LatestVersion delegates to the provider registered under source. A source with
// no registered provider is not an error: it reports an empty version and a nil
// error, meaning no update information is available.
func (pr ProviderRegistry) LatestVersion(ctx context.Context, source string) (string, error) {
	provider, ok := pr[source]
	if !ok {
		return "", nil
	}
	return provider.LatestVersion(ctx)
}
