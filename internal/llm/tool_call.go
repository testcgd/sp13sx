package llm

type ToolCall struct {
	ID        string
	CallID    string
	Name      string
	Arguments map[string]any
}
