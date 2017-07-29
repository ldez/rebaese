package gh

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/go-github/github"
)

type GHub struct {
	ctx    context.Context
	client *github.Client
}

func NewGHub(ctx context.Context, client *github.Client) *GHub {
	return &GHub{ctx: ctx, client: client}
}

func (g *GHub) IsFullyMergeable(pr *github.PullRequest, minReview int) error {

	prNumber := pr.GetNumber()

	if pr.GetMerged() {
		return fmt.Errorf("The PR #%v is already merged.", prNumber)
	}
	if !pr.GetMergeable() {
		return fmt.Errorf("Conflicts must be resolve in the PR #%v", prNumber)
	}

	err := g.HasSuccessStatus(pr)
	if err != nil {
		return err
	}

	err = g.HasReviewsApprove(pr, minReview)
	if err != nil {
		return err
	}

	return nil
}

func (g *GHub) HasReviewsApprove(pr *github.PullRequest, minReview int) error {

	owner := pr.Base.Repo.Owner.GetLogin()
	repositoryName := pr.Base.Repo.GetName()
	prNumber := pr.GetNumber()

	reviews, _, err := g.client.PullRequests.ListReviews(g.ctx, owner, repositoryName, prNumber, nil)
	if err != nil {
		return err
	}

	reviewsState := make(map[string]string)
	for _, review := range reviews {
		if *review.State != "COMMENTED" {
			reviewsState[review.User.GetLogin()] = review.GetState()
			log.Printf("%s: %s\n", review.User.GetLogin(), review.GetState())
		}
	}

	if len(reviewsState) < minReview {
		return fmt.Errorf("Need more review [%v/2]", len(reviewsState))
	}

	for login, state := range reviewsState {
		if state != "APPROVED" {
			return fmt.Errorf("%s by %s", state, login)
		}
	}

	return nil
}

func (g *GHub) HasSuccessStatus(pr *github.PullRequest) error {

	owner := pr.Base.Repo.Owner.GetLogin()
	repositoryName := pr.Base.Repo.GetName()
	prRef := *pr.Head.SHA

	sts, _, err := g.client.Repositories.GetCombinedStatus(g.ctx, owner, repositoryName, prRef, nil)
	if err != nil {
		return err
	}

	if *sts.State != "success" {
		statuses, _, err := g.client.Repositories.ListStatuses(g.ctx, owner, repositoryName, prRef, nil)
		if err != nil {
			return err
		}
		var summary string
		for _, stat := range statuses {
			if stat.GetState() != "success" {
				summary += stat.GetDescription() + "\n"
			}
		}
		return errors.New(summary)
	}
	return nil
}
