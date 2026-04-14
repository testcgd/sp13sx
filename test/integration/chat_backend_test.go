//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"sp13sx/internal/llm"
	"sp13sx/internal/llm/openai"
)

func TestChatBackendBasicChat(t *testing.T) {
	skipIfNoIntegration(t)

	cfg := loadTestConfig(t)
	backendCfg := cfg.Backends[cfg.Defaults.Backend]

	backend, err := openai.NewChatBackend(backendCfg)
	if err != nil {
		t.Fatalf("create backend: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := llm.GenerateRequest{
		Model: cfg.Defaults.Model,
		Input: []llm.InputItem{
			llm.UserTextInput("请回复：测试成功"),
		},
	}

	stream, err := backend.Generate(ctx, req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var response string
	for event := range stream {
		if event.Type == "message" {
			response += event.Content
		}
		if event.Error != nil {
			t.Fatalf("stream error: %v", event.Error)
		}
	}

	if response == "" {
		t.Fatal("expected non-empty response")
	}

	t.Logf("Response: %s", response)
}

func TestChatBackendToolCall(t *testing.T) {
	skipIfNoIntegration(t)

	cfg := loadTestConfig(t)
	backendCfg := cfg.Backends[cfg.Defaults.Backend]

	backend, err := openai.NewChatBackend(backendCfg)
	if err != nil {
		t.Fatalf("create backend: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := llm.GenerateRequest{
		Model: cfg.Defaults.Model,
		Input: []llm.InputItem{
			llm.UserTextInput("读取 README.md 的第一行"),
		},
		Tools: []llm.ToolDefinition{{
			Name:        "read_file",
			Description: "读取文件内容",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{"type": "string"},
				},
				"required": []string{"path"},
			},
		}},
	}

	stream, err := backend.Generate(ctx, req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var toolCalls []*llm.ToolCall
	var response string

	for event := range stream {
		if event.Type == "tool_call" && event.ToolCall != nil {
			toolCalls = append(toolCalls, event.ToolCall)
		}
		if event.Type == "message" {
			response += event.Content
		}
		if event.Error != nil {
			t.Fatalf("stream error: %v", event.Error)
		}
	}

	if len(toolCalls) == 0 && response == "" {
		t.Fatal("expected either tool call or response")
	}

	if len(toolCalls) > 0 {
		t.Logf("Tool call: %s, args: %v", toolCalls[0].Name, toolCalls[0].Arguments)
	}
	if response != "" {
		t.Logf("Response: %s", response)
	}
}

func TestChatBackendMultiTurn(t *testing.T) {
	skipIfNoIntegration(t)

	cfg := loadTestConfig(t)
	backendCfg := cfg.Backends[cfg.Defaults.Backend]

	backend, err := openai.NewChatBackend(backendCfg)
	if err != nil {
		t.Fatalf("create backend: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := llm.GenerateRequest{
		Model: cfg.Defaults.Model,
		Input: []llm.InputItem{
			llm.UserTextInput("我的名字是测试者"),
		},
	}

	stream, err := backend.Generate(ctx, req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for event := range stream {
		if event.Error != nil {
			t.Fatalf("stream error: %v", event.Error)
		}
	}

	req = llm.GenerateRequest{
		Model: cfg.Defaults.Model,
		Input: []llm.InputItem{
			llm.UserTextInput("你还记得我的名字吗？"),
		},
	}

	stream, err = backend.Generate(ctx, req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var response string
	for event := range stream {
		if event.Type == "message" {
			response += event.Content
		}
		if event.Error != nil {
			t.Fatalf("stream error: %v", event.Error)
		}
	}

	if response == "" {
		t.Fatal("expected non-empty response")
	}

	if !strings.Contains(response, "测试") {
		t.Logf("Warning: response may not contain the name, got: %s", response)
	}

	t.Logf("Response: %s", response)
}

func TestChatBackendWithReasoning(t *testing.T) {
	skipIfNoIntegration(t)

	cfg := loadTestConfig(t)
	backendCfg := cfg.Backends[cfg.Defaults.Backend]

	backend, err := openai.NewChatBackend(backendCfg)
	if err != nil {
		t.Fatalf("create backend: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := llm.GenerateRequest{
		Model: cfg.Defaults.Model,
		Input: []llm.InputItem{
			llm.UserTextInput("请思考一下：1+1等于多少？"),
		},
	}

	stream, err := backend.Generate(ctx, req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var response string
	var reasoning string

	for event := range stream {
		if event.Type == "message" {
			response += event.Content
		}
		if event.Type == "reasoning" {
			reasoning += event.ReasoningContent
		}
		if event.Error != nil {
			t.Fatalf("stream error: %v", event.Error)
		}
	}

	if response == "" {
		t.Fatal("expected non-empty response")
	}

	t.Logf("Response: %s", response)
	if reasoning != "" {
		t.Logf("Reasoning: %s", reasoning)
	} else {
		t.Log("No reasoning content returned (model may not support it)")
	}
}
