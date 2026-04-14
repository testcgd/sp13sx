---
id: "003-fix-tool-calls-index"
title: "Fix tool_calls index handling in streaming response"
status: "passed"
review_attempts: 0
max_review_attempts: 5
labels: ["bug", "streaming", "openai"]
dependencies: []
---

# Fix tool_calls index handling in streaming response

## Description

在 `internal/llm/openai/chat.go` 的 `processStream` 函数中，处理流式响应的 `tool_calls` 时，index 被硬编码为 `0`，这会导致当模型返回多个 tool_calls 时无法正确累积参数。

### 当前问题代码

```go
for _, tc := range delta.ToolCalls {
    idx := 0  // BUG: 硬编码为 0
    if tc.ID != "" {
        // ...
    }
}
```

### 预期行为

根据 OpenAI Chat Completions API 规范，`tool_calls` 数组中的每个元素都有一个 `index` 字段，用于标识这是第几个 tool call。应该使用这个 index 来正确累积参数。

## Acceptance Criteria

- [ ] `toolCallMsg` 结构体添加 `Index int` 字段
- [ ] `processStream` 使用 `tc.Index` 而非硬编码 `0`
- [ ] 添加多 tool_calls 场景的测试用例
- [ ] 现有测试全部通过

## Notes

### 相关文件

- `internal/llm/openai/chat.go` - 需要修复的主要文件
- `internal/llm/backend.go` - `ToolCall` 结构体定义

### API 参考

OpenAI Chat Completions API 中 `ChoiceDeltaToolCall` 结构：

```json
{
  "index": 0,
  "id": "call_xxx",
  "type": "function",
  "function": {
    "name": "read_file",
    "arguments": "{\"path\":"
  }
}
```

### 额外问题：tool_call 事件延迟发送

当前 `tool_call` 事件在流结束后才发送（第 214-226 行），这意味着：

1. 无法在推理过程中执行 tool
2. 用户体验不佳，需要等待整个流结束才能看到 tool 调用

未来可以考虑：
- 当 tool_call 参数完整时立即发送事件
- 在 runtime 层支持立即执行 tool_call

## Review Feedback

验收失败时，编排器会在这里追加反馈。
