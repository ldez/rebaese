package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/ldez/rebaese/git"
)

type repositoryInformation struct {
	URL        string
	BranchName string
}

func Process(pr *github.PullRequest, ssh bool, gitHubToken string, dryRun bool) error {

	log.Println("Base branch: ", *pr.Base.Ref, "- Fork branch: ", *pr.Head.Ref)

	forkInformation := &repositoryInformation{
		URL:        createRepositoryURL(*pr.Head.Repo.GitURL, ssh, gitHubToken),
		BranchName: *pr.Head.Ref,
	}

	baseInformation := &repositoryInformation{
		URL:        createRepositoryURL(*pr.Base.Repo.GitURL, ssh, ""),
		BranchName: *pr.Base.Ref,
	}

	dir, err := ioutil.TempDir("", "rebaese")
	if err != nil {
		return err
	}

	// clean up
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	fmt.Println(os.Getwd())

	remoteName, err := prepare(forkInformation, baseInformation)
	if err != nil {
		return err
	}

	output, err := git.Rebase(remoteName, baseInformation.BranchName)
	if err != nil {
		log.Print(err)
		return errors.New(output)
	}

	if dryRun {
		log.Println("Fake push force.")
	} else {
		output, err = git.PushForce("origin", forkInformation.BranchName)
		if err != nil {
			log.Print(err)
			return errors.New(output)
		}
	}

	return nil
}

func prepare(forkInformation *repositoryInformation, baseInformation *repositoryInformation) (string, error) {

	remoteName := "upstream"

	if forkInformation.URL == baseInformation.URL {
		log.Print("It's not a fork, it's a branch on the main repository.")

		if forkInformation.BranchName == "master" {
			return "", errors.New("Master branch cannot be rebase.")
		}

		remoteName = "origin"

		output, err := prepareMainRepository(forkInformation, baseInformation)
		if err != nil {
			log.Print(err)
			return "", errors.New(output)
		}
	} else {
		output, err := prepareFork(forkInformation, remoteName, baseInformation)
		if err != nil {
			log.Print(err)
			return "", errors.New(output)
		}
	}

	return remoteName, nil
}

func prepareMainRepository(forkInformation *repositoryInformation, baseInformation *repositoryInformation) (string, error) {

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

func prepareFork(forkInformation *repositoryInformation, remoteName string, baseInformation *repositoryInformation) (string, error) {

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

func createRepositoryURL(cloneURL string, ssh bool, token string) string {
	if ssh {
		return strings.Replace(cloneURL, "git://github.com/", "git@github.com:", -1)
	}

	prefix := "https://"
	if len(token) > 0 {
		prefix += token + "@"
	}
	return strings.Replace(cloneURL, "git://", prefix, -1)
}
