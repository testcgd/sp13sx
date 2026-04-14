package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"sp13sx/internal/domain"
	"sp13sx/internal/util"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = max(20, msg.Width-m.runtime.RightPaneWidth()-6)
		m.viewport.Height = max(10, msg.Height-8)
		m.input.SetWidth(m.viewport.Width)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			return m, m.submitInput()
		case tea.KeyRunes:
			if string(msg.Runes) == "r" {
				m.toggleLastReasoning()
				return m, nil
			}
		}
	case responseMsg:
		if msg.event.Error != nil {
			m.err = msg.event.Error
			m.status = "error"
			return m, nil
		}
		switch msg.event.Type {
		case "message":
			if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "assistant" {
				last := &m.messages[len(m.messages)-1]
				last.Content[0].Text += msg.event.Content
			} else {
				m.appendMessage("assistant", msg.event.Content)
			}
			m.viewport.SetContent(renderMessages(m.messages, m.reasoningExpanded))
			m.viewport.GotoBottom()
			m.status = "ready"
		case "reasoning":
			if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "assistant" {
				last := &m.messages[len(m.messages)-1]
				if len(last.Reasoning) == 0 {
					last.Reasoning = []domain.ContentPart{{Type: "text", Text: msg.event.ReasoningContent}}
				} else {
					last.Reasoning[0].Text += msg.event.ReasoningContent
				}
			} else {
				m.messages = append(m.messages, domain.Message{
					ID:        util.NewID("msg"),
					Role:      "assistant",
					Reasoning: []domain.ContentPart{{Type: "text", Text: msg.event.ReasoningContent}},
				})
			}
			m.viewport.SetContent(renderMessages(m.messages, m.reasoningExpanded))
			m.viewport.GotoBottom()
			m.status = "streaming"
		case "tool_call":
			if msg.event.ToolCall != nil {
				m.appendMessage("system", "Tool requested: "+msg.event.ToolCall.Name)
				m.status = "running tool"
			}
		case "status":
			m.appendMessage("system", msg.event.Content)
		case "response_id":
			m.status = "streaming"
		}
		return m, nil
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) toggleLastReasoning() {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if len(m.messages[i].Reasoning) > 0 {
			if m.messages[i].ID == "" {
				m.messages[i].ID = util.NewID("msg")
			}
			m.reasoningExpanded[m.messages[i].ID] = !m.reasoningExpanded[m.messages[i].ID]
			m.viewport.SetContent(renderMessages(m.messages, m.reasoningExpanded))
			break
		}
	}
}

func (m *Model) handleCommand(command string) tea.Cmd {
	switch strings.TrimSpace(command) {
	case "/help":
		m.appendMessage("system", "Commands: /help, /session list, /skill list, /mcp list")
	case "/session list":
		m.appendMessage("system", "Current session: "+m.runtime.SessionTitle())
	case "/skill list":
		skills := m.runtime.DiscoveredSkillNames()
		if len(skills) == 0 {
			m.appendMessage("system", "No skills discovered.")
			break
		}
		m.appendMessage("system", "Skills: "+strings.Join(skills, ", "))
	case "/mcp list":
		lines := m.runtime.MCPStatusLines()
		if len(lines) == 0 {
			lines = append(lines, "no configured servers")
		}
		m.appendMessage("system", "MCP: "+strings.Join(lines, ", "))
	default:
		m.appendMessage("system", "Unknown command: "+command)
	}
	return nil
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
