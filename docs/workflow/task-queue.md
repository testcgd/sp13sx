# Task 队列设计

## 目标

队列需要支持两件事：

- 编排器运行时继续添加任务。
- 编排器运行时直接调整任务顺序。

因此，队列索引使用单独文件维护，编排器每轮循环重新读取，不缓存整个队列状态。

## 队列文件

文件路径：`docs/tasks/queue.yaml`

示例：

```yaml
version: 1

current:
  id: ""
  started_at: ""
  worktree: ""

pending:
  - "001-bootstrap-workflow"
  - "002-add-ci"

history:
  - id: "000-initial-setup"
    status: completed
    completed_at: "2026-04-09T08:00:00Z"
```

## 字段说明

- `current`：当前正在执行的任务。编排器独占写入。
- `pending`：等待执行的任务 ID 列表。用户可以直接编辑顺序。
- `history`：历史记录。编排器只追加，不回写旧项。

## 任务定义

任务定义放在 `docs/tasks/definitions/<task-id>.md`。

示例：

```markdown
---
id: "001-bootstrap-workflow"
title: "Bootstrap autonomous workflow"
status: "pending"
review_attempts: 0
max_review_attempts: 5
labels: ["workflow"]
dependencies: []
---

# Bootstrap autonomous workflow

## Description

实现长期无人监督开发工作流的第一版。

## Acceptance Criteria

- [ ] 文档补齐
- [ ] 脚本可运行
- [ ] `go test ./...` 通过

## Review Feedback

验收失败时，编排器会在本节后面追加新的反馈块。
```

## 动态调整方式

推荐直接编辑 `docs/tasks/queue.yaml`。

支持的操作：

- 往 `pending` 末尾追加新 task。
- 调整 `pending` 中 task 顺序。
- 删除尚未执行的 task。

不建议手动编辑 `current` 或 `history`，这两部分由编排器维护。

## 编排器读取规则

1. 每轮循环重新读取 `queue.yaml`。
2. 只处理 `pending` 的第一项。
3. 当前 task 处理期间，不会抢占切换到新的 task。
4. 只有当前 task 完整结束后，新的顺序调整才会影响后续执行。

## 失败重试

- 验收失败：追加反馈到 task 文档，继续同一 task。
- 合入后主分支验证失败：派发修复 agent，仍然归属于同一 task。
- 超过最大重试次数：编排器停止并等待人工处理。
