package llm

func UserTextInput(content string) InputItem {
	return InputItem{
		Type:    "message",
		Role:    "user",
		Content: content,
	}
}

func FunctionOutputInput(callID string, output string) InputItem {
	return InputItem{
		Type:   "function_call_output",
		CallID: callID,
		Output: output,
	}
}
