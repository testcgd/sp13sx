package mcp

import "context"

type Client interface {
	ListTools(ctx context.Context) ([]RemoteDescriptor, error)
	CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error)
	Close() error
}

type RemoteDescriptor struct {
	Name        string
	Description string
	Schema      map[string]any
}
