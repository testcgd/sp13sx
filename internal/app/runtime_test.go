package app

import (
	"context"
	"encoding/json"
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
	if requests[1].PreviousResponseID != "resp_1" {
		t.Fatalf("expected previous_response_id resp_1, got %q", requests[1].PreviousResponseID)
	}
	if len(requests[1].Input) != 1 || requests[1].Input[0].Type != "function_call_output" {
		t.Fatalf("expected function_call_output input, got %#v", requests[1].Input)
	}
	if !strings.Contains(requests[1].Input[0].Output, "\"echo\":\"hello\"") {
		t.Fatalf("expected tool output to be passed back, got %q", requests[1].Input[0].Output)
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

type fakeTool struct {
	name string
	run  func(context.Context, map[string]any) (map[string]any, error)
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
