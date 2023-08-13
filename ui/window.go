package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v53/github"
)

func (w *Window) Run() (Result, error) {
	filter := func(m tea.Model, msg tea.Msg) tea.Msg {
		if _, ok := msg.(tea.QuitMsg); ok {
			return windowPreQuit()
		}
		if _, ok := msg.(windowQuitMsg); ok {
			return tea.Quit()
		}
		return msg
	}
	m, err := tea.NewProgram(w, tea.WithAltScreen(), tea.WithFilter(filter)).Run()
	if err != nil {
		return Result{}, err
	}
	return m.(Window).Result, nil
}

type windowPreQuitMsg struct{}

func windowPreQuit() tea.Msg { return windowPreQuitMsg{} }

type windowQuitMsg struct{}

func windowQuit() tea.Msg { return windowQuitMsg{} }

type windowErrorMsg error

func windowError(err error) tea.Cmd { return func() tea.Msg { return windowErrorMsg(err) } }

type Window struct {
	header          string
	isRepoSelection bool
	repoSelector    repoSelector
	fileSelector    fileSelector
	footer          string
	width           int
	height          int
	Result          Result
}

type Result struct {
	Repo         *github.Repository
	Entry        *github.TreeEntry
	Branch, Path string
	Canceled     bool
	Error        error
}

func NewWindow(client *github.Client, owner string) *Window {
	w := &Window{
		isRepoSelection: true,
		repoSelector:    newRepoSelector(client, owner),
	}
	return w
}

var _ tea.Model = Window{}

func (w Window) Init() tea.Cmd {
	return w.repoSelector.Init()
}

func (w Window) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height

	case windowErrorMsg:
		w.Result.Error = error(msg)
		return w, windowQuit

	case windowPreQuitMsg:
		if w.repoSelector.Canceled || w.fileSelector.Canceled {
			w.Result.Canceled = true
			return w, windowQuit
		}

		if w.isRepoSelection {
			w.Result.Repo = w.repoSelector.Result
			w.isRepoSelection = false
			w.fileSelector = newFileSelector(w.Result.Repo, w.repoSelector.client)

			cmds = append(
				cmds,
				w.fileSelector.Init(),
				func() tea.Msg {
					return tea.WindowSizeMsg{
						Width:  w.width,
						Height: w.height,
					}
				})
		} else {
			if w.fileSelector.Result.GetType() == "blob" {
				w.Result.Entry = w.fileSelector.Result
				return w, windowQuit
			}

			return w, w.fileSelector.fetchTree(w.fileSelector.Result.GetSHA())
		}
	}

	if w.isRepoSelection {
		selector, cmd := w.repoSelector.Update(msg)
		w.repoSelector = selector.(repoSelector)
		cmds = append(cmds, cmd)
	} else {
		selector, cmd := w.fileSelector.Update(msg)
		w.fileSelector = selector.(fileSelector)
		cmds = append(cmds, cmd)
	}

	return w, tea.Batch(cmds...)
}

func (w Window) View() string {
	if w.isRepoSelection {
		return w.repoSelector.View()
	} else {
		return w.fileSelector.View()
	}
}
