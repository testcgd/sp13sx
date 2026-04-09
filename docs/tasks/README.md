# Task 文档

## 结构

- `queue.yaml`：任务队列索引。
- `definitions/*.md`：待执行任务定义。
- `reviews/<task-id>/review-*.md`：验收记录。
- `completed/*.md`：已完成任务归档。

## 约定

- task 文件名应与 front matter 中的 `id` 一致。
- 编排器只从 `queue.yaml` 的 `pending` 列表取任务。
- 运行中允许直接编辑 `queue.yaml` 调整后续顺序。

## 如何添加 Task

方式一：使用脚本创建。

```bash
./scripts/task-create.sh 002 add-config-tests
```

脚本会生成 `docs/tasks/definitions/002.md`。生成后，把 task id 追加到 `docs/tasks/queue.yaml` 的 `pending` 列表。

```yaml
pending:
  - "001-bootstrap-autonomous-workflow"
  - "002"
```

方式二：直接复制 `docs/tasks/definitions/template.md` 并手工填写。

## 如何启动工作流

```bash
./scripts/orchestrator.sh
```

常用命令：

```bash
./scripts/task-queue.sh status
./scripts/task-queue.sh validate
./scripts/orchestrator.sh pause
./scripts/orchestrator.sh resume
```

## 主要工作流程

1. 编排器读取 `queue.yaml` 的 `pending` 首项。
2. 为该 task 创建独立 worktree。
3. 实现 agent 执行 task。
4. 新的 review agent 独立验收。
5. 若验收失败，失败内容回写到 task 文档，并重新进入实现阶段。
6. 若验收通过，在 `.worktrees/_integration` 中把 task 分支本地合入 `master`。
7. 合入后重新执行格式和测试验证。
8. 若验证失败，启动 fix agent 修复并再次集成。
9. 成功后归档 task 并继续下一个任务。
