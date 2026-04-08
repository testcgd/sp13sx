package domain

import "time"

type ToolInvocation struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	ToolName  string         `json:"tool_name"`
	Status    string         `json:"status"`
	Arguments map[string]any `json:"arguments,omitempty"`
	Output    map[string]any `json:"output,omitempty"`
	Error     string         `json:"error,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}
