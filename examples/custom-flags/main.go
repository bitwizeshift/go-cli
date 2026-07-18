package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/arg"
	"github.com/google/go-github/v88/github"
)

type GitHubFlags struct {
	token string

	githubBaseURL   string
	githubUploadURL string

	owner, repo string

	githubPR int
}

func (f *GitHubFlags) defaultRepoOwner() (owner, repo string, has bool) {
	val, ok := os.LookupEnv("GITHUB_REPO")
	if !ok {
		return "", "", false
	}
	parts := strings.SplitN(val, "/", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func (f *GitHubFlags) defaultOwner(context.Context) (string, error) {
	owner, _, _ := f.defaultRepoOwner()
	return owner, nil
}

func (f *GitHubFlags) defaultRepo(context.Context) (string, error) {
	_, repo, _ := f.defaultRepoOwner()
	return repo, nil
}

func (f *GitHubFlags) RegisterArgs(cl *arg.CommandLine) {
	token := arg.Flag("github-token", &f.token,
		arg.Shorthand("T"),
		arg.Type("api-token"),
		arg.Usage("the GitHub API token to use for communication"),
		arg.DefaultFromEnv("GITHUB_TOKEN"),
		arg.DefaultFromEnv("GH_TOKEN"),
	)
	pr := arg.Flag("github-pr", &f.githubPR,
		arg.Usage("the pull request number"),
		arg.Type("pr-number"),
	)
	owner := arg.Flag("github-owner", &f.owner,
		arg.Usage("the user or org that owns the repostiory"),
		arg.DefaultFromEnv("GITHUB_REPOSITORY_OWNER"),
		arg.DefaultFromFunc(f.defaultOwner),
	)
	repo := arg.Flag("github-repo", &f.repo,
		arg.Usage("the user or org that owns the repostiory"),
		arg.DefaultFromFunc(f.defaultRepo),
	)
	apiURL := arg.Flag("github-api-url", &f.githubBaseURL,
		arg.Type("url"),
		arg.Usage("the base URL for the GitHub API"),
		arg.DefaultFromEnv("GITHUB_API_URL"),
		arg.Hidden(),
	)
	uploadURL := arg.Flag("github-upload-url", &f.githubUploadURL,
		arg.Type("url"),
		arg.Usage("the upload URL for the GitHub API"),
		arg.DefaultFromEnv("GITHUB_API_URL"),
		arg.Hidden(),
	)
	foo := arg.Positional("foo", 0, &f.githubBaseURL,
		arg.Usage("Mehhhh"),
	)

	cl.Add(token, pr, owner, repo, apiURL, uploadURL, foo)
	arg.Group("GitHub Flags", token, pr, owner, repo, apiURL, uploadURL)
}

var _ arg.Registrar = (*GitHubFlags)(nil)

// Client returns a constructed GitHub Client from the given flags
func (f *GitHubFlags) Client() *github.Client {
	var opts []github.ClientOptionsFunc
	if f.token != "" {
		opts = append(opts, github.WithAuthToken(f.token))
	}
	if f.githubBaseURL != "" {
		opts = append(opts, github.WithEnterpriseURLs(f.githubBaseURL, f.githubBaseURL))
	}
	opts = append(opts, github.WithTimeout(5*time.Second))
	client, _ := github.NewClient(opts...)

	return client
}

func (f *GitHubFlags) Repo() (string, string) {
	return f.owner, f.repo
}

func (f *GitHubFlags) PullRequest() int {
	return f.githubPR
}

type ClientSource interface {
	Client() *github.Client
}

type RepoSource interface {
	Repo() (owner, repo string)
}

type PRSource interface {
	PullRequest() int
}

type rootRunner struct {
	ClientSource ClientSource
	RepoSource   RepoSource
	PRSource     PRSource
}

func (ghr *rootRunner) Run(ctx context.Context) error {
	client := ghr.ClientSource.Client()
	owner, repo := ghr.RepoSource.Repo()
	pr := ghr.PRSource.PullRequest()

	fmt.Println("owner:", owner)
	fmt.Println("repo:", repo)
	fmt.Println("pr:", pr)

	resp, _, err := client.PullRequests.Get(ctx, owner, repo, pr)
	if err != nil {
		return err
	}
	fmt.Printf("Title:\n\n%s\n", resp.GetTitle())
	fmt.Printf("Body:\n\n%s\n", resp.GetBody())
	return nil
}

//go:embed app.yaml
var configYAML []byte

func main() {
	flags := &GitHubFlags{}
	cli.FromBytes(configYAML,
		cli.BindRunner("root", &rootRunner{
			ClientSource: flags,
			RepoSource:   flags,
			PRSource:     flags,
		}),
	).Execute()
}
