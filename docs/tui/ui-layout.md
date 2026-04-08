# TUI 布局

当前 TUI 使用双栏布局：

- 左栏：对话流与输入框
- 右栏：backend、model、session、skills、MCP server 状态、错误信息

## 当前命令

- `/help`
- `/session list`
- `/skill list`
- `/mcp list`

## 当前运行反馈

左侧对话流中除了用户和 assistant 消息，还会插入系统反馈，例如：

- 工具被请求
- 工具进入队列
- 工具执行完成
- 工具执行失败

这让当前版本即使右侧面板还比较轻，也能看见完整运行链路。
