package llm

import "context"

type InputItem struct {
	Type    string
	Role    string
	Content string
	CallID  string
	Output  string
}

type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type StreamEvent struct {
	Type       string
	Content    string
	ResponseID string
	ToolCall   *ToolCall
	Error      error
}

type GenerateRequest struct {
	Model              string
	Instructions       string
	Input              []InputItem
	Tools              []ToolDefinition
	PreviousResponseID string
}

type Backend interface {
	Name() string
	Generate(ctx context.Context, req GenerateRequest) (<-chan StreamEvent, error)
}
