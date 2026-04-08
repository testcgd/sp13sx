# 架构总览

`sp13sx` 是一个终端优先的 Go agent，运行时由几个清晰分层的模块组成：

- `config`：加载并校验 YAML 配置
- `store`：负责 JSONL 持久化与回放
- `skills`：发现标准 `SKILL.md` 目录并构造有效指令
- `tools`：统一封装内置工具与 MCP 远程工具
- `llm`：抽象模型后端
- `mcp`：管理 MCP client 连接、拉取工具并适配
- `tui`：渲染双栏界面并处理事件

## 当前运行时主流程

1. 加载 `config.yml`
2. 恢复或创建会话
3. 发现本地 skills
4. 连接已启用的 MCP server
5. 注册内置工具和 MCP 远程工具
6. 启动 Bubble Tea 事件循环
7. 用户输入后发起 Responses 请求
8. 若模型请求工具，则执行工具并继续同一轮 Responses 调用

## 当前实现特征

- 配置与会话持久化已经分离
- 工具执行链路已经统一，不区分本地工具和 MCP 工具
- OpenAI backend 已支持流式文本与工具调用事件
- TUI 已能展示基础的运行状态和工具反馈
