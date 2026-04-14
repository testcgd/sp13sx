package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	rightWidth := m.runtime.RightPaneWidth()
	if rightWidth <= 0 {
		rightWidth = 40
	}

	leftContent := m.viewport.View()
	if len(m.pendingInputs) > 0 {
		leftContent += "\n" + m.renderPendingInputs()
	}
	leftContent += "\n\n" + m.input.View()

	left := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		Width(max(20, m.width-rightWidth-4)).
		Render(leftContent)

	status := "status: " + m.status
	if m.err != nil {
		status += "\nerror: " + m.err.Error()
	}
	if len(m.pendingInputs) > 0 {
		status += fmt.Sprintf("\nqueued: %d", len(m.pendingInputs))
	}

	rightLines := append([]string{status, "", "context:"}, m.rightPane...)

	if len(m.events) > 0 {
		rightLines = append(rightLines, "", "recent events:")
		for _, event := range m.events {
			if len(event) > 35 {
				event = event[:32] + "..."
			}
			rightLines = append(rightLines, "  "+event)
		}
	}

	right := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		Width(rightWidth).
		Render(strings.Join(rightLines, "\n"))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) renderPendingInputs() string {
	var b strings.Builder
	b.WriteString("┌─ Queued Inputs ─────────────────────┐\n")
	for i, input := range m.pendingInputs {
		prefix := "⏳ "
		if m.interruptMode && i == 0 {
			prefix = "⚡ "
		}
		display := input
		if len(display) > 36 {
			display = display[:33] + "..."
		}
		b.WriteString(fmt.Sprintf("│ %s%s\n", prefix, display))
	}
	b.WriteString("└─────────────────────────────────────┘")
	return b.String()
}
