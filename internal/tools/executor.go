package tools

import "context"

type Executor struct {
	registry *Registry
}

func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

func (e *Executor) Run(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	tool, ok := e.registry.Get(name)
	if !ok {
		return nil, ErrToolNotFound(name)
	}
	return tool.Run(ctx, args)
}
