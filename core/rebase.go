package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/ldez/go-git-cmd-wrapper/checkout"
	"github.com/ldez/go-git-cmd-wrapper/clone"
	"github.com/ldez/go-git-cmd-wrapper/config"
	"github.com/ldez/go-git-cmd-wrapper/fetch"
	"github.com/ldez/go-git-cmd-wrapper/git"
	"github.com/ldez/go-git-cmd-wrapper/push"
	"github.com/ldez/go-git-cmd-wrapper/rebase"
	"github.com/ldez/go-git-cmd-wrapper/remote"
)

type repositoryInformation struct {
	URL        string
	BranchName string
}

func Process(pr *github.PullRequest, ssh bool, gitHubToken string, dryRun bool, debug bool) error {

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

	remoteName, err := prepare(forkInformation, baseInformation, debug)
	if err != nil {
		return err
	}

	output, err := git.Rebase(rebase.PreserveMerges, rebase.Branch(fmt.Sprintf("%s/%s", remoteName, baseInformation.BranchName)))
	if err != nil {
		log.Print(err)
		return errors.New(output)
	}

	if dryRun {
		log.Println("Fake push force.")
		output, err = git.Push(push.ForceWithLease, push.DryRun, push.Remote("origin"), push.RefSpec(forkInformation.BranchName), git.Debugger(debug))
		if err != nil {
			log.Print(err)
			return errors.New(output)
		} else {
			log.Println(output)
		}
	} else {
		output, err = git.Push(push.ForceWithLease, push.Remote("origin"), push.RefSpec(forkInformation.BranchName), git.Debugger(debug))
		if err != nil {
			log.Print(err)
			return errors.New(output)
		}
	}

	return nil
}

func prepare(forkInformation *repositoryInformation, baseInformation *repositoryInformation, debug bool) (string, error) {

	remoteName := "upstream"

	if forkInformation.URL == baseInformation.URL {
		log.Print("It's not a fork, it's a branch on the main repository.")

		if forkInformation.BranchName == "master" {
			return "", errors.New("Master branch cannot be rebase.")
		}

		remoteName = "origin"

		output, err := prepareMainRepository(forkInformation, baseInformation, debug)
		if err != nil {
			log.Print(err)
			return "", errors.New(output)
		}
	} else {
		output, err := prepareFork(forkInformation, remoteName, baseInformation, debug)
		if err != nil {
			log.Print(err)
			return "", errors.New(output)
		}
	}

	return remoteName, nil
}

func prepareMainRepository(forkInformation *repositoryInformation, baseInformation *repositoryInformation, debug bool) (string, error) {

	output, err := git.Clone(clone.Repository(baseInformation.URL), git.Debugger(debug))
	if err != nil {
		return output, err
	}

	git.Config(config.Entry("rebase.autoSquash", "true"))
	git.Config(config.Entry("push.default", "current"))

	output, err = git.Checkout(checkout.Branch(forkInformation.BranchName), git.Debugger(debug))
	if err != nil {
		return output, err
	}

	return "", nil
}

func prepareFork(forkInformation *repositoryInformation, remoteName string, baseInformation *repositoryInformation, debug bool) (string, error) {

	output, err := git.Clone(clone.Repository(forkInformation.URL), clone.Branch(forkInformation.BranchName), clone.Directory("."), git.Debugger(debug))
	if err != nil {
		return output, err
	}

	git.Config(config.Entry("rebase.autoSquash", "true"), git.Debugger(debug))
	git.Config(config.Entry("push.default", "current"), git.Debugger(debug))

	output, err = git.Remote(remote.Add, remote.Name(remoteName), remote.URL(baseInformation.URL), git.Debugger(debug))
	if err != nil {
		return output, err
	}

	output, err = git.Fetch(fetch.NoTags, fetch.Remote(remoteName), fetch.RefSpec(baseInformation.BranchName), git.Debugger(debug))
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
