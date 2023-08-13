package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v53/github"
	"github.com/tbistr/inc"
	"github.com/tbistr/inc/ui"
)

type repoSelector struct {
	client        *github.Client
	owner         string
	engine        *inc.Engine
	selector      ui.Model
	preview       string
	width, height int
	Canceled      bool
	Result        *github.Repository
	Error         error
}

var _ tea.Model = repoSelector{}

type fetchRepoMsg struct {
	repos    []*github.Repository
	nextPage int
	ok       bool
	err      error
}

func (rs repoSelector) fetchRepo(page int) tea.Cmd {
	return func() tea.Msg {
		opt := &github.RepositoryListOptions{
			Type:        "public",
			ListOptions: github.ListOptions{Page: page},
		}
		repos, resp, err := rs.client.Repositories.List(context.Background(), rs.owner, opt)
		return fetchRepoMsg{repos, resp.NextPage, resp.NextPage != 0, err}
	}
}

func newRepoSelector(client *github.Client, owner string) repoSelector {
	e := inc.New("", nil)
	rs := repoSelector{
		client:   client,
		owner:    owner,
		engine:   e,
		selector: ui.NewModel(e),
	}

	return rs
}

func (rs repoSelector) Init() tea.Cmd {
	return tea.Batch(rs.selector.Init(), rs.fetchRepo(0))
}

func (rs repoSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msgT := msg.(type) {
	case tea.WindowSizeMsg:
		rs.width = msgT.Width
		rs.height = msgT.Height
		msg = tea.WindowSizeMsg{
			Width:  msgT.Width/2 - 2,
			Height: msgT.Height - 2,
		}

	case fetchRepoMsg:
		if msgT.err != nil {
			rs.Error = msgT.err
			return rs, windowPreQuit
		}

		items := make([]inc.Candidate, len(msgT.repos))
		for i, repo := range msgT.repos {
			items[i] = inc.Candidate{
				Text: []rune(repo.GetName()),
				Ptr:  repo,
			}
		}
		rs.engine.AppendCands(items)
		if msgT.ok {
			return rs, rs.fetchRepo(msgT.nextPage)
		} else {
			return rs, nil
		}
	}

	selector, cmd := rs.selector.Update(msg)
	rs.selector = selector.(ui.Model)
	rs.Canceled = rs.selector.Canceled
	if s, ok := rs.selector.Selected.Ptr.(*github.Repository); ok {
		rs.Result = s
		rs.preview = rs.repoPreview(s)
	} else {
		rs.preview = ""
	}
	return rs, cmd
}

func (rs repoSelector) View() string {
	contentStyle := lipgloss.NewStyle().
		Height(rs.height - 2).
		Width(rs.width/2 - 2).
		MaxHeight(rs.height - 2).
		MaxWidth(rs.width/2 - 2)

	paneStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder())

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		paneStyle.Render(contentStyle.Render(rs.selector.View())),
		paneStyle.Render(contentStyle.Render(rs.preview)),
	)
}

func (rs repoSelector) repoPreview(repo *github.Repository) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		repo.GetName(),
		repo.GetDescription(),
		repo.GetHTMLURL(),
	)
}
