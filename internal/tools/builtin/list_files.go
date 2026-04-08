package builtin

import (
	"context"
	"os"
)

type ListFiles struct{}

func (ListFiles) Name() string { return "list_files" }

func (ListFiles) Description() string { return "List directory entries for a path." }

func (ListFiles) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string"},
		},
	}
}

func (ListFiles) Run(_ context.Context, args map[string]any) (map[string]any, error) {
	path, _ := args["path"].(string)
	if path == "" {
		path = "."
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return map[string]any{"entries": files}, nil
}
