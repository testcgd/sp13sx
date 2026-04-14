package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

func TestInputQueuedEvent(t *testing.T) {
	model := NewModel(fakeRuntime{})

	updated, _ := model.Update(responseMsg{
		event: llm.StreamEvent{
			Type:    "input_queued",
			Content: "user input",
		},
	})
	m := updated.(Model)
	if len(m.pendingInputs) != 1 {
		t.Fatalf("expected 1 pending input, got %d", len(m.pendingInputs))
	}
	if m.pendingInputs[0] != "user input" {
		t.Fatalf("expected 'user input', got %q", m.pendingInputs[0])
	}
	if m.interruptMode {
		t.Fatal("expected interrupt mode to be false")
	}
}

func TestMultipleQueuedInputs(t *testing.T) {
	model := NewModel(fakeRuntime{})

	updated, _ := model.Update(responseMsg{
		event: llm.StreamEvent{Type: "input_queued", Content: "first"},
	})
	updated, _ = updated.(Model).Update(responseMsg{
		event: llm.StreamEvent{Type: "input_queued", Content: "second"},
	})
	m := updated.(Model)
	if len(m.pendingInputs) != 2 {
		t.Fatalf("expected 2 pending inputs, got %d", len(m.pendingInputs))
	}
	if m.pendingInputs[0] != "first" || m.pendingInputs[1] != "second" {
		t.Fatalf("expected FIFO order, got %v", m.pendingInputs)
	}
}

func TestEscapeTogglesInterruptMode(t *testing.T) {
	cancelCalled := false
	runtime := fakeRuntime{
		cancel: func() { cancelCalled = true },
	}
	model := NewModel(runtime)
	model.pendingInputs = []string{"queued input"}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m := updated.(Model)
	if !m.interruptMode {
		t.Fatal("expected interrupt mode to be true")
	}
	if !cancelCalled {
		t.Fatal("expected Cancel to be called")
	}
}

func TestEscapeTogglesBackToQueueMode(t *testing.T) {
	model := NewModel(fakeRuntime{})
	model.pendingInputs = []string{"queued input"}
	model.interruptMode = true

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m := updated.(Model)
	if m.interruptMode {
		t.Fatal("expected interrupt mode to be false")
	}
}

func TestRenderPendingInputs(t *testing.T) {
	model := NewModel(fakeRuntime{})
	model.pendingInputs = []string{"first input", "second input with longer text"}

	view := model.renderPendingInputs()
	if !strings.Contains(view, "first input") {
		t.Fatal("expected first input in view")
	}
	if !strings.Contains(view, "second input") {
		t.Fatal("expected second input in view")
	}
	if !strings.Contains(view, "⏳") {
		t.Fatal("expected queue icon in view")
	}
}

func TestRenderPendingInputsWithInterruptMode(t *testing.T) {
	model := NewModel(fakeRuntime{})
	model.pendingInputs = []string{"first input", "second input"}
	model.interruptMode = true

	view := model.renderPendingInputs()
	if !strings.Contains(view, "⚡") {
		t.Fatal("expected interrupt icon in view")
	}
}

type fakeRuntime struct {
	send   func(context.Context, string) (<-chan llm.StreamEvent, error)
	cancel func()
	status llm.RuntimeStatus
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
func (f fakeRuntime) Cancel() {
	if f.cancel != nil {
		f.cancel()
	}
}
func (f fakeRuntime) Status() llm.RuntimeStatus {
	if f.status.IsRunningTool || len(f.status.PendingInputs) > 0 {
		return f.status
	}
	return llm.RuntimeStatus{IsRunningTool: false, PendingInputs: nil}
}
