package openai

import (
	"encoding/json"
	"testing"

	"sp13sx/internal/llm"
)

func TestProcessStreamMultipleToolCalls(t *testing.T) {
	chunks := []chatCompletionChunk{
		{
			ID:      "chat_1",
			Object:  "chat.completion.chunk",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []struct {
				Index        int          `json:"index"`
				Delta        deltaContent `json:"delta"`
				FinishReason string       `json:"finish_reason"`
			}{
				{
					Index: 0,
					Delta: deltaContent{
						Role: "assistant",
						ToolCalls: []toolCallMsg{
							{Index: 0, ID: "call_1", Type: "function", Function: functionMsg{Name: "read_file", Arguments: `{"path":"README.md"}`}},
						},
					},
				},
			},
		},
		{
			ID:      "chat_1",
			Object:  "chat.completion.chunk",
			Created: 1234567891,
			Model:   "gpt-4",
			Choices: []struct {
				Index        int          `json:"index"`
				Delta        deltaContent `json:"delta"`
				FinishReason string       `json:"finish_reason"`
			}{
				{
					Index: 0,
					Delta: deltaContent{
						ToolCalls: []toolCallMsg{
							{Index: 1, ID: "call_2", Type: "function", Function: functionMsg{Name: "write_file", Arguments: `{"path":"output.txt"}`}},
						},
					},
				},
			},
		},
		{
			ID:      "chat_1",
			Object:  "chat.completion.chunk",
			Created: 1234567892,
			Model:   "gpt-4",
			Choices: []struct {
				Index        int          `json:"index"`
				Delta        deltaContent `json:"delta"`
				FinishReason string       `json:"finish_reason"`
			}{
				{
					Index:        0,
					FinishReason: "tool_calls",
				},
			},
		},
	}

	stream := make(chan []byte, len(chunks))
	errs := make(chan error, 1)
	out := make(chan llm.StreamEvent, 16)

	for _, chunk := range chunks {
		data, err := json.Marshal(chunk)
		if err != nil {
			t.Fatalf("marshal chunk: %v", err)
		}
		stream <- data
	}
	close(stream)
	errs <- nil
	close(errs)

	b := &ChatBackend{
		name:     "test",
		messages: make([]chatMessage, 0),
	}
	b.processStream(stream, errs, out, "")

	var toolCalls []*llm.ToolCall
	for event := range out {
		if event.Type == "error" {
			t.Fatalf("unexpected error: %v", event.Error)
		}
		if event.Type == "tool_call" && event.ToolCall != nil {
			toolCalls = append(toolCalls, event.ToolCall)
		}
	}

	if len(toolCalls) != 2 {
		t.Fatalf("expected 2 tool calls, got %d", len(toolCalls))
	}

	tc0 := toolCalls[0]
	if tc0.ID != "call_1" {
		t.Errorf("expected tool call 0 ID 'call_1', got %q", tc0.ID)
	}
	if tc0.Name != "read_file" {
		t.Errorf("expected tool call 0 Name 'read_file', got %q", tc0.Name)
	}
	if tc0.Arguments["path"] != "README.md" {
		t.Errorf("expected tool call 0 Arguments[path] 'README.md', got %v", tc0.Arguments["path"])
	}

	tc1 := toolCalls[1]
	if tc1.ID != "call_2" {
		t.Errorf("expected tool call 1 ID 'call_2', got %q", tc1.ID)
	}
	if tc1.Name != "write_file" {
		t.Errorf("expected tool call 1 Name 'write_file', got %q", tc1.Name)
	}
	if tc1.Arguments["path"] != "output.txt" {
		t.Errorf("expected tool call 1 Arguments[path] 'output.txt', got %v", tc1.Arguments["path"])
	}
}

func TestProcessStreamSingleToolCall(t *testing.T) {
	chunks := []chatCompletionChunk{
		{
			ID:      "chat_1",
			Object:  "chat.completion.chunk",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []struct {
				Index        int          `json:"index"`
				Delta        deltaContent `json:"delta"`
				FinishReason string       `json:"finish_reason"`
			}{
				{
					Index: 0,
					Delta: deltaContent{
						Role: "assistant",
						ToolCalls: []toolCallMsg{
							{Index: 0, ID: "call_single", Type: "function", Function: functionMsg{Name: "read_file", Arguments: `{"path":"test.txt"}`}},
						},
					},
				},
			},
		},
		{
			ID:      "chat_1",
			Object:  "chat.completion.chunk",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []struct {
				Index        int          `json:"index"`
				Delta        deltaContent `json:"delta"`
				FinishReason string       `json:"finish_reason"`
			}{
				{
					Index:        0,
					FinishReason: "tool_calls",
				},
			},
		},
	}

	stream := make(chan []byte, len(chunks))
	errs := make(chan error, 1)
	out := make(chan llm.StreamEvent, 16)

	for _, chunk := range chunks {
		data, err := json.Marshal(chunk)
		if err != nil {
			t.Fatalf("marshal chunk: %v", err)
		}
		stream <- data
	}
	close(stream)
	errs <- nil
	close(errs)

	b := &ChatBackend{
		name:     "test",
		messages: make([]chatMessage, 0),
	}
	b.processStream(stream, errs, out, "")

	var toolCalls []*llm.ToolCall
	for event := range out {
		if event.Type == "error" {
			t.Fatalf("unexpected error: %v", event.Error)
		}
		if event.Type == "tool_call" && event.ToolCall != nil {
			toolCalls = append(toolCalls, event.ToolCall)
		}
	}

	if len(toolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(toolCalls))
	}

	tc := toolCalls[0]
	if tc.ID != "call_single" {
		t.Errorf("expected tool call ID 'call_single', got %q", tc.ID)
	}
	if tc.Name != "read_file" {
		t.Errorf("expected tool call Name 'read_file', got %q", tc.Name)
	}
	if tc.Arguments["path"] != "test.txt" {
		t.Errorf("expected tool call Arguments[path] 'test.txt', got %v", tc.Arguments["path"])
	}
}
