# 编排器设计

## 职责

编排器负责把任务队列转化成一条可执行流水线：

1. 读取队列。
2. 创建 worktree。
3. 派发实现 agent。
4. 派发验收 agent。
5. 在独立集成 worktree 中执行本地 merge。
6. 在 `master` 上做最终验证。
7. 清理现场并推进下一个任务。

## 状态机

```text
pending -> in_progress -> review -> passed -> merged -> completed
                           |
                           v
                         failed -> in_progress
```

如果在 `merged` 后主分支验证失败：

```text
merged -> fix_in_progress -> merged -> completed
```

## Worktree 策略

- 根目录：`.worktrees/`
- 分支命名：`task/<task-id>`
- 路径命名：`.worktrees/<task-id>`

每个 task 只使用一个 worktree。验收失败后的重做、主分支验证失败后的修复，都回到同一个 worktree 继续。

## 外部依赖

- `git`
- `opencode`，或兼容的 agent CLI
- `go`
- `gofmt`

可通过环境变量替换：

- `OPENCODE_BIN`
- `MASTER_BRANCH`

## 集成策略

- 分支从 `master` 创建。
- 使用 squash merge。
- 不依赖任何远程代码托管平台。
- 集成在 `.worktrees/_integration` 中完成，使用专用分支 `integration/master`，避免与其他 worktree 上的 `master` 冲突。

推荐标题格式：

```text
<task-id>: <task-title>
```

## 验证策略

合入前：

- worktree 内执行 `gofmt -l .`
- worktree 内执行 `go test ./...`

合入后：

- `master` 上再次执行 `gofmt -l .`
- `master` 上再次执行 `go test ./...`

这样可以避免 worktree 本地通过但主干集成失败。

## 暂停与恢复

编排器通过 `.worktrees/.pause` 文件控制暂停。

- `./scripts/orchestrator.sh pause`
- `./scripts/orchestrator.sh resume`

暂停只影响下一个循环，不会强杀当前正在运行的 agent。
