package update

import (
	"context"
	"fmt"
)

// gitlabBaseURL is the default host for the GitLab API.
const gitlabBaseURL = "https://gitlab.com"

// GitLabProvider reports the latest version from a GitLab project's releases,
// using the tag of the project's most recent release.
type GitLabProvider struct {
	// Project is the numeric project ID or the URL-encoded "namespace/project"
	// path.
	Project string `yaml:"project"`

	// BaseURL overrides the GitLab host. It defaults to [gitlabBaseURL] when empty
	// and exists primarily so tests can target a local server.
	BaseURL string `yaml:"-"`
}

// LatestVersion returns the tag of the project's most recent release as
// canonical, v-prefixed semver, or an empty version when the project has no
// releases. It returns [ErrUnexpectedStatus], [ErrDecodeResponse], or
// [ErrInvalidVersion] when the release cannot be resolved.
func (p *GitLabProvider) LatestVersion(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/releases", baseURL(p.BaseURL, gitlabBaseURL), p.Project)
	var releases []struct {
		TagName string `json:"tag_name"`
	}
	if err := fetchJSON(ctx, url, &releases); err != nil {
		return "", err
	}
	if len(releases) == 0 {
		return "", nil
	}
	return canonicalVersion(releases[0].TagName)
}

var _ Provider = (*GitLabProvider)(nil)
