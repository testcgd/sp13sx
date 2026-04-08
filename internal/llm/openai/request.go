package openai

import "sp13sx/internal/llm"

type responseRequest struct {
	Model              string         `json:"model"`
	Instructions       string         `json:"instructions,omitempty"`
	Input              []inputItem    `json:"input"`
	Tools              []responseTool `json:"tools,omitempty"`
	PreviousResponseID string         `json:"previous_response_id,omitempty"`
	Stream             bool           `json:"stream"`
}

type inputItem struct {
	Type    string `json:"type,omitempty"`
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
	CallID  string `json:"call_id,omitempty"`
	Output  string `json:"output,omitempty"`
}

type responseTool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

func buildRequest(req llm.GenerateRequest) responseRequest {
	out := responseRequest{
		Model:              req.Model,
		Instructions:       req.Instructions,
		Input:              make([]inputItem, 0, len(req.Input)),
		Tools:              make([]responseTool, 0, len(req.Tools)),
		PreviousResponseID: req.PreviousResponseID,
		Stream:             true,
	}
	for _, item := range req.Input {
		out.Input = append(out.Input, inputItem{
			Type:    item.Type,
			Role:    item.Role,
			Content: item.Content,
			CallID:  item.CallID,
			Output:  item.Output,
		})
	}
	for _, tool := range req.Tools {
		out.Tools = append(out.Tools, responseTool{
			Type:        "function",
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.InputSchema,
		})
	}
	return out
}
