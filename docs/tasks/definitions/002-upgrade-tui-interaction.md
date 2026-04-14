---
id: "002-upgrade-tui-interaction"
title: "Upgrade TUI interaction to a complete usable flow"
status: "review"
review_attempts: 2
max_review_attempts: 5
labels: ["tui", "ux", "agent"]
dependencies: []
---

# Upgrade TUI interaction to a complete usable flow

## Description

把当前 `sp13sx` 的 MVP 双栏 TUI 升级为更完整、可持续使用的交互版本，重点解决输入体验薄弱、对话与运行日志混杂、状态表达不准确、右栏信息静态化、错误恢复不清晰等问题。

本任务只改进终端交互层与必要的 runtime/TUI 协作边界，不扩展为复杂 GUI，也不引入与当前架构不匹配的重型组件。目标是在保持 Bubble Tea 双栏结构的前提下，把当前原型提升到“适合日常 agent 使用”的水平。

## Acceptance Criteria

- [x] TUI 支持清晰可用的输入交互：保留发送快捷键，并支持显式多行输入
- [x] `/help` 能展示当前命令与关键快捷键，不再只列出少量命令名
- [x] TUI 明确区分至少五种运行状态：`idle`、`waiting_model`、`streaming`、`running_tool`、`error`
- [x] 状态切换与真实运行阶段一致，assistant 流式输出未结束前不会错误回到 ready/idle
- [x] 左侧主对话区优先呈现 user/assistant 对话，不再让工具事件把主回复切碎成日志墙
- [x] 工具请求、排队、完成、失败等运行事件被收纳到独立区域或以明显弱化方式呈现，保证长对话可读性
- [x] 右栏改为实时上下文面板，至少动态展示当前状态、backend、model、session、最近工具事件摘要、MCP 状态、当前错误摘要
- [x] user、assistant、system、error、tool 状态在视觉上有明确区分，终端内能快速扫读
- [x] 错误出现时，用户可以在界面中清楚知道当前失败状态和后续可继续操作的路径
- [x] 现有命令 `/help`、`/session list`、`/skill list`、`/mcp list` 保持可用
- [x] 新增或重构后的 TUI 行为有相应测试覆盖，重点覆盖输入、状态流转、工具事件展示、错误展示
- [x] `go test ./...` 通过

## Notes

建议实现范围和方向如下：

1. 输入模型
保留当前 `textarea`，但重新定义发送与换行行为。
建议采用：
- `Enter` 发送
- `Ctrl+J` 或 `Shift+Enter` 插入换行
- `Ctrl+C` 退出

如果 Bubble Tea/textarea 对 `Shift+Enter` 支持不稳定，可优先选 `Ctrl+J` 作为显式换行方案。

2. 状态模型
当前 `response_id`、`message`、`tool_call`、`status` 的处理会导致状态表达失真。
建议把 TUI 状态明确建模，并基于 runtime 事件结束点恢复到 `idle`，不要在收到首个 message chunk 后立即回到 ready。

3. 视图分层
保持双栏，但要把“对话”和“运行事件”分层展示。
推荐方向：
- 左栏主区域显示用户与 assistant 的 turn
- 左栏底部或右栏部分区域显示最近运行事件
- system/tool 事件不再无差别插入主对话正文

4. 右栏动态刷新
当前右栏内容在 `NewModel` 时一次性生成，后续不更新。
建议将右栏渲染改为基于当前 model 状态实时生成，而不是缓存静态字符串切片。

5. 视觉样式
在不脱离现有 Lip Gloss 风格的前提下，至少增加：
- role 区分
- 错误高亮
- 工具状态弱化或单独样式
- 当前状态栏更醒目

6. 测试建议
至少补齐：
- 输入发送与多行输入行为
- 流式消息期间状态流转
- tool_call/status 对主视图与事件区的影响
- error 展示与状态恢复
- `/help` 输出包含快捷键说明

7. 参考文件
- `internal/tui/model.go`
- `internal/tui/update.go`
- `internal/tui/view.go`
- `internal/tui/keymap.go`
- `internal/tui/model_test.go`
- `docs/tui/ui-layout.md`
- `docs/tui/event-model.md`

## Review Feedback

验收失败时，编排器会在这里追加反馈。

### Review Attempt 1

## Findings

1. `internal/tui/update.go:122-125` 在收到 `tool error:` 状态时，除了把事件写入事件区，还会额外 `appendMessage("error", statusLine)` 把工具失败重新插入主对话区。任务要求明确包含“工具请求、排队、完成、失败”等运行事件都应被收纳到独立区域或弱化呈现；当前实现下，失败类工具事件仍会打断主对话阅读流，未完成该验收项。
2. `internal/tui/model_test.go` 目前覆盖了输入、多行、状态流转和事件记录，但没有断言 `Model.View()` 的实际渲染结果，因此新的“展示”行为仍缺少直接测试：例如独立 Events 区、右栏最近工具事件摘要、MCP/error 摘要、错误后的继续操作提示都没有被验证。这与 task 对“工具事件展示、错误展示”重点覆盖的要求仍有差距。

### Review Attempt 1


1. `internal/tui/update.go:122-125` 在收到 `tool error:` 状态时，除了把事件写入事件区，还会额外 `appendMessage("error", statusLine)` 把工具失败重新插入主对话区。任务要求明确包含“工具请求、排队、完成、失败”等运行事件都应被收纳到独立区域或弱化呈现；当前实现下，失败类工具事件仍会打断主对话阅读流，未完成该验收项。
2. `internal/tui/model_test.go` 目前覆盖了输入、多行、状态流转和事件记录，但没有断言 `Model.View()` 的实际渲染结果，因此新的“展示”行为仍缺少直接测试：例如独立 Events 区、右栏最近工具事件摘要、MCP/error 摘要、错误后的继续操作提示都没有被验证。这与 task 对“工具事件展示、错误展示”重点覆盖的要求仍有差距。

### Fix Applied (2026-04-15)

1. **工具事件分离**: 添加 `events []string` 字段存储工具事件，`tool_call` 和 `status` 事件现在通过 `appendEvent()` 添加到独立区域，不再插入主对话区。
2. **右栏事件展示**: 在右栏添加 "recent events:" 区域显示最近10条工具事件。
3. **View 渲染测试**: 添加以下测试：
   - `TestViewContainsEventsSection`: 验证事件区域存在
   - `TestViewShowsErrorWithRecoveryHint`: 验证错误状态显示
   - `TestViewShowsStatus`: 验证状态显示
   - `TestEventsAreSeparatedFromMessages`: 验证事件与消息分离
