package update

import (
	"context"
	"fmt"

	"github.com/bitwizeshift/go-cli/internal/updatecheck"
)

// githubBaseURL is the default host for the GitHub REST API.
const githubBaseURL = "https://api.github.com"

// GitHubProvider reports the latest version from a GitHub repository's releases,
// using the tag of the repository's latest release.
type GitHubProvider struct {
	Owner string `yaml:"owner"`
	Repo  string `yaml:"repo"`

	// BaseURL overrides the GitHub API host. It defaults to [githubBaseURL] when
	// empty and exists primarily so tests can target a local server.
	BaseURL string `yaml:"-"`
}

// LatestVersion returns the tag of the repository's latest release as canonical,
// v-prefixed semver. It returns [ErrUnexpectedStatus], [ErrDecodeResponse], or
// [ErrInvalidVersion] when the release cannot be resolved.
func (p *GitHubProvider) LatestVersion(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", baseURL(p.BaseURL, githubBaseURL), p.Owner, p.Repo)
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := fetchJSON(ctx, url, &release); err != nil {
		return "", err
	}
	return updatecheck.CanonicalVersion(release.TagName)
}

var _ Provider = (*GitHubProvider)(nil)
