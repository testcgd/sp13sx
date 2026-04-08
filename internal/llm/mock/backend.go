package mock

import (
	"context"

	"sp13sx/internal/llm"
)

type Backend struct {
	NameValue  string
	GenerateFn func(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error)
}

func NewBackend(fn func(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error)) *Backend {
	return &Backend{
		NameValue:  "mock",
		GenerateFn: fn,
	}
}

func (b *Backend) Name() string {
	if b.NameValue == "" {
		return "mock"
	}
	return b.NameValue
}

func (b *Backend) Generate(ctx context.Context, req llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	if b.GenerateFn == nil {
		ch := make(chan llm.StreamEvent)
		close(ch)
		return ch, nil
	}
	return b.GenerateFn(ctx, req)
}
