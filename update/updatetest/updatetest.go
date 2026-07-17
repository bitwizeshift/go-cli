package updatetest

import (
	"context"

	"github.com/bitwizeshift/go-cli/update"
)

// Provider returns an [update.Provider] that always reports version as the latest
// available version.
func Provider(version string) update.Provider {
	return providerFunc(func(context.Context) (string, error) {
		return version, nil
	})
}

// ErrProvider returns an [update.Provider] that always fails with err.
func ErrProvider(err error) update.Provider {
	return providerFunc(func(context.Context) (string, error) {
		return "", err
	})
}

// providerFunc adapts a function into an [update.Provider].
type providerFunc func(ctx context.Context) (string, error)

// LatestVersion calls the underlying function.
func (f providerFunc) LatestVersion(ctx context.Context) (string, error) {
	return f(ctx)
}

var _ update.Provider = (providerFunc)(nil)
