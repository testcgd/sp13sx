# MCP 客户端生命周期

当前实现已经使用 Go MCP SDK 建立真实的 stdio MCP client。

## 当前职责

- 从 YAML 加载 MCP server 配置
- 对已启用 server 启动子进程
- 通过 `mcp.CommandTransport` 建立连接
- 初始化 client session
- 调用 `ListTools`
- 将远程工具注册进统一工具系统
- 在 TUI 右栏中显示连接状态

## 生命周期流程

对每个启用的 MCP server，运行时执行：

1. 根据配置构造 `exec.Cmd`
2. 创建 `mcp.NewClient(...)`
3. 调用 `client.Connect(...)`
4. 调用 `session.ListTools(...)`
5. 将返回工具适配成内部 `tools.Tool`
6. 将 server 状态记为 `connected`

如果某一步失败：

- 状态会变成 `error`
- 错误文本会进入右侧状态栏

## 当前限制

- 目前只实现 stdio transport
- 还没有自动重连
- 还没有接入 prompt/resource 相关能力
- 还没有处理 tool list changed 通知
