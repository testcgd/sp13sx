# JSONL 模型

当前对话相关数据采用 append-only JSONL：

- `sessions.jsonl`：会话元数据
- `messages.jsonl`：用户与 assistant 消息
- `tool_calls.jsonl`：工具生命周期事件

## 回放策略

- session 使用 latest-record-wins 思路恢复最终状态
- message 采用追加顺序恢复
- tool call 采用事件流方式恢复

## 当前持久化策略

- 用户消息在发送前写入
- assistant 消息在整轮结束后写入最终聚合文本
- 工具队列、成功、失败都会写入 `tool_calls.jsonl`
