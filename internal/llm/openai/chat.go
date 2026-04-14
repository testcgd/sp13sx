package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"sp13sx/internal/config"
	"sp13sx/internal/domain"
	"sp13sx/internal/llm"
)

type ChatBackend struct {
	name     string
	client   *Client
	messages []chatMessage
	mu       sync.Mutex
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Tools    []chatTool    `json:"tools,omitempty"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role             string        `json:"role"`
	Content          any           `json:"content,omitempty"`
	ReasoningContent string        `json:"reasoning_content,omitempty"`
	ToolCalls        []toolCallMsg `json:"tool_calls,omitempty"`
	ToolCallID       string        `json:"tool_call_id,omitempty"`
}

type toolCallMsg struct {
	Index    int         `json:"index"`
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Function functionMsg `json:"function"`
}

type functionMsg struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatTool struct {
	Type     string      `json:"type"`
	Function functionDef `json:"function"`
}

type functionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type chatCompletionChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int          `json:"index"`
		Delta        deltaContent `json:"delta"`
		FinishReason string       `json:"finish_reason"`
	} `json:"choices"`
}

type deltaContent struct {
	Role             string        `json:"role,omitempty"`
	Content          string        `json:"content,omitempty"`
	ReasoningContent string        `json:"reasoning_content,omitempty"`
	ToolCalls        []toolCallMsg `json:"tool_calls,omitempty"`
}

func NewChatBackend(cfg config.Backend) (*ChatBackend, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ChatBackend{
		name:     "openai-chat",
		client:   client,
		messages: make([]chatMessage, 0),
	}, nil
}

func (b *ChatBackend) Name() string {
	return b.name
}

func (b *ChatBackend) Generate(ctx context.Context, req llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	b.mu.Lock()
	for _, item := range req.Input {
		switch item.Type {
		case "user_message":
			b.messages = append(b.messages, chatMessage{
				Role:    "user",
				Content: item.Content,
			})
		case "assistant_context":
			toolCalls := make([]toolCallMsg, len(item.ToolCalls))
			for i, tc := range item.ToolCalls {
				argsJSON, _ := json.Marshal(tc.Arguments)
				toolCalls[i] = toolCallMsg{
					ID:   tc.ID,
					Type: "function",
					Function: functionMsg{
						Name:      tc.Name,
						Arguments: string(argsJSON),
					},
				}
			}
			b.messages = append(b.messages, chatMessage{
				Role:             "assistant",
				Content:          item.Content,
				ReasoningContent: item.ReasoningContent,
				ToolCalls:        toolCalls,
			})
		case "tool_result":
			b.messages = append(b.messages, chatMessage{
				Role:       "tool",
				ToolCallID: item.CallID,
				Content:    item.Content,
			})
		}
	}
	messagesCopy := make([]chatMessage, len(b.messages))
	copy(messagesCopy, b.messages)
	b.mu.Unlock()

	chatReq := chatRequest{
		Model:    req.Model,
		Messages: messagesCopy,
		Stream:   true,
	}

	if len(req.Tools) > 0 {
		chatReq.Tools = make([]chatTool, len(req.Tools))
		for i, tool := range req.Tools {
			chatReq.Tools[i] = chatTool{
				Type: "function",
				Function: functionDef{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			}
		}
	}

	stream, errs, err := b.client.Stream(ctx, "/chat/completions", chatReq)
	if err != nil {
		return nil, err
	}

	ch := make(chan llm.StreamEvent, 16)
	go b.processStream(stream, errs, ch, req.Instructions)

	return ch, nil
}

func (b *ChatBackend) processStream(stream <-chan []byte, errs <-chan error, out chan<- llm.StreamEvent, instructions string) {
	defer close(out)

	partialToolCalls := make(map[int]*llm.ToolCall)
	emittedToolCalls := make(map[int]bool)
	var assistantContent string
	var reasoningContent string

	for raw := range stream {
		var chunk chatCompletionChunk
		if err := json.Unmarshal(raw, &chunk); err != nil {
			out <- llm.StreamEvent{Type: "error", Error: fmt.Errorf("decode chunk: %w (raw: %s)", err, string(raw))}
			return
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		delta := choice.Delta

		if delta.Role == "assistant" && instructions != "" {
		}

		if delta.ReasoningContent != "" {
			reasoningContent += delta.ReasoningContent
			out <- llm.StreamEvent{Type: "reasoning", ReasoningContent: delta.ReasoningContent}
		}

		if delta.Content != "" {
			assistantContent += delta.Content
			out <- llm.StreamEvent{Type: "message", Content: delta.Content}
		}

		for _, tc := range delta.ToolCalls {
			idx := tc.Index
			if tc.ID != "" {
				if existing, ok := partialToolCalls[idx]; ok {
					existing.ID = tc.ID
					if tc.Function.Name != "" {
						existing.Name = tc.Function.Name
					}
					existing.Arguments = appendArgs(existing, tc.Function.Arguments)
				} else {
					partialToolCalls[idx] = &llm.ToolCall{
						ID:        tc.ID,
						CallID:    tc.ID,
						Name:      tc.Function.Name,
						Arguments: parseArgs(tc.Function.Arguments),
					}
				}
			} else if tc.Function.Arguments != "" {
				if existing, ok := partialToolCalls[idx]; ok {
					existing.Arguments = appendArgs(existing, tc.Function.Arguments)
				}
			}

			if tc, ok := partialToolCalls[idx]; ok && isCompleteToolCall(tc) && !emittedToolCalls[idx] {
				emittedToolCalls[idx] = true
				out <- llm.StreamEvent{Type: "tool_call", ToolCall: tc}
			}
		}

		if choice.FinishReason != "" {
			_ = choice.FinishReason
		}
	}

	if err := <-errs; err != nil {
		out <- llm.StreamEvent{Type: "error", Error: err}
		return
	}

	if len(partialToolCalls) > 0 {
		toolCalls := make([]toolCallMsg, 0, len(partialToolCalls))
		for _, tc := range partialToolCalls {
			argsJSON, _ := json.Marshal(tc.Arguments)
			toolCalls = append(toolCalls, toolCallMsg{
				ID:   tc.ID,
				Type: "function",
				Function: functionMsg{
					Name:      tc.Name,
					Arguments: string(argsJSON),
				},
			})
		}

		b.mu.Lock()
		b.messages = append(b.messages, chatMessage{
			Role:             "assistant",
			Content:          assistantContent,
			ReasoningContent: reasoningContent,
			ToolCalls:        toolCalls,
		})
		b.mu.Unlock()
	} else if assistantContent != "" || reasoningContent != "" {
		b.mu.Lock()
		b.messages = append(b.messages, chatMessage{
			Role:             "assistant",
			Content:          assistantContent,
			ReasoningContent: reasoningContent,
		})
		b.mu.Unlock()
	}
}

func isCompleteToolCall(tc *llm.ToolCall) bool {
	if tc == nil || tc.ID == "" || tc.Name == "" {
		return false
	}
	if tc.Arguments == nil {
		return false
	}
	_, err := json.Marshal(tc.Arguments)
	return err == nil
}

func parseArgs(args string) map[string]any {
	if args == "" {
		return make(map[string]any)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(args), &result); err != nil {
		return make(map[string]any)
	}
	return result
}

func appendArgs(tc *llm.ToolCall, args string) map[string]any {
	if tc.Arguments == nil {
		tc.Arguments = make(map[string]any)
	}
	if args == "" {
		return tc.Arguments
	}
	var additional map[string]any
	if err := json.Unmarshal([]byte(args), &additional); err != nil {
		return tc.Arguments
	}
	for k, v := range additional {
		tc.Arguments[k] = v
	}
	return tc.Arguments
}

func (b *ChatBackend) LoadHistory(messages []domain.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages = make([]chatMessage, 0, len(messages))
	for _, msg := range messages {
		cm := chatMessage{Role: msg.Role}

		switch msg.Role {
		case "user":
			cm.Content = extractTextContent(msg.Content)
		case "assistant":
			cm.Content = extractTextContent(msg.Content)
			if len(msg.ToolCalls) > 0 {
				cm.ToolCalls = make([]toolCallMsg, len(msg.ToolCalls))
				for i, tc := range msg.ToolCalls {
					argsJSON, _ := json.Marshal(tc.Arguments)
					cm.ToolCalls[i] = toolCallMsg{
						ID:   tc.ID,
						Type: "function",
						Function: functionMsg{
							Name:      tc.Name,
							Arguments: string(argsJSON),
						},
					}
				}
			}
		case "tool":
			cm.ToolCallID = msg.ToolCallID
			cm.Content = extractTextContent(msg.Content)
		}

		b.messages = append(b.messages, cm)
	}
}

func (b *ChatBackend) ClearHistory() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = make([]chatMessage, 0)
}

func extractTextContent(parts []domain.ContentPart) string {
	var b strings.Builder
	for _, part := range parts {
		if part.Type == "text" {
			b.WriteString(part.Text)
		}
	}
	return b.String()
}
