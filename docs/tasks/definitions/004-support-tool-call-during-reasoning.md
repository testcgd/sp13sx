---
id: "004-support-tool-call-during-reasoning"
title: "Support tool execution during reasoning with continued inference"
status: "in_progress"
review_attempts: 0
max_review_attempts: 5
labels: ["feature", "streaming", "reasoning", "tools"]
dependencies: ["003-fix-tool-calls-index"]
---

# Support tool execution during reasoning with continued inference

## Description

当前实现中，`tool_call` 事件在流式响应结束后才发送，runtime 收集所有 tool_calls 后才执行。这意味着无法在推理过程中调用 tool，也无法在 tool 执行结束后继续推理。此外，`reasoning_content` 没有正确传递到下一轮请求，导致模型丢失推理上下文。

### 当前流程

```
流式输出: reasoning → message → tool_call → [流结束]
                                              ↓
                                         执行 tool
                                              ↓
                                         新一轮请求（丢失 reasoning_content）
```

### 当前问题

1. **tool_call 延迟**: `tool_call` 事件在流结束后才发送
2. **reasoning_content 丢失**: tool_result 后的新请求没有包含之前的 `reasoning_content`
3. **推理链断裂**: 模型无法基于之前的推理继续思考

### 期望流程

```
流式输出: reasoning → tool_call (完整) → [立即执行 tool]
                    ↓
              发送 tool_result
                    ↓
         新请求（包含 reasoning_content）
                    ↓
          继续推理 reasoning → message → [流结束或下一个 tool_call]
```

## Acceptance Criteria

### 核心功能

- [x] 当流式响应中收到完整的 tool_call 时，立即发送 `tool_call` 事件
- [ ] Runtime 收到 `tool_call` 事件时，立即执行 tool（不等流结束）
- [x] Tool 执行完成后，发起新请求，包含 `reasoning_content`
- [x] 支持多个 tool_calls 串行执行，每次都传递累积的 `reasoning_content`

### reasoning_content 传递规则

- [x] **同一 thinking + tool 循环中**: 保留并发送 `reasoning_content`
- [x] **assistant message 格式**: 包含 `content`、`reasoning_content`、`tool_calls`
- [x] **新用户问题**: 清除旧的 `reasoning_content`，开始新的推理循环
- [x] **API 兼容**: 支持 DeepSeek V3.2、GLM 4.6、Minimax M2、Kimi-k2-thinking 等模型

### 其他

- [x] TUI 实时显示 tool 执行状态和推理过程
- [x] 现有测试全部通过
- [x] 添加相关集成测试

### 未完成项说明

**Runtime 收到 tool_call 事件时，立即执行 tool（不等流结束）**

此功能受 Chat Completions API 限制，无法在流中途追加输入。当前实现采用方案 C：
- 流结束后执行所有 tool_calls
- 新请求包含累积的 reasoning_content
- 用户感知上接近"推理中执行 tool"的效果

如需真正的"流中途执行"，需要切换到 OpenAI Responses API。

## Technical Design

### 1. 消息格式定义

assistant 消息需要包含 `reasoning_content`：

```go
type chatMessage struct {
    Role             string        `json:"role"`
    Content          any           `json:"content,omitempty"`
    ReasoningContent string        `json:"reasoning_content,omitempty"`
    ToolCalls        []toolCallMsg `json:"tool_calls,omitempty"`
    ToolCallID       string        `json:"tool_call_id,omitempty"`
}
```

### 2. runtime.go 修改

在 tool_result 后发起的新请求中，必须包含 `reasoning_content`：

