package update

import "context"

// Provider reports the latest released version available from a single
// distribution channel.
type Provider interface {
	// LatestVersion returns the latest available version as canonical, v-prefixed
	// semver (such as "v1.4.0"), or an error when the channel cannot be queried.
	LatestVersion(ctx context.Context) (string, error)
}
