package update

import (
	"context"
	"fmt"

	"github.com/bitwizeshift/go-cli/internal/updatecheck"
)

// goProxyBaseURL is the default host for the Go module proxy.
const goProxyBaseURL = "https://proxy.golang.org"

// GoProxyProvider reports the latest version of a module published to a Go
// module proxy.
type GoProxyProvider struct {
	Module string `yaml:"module"`

	// BaseURL overrides the module proxy host. It defaults to [goProxyBaseURL]
	// when empty and exists primarily so tests can target a local server.
	BaseURL string `yaml:"-"`
}

// LatestVersion returns the module's latest version as canonical, v-prefixed
// semver. It returns [ErrUnexpectedStatus], [ErrDecodeResponse], or
// [ErrInvalidVersion] when the version cannot be resolved.
func (p *GoProxyProvider) LatestVersion(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/%s/@latest", baseURL(p.BaseURL, goProxyBaseURL), p.Module)
	var info struct {
		Version string `json:"Version"`
	}
	if err := fetchJSON(ctx, url, &info); err != nil {
		return "", err
	}
	return updatecheck.CanonicalVersion(info.Version)
}

var _ Provider = (*GoProxyProvider)(nil)
