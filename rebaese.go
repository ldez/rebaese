package main

import (
	"context"
	"log"

	"github.com/google/go-github/github"
	"github.com/ldez/rebaese/core"
	"github.com/ldez/rebaese/gh"
	"golang.org/x/oauth2"
)

type Rebaese struct {
	Owner          string
	RepositoryName string
	GitHubToken    string
	PRNumber       int
	MinReview      int
	DryRun         bool
}

func main() {

	// 1504
	// same remote: 1589
	rebaese := &Rebaese{
		Owner:          "containous",
		RepositoryName: "traefik",
		GitHubToken:    "",
		PRNumber:       1635,
		DryRun:         false,
	}

	ctx := context.Background()
	client := newGitHubClient(ctx, rebaese.GitHubToken)

	rebaese.rebase(ctx, client)
}

func (r *Rebaese) rebase(ctx context.Context, client *github.Client) {

	pr, _, err := client.PullRequests.Get(ctx, r.Owner, r.RepositoryName, r.PRNumber)
	if err != nil {
		log.Fatal(err)
	}

	// Check status
	ghub := gh.NewGHub(ctx, client)

	err = ghub.IsFullyMergeable(pr, r.MinReview)
	if err != nil {
		log.Fatal(err)
	}

	err = core.Process(pr, r.GitHubToken, r.DryRun)
	if err != nil {
		log.Fatal(err)
	}
}

func newGitHubClient(ctx context.Context, token string) *github.Client {
	var client *github.Client
	if len(token) == 0 {
		client = github.NewClient(nil)
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}
	return client
}
