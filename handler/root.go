package handler

import (
	"context"
	"fmt"

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
	allRepos := []*github.Repository{}
	e := inc.New("", nil)

	fetch := func(page int) (next int, ok bool, err error) {
		opt := &github.RepositoryListOptions{
			Type:        "owner",
			ListOptions: github.ListOptions{Page: page},
		}

		repos, resp, err := h.github.Repositories.List(context.Background(), "", opt)
		if err != nil {
			return 0, false, err
		}

		cands := []inc.Candidate{}
		for _, repo := range repos {
			cands = append(cands, inc.Candidate{
				Text: []rune(repo.GetName()),
				Ptr:  repo,
			})
		}
		e.AppendCands(cands)
		allRepos = append(allRepos, repos...)

		return resp.NextPage, resp.NextPage != 0, nil
	}

	next, ok, err := fetch(0)
	go func() {
		for ok {
			next, ok, err = fetch(next)
		}
	}()
	canceled, selected, err := ui.RunSelector(e)
	if canceled || (err != nil) {
		return err
	}
	fmt.Println(selected.Ptr.(*github.Repository).GetURL())

	return nil
}
