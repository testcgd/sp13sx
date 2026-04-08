# MCP 工具适配

MCP 远程工具在运行时会被转换成内部统一的 `tools.Tool` 接口。

## 命名规则

每个远程工具会注册成：

- `server_name.tool_name`

这样可以避免不同 MCP server 暴露出同名工具时发生冲突。

## 当前映射内容

适配时会保留：

- 工具名
- 描述
- 输入 schema

执行时会调用：

- `session.CallTool(ctx, &mcp.CallToolParams{...})`

## 返回结果的统一化

当前 MCP 工具结果会被归一化成一个通用 map，可能包含：

- `is_error`
- `structured`
- `content`

这个结果随后会：

- 写入 `tool_calls.jsonl`
- 序列化为 `function_call_output` 回传给模型

## 当前限制

目前没有完整保留 MCP 原始 content model，而是优先转换成适合 agent 继续推理的通用结果结构。
