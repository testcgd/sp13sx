package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sp13sx/internal/config"
	"sp13sx/internal/llm"
)

type Backend struct {
	name   string
	client *Client
}

func NewBackend(cfg config.Backend) (*Backend, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Backend{name: "openai", client: client}, nil
}

func (b *Backend) Name() string {
	return b.name
}

func (b *Backend) Generate(ctx context.Context, req llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 8)
	stream, errs, err := b.client.Stream(ctx, "/responses", buildRequest(req))
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(ch)
		partialCalls := map[string]*llm.ToolCall{}

		for raw := range stream {
			var event responseEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				ch <- llm.StreamEvent{Type: "error", Error: fmt.Errorf("decode stream event: %w", err)}
				return
			}
			switch event.Type {
			case "response.created":
				ch <- llm.StreamEvent{Type: "response_id", ResponseID: event.Response.ID}
			case "response.output_text.delta":
				ch <- llm.StreamEvent{Type: "message", Content: event.Delta}
			case "response.output_item.added":
				if event.Item.Type != "function_call" {
					continue
				}
				partialCalls[event.Item.ID] = &llm.ToolCall{
					ID:        event.Item.ID,
					CallID:    event.Item.CallID,
					Name:      event.Item.Name,
					Arguments: map[string]any{},
				}
			case "response.function_call_arguments.done":
				call := partialCalls[event.ItemID]
				if call == nil {
					call = &llm.ToolCall{
						ID:        event.ItemID,
						Name:      event.Name,
						Arguments: map[string]any{},
					}
				}
				call.Name = firstNonEmpty(call.Name, event.Name)
				var args map[string]any
				if strings.TrimSpace(event.Arguments) != "" {
					if err := json.Unmarshal([]byte(event.Arguments), &args); err != nil {
						ch <- llm.StreamEvent{Type: "error", Error: fmt.Errorf("decode tool arguments: %w", err)}
						return
					}
				}
				if args == nil {
					args = map[string]any{}
				}
				call.Arguments = args
				ch <- llm.StreamEvent{Type: "tool_call", ToolCall: call}
			case "response.completed":
				if event.Response.ID != "" {
					ch <- llm.StreamEvent{Type: "response_id", ResponseID: event.Response.ID}
				}
			case "response.failed":
				ch <- llm.StreamEvent{Type: "error", Error: fmt.Errorf("response failed")}
				return
			}
		}
		if err := <-errs; err != nil {
			ch <- llm.StreamEvent{Type: "error", Error: err}
		}
	}()
	return ch, nil
}

type responseEvent struct {
	Type      string `json:"type"`
	Delta     string `json:"delta"`
	Arguments string `json:"arguments"`
	ItemID    string `json:"item_id"`
	Name      string `json:"name"`
	Response  struct {
		ID string `json:"id"`
	} `json:"response"`
	Item struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Name   string `json:"name"`
		CallID string `json:"call_id"`
	} `json:"item"`
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
