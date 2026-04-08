package mock

import (
	"context"
	"testing"

	"sp13sx/internal/llm"
)

func TestScenarioBackendMatchesAppliedToolNames(t *testing.T) {
	scenario, err := LoadScenarioFromString(`
name: tool_match
turns:
  - match_tools_applied:
      - read_file
    response_id: "resp_1"
    events:
      - type: message
        content: "matched"
`)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}

	backend := NewScenarioBackend(scenario)

	stream, err := backend.Generate(context.Background(), llm.GenerateRequest{
		Input: []llm.InputItem{llm.ToolResultInput("call_1", "read_file", `{}`)},
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	found := false
	for event := range stream {
		if event.Type == "message" && event.Content == "matched" {
			found = true
		}
		if event.Error != nil {
			t.Fatalf("unexpected error: %v", event.Error)
		}
	}

	if !found {
		t.Fatal("expected tool-matched turn to emit final message")
	}
}
