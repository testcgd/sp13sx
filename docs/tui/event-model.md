# TUI 事件模型

TUI 基于 Bubble Tea 的 `Update` 循环驱动。

## 事件来源

- 终端窗口尺寸变化
- 键盘输入
- runtime 返回的流式事件

## 当前处理的 runtime 事件

- `message`
- `tool_call`
- `status`
- `response_id`
- `error`

## 当前行为

- `message`：增量追加 assistant 文本
- `tool_call`：在对话区显示模型请求了哪个工具
- `status`：显示工具执行状态
- `response_id`：把 UI 状态切到 streaming
- `error`：在右栏中显示错误

## 设计边界

TUI 负责展示，不直接负责持久化。
消息写盘和工具事件写盘由 runtime/store 负责。
