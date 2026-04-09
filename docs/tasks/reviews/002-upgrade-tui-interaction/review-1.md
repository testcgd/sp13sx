---
status: FAILED
task_id: "002-upgrade-tui-interaction"
review_number: 1
---

# Review Report

## Summary

实现完成了大部分 TUI 升级目标，但仍有阻止合入的问题：工具失败事件会重新插入主对话区，未完全满足事件分层展示要求。

## Acceptance Criteria Check

- [x] TUI 支持清晰可用的输入交互：保留发送快捷键，并支持显式多行输入
- [x] `/help` 能展示当前命令与关键快捷键，不再只列出少量命令名
- [x] TUI 明确区分至少五种运行状态：`idle`、`waiting_model`、`streaming`、`running_tool`、`error`
- [x] 状态切换与真实运行阶段一致，assistant 流式输出未结束前不会错误回到 ready/idle
- [x] 左侧主对话区优先呈现 user/assistant 对话，不再让工具事件把主回复切碎成日志墙
- [ ] 工具请求、排队、完成、失败等运行事件被收纳到独立区域或以明显弱化方式呈现，保证长对话可读性
- [x] 右栏改为实时上下文面板，至少动态展示当前状态、backend、model、session、最近工具事件摘要、MCP 状态、当前错误摘要
- [x] user、assistant、system、error、tool 状态在视觉上有明确区分，终端内能快速扫读
- [x] 错误出现时，用户可以在界面中清楚知道当前失败状态和后续可继续操作的路径
- [x] 现有命令 `/help`、`/session list`、`/skill list`、`/mcp list` 保持可用
- [ ] 新增或重构后的 TUI 行为有相应测试覆盖，重点覆盖输入、状态流转、工具事件展示、错误展示
- [x] `go test ./...` 通过

## Findings

1. `internal/tui/update.go:122-125` 在收到 `tool error:` 状态时，除了把事件写入事件区，还会额外 `appendMessage("error", statusLine)` 把工具失败重新插入主对话区。任务要求明确包含“工具请求、排队、完成、失败”等运行事件都应被收纳到独立区域或弱化呈现；当前实现下，失败类工具事件仍会打断主对话阅读流，未完成该验收项。
2. `internal/tui/model_test.go` 目前覆盖了输入、多行、状态流转和事件记录，但没有断言 `Model.View()` 的实际渲染结果，因此新的“展示”行为仍缺少直接测试：例如独立 Events 区、右栏最近工具事件摘要、MCP/error 摘要、错误后的继续操作提示都没有被验证。这与 task 对“工具事件展示、错误展示”重点覆盖的要求仍有差距。
