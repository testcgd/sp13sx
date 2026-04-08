package domain

import "time"

type ContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type Message struct {
	ID         string        `json:"id"`
	SessionID  string        `json:"session_id"`
	Role       string        `json:"role"`
	Content    []ContentPart `json:"content,omitempty"`
	ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}