```go
func (r *Runtime) runTurnLoop(...) {
    var accumulatedReasoning string
    
    for {
        stream, _ := r.Backend.Generate(ctx, current)
        
        for event := range stream {
            if event.Type == "reasoning" {
                accumulatedReasoning += event.ReasoningContent
            }
            if event.Type == "tool_call" && isComplete(event.ToolCall) {
                // 立即执行 tool
                result := executeTool(event.ToolCall)
                
                // 构建下一轮消息（关键：包含 reasoning_content）
                messages = append(messages, chatMessage{
                    Role:             "assistant",
                    Content:          accumulatedContent,
                    ReasoningContent: accumulatedReasoning,  // 必须传递！
                    ToolCalls:        []toolCallMsg{...},
                })
                messages = append(messages, chatMessage{
                    Role:       "tool",
                    ToolCallID: event.ToolCall.ID,
                    Content:    result,
                })
                
                // 发起新请求
                current = llm.GenerateRequest{
                    Model:    r.Session.Model,
                    Input:    messages,
                    Tools:    buildToolDefinitions(r.Tools),
                }
                break  // 退出内层循环，开始新一轮请求
            }
            out <- event
        }
    }
}
```

### 3. 关键实现细节

#### 3.1 reasoning_content 清除时机

```go
// 新用户问题时，清除旧的 reasoning_content
func (r *Runtime) Send(ctx context.Context, text string) (<-chan llm.StreamEvent, error) {
    // 新的用户输入，重置推理上下文
    r.clearReasoningContext()
    
    // 开始新的推理循环
    ...
}
```

#### 3.2 流式响应处理

```go
// 在 processStream 中
// delta 可能同时包含 reasoning_content、content、tool_calls
for _, tc := range delta.ToolCalls {
    // 累积参数
    if isCompleteToolCall(tc) {
        // 参数完整，立即发送 tool_call 事件
        out <- llm.StreamEvent{Type: "tool_call", ToolCall: tc}
    }
}
```

#### 3.3 检测完整 tool_call

```go
func isCompleteToolCall(tc *llm.ToolCall) bool {
    if tc.ID == "" || tc.Name == "" {
        return false
    }
    // 尝试解析 arguments 为有效 JSON
    if tc.Arguments == nil {
        return false
    }
    _, err := json.Marshal(tc.Arguments)
    return err == nil
}
```

### 4. API 行为差异

| 模型/API | reasoning_content 传递要求 |
|----------|---------------------------|
| DeepSeek V3.2 | 必须传递，否则 API 返回 400 错误 |
| GLM 4.6 | 建议传递，支持 interleaved thinking |
| Minimax M2 | 建议传递 |
| Kimi-k2-thinking | 建议传递 |
| OpenAI Chat Completions | 不要求，但传递可保持推理一致性 |
| OpenAI Responses API | 原生支持，自动管理推理状态 |

### 5. 错误处理

如果 API 返回错误提示 `reasoning_content is missing in assistant tool call message`：

```
{
  "error": {
    "message": "thinking is enabled but reasoning_content is missing in assistant tool call message at index 4",
    "type": "invalid_request_error"
  }
}
```

需要检查 assistant 消息是否正确包含了 `reasoning_content` 字段。

## Notes

### 相关文件

- `internal/llm/openai/chat.go` - 检测完整 tool_call，累积 reasoning_content
- `internal/app/runtime.go` - 执行 tool 并传递 reasoning_content 发起继续请求
- `internal/llm/backend.go` - 可能需要新增接口方法
- `internal/domain/message.go` - Message 结构体已有 Reasoning 字段

### API 参考

- OpenAI Responses API: https://developers.openai.com/cookbook/examples/reasoning_function_calls
- DeepSeek Thinking Mode: https://api-docs.deepseek.com/guides/thinking_mode
- LangChain PR #34177 (interleaved reasoning): https://github.com/langchain-ai/langchain/pull/34177

### 依赖

需要先完成 `003-fix-tool-calls-index`，修复 index 处理问题。

### 测试场景

1. 单 tool 调用 + reasoning_content 传递
2. 多 tool 串行调用 + reasoning_content 累积
3. 新用户问题清除旧 reasoning_content
4. Interleaved thinking 模型支持

## Review Feedback

验收失败时，编排器会在这里追加反馈。
