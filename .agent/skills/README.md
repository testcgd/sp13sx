# 共享 Skills 目录

这个目录用于集中存放仓库级 skills，并通过软链接同时暴露给不同 agent 工具。

## 约定

- 公共 skill 放在 `.agent/skills/<skill-name>/`
- skill 内容尽量遵循标准 `SKILL.md` 结构
- 需要兼容多个 agent 时，以这里作为唯一源目录

## 当前软链接适配

- `.claude/skills` -> `.agent/skills`
- `.codex/skills` -> `.agent/skills`
- `.opencode/skills` -> `.agent/skills`

## 建议

- 每个 skill 使用独立目录
- 在 skill 目录内保留 `SKILL.md`
- 如有附加资料，可放 `references/`、`assets/`、`scripts/`
