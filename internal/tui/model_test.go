package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"sp13sx/internal/llm"
)

func TestSubmitInputStartsAssistantStream(t *testing.T) {
	runtime := fakeRuntime{
		send: func(_ context.Context, text string) (<-chan llm.StreamEvent, error) {
			if text != "hello" {
				t.Fatalf("expected input hello, got %q", text)
			}
			ch := make(chan llm.StreamEvent, 2)
			ch <- llm.StreamEvent{Type: "message", Content: "hi"}
			close(ch)
			return ch, nil
		},
	}

	model := NewModel(runtime)
	model.input.SetValue("hello")

	cmd := model.submitInput()
	if cmd == nil {
		t.Fatalf("expected command from submitInput")
	}
	msg := cmd()
	resp, ok := msg.(responseMsg)
	if !ok {
		t.Fatalf("expected responseMsg, got %T", msg)
	}
	if resp.event.Content != "hi" {
		t.Fatalf("expected first streamed content hi, got %q", resp.event.Content)
	}
}

func TestUpdateHandlesToolCallAndStatus(t *testing.T) {
	model := NewModel(fakeRuntime{})

	updated, _ := model.Update(responseMsg{
		event: llm.StreamEvent{
			Type: "tool_call",
			ToolCall: &llm.ToolCall{
				Name: "demo_tool",
			},
		},
	})
	m := updated.(Model)
	if m.status != "running tool" {
		t.Fatalf("expected running tool status, got %q", m.status)
	}
	if len(m.messages) == 0 || !strings.Contains(m.messages[len(m.messages)-1].Content[0].Text, "demo_tool") {
		t.Fatalf("expected tool call message to be appended")
	}

	updated, _ = m.Update(responseMsg{
		event: llm.StreamEvent{
			Type:    "status",
			Content: "tool completed: demo_tool",
		},
	})
	m = updated.(Model)
	if len(m.messages) < 2 {
		t.Fatalf("expected status message to be appended")
	}
}

func TestUpdateHandlesError(t *testing.T) {
	model := NewModel(fakeRuntime{})

	updated, _ := model.Update(responseMsg{
		event: llm.StreamEvent{
			Error: errors.New("boom"),
		},
	})
	m := updated.(Model)
	if m.status != "error" {
		t.Fatalf("expected error status, got %q", m.status)
	}
	if m.err == nil || m.err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", m.err)
	}
}

func TestHandleCommandHelp(t *testing.T) {
	model := NewModel(fakeRuntime{})

	cmd := model.handleCommand("/help")
	if cmd != nil {
		t.Fatalf("expected nil command for /help")
	}
	if len(model.messages) == 0 {
		t.Fatalf("expected help message to be appended")
	}
}

type fakeRuntime struct {
	send func(context.Context, string) (<-chan llm.StreamEvent, error)
}

func (f fakeRuntime) Send(ctx context.Context, text string) (<-chan llm.StreamEvent, error) {
	if f.send == nil {
		ch := make(chan llm.StreamEvent)
		close(ch)
		return ch, nil
	}
	return f.send(ctx, text)
}

func (fakeRuntime) BackendName() string            { return "mock" }
func (fakeRuntime) ModelName() string              { return "mock-model" }
func (fakeRuntime) SessionTitle() string           { return "test-session" }
func (fakeRuntime) EnabledSkillNames() []string    { return []string{"skill-a"} }
func (fakeRuntime) DiscoveredSkillNames() []string { return []string{"skill-a", "skill-b"} }
func (fakeRuntime) MCPStatusLines() []string       { return []string{"filesystem [connected]"} }
func (fakeRuntime) RightPaneWidth() int            { return 40 }
