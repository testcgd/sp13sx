package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"sp13sx/internal/config"
	"sp13sx/internal/llm"
	"sp13sx/internal/llm/mock"
	"sp13sx/internal/tui"
)

// TestBasicChatScenario 测试基础对话场景
func TestBasicChatScenario(t *testing.T) {
	scenario, err := mock.LoadScenario("../scenarios/basic_chat.yaml")
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}

	if err := scenario.Validate(); err != nil {
		t.Fatalf("validate scenario: %v", err)
	}

	backend := mock.NewScenarioBackend(scenario)

	// 验证初始状态
	if backend.Name() == "" {
		t.Error("expected backend name to be non-empty")
	}

	// 测试场景匹配
	if !backend.MatchesInput("hello") {
		t.Error("expected backend to match 'hello' input")
	}
}

// TestScenarioBackendFlow 测试场景后端的完整流程
func TestScenarioBackendFlow(t *testing.T) {
	scenario, err := mock.LoadScenarioFromString(`
name: test_flow
description: 测试流程
turns:
  - match_input: "test"
    response_id: "resp_001"
    events:
      - type: message
        content: "Hello"
        delay: 10ms
      - type: message
        content: " World"
        delay: 10ms
`)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}

	backend := mock.NewScenarioBackend(scenario)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := mock.CreateTestRequest("test")
	stream, err := backend.Generate(ctx, req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var messages []string
	for event := range stream {
		if event.Type == "message" {
			messages = append(messages, event.Content)
		}
		if event.Error != nil {
			t.Fatalf("unexpected error: %v", event.Error)
		}
	}

	combined := strings.Join(messages, "")
	if combined != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", combined)
	}
}

// TestToolChainScenario 测试工具链场景
func TestToolChainScenario(t *testing.T) {
	scenario, err := mock.LoadScenario("../scenarios/tool_chain.yaml")
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}

	backend := mock.NewScenarioBackend(scenario)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 第一轮：触发工具调用
	req := mock.CreateTestRequest("读取 main.go 并分析")
	stream, err := backend.Generate(ctx, req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var foundToolCall bool
	var responseID string
	for event := range stream {
		if event.Type == "tool_call" && event.ToolCall != nil {
			foundToolCall = true
			if event.ToolCall.Name != "read_file" {
				t.Errorf("expected tool_call 'read_file', got %q", event.ToolCall.Name)
			}
		}
		if event.Type == "response_id" {
			responseID = event.ResponseID
		}
	}

	if !foundToolCall {
		t.Error("expected tool_call event")
	}
	if responseID != "resp_001" {
		t.Errorf("expected response_id 'resp_001', got %q", responseID)
	}
}

// TestErrorRecoveryScenario 测试错误恢复场景
func TestErrorRecoveryScenario(t *testing.T) {
	scenario, err := mock.LoadScenario("../scenarios/error_recovery.yaml")
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}

	backend := mock.NewScenarioBackend(scenario)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := mock.CreateTestRequest("trigger error")
	stream, err := backend.Generate(ctx, req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var foundError bool
	for event := range stream {
		if event.Type == "error" {
			foundError = true
			if event.Error == nil {
				t.Error("expected error to have error value")
			}
		}
	}

	if !foundError {
		t.Error("expected error event")
	}
}

// TestTUIWithMockBackend 测试 TUI 与 Mock Backend 的集成
func TestTUIWithMockBackend(t *testing.T) {
	scenario, err := mock.LoadScenarioFromString(`
name: tui_test
description: TUI 测试
turns:
  - match_input: "hello"
    response_id: "resp_001"
    events:
      - type: message
        content: "你好！"
        delay: 10ms
`)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}

	backend := mock.NewScenarioBackend(scenario)
	runtime := &testRuntime{backend: backend}
	model := tui.NewModel(runtime)

	// 验证初始状态
	if model.Status() != "ready" {
		t.Errorf("expected initial status 'ready', got %q", model.Status())
	}
}

// testRuntime 用于测试的 Runtime 实现
type testRuntime struct {
	backend *mock.ScenarioBackend
}

func (r *testRuntime) Send(ctx context.Context, text string) (<-chan llm.StreamEvent, error) {
	req := mock.CreateTestRequest(text)
	return r.backend.Generate(ctx, req)
}

func (r *testRuntime) BackendName() string            { return "mock" }
func (r *testRuntime) ModelName() string              { return "mock-model" }
func (r *testRuntime) SessionTitle() string           { return "test-session" }
func (r *testRuntime) EnabledSkillNames() []string    { return []string{} }
func (r *testRuntime) DiscoveredSkillNames() []string { return []string{} }
func (r *testRuntime) MCPStatusLines() []string       { return []string{} }
func (r *testRuntime) RightPaneWidth() int            { return 40 }

// createTestConfig 创建测试配置
func createTestConfig() config.Config {
	return config.Config{
		UI: config.UIConfig{
			RightPaneWidth: 40,
		},
		Defaults: config.DefaultsConfig{
			Backend:       "mock",
			Model:         "mock-model",
			EnabledSkills: []string{},
		},
		Tools: config.ToolsConfig{
			Builtin: config.BuiltinToolsConfig{
				ReadFile:  config.ToolToggle{Enabled: true},
				ListFiles: config.ToolToggle{Enabled: true},
			},
		},
	}
}
