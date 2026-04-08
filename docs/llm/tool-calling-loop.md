# 工具调用回环

`sp13sx` 把工具调用回环放在 runtime 层实现，而不是直接塞进 OpenAI backend。

## 执行流程

1. 用户消息先写入 `messages.jsonl`
2. runtime 组装一轮 Responses 请求
3. OpenAI 流式返回文本增量与可能的 tool call
4. runtime 收集本轮响应里的所有工具调用
5. 如果没有工具调用，则本轮结束
6. 如果有工具调用：
   - 写入 `tool.queued`
   - 通过统一工具执行器调用本地工具或 MCP 工具
   - 写入 `tool.completed` 或 `tool.error`
   - 将结果转成 `function_call_output`
7. runtime 使用 `previous_response_id` 发起下一轮 Responses 请求
8. 循环直到模型不再请求工具

## 工具命名规则

- 内置工具保持原名
- MCP 工具使用 `server_name.tool_name`

这样模型看到的是一套统一的工具空间，执行时也走统一链路。

## 持久化边界

当前会持久化：

- 用户消息
- 最终 assistant 消息
- 工具执行事件

assistant 的增量 token 只在 TUI 中显示，不单独落盘。
