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
