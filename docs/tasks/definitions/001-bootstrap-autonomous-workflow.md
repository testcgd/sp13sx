---
id: "001-bootstrap-autonomous-workflow"
title: "Bootstrap autonomous workflow"
status: "pending"
review_attempts: 0
max_review_attempts: 5
labels: ["workflow", "automation"]
dependencies: []
---

# Bootstrap autonomous workflow

## Description

为当前仓库建立一套基于 task、repo、agent 和 `git worktree` 的长期无人监督开发工作流。

## Acceptance Criteria

- [ ] `docs/workflow/` 下补齐工作流文档
- [ ] `docs/tasks/` 下建立任务队列和任务模板
- [ ] `.agent/review.md` 和 `.agent/orchestrator.yaml` 建立完成
- [ ] `scripts/orchestrator.sh` 及相关脚本可用于驱动该流程
- [ ] 编排器只依赖 `git`，不依赖 `gh`
- [ ] `go test ./...` 通过

## Review Feedback

验收失败时，编排器会在这里追加反馈。
