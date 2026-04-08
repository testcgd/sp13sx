# Agent 执行协议

## 实现 Agent

实现 agent 的输入：

1. 任务定义文件路径。
2. 当前 worktree 路径。
3. 仓库约束文件 `.agent/agent.md`。
4. 如果不是第一次实现，还会包含最近一次验收报告路径。

实现 agent 的职责：

1. 阅读 task 文档和最近反馈。
2. 在 worktree 内完成实现。
3. 运行 `gofmt` 和 `go test ./...`。
4. 生成一个可提交的变更集。
5. 提交并推送 `task/<task-id>` 分支。

实现 agent 的输出：

- 已提交的代码。
- 最新提交 SHA。
- 命令执行日志。

## 验收 Agent

验收 agent 必须是新的 agent 进程，不能复用实现 agent 的上下文。

输入：

1. 任务定义文件。
2. 当前 worktree 的 diff 和提交历史。
3. 验证命令输出。
4. `.agent/review.md`。

职责：

1. 对照 task 的 `Acceptance Criteria` 做逐项检查。
2. 判断实现是否满足可合入 `master` 的标准。
3. 输出 `PASSED` 或 `FAILED` 报告。
4. 若失败，明确列出必须修复的问题。

输出文件统一写入：

`docs/tasks/reviews/<task-id>/review-<n>.md`

## 修复 Agent

修复 agent 只在 task 已合入 `master` 但主干验证失败时触发。

输入：

1. 当前失败日志。
2. 触发失败的 task 定义。
3. 主干最新状态。

职责：

1. 在同一 task worktree 中修复集成问题。
2. 重新提交到同一分支。
3. 让编排器继续发起新的本地修复集成。

## 不变约束

- 实现、验收、修复三个阶段必须使用独立 agent 进程。
- 验收失败不能直接合并。
- 所有自动提交都必须可通过 `go test ./...`。
- 编排器只按队列顺序推进，不并行执行多个 task。
