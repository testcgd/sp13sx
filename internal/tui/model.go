package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"sp13sx/internal/domain"
	"sp13sx/internal/llm"
)

type Runtime interface {
	Send(ctx context.Context, text string) (<-chan llm.StreamEvent, error)
	BackendName() string
	ModelName() string
	SessionTitle() string
	EnabledSkillNames() []string
	DiscoveredSkillNames() []string
	MCPStatusLines() []string
	RightPaneWidth() int
}

type Model struct {
	runtime   Runtime
	viewport  viewport.Model
	input     textarea.Model
	width     int
	height    int
	messages  []domain.Message
	status    string
	err       error
	rightPane []string
}

type responseMsg struct {
	event llm.StreamEvent
}

func NewModel(runtime Runtime) Model {
	vp := viewport.New(80, 20)
	ta := textarea.New()
	ta.Placeholder = "Ask the agent or run /help"
	ta.Focus()
	ta.SetHeight(3)
	return Model{
		runtime:   runtime,
		viewport:  vp,
		input:     ta,
		status:    "ready",
		rightPane: buildRightPane(runtime),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func streamCmd(ch <-chan llm.StreamEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return nil
		}
		return responseMsg{event: event}
	}
}

func (m *Model) appendMessage(role string, content string) {
	m.messages = append(m.messages, domain.Message{
		Role:    role,
		Content: []domain.ContentPart{{Type: "text", Text: content}},
	})
	m.viewport.SetContent(renderMessages(m.messages))
	m.viewport.GotoBottom()
}

func renderMessages(messages []domain.Message) string {
	var b strings.Builder
	for _, msg := range messages {
		b.WriteString(strings.ToUpper(msg.Role))
		b.WriteString("\n")
		for _, part := range msg.Content {
			b.WriteString(part.Text)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func buildRightPane(runtime Runtime) []string {
	lines := []string{
		"backend: " + runtime.BackendName(),
		"model: " + runtime.ModelName(),
		"session: " + runtime.SessionTitle(),
		"skills:",
	}
	for _, skill := range runtime.EnabledSkillNames() {
		lines = append(lines, "  - "+skill)
	}
	lines = append(lines, "mcp:")
	for _, server := range runtime.MCPStatusLines() {
		lines = append(lines, "  - "+server)
	}
	return lines
}

func (m *Model) submitInput() tea.Cmd {
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		return nil
	}
	m.input.Reset()
	if strings.HasPrefix(text, "/") {
		return m.handleCommand(text)
	}

	m.appendMessage("user", text)
	m.status = "waiting for assistant"
	stream, err := m.runtime.Send(context.Background(), text)
	if err != nil {
		m.err = err
		m.status = "error"
		return nil
	}
	return streamCmd(stream)
}

// Status 返回当前状态（用于测试）
func (m Model) Status() string {
	return m.status
}

// Messages 返回消息列表（用于测试）
func (m Model) Messages() []domain.Message {
	return m.messages
}

// Error 返回当前错误（用于测试）
func (m Model) Error() error {
	return m.err
}
