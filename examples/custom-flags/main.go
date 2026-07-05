package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/flag"
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

func (f *GitHubFlags) RegisterFlags(registry *flag.Registry) {
	flag.AddToGroup("GitHub Flags",
		flag.Add(registry, "github-token", &f.token,
			flag.Shorthand("T"),
			flag.Type("api-token"),
			flag.Usage("the GitHub API token to use for communication"),
			flag.DefaultFromEnv("GITHUB_TOKEN"),
			flag.DefaultFromEnv("GH_TOKEN"),
		),
		flag.Add(registry, "github-pr", &f.githubPR,
			flag.Usage("the pull request number"),
			flag.Type("pr-number"),
		),
		flag.Add(registry, "github-owner", &f.owner,
			flag.Usage("the user or org that owns the repostiory"),
			flag.DefaultFromEnv("GITHUB_REPOSITORY_OWNER"),
			flag.DefaultFromFunc(f.defaultOwner),
		),
		flag.Add(registry, "github-repo", &f.repo,
			flag.Usage("the user or org that owns the repostiory"),
			flag.DefaultFromFunc(f.defaultRepo),
		),
		flag.Add(registry, "github-api-url", &f.githubBaseURL,
			flag.Type("url"),
			flag.Usage("the base URL for the GitHub API"),
			flag.DefaultFromEnv("GITHUB_API_URL"),
			flag.Hidden(),
		),
		flag.Add(registry, "github-upload-url", &f.githubUploadURL,
			flag.Type("url"),
			flag.Usage("the upload URL for the GitHub API"),
			flag.DefaultFromEnv("GITHUB_API_URL"),
			flag.Hidden(),
		),
	)
}

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

func (ghr *rootRunner) Run(ctx context.Context, _ ...string) error {
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
