package update

import (
	"context"
	"fmt"
)

// brewBaseURL is the default host for the Homebrew formulae API.
const brewBaseURL = "https://formulae.brew.sh"

// BrewProvider reports the latest version of a homebrew-core formula, using the
// formula's stable version. Custom taps are not supported.
type BrewProvider struct {
	Name string `yaml:"name"`

	// BaseURL overrides the Homebrew formulae host. It defaults to [brewBaseURL]
	// when empty and exists primarily so tests can target a local server.
	BaseURL string `yaml:"-"`
}

// LatestVersion returns the formula's stable version as canonical, v-prefixed
// semver. It returns [ErrUnexpectedStatus], [ErrDecodeResponse], or
// [ErrInvalidVersion] when the version cannot be resolved.
func (p *BrewProvider) LatestVersion(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/api/formula/%s.json", baseURL(p.BaseURL, brewBaseURL), p.Name)
	var formula struct {
		Versions struct {
			Stable string `json:"stable"`
		} `json:"versions"`
	}
	if err := fetchJSON(ctx, url, &formula); err != nil {
		return "", err
	}
	return canonicalVersion(formula.Versions.Stable)
}

var _ Provider = (*BrewProvider)(nil)
