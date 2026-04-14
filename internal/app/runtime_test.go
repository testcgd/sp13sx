package app

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"sp13sx/internal/config"
	"sp13sx/internal/domain"
	"sp13sx/internal/llm"
	mockllm "sp13sx/internal/llm/mock"
	"sp13sx/internal/store"
	"sp13sx/internal/tools"
)

func TestRunTurnLoopToolRoundTrip(t *testing.T) {
	tempDir := t.TempDir()

	var requests []llm.GenerateRequest
	backend := mockllm.NewBackend(func(_ context.Context, req llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
		requests = append(requests, req)
		ch := make(chan llm.StreamEvent, 8)
		if len(requests) == 1 {
			ch <- llm.StreamEvent{Type: "response_id", ResponseID: "resp_1"}
			ch <- llm.StreamEvent{
				Type: "tool_call",
				ToolCall: &llm.ToolCall{
					ID:        "item_1",
					CallID:    "call_1",
					Name:      "echo_tool",
					Arguments: map[string]any{"value": "hello"},
				},
			}
			close(ch)
			return ch, nil
		}
		ch <- llm.StreamEvent{Type: "response_id", ResponseID: "resp_2"}
		ch <- llm.StreamEvent{Type: "message", Content: "done"}
		close(ch)
		return ch, nil
	})

	registry := tools.NewRegistry()
	registry.Register(fakeTool{
		name: "echo_tool",
		run: func(_ context.Context, args map[string]any) (map[string]any, error) {
			return map[string]any{"echo": args["value"]}, nil
		},
	})

	runtime := &Runtime{
		Config: config.Config{
			UI: config.UIConfig{RightPaneWidth: 40},
			Defaults: config.DefaultsConfig{
				EnabledSkills: []string{},
			},
		},
		StorePaths: store.Paths{
			MessagesPath:  filepath.Join(tempDir, "messages.jsonl"),
			ToolCallsPath: filepath.Join(tempDir, "tool_calls.jsonl"),
		},
		Backend: backend,
		Tools:   registry,
		Session: testSession(),
	}

	req := llm.GenerateRequest{
		Model:        "mock-model",
		Instructions: "base",
		Input:        []llm.InputItem{llm.UserTextInput("hi")},
		Tools:        buildToolDefinitions(registry),
	}
	out := make(chan llm.StreamEvent, 16)
	go runtime.runTurnLoop(context.Background(), req, out)

	var events []llm.StreamEvent
	for event := range out {
		events = append(events, event)
	}

	if len(requests) != 2 {
		t.Fatalf("expected 2 backend requests, got %d", len(requests))
	}

	if len(requests[1].Input) != 2 {
		t.Fatalf("expected 2 input items (assistant_context + tool_result), got %d", len(requests[1].Input))
	}
	if requests[1].Input[0].Type != "assistant_context" {
		t.Fatalf("expected first input to be assistant_context, got %q", requests[1].Input[0].Type)
	}
	if requests[1].Input[1].Type != "tool_result" {
		t.Fatalf("expected second input to be tool_result, got %q", requests[1].Input[1].Type)
	}
	if requests[1].Input[1].ToolName != "echo_tool" {
		t.Fatalf("expected tool name echo_tool, got %q", requests[1].Input[1].ToolName)
	}

	if !strings.Contains(requests[1].Input[1].Content, "\"echo\":\"hello\"") {
		t.Fatalf("expected tool output to be passed back, got %q", requests[1].Input[1].Content)
	}

	foundToolCall := false
	foundDone := false
	for _, event := range events {
		if event.Type == "tool_call" && event.ToolCall != nil && event.ToolCall.Name == "echo_tool" {
			foundToolCall = true
		}
		if event.Type == "message" && event.Content == "done" {
			foundDone = true
		}
	}
	if !foundToolCall {
		t.Fatalf("expected tool_call event in output")
	}
	if !foundDone {
		t.Fatalf("expected final assistant message in output")
	}

	toolRows, err := store.ReadAll(runtime.StorePaths.ToolCallsPath)
	if err != nil {
		t.Fatalf("read tool calls: %v", err)
	}
	if len(toolRows) != 2 {
		t.Fatalf("expected 2 tool call records, got %d", len(toolRows))
	}
}

