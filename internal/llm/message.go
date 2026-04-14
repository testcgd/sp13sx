package llm

func UserTextInput(content string) InputItem {
	return InputItem{
		Type:    "user_message",
		Content: content,
	}
}

func ToolResultInput(callID string, toolName string, content string) InputItem {
	return InputItem{
		Type:     "tool_result",
		CallID:   callID,
		ToolName: toolName,
		Content:  content,
	}
}

func AssistantContextInput(content string, reasoningContent string, toolCalls []ToolCall) InputItem {
	return InputItem{
		Type:             "assistant_context",
		Content:          content,
		ReasoningContent: reasoningContent,
		ToolCalls:        toolCalls,
	}
}
