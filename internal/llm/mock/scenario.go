package mock

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"sp13sx/internal/llm"
)

// Scenario 定义一个完整的测试场景
type Scenario struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Turns       []Turn `yaml:"turns"`
}

// Turn 定义一轮对话的匹配条件和响应
type Turn struct {
	// 匹配条件（至少一个）
	MatchInput          string   `yaml:"match_input,omitempty"`             // 精确匹配用户输入
	MatchPattern        string   `yaml:"match_pattern,omitempty"`           // 正则匹配用户输入
	MatchPreviousRespID string   `yaml:"match_previous_response,omitempty"` // 匹配上一轮的 response_id
	MatchToolsApplied   []string `yaml:"match_tools_applied,omitempty"`     // 匹配已执行的工具

	// 响应内容
	ResponseID string  `yaml:"response_id"`
	Events     []Event `yaml:"events"`
}

// Event 定义单个流事件
type Event struct {
	Type    string         `yaml:"type"`
	Content string         `yaml:"content,omitempty"`
	Delay   time.Duration  `yaml:"delay,omitempty"`
	Tool    *ToolCallEvent `yaml:"tool_call,omitempty"`
	Error   string         `yaml:"error,omitempty"`
}

// ToolCallEvent 定义工具调用事件
type ToolCallEvent struct {
	ID        string         `yaml:"id"`
	CallID    string         `yaml:"call_id"`
	Name      string         `yaml:"name"`
	Arguments map[string]any `yaml:"arguments"`
}

// LoadScenario 从 YAML 文件加载场景
func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario file: %w", err)
	}

	var scenario Scenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("parse scenario: %w", err)
	}

	if scenario.Name == "" {
		return nil, fmt.Errorf("scenario name is required")
	}

	return &scenario, nil
}

// LoadScenarioFromString 从 YAML 字符串加载场景
func LoadScenarioFromString(data string) (*Scenario, error) {
	var scenario Scenario
	if err := yaml.Unmarshal([]byte(data), &scenario); err != nil {
		return nil, fmt.Errorf("parse scenario: %w", err)
	}

	if scenario.Name == "" {
		return nil, fmt.Errorf("scenario name is required")
	}

	return &scenario, nil
}

// ToStreamEvent 将 Event 转换为 llm.StreamEvent
func (e *Event) ToStreamEvent() llm.StreamEvent {
	event := llm.StreamEvent{
		Type: e.Type,
	}

	switch e.Type {
	case "message", "status":
		event.Content = e.Content
	case "response_id":
		event.ResponseID = e.Content
	case "tool_call":
		if e.Tool != nil {
			event.ToolCall = &llm.ToolCall{
				ID:        e.Tool.ID,
				CallID:    e.Tool.CallID,
				Name:      e.Tool.Name,
				Arguments: e.Tool.Arguments,
			}
		}
	case "error":
		if e.Error != "" {
			event.Error = fmt.Errorf("%s", e.Error)
		}
	}

	return event
}

// Validate 验证场景是否有效
func (s *Scenario) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("scenario name is required")
	}

	for i, turn := range s.Turns {
		if turn.ResponseID == "" {
			return fmt.Errorf("turn %d: response_id is required", i)
		}
		if len(turn.Events) == 0 {
			return fmt.Errorf("turn %d: at least one event is required", i)
		}

		for j, event := range turn.Events {
			if event.Type == "" {
				return fmt.Errorf("turn %d event %d: type is required", i, j)
			}
		}
	}

	return nil
}

// CreateTestRequest 创建测试用的 GenerateRequest
func CreateTestRequest(input string) llm.GenerateRequest {
	return llm.GenerateRequest{
		Model: "test-model",
		Input: []llm.InputItem{
			{
				Type:    "user_message",
				Content: input,
			},
		},
	}
}
