package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/ldez/rebaese/git"
	"golang.org/x/oauth2"
)

type RepositoryInformation struct {
	URL        string
	BranchName string
}

type Configuration struct {
	Owner          string
	RepositoryName string
	GitHubToken    string
	PRNumber       int
}

func main() {

	// 1504
	// same remote: 1589
	configuration := &Configuration{
		Owner:          "containous",
		RepositoryName: "traefik",
		GitHubToken:    "",
		PRNumber:       1567,
	}

	ctx := context.Background()
	client := newGitHubClient(ctx, configuration.GitHubToken)

	pr, _, err := client.PullRequests.Get(ctx, configuration.Owner, configuration.RepositoryName, configuration.PRNumber)

	if err != nil {
		log.Panic(err)
	}

	log.Println("Base branch: ", *pr.Base.Ref, "- Fork branch: ", *pr.Head.Ref)

	forkInformation := &RepositoryInformation{
		URL:        createRepositoryURL(*pr.Head.Repo.GitURL, configuration.GitHubToken),
		BranchName: *pr.Head.Ref,
	}

	baseInformation := &RepositoryInformation{
		URL:        createRepositoryURL(*pr.Base.Repo.GitURL, ""),
		BranchName: *pr.Base.Ref,
	}

	pullRequestRebase(*forkInformation, *baseInformation)
}

func pullRequestRebase(forkInformation RepositoryInformation, baseInformation RepositoryInformation) {

	dir, err := ioutil.TempDir("", "forker")
	if err != nil {
		log.Panic(err)
	}

	// clean up
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	fmt.Println(os.Getwd())

	remoteName := "upstream"

	if forkInformation.URL == baseInformation.URL {
		log.Print("It's not a fork, it's a branch on the main repository.")

		if forkInformation.BranchName == "master" {
			log.Fatal("Master branch cannot be rebase.")
		}

		remoteName = "origin"

		output, err := prepareMainRepository(forkInformation, baseInformation)
		if err != nil {
			log.Fatal(output, err)
		}
	} else {
		output, err := prepareFork(forkInformation, remoteName, baseInformation)
		if err != nil {
			log.Fatal(output, err)
		}
	}

	output, err := git.Rebase(remoteName, baseInformation.BranchName)
	if err != nil {
		log.Fatal(output, err)
	}

	output, err = git.PushForce("origin", forkInformation.BranchName)
	if err != nil {
		log.Fatal(output, err)
	}
}
func prepareMainRepository(forkInformation RepositoryInformation, baseInformation RepositoryInformation) (string, error) {

	output, err := git.Clone(baseInformation.URL)
	if err != nil {
		return output, err
	}

	git.Config("rebase.autoSquash", "true")
	git.Config("push.default", "current")

	output, err = git.Checkout(forkInformation.BranchName)
	if err != nil {
		return output, err
	}

	return "", nil
}

func prepareFork(forkInformation RepositoryInformation, remoteName string, baseInformation RepositoryInformation) (string, error) {

	output, err := git.CloneBranch(forkInformation.URL, forkInformation.BranchName)
	if err != nil {
		return output, err
	}

	git.Config("rebase.autoSquash", "true")
	git.Config("push.default", "current")

	output, err = git.AddRemote(remoteName, baseInformation.URL)
	if err != nil {
		return output, err
	}

	output, err = git.Fetch(remoteName, baseInformation.BranchName)
	if err != nil {
		return output, err
	}

	return "", nil
}

func createRepositoryURL(cloneURL string, token string) string {
	if len(token) > 0 {
		return strings.Replace(cloneURL, "git://", "https://"+token+"@", -1)
	}
	return strings.Replace(cloneURL, "git://", "https://", -1)
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
