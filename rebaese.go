package main

import (
	"context"
	"log"
	"os"

	"github.com/containous/flaeg"
	"github.com/google/go-github/github"
	"github.com/ldez/rebaese/core"
	"github.com/ldez/rebaese/gh"
	"github.com/ldez/rebaese/meta"
	"golang.org/x/oauth2"
)

type Rebaese struct {
	Owner          string `short:"o" description:"Repository owner."`
	RepositoryName string `long:"repo-name" short:"r" description:"Repository name."`
	GitHubToken    string `long:"token" short:"t" description:"GitHub Token."`
	SSH            bool   `description:"Enable SSH support."`
	PRNumber       int    `long:"pr" description:"PR number."`
	MinReview      int    `long:"min-review" description:"Minimum number of required reviews."`
	DryRun         bool   `long:"dry-run" description:"Dry run mode."`
	Debug          bool   `description:"Debug mode."`
	Version        bool   `short:"v" description:"Display the current version."`
}

func main() {

	rebaese := &Rebaese{
		DryRun: true,
	}

	rootCmd := &flaeg.Command{
		Name:                  "rebaese",
		Description:           "Rebaese is a tool made for rebase PR from GitHub.",
		Config:                rebaese,
		DefaultPointersConfig: &Rebaese{},
		Run: func() error {

			if rebaese.Version {
				meta.DisplayVersion()
				return nil
			}

			if rebaese.Debug {
				log.Printf("Run Rebaese command with config : %+v\n", rebaese)
			}
			if rebaese.DryRun {
				log.Print("IMPORTANT: you are using the dry-run mode. Use `--dry-run=false` to disable this mode.")
			}
			requiredStringField(rebaese.Owner, "owner")
			requiredStringField(rebaese.RepositoryName, "repo-name")
			requiredIntField(rebaese.PRNumber, "pr")

			ctx := context.Background()
			client := newGitHubClient(ctx, rebaese.GitHubToken)

			err := rebaese.rebase(ctx, client)
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	flag := flaeg.New(rootCmd, os.Args[1:])
	flag.Run()
}

func (r *Rebaese) rebase(ctx context.Context, client *github.Client) error {

	pr, _, err := client.PullRequests.Get(ctx, r.Owner, r.RepositoryName, r.PRNumber)
	if err != nil {
		return err
	}

	// Check status
	ghub := gh.NewGHub(ctx, client)

	err = ghub.IsFullyMergeable(pr, r.MinReview)
	if err != nil {
		return err
	}

	err = core.Process(pr, r.SSH, r.GitHubToken, r.DryRun, r.Debug)
	if err != nil {
		return err
	}

	return nil
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

func requiredStringField(field string, fieldName string) error {
	if len(field) == 0 {
		log.Fatalf("%s is mandatory.", fieldName)
	}
	return nil
}

func requiredIntField(field int, fieldName string) error {
	if field < 0 {
		log.Fatalf("%s is mandatory.", fieldName)
	}
	return nil
}
