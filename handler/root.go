package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v53/github"
	"github.com/tbistr/ghe/ui"
	"golang.org/x/oauth2"
)

type Handler struct {
	github *github.Client
}

func New(token string) *Handler {
	var tc *http.Client
	if token != "" {
		tc = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))
	}
	return &Handler{
		github: github.NewClient(tc),
	}
}

// Root is the handler for the root command.
func (h *Handler) Root() error {
	w := ui.NewWindow(h.github, "tbistr")
	result, err := w.Run()
	if err != nil {
		return err
	}
	if result.Canceled {
		return nil
	}

	b, _, _ := h.github.Git.GetBlobRaw(context.Background(), result.Repo.GetOwner().GetLogin(), result.Repo.GetName(), result.Entry.GetSHA())
	fmt.Println(string(b))

	return nil
}
