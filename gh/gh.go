package gh

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/go-github/github"
)

type helper struct {
	ctx       context.Context
	client    *github.Client
	minReview int
}

func NewHelper(ctx context.Context, client *github.Client, minReview int) *helper {
	return &helper{ctx: ctx, client: client, minReview: minReview}
}

func (h *helper) IsFullMergeable(owner string, repositoryName string, prNumber int) error {
	pr, _, err := h.client.PullRequests.Get(h.ctx, owner, repositoryName, prNumber)
	if err != nil {
		return err
	}
	if *pr.Merged {
		return fmt.Errorf("The PR #%v is already merged.", prNumber)
	}
	if !*pr.Mergeable {
		return fmt.Errorf("Conflicts must be resolve in the PR #%v", prNumber)
	}

	err = h.HasSuccessStatus(owner, repositoryName, *pr.Head.SHA)
	if err != nil {
		return err
	}

	err = h.HasReviewsApprove(owner, repositoryName, prNumber)
	if err != nil {
		return err
	}
	return nil
}

func (h *helper) HasReviewsApprove(owner string, repositoryName string, prNumber int) error {

	reviews, _, err := h.client.PullRequests.ListReviews(h.ctx, owner, repositoryName, prNumber)
	if err != nil {
		return err
	}

	reviewsState := make(map[string]string)
	for _, review := range reviews {
		if *review.State != "COMMENTED" {
			reviewsState[*review.User.Login] = *review.State
			log.Printf("%s: %s\n", *review.User.Login, *review.State)
		}
	}

	if len(reviewsState) < h.minReview {
		return fmt.Errorf("Need more review [%v/2]", len(reviewsState))
	}

	for login, state := range reviewsState {
		if state != "APPROVED" {
			return fmt.Errorf("%s by %s", state, login)
		}
	}

	return nil
}

func (h *helper) HasSuccessStatus(owner string, repositoryName string, prRef string) error {

	sts, _, err := h.client.Repositories.GetCombinedStatus(h.ctx, owner, repositoryName, prRef, nil)
	if err != nil {
		return err
	}

	if *sts.State != "success" {
		statuses, _, err := h.client.Repositories.ListStatuses(h.ctx, owner, repositoryName, prRef, nil)
		if err != nil {
			return err
		}
		var summary string
		for _, stat := range statuses {
			if *stat.State != "success" {
				summary += *stat.Description + "\n"
			}
		}
		return errors.New(summary)
	}
	return nil
}
