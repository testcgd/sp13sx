package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	rightWidth := m.runtime.RightPaneWidth()
	if rightWidth <= 0 {
		rightWidth = 40
	}

	left := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		Width(max(20, m.width-rightWidth-4)).
		Render(m.viewport.View() + "\n\n" + m.input.View())

	status := "status: " + m.status
	if m.err != nil {
		status += "\nerror: " + m.err.Error()
	}

	rightLines := append([]string{status, "", "context:"}, m.rightPane...)
	right := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		Width(rightWidth).
		Render(strings.Join(rightLines, "\n"))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}