func TestRunTurnLoopPassesReasoningContent(t *testing.T) {
	tempDir := t.TempDir()

	var requests []llm.GenerateRequest
	backend := mockllm.NewBackend(func(_ context.Context, req llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
		requests = append(requests, req)
		ch := make(chan llm.StreamEvent, 16)
		if len(requests) == 1 {
			ch <- llm.StreamEvent{Type: "reasoning", ReasoningContent: "Let me think..."}
			ch <- llm.StreamEvent{Type: "reasoning", ReasoningContent: " I need to use a tool."}
			ch <- llm.StreamEvent{
				Type: "tool_call",
				ToolCall: &llm.ToolCall{
					ID:        "call_1",
					CallID:    "call_1",
					Name:      "test_tool",
					Arguments: map[string]any{"query": "test"},
				},
			}
			close(ch)
			return ch, nil
		}
		ch <- llm.StreamEvent{Type: "message", Content: "final answer"}
		close(ch)
		return ch, nil
	})

	registry := tools.NewRegistry()
	registry.Register(fakeTool{
		name: "test_tool",
		run: func(_ context.Context, args map[string]any) (map[string]any, error) {
			return map[string]any{"result": args["query"]}, nil
		},
	})

	runtime := &Runtime{
		Config: config.Config{
			UI: config.UIConfig{RightPaneWidth: 40},
			Defaults: config.DefaultsConfig{
				EnabledSkills: []string{},
			},
		},
		StorePaths: store.Paths{
			MessagesPath:  filepath.Join(tempDir, "messages.jsonl"),
			ToolCallsPath: filepath.Join(tempDir, "tool_calls.jsonl"),
		},
		Backend: backend,
		Tools:   registry,
		Session: testSession(),
	}

	req := llm.GenerateRequest{
		Model:        "mock-model",
		Instructions: "base",
		Input:        []llm.InputItem{llm.UserTextInput("test")},
		Tools:        buildToolDefinitions(registry),
	}
	out := make(chan llm.StreamEvent, 32)
	go runtime.runTurnLoop(context.Background(), req, out)

	for range out {
	}

	if len(requests) != 2 {
		t.Fatalf("expected 2 backend requests, got %d", len(requests))
	}

	if len(requests[1].Input) != 2 {
		t.Fatalf("expected 2 input items in second request, got %d", len(requests[1].Input))
	}

	assistantCtx := requests[1].Input[0]
	if assistantCtx.Type != "assistant_context" {
		t.Fatalf("expected first input to be assistant_context, got %q", assistantCtx.Type)
	}
	if assistantCtx.ReasoningContent != "Let me think... I need to use a tool." {
		t.Fatalf("expected reasoning_content to be accumulated, got %q", assistantCtx.ReasoningContent)
	}
	if len(assistantCtx.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool_call in assistant_context, got %d", len(assistantCtx.ToolCalls))
	}
	if assistantCtx.ToolCalls[0].Name != "test_tool" {
		t.Fatalf("expected tool_call name 'test_tool', got %q", assistantCtx.ToolCalls[0].Name)
	}
}

func TestPersistAssistantStreamWritesFinalMessage(t *testing.T) {
	tempDir := t.TempDir()
	msgPath := filepath.Join(tempDir, "messages.jsonl")

	in := make(chan llm.StreamEvent, 4)
	in <- llm.StreamEvent{Type: "message", Content: "hello "}
	in <- llm.StreamEvent{Type: "message", Content: "world"}
	close(in)

	out := persistAssistantStream(msgPath, "sess_1", in)
	for range out {
	}

	rows, err := store.ReadAll(msgPath)
	if err != nil {
		t.Fatalf("read messages: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 persisted assistant message, got %d", len(rows))
	}

	var rec store.MessageRecord
	if err := json.Unmarshal(rows[0], &rec); err != nil {
		t.Fatalf("unmarshal record: %v", err)
	}
	if got := rec.Payload.Content[0].Text; got != "hello world" {
		t.Fatalf("expected aggregated assistant text, got %q", got)
	}
}

func TestRuntimeCloseClosesBackend(t *testing.T) {
	backend := &closerBackend{}
	runtime := &Runtime{Backend: backend}

	if err := runtime.Close(); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
	if !backend.closed {
		t.Fatal("expected runtime close to close backend")
	}
}

