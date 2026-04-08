package builtin

import (
	"context"
	"os"
)

type ReadFile struct{}

func (ReadFile) Name() string { return "read_file" }

func (ReadFile) Description() string { return "Read a UTF-8 text file from disk." }

func (ReadFile) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string"},
		},
		"required": []string{"path"},
	}
}

func (ReadFile) Run(_ context.Context, args map[string]any) (map[string]any, error) {
	path, _ := args["path"].(string)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return map[string]any{"content": string(data)}, nil
}
