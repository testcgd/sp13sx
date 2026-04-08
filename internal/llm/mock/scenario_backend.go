package mock

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"sp13sx/internal/llm"
)

// ScenarioBackend 基于场景脚本的 LLM Backend
type ScenarioBackend struct {
	scenario       *Scenario
	currentTurnIdx int
	options        ScenarioOptions
}

// ScenarioOptions 场景后端配置选项
type ScenarioOptions struct {
	// RandomDelay 是否随机化延迟（在配置的延迟基础上增加 0-50% 的随机值）
	RandomDelay bool

	// DefaultDelay 默认延迟（当事件未配置延迟时使用）
	DefaultDelay time.Duration

	// OnNoMatch 当没有匹配的 Turn 时的行为
	OnNoMatch NoMatchBehavior
}

// NoMatchBehavior 定义无匹配时的行为
type NoMatchBehavior int

const (
	// NoMatchError 返回错误
	NoMatchError NoMatchBehavior = iota
	// NoMatchEmpty 返回空响应
	NoMatchEmpty
	// NoMatchContinue 尝试下一个 Turn
	NoMatchContinue
)

// NewScenarioBackend 创建基于场景的 Backend
func NewScenarioBackend(scenario *Scenario, opts ...ScenarioOptions) *ScenarioBackend {
	options := ScenarioOptions{
		DefaultDelay: 50 * time.Millisecond,
		OnNoMatch:    NoMatchError,
	}
	if len(opts) > 0 {
		options = opts[0]
	}

	return &ScenarioBackend{
		scenario: scenario,
		options:  options,
	}
}

// Name 返回 Backend 名称
func (b *ScenarioBackend) Name() string {
	return "scenario:" + b.scenario.Name
}

// Generate 实现 llm.Backend 接口
func (b *ScenarioBackend) Generate(ctx context.Context, req llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 16)

	go func() {
		defer close(ch)

		turn := b.matchTurn(req)
		if turn == nil {
			switch b.options.OnNoMatch {
			case NoMatchEmpty:
				return
			case NoMatchContinue:
				if b.currentTurnIdx < len(b.scenario.Turns) {
					turn = &b.scenario.Turns[b.currentTurnIdx]
					b.currentTurnIdx++
				} else {
					return
				}
			default:
				ch <- llm.StreamEvent{
					Type:  "error",
					Error: fmt.Errorf("no matching turn for request"),
				}
				return
			}
		} else {
			b.currentTurnIdx++
		}

		// 发送 response_id 事件
		if turn.ResponseID != "" {
			ch <- llm.StreamEvent{
				Type:       "response_id",
				ResponseID: turn.ResponseID,
			}
		}

		// 按顺序发送事件
		for _, event := range turn.Events {
			delay := event.Delay
			if delay == 0 {
				delay = b.options.DefaultDelay
			}
			if b.options.RandomDelay && delay > 0 {
				delay = time.Duration(float64(delay) * (1 + 0.5*float64(time.Now().UnixNano()%100)/100))
			}

			select {
			case <-ctx.Done():
				ch <- llm.StreamEvent{Type: "error", Error: ctx.Err()}
				return
			case <-time.After(delay):
			}

			ch <- event.ToStreamEvent()
		}
	}()

	return ch, nil
}

// matchTurn 匹配当前请求对应的 Turn
func (b *ScenarioBackend) matchTurn(req llm.GenerateRequest) *Turn {
	for i := b.currentTurnIdx; i < len(b.scenario.Turns); i++ {
		turn := &b.scenario.Turns[i]
		if b.turnMatches(turn, req) {
			return turn
		}
	}
	return nil
}

// turnMatches 检查 Turn 是否匹配当前请求
func (b *ScenarioBackend) turnMatches(turn *Turn, req llm.GenerateRequest) bool {
	// 检查用户输入匹配
	if turn.MatchInput != "" {
		userInput := extractUserInput(req)
		if userInput != turn.MatchInput {
			return false
		}
	}

	// 检查正则匹配
	if turn.MatchPattern != "" {
		userInput := extractUserInput(req)
		matched, err := regexp.MatchString(turn.MatchPattern, userInput)
		if err != nil || !matched {
			return false
		}
	}

	if len(turn.MatchToolsApplied) > 0 {
		appliedTools := extractAppliedTools(req)
		for _, required := range turn.MatchToolsApplied {
			if !contains(appliedTools, required) {
				return false
			}
		}
	}

	return true
}

func extractUserInput(req llm.GenerateRequest) string {
	for _, item := range req.Input {
		if item.Type == "user_message" {
			return item.Content
		}
	}
	return ""
}

func extractAppliedTools(req llm.GenerateRequest) []string {
	var tools []string
	for _, item := range req.Input {
		if item.Type == "tool_result" && item.ToolName != "" {
			tools = append(tools, item.ToolName)
		}
	}
	return tools
}

// contains 检查字符串切片是否包含指定字符串
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// Reset 重置场景状态，从头开始
func (b *ScenarioBackend) Reset() {
	b.currentTurnIdx = 0
}

// SetTurnIndex 设置当前 Turn 索引（用于跳过某些 Turn）
func (b *ScenarioBackend) SetTurnIndex(idx int) {
	b.currentTurnIdx = idx
}

// RemainingTurns 返回剩余未使用的 Turn 数量
func (b *ScenarioBackend) RemainingTurns() int {
	return len(b.scenario.Turns) - b.currentTurnIdx
}

// MatchesInput 检查输入是否匹配场景的任意 Turn（用于测试和调试）
func (b *ScenarioBackend) MatchesInput(input string) bool {
	for _, turn := range b.scenario.Turns {
		if turn.MatchInput == input {
			return true
		}
		if turn.MatchPattern != "" {
			if matched, _ := regexp.MatchString(turn.MatchPattern, input); matched {
				return true
			}
		}
	}
	return false
}

// DumpScenario 返回场景的调试字符串
func (b *ScenarioBackend) DumpScenario() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Scenario: %s\n", b.scenario.Name))
	sb.WriteString(fmt.Sprintf("Description: %s\n", b.scenario.Description))
	sb.WriteString(fmt.Sprintf("Total Turns: %d\n", len(b.scenario.Turns)))
	sb.WriteString(fmt.Sprintf("Current Turn Index: %d\n", b.currentTurnIdx))

	for i, turn := range b.scenario.Turns {
		marker := " "
		if i == b.currentTurnIdx {
			marker = ">"
		}
		sb.WriteString(fmt.Sprintf("  %s Turn %d: match_input=%q, response_id=%s, events=%d\n",
			marker, i, turn.MatchInput, turn.ResponseID, len(turn.Events)))
	}

	return sb.String()
}