func TestRuntimeCloseReturnsBackendError(t *testing.T) {
	backend := &closerBackend{err: errors.New("close failed")}
	runtime := &Runtime{Backend: backend}

	err := runtime.Close()
	if err == nil || !strings.Contains(err.Error(), "close failed") {
		t.Fatalf("expected close error, got %v", err)
	}
}

func TestRuntimeSendQueuesInputWhenToolRunning(t *testing.T) {
	tempDir := t.TempDir()

	backend := mockllm.NewBackend(func(ctx context.Context, req llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
		ch := make(chan llm.StreamEvent, 8)
		go func() {
			defer close(ch)
			ch <- llm.StreamEvent{Type: "response_id", ResponseID: "resp_1"}
			ch <- llm.StreamEvent{
				Type: "tool_call",
				ToolCall: &llm.ToolCall{
					ID:        "item_1",
					CallID:    "call_1",
					Name:      "slow_tool",
					Arguments: map[string]any{},
				},
			}
			<-ctx.Done()
		}()
		return ch, nil
	})

	registry := tools.NewRegistry()
	registry.Register(fakeTool{
		name: "slow_tool",
		run: func(ctx context.Context, args map[string]any) (map[string]any, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	})

	runtime := &Runtime{
		Config: config.Config{
			UI: config.UIConfig{RightPaneWidth: 40},
			Defaults: config.DefaultsConfig{
				EnabledSkills: []string{},
			},
		},
		StorePaths: store.Paths{
			MessagesPath:  filepath.Join(tempDir, "messages.jsonl"),
			ToolCallsPath: filepath.Join(tempDir, "tool_calls.jsonl"),
		},
		Backend: backend,
		Tools:   registry,
		Session: testSession(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := runtime.Send(ctx, "first message")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	<-stream

	runtime.mu.Lock()
	runtime.isRunningTool = true
	runtime.mu.Unlock()

	queuedStream, err := runtime.Send(context.Background(), "queued message")
	if err != nil {
		t.Fatalf("Send during tool execution failed: %v", err)
	}

	event := <-queuedStream
	if event.Type != "input_queued" {
		t.Fatalf("expected input_queued event, got %q", event.Type)
	}
	if event.Content != "queued message" {
		t.Fatalf("expected queued message content, got %q", event.Content)
	}

	status := runtime.Status()
	if len(status.PendingInputs) != 1 {
		t.Fatalf("expected 1 pending input, got %d", len(status.PendingInputs))
	}
	if status.PendingInputs[0] != "queued message" {
		t.Fatalf("expected 'queued message', got %q", status.PendingInputs[0])
	}
}

func TestRuntimeCancelCancelsCurrentContext(t *testing.T) {
	cancelCalled := false
	runtime := &Runtime{
		currentCancel: func() { cancelCalled = true },
	}

	runtime.Cancel()
	if !cancelCalled {
		t.Fatal("expected Cancel to call currentCancel")
	}
}

func TestRuntimeStatusReturnsCorrectState(t *testing.T) {
	runtime := &Runtime{
		isRunningTool: true,
		pendingInputs: []string{"pending1", "pending2"},
	}

	status := runtime.Status()
	if !status.IsRunningTool {
		t.Fatal("expected IsRunningTool to be true")
	}
	if len(status.PendingInputs) != 2 {
		t.Fatalf("expected 2 pending inputs, got %d", len(status.PendingInputs))
	}
}

type fakeTool struct {
	name string
	run  func(context.Context, map[string]any) (map[string]any, error)
}

type closerBackend struct {
	closed bool
	err    error
}

func (b *closerBackend) Name() string { return "closer" }

func (b *closerBackend) Generate(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent)
	close(ch)
	return ch, nil
}

func (b *closerBackend) Close() error {
	b.closed = true
	return b.err
}

func (f fakeTool) Name() string           { return f.name }
func (f fakeTool) Description() string    { return "fake tool" }
func (f fakeTool) Schema() map[string]any { return map[string]any{"type": "object"} }
func (f fakeTool) Run(ctx context.Context, args map[string]any) (map[string]any, error) {
	return f.run(ctx, args)
}

func testSession() domain.Session {
	return domain.Session{
		ID:      "sess_test",
		Title:   "Test Session",
		Backend: "mock",
		Model:   "mock-model",
	}
}
