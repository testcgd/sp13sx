package store

import "path/filepath"

type Paths struct {
	BaseDir       string
	SessionsPath  string
	MessagesPath  string
	ToolCallsPath string
}

func NewPaths(baseDir string) Paths {
	return Paths{
		BaseDir:       baseDir,
		SessionsPath:  filepath.Join(baseDir, "sessions.jsonl"),
		MessagesPath:  filepath.Join(baseDir, "messages.jsonl"),
		ToolCallsPath: filepath.Join(baseDir, "tool_calls.jsonl"),
	}
}
