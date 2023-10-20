package ui

import (
	"context"
	"errors"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v53/github"
	"github.com/tbistr/inc"
	"github.com/tbistr/inc/ui"
)

type fileSelector struct {
	client                        *github.Client
	repo                          *github.Repository
	engine                        *inc.Engine
	selector                      ui.Model
	preview                       string
	leftPaneStyle, rightPaneStyle Pane
	Canceled                      bool
	Result                        *github.TreeEntry
}

var _ tea.Model = fileSelector{}

type treeMsg struct {
	tree *github.Tree
	err  error
}

func (fs fileSelector) fetchTree(sha string) tea.Cmd {
	return func() tea.Msg {
		tree, _, err := fs.client.Git.GetTree(
			context.Background(),
			fs.repo.GetOwner().GetLogin(),
			fs.repo.GetName(),
			sha,
			false,
		)
		return treeMsg{tree, err}
	}
}

type contentMsg struct {
	content string
	err     error
}

func (fs fileSelector) fetchContent(entry *github.TreeEntry) tea.Cmd {
	return func() tea.Msg {
		if entry.GetType() != "blob" {
			return contentMsg{"", nil}
		}
		b, _, err := fs.client.Git.GetBlobRaw(
			context.Background(),
			fs.repo.GetOwner().GetLogin(),
			fs.repo.GetName(),
			entry.GetSHA(),
		)
		if err != nil {
			return contentMsg{"", err}
		}
		if !utf8.Valid(b) {
			return contentMsg{"", errors.New("invalid utf8 content")}
		}

		return contentMsg{string(b), nil}
	}
}

func newFileSelector(repo *github.Repository, client *github.Client) fileSelector {
	e := inc.New("", nil)
	return fileSelector{
		client:   client,
		repo:     repo,
		engine:   e,
		selector: ui.NewModel(e),
	}
}

func (fs fileSelector) Init() tea.Cmd {
	return tea.Batch(
		fs.selector.Init(),
		fs.fetchTree(fs.repo.GetDefaultBranch()),
	)
}

func (fs fileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msgT := msg.(type) {
	case tea.WindowSizeMsg:
		leftW := max(msgT.Width/3, 40)
		fs.leftPaneStyle.SetSize(leftW, msgT.Height)
		fs.rightPaneStyle.SetSize(msgT.Width-leftW, msgT.Height)
		w, h := fs.leftPaneStyle.GetContentSize()
		msg = tea.WindowSizeMsg{
			Width: w, Height: h,
		}

	case treeMsg:
		if msgT.err != nil {
			return fs, windowError(msgT.err)
		}
		items := make([]inc.Candidate, len(msgT.tree.Entries))
		for i, entry := range msgT.tree.Entries {
			var text string
			if entry.GetType() == "blob" {
				text = "📄 " + entry.GetPath()
			} else {
				text = "📁 " + entry.GetPath() + "/"
			}
			items[i] = inc.Candidate{
				Text: []rune(text),
				Ptr:  entry,
			}
		}
		fs.engine.DelQuery()
		fs.engine.DeleteCands()
		fs.engine.AppendCands(items)

	case contentMsg:
		if msgT.err != nil {
			return fs, windowError(msgT.err)
		}
		fs.preview = msgT.content
	}

	old := fs.selector.Selected.Ptr
	selector, cmd := fs.selector.Update(msg)
	fs.selector = selector.(ui.Model)
	fs.Canceled = fs.selector.Canceled
	if s, ok := fs.selector.Selected.Ptr.(*github.TreeEntry); ok {
		fs.Result = s
	}
	if old != fs.selector.Selected.Ptr {
		return fs, tea.Batch(cmd, fs.fetchContent(fs.Result))
	}

	return fs, cmd
}

func (fs fileSelector) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		fs.leftPaneStyle.Render(fs.selector.View()),
		fs.rightPaneStyle.Render(fs.preview),
	)
}
