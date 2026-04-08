package mcp

import (
	"context"

	"sp13sx/internal/tools"
)

type RemoteTool struct {
	NameValue        string
	DescriptionValue string
	SchemaValue      map[string]any
	Invoke           func(ctx context.Context, args map[string]any) (map[string]any, error)
}

func (t RemoteTool) Name() string           { return t.NameValue }
func (t RemoteTool) Description() string    { return t.DescriptionValue }
func (t RemoteTool) Schema() map[string]any { return t.SchemaValue }
func (t RemoteTool) Run(ctx context.Context, args map[string]any) (map[string]any, error) {
	if t.Invoke == nil {
		return map[string]any{"status": "unimplemented"}, nil
	}
	return t.Invoke(ctx, args)
}

var _ tools.Tool = RemoteTool{}
