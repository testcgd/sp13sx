# 无人监督长期开发工作流

本文档描述一个基于 task、repo、agent 和 `git worktree` 的串行开发流程。目标是让系统在当前仓库中长期运行，持续消费任务队列，并把每个任务产出的代码以可直接合入 `master` 的本地分支变更形式落地。

## 核心原则

- 一个 task 只由一个实现 agent 负责落地。
- 每个 task 完成后，必须由一个新的验收 agent 独立验收。
- 只有验收通过的 task 才允许进入本地集成并合入 `master`。
- 合入后必须在 `master` 上重新执行完整验证。
- 若验证失败，派发修复 agent 修复，直到验证通过或达到最大重试次数。
- 编排器始终串行处理 task，但队列支持运行时继续追加和手工调整顺序。

## 目录约定

- `docs/tasks/queue.yaml`：任务队列和状态索引。
- `docs/tasks/definitions/*.md`：任务定义，使用 Markdown + YAML front matter。
- `docs/tasks/reviews/<task-id>/review-*.md`：验收报告。
- `docs/tasks/completed/*.md`：已完成任务归档。
- `.agent/agent.md`：实现 agent 约束。
- `.agent/review.md`：验收 agent 约束。
- `.agent/orchestrator.yaml`：编排器配置。
- `scripts/orchestrator.sh`：主循环入口。

## 主流程

1. 编排器轮询读取 `docs/tasks/queue.yaml`。
2. 取 `pending` 队首作为当前任务，并创建独立 worktree。
3. 派发实现 agent 在 worktree 中实现任务。
4. 派发新的验收 agent 对实现结果做独立验收。
5. 若验收失败，把失败原因追加回 task 文档，并重新派发实现 agent。
6. 若验收通过，在独立的集成 worktree 中执行本地 squash merge。
7. 在 `master` 上执行 `gofmt` 检查和 `go test ./...`。
8. 若验证失败，派发修复 agent 修复并再次合入，直到通过。
9. 清理 worktree，归档 task，继续处理下一个任务。

## 运行方式

```bash
./scripts/orchestrator.sh
```

常用辅助命令：

```bash
./scripts/task-queue.sh status
./scripts/task-queue.sh validate
./scripts/task-create.sh 001 add-runtime-logging
./scripts/orchestrator.sh pause
./scripts/orchestrator.sh resume
```

## 相关文档

- `docs/workflow/task-queue.md`
- `docs/workflow/orchestrator.md`
- `docs/workflow/agent-protocol.md`
- `docs/workflow/review-process.md`
