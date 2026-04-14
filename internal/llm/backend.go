package llm

import "context"

type InputItem struct {
	Type             string
	Content          string
	CallID           string
	ToolName         string
	ReasoningContent string
	ToolCalls        []ToolCall
}

type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type StreamEvent struct {
	Type             string
	Content          string
	ReasoningContent string
	ResponseID       string
	ToolCall         *ToolCall
	Error            error
}

type RuntimeStatus struct {
	IsRunningTool bool     `json:"is_running_tool"`
	PendingInputs []string `json:"pending_inputs"`
}

type GenerateRequest struct {
	Model        string
	Instructions string
	Input        []InputItem
	Tools        []ToolDefinition
}

type Backend interface {
	Name() string
	Generate(ctx context.Context, req GenerateRequest) (<-chan StreamEvent, error)
}

type ReasoningContentClearer interface {
	ClearReasoningContent()
}
