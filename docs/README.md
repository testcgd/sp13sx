# 文档目录

- `architecture`：运行时架构、模块边界、主流程
- `config`：`config.yml` 字段说明、约束与示例
- `storage`：JSONL 文件结构、回放策略、持久化边界
- `llm`：模型后端抽象、OpenAI Responses 接入、工具调用回环
- `tui`：界面布局、事件流、交互反馈
- `mcp`：MCP 客户端生命周期、工具适配方式
- `skills`：标准 `SKILL.md` 技能发现与指令拼装
- `adr`：关键架构决策记录
- `roadmap`：当前 MVP 与后续迭代方向
- `workflow`：无人监督长期开发流程、队列、验收与编排器设计
- `tasks`：任务队列、任务定义、验收报告与归档

## 当前实现状态

代码已经实现并通过编译验证的能力包括：

- OpenAI Responses API 接入
- SSE 流式响应消费
- tool call 识别与 `function_call_output` 回传
- MCP stdio 客户端接入
- 远程 MCP 工具注册进统一工具执行链路
- 双栏 TUI 中展示工具执行状态与 MCP 连接状态

实现以代码为准，最关键的入口文件如下：

- `internal/app/runtime.go`
- `internal/llm/openai/*`
- `internal/mcp/manager.go`
- `internal/tui/*`

## 工作流入口

- 启动编排器：`./scripts/orchestrator.sh`
- 查看队列：`./scripts/task-queue.sh status`
- 创建 task：`./scripts/task-create.sh <task-id> <title>`
- 工作流说明：`docs/workflow/README.md`
- task 说明：`docs/tasks/README.md`
