package handler

import (
	"context"

	"github.com/google/go-github/v53/github"
	"github.com/tbistr/inc"
	"github.com/tbistr/inc/ui"
	"golang.org/x/oauth2"
)

type Handler struct {
	github *github.Client
}

func New(token string) *Handler {
	tc := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	return &Handler{
		github: github.NewClient(tc),
	}
}

// Root is the handler for the root command.
func (h *Handler) Root() error {
	opt := &github.RepositoryListOptions{
		Type:        "owner",
		ListOptions: github.ListOptions{},
	}
	// get all allRepos
	allRepos := make([]*github.Repository, 0)
	for i := 0; ; i++ {
		repos, resp, err := h.github.Repositories.List(context.Background(), "", opt)
		if err != nil {
			return err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	cands := make([]inc.Candidate, len(allRepos))
	for i, repo := range allRepos {
		cands[i] = inc.Candidate{
			Text: repo.GetName(),
			Ptr:  repo,
		}
	}

	e := inc.New("", cands)
	ui.RunSelector(e)

	return nil
}
