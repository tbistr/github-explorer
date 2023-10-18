package ui

import "github.com/charmbracelet/lipgloss"

type Pane struct {
	width, height               int
	contentWidth, contentHeight int
	paneStyle, contentStyle     lipgloss.Style
}

func (p Pane) Render(s string) string {
	return p.paneStyle.Render(p.contentStyle.Render(s))
}

func (p *Pane) SetSize(w, h int) {
	p.width = w
	p.height = h

	// -2 for the borders
	p.contentWidth = w - 2
	p.contentHeight = h - 2

	p.contentStyle = lipgloss.NewStyle().
		Height(p.contentHeight).Width(p.contentWidth).
		MaxHeight(p.contentHeight).MaxWidth(p.contentWidth)
	p.paneStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder())
}

func (p Pane) GetContentSize() (w int, h int) {
	return p.contentWidth, p.contentHeight
}
