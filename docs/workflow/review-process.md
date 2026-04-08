# 验收流程

## 目标

验收流程把“实现完成”与“允许合入”拆成两个独立决策，避免实现 agent 自评通过。

## 流程

1. 实现 agent 完成任务并推送分支。
2. 编排器启动新的验收 agent。
3. 验收 agent 读取 task、diff、测试结果。
4. 输出验收报告。
5. 若 `PASSED`，编排器继续执行本地 merge。
6. 若 `FAILED`，编排器把报告摘要追加到 task 文档，再次启动实现 agent。

## 验收报告格式

```markdown
---
status: PASSED
task_id: "001-bootstrap-workflow"
review_number: 2
---

# Review Report

## Summary

所有验收标准均满足。

## Acceptance Criteria Check

- [x] 文档补齐
- [x] 脚本可运行
- [x] `go test ./...` 通过

## Findings

无。
```

失败示例：

```markdown
---
status: FAILED
task_id: "001-bootstrap-workflow"
review_number: 1
---

# Review Report

## Findings

1. `scripts/orchestrator.sh` 未在主干上执行最终验证。
2. `docs/tasks/queue.yaml` 没有说明运行时如何改顺序。
```

## 反馈回写策略

验收失败后，编排器会把失败摘要追加到任务文档的 `## Review Feedback` 下，例如：

```markdown
### Review Attempt 1

- `scripts/orchestrator.sh` 未在主干上执行最终验证。
- `docs/tasks/queue.yaml` 缺少顺序调整说明。
```

这样下一次实现 agent 可以直接把 task 文档当成完整输入，而不依赖外部上下文。
