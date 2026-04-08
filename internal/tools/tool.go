package tools

import "context"

type Tool interface {
	Name() string
	Description() string
	Schema() map[string]any
	Run(ctx context.Context, args map[string]any) (map[string]any, error)
}
