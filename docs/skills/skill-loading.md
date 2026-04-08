# Skill 加载

skills 采用标准 Codex 结构：

- 每个 skill 一个目录
- 必须包含 `SKILL.md`
- 可以带 frontmatter，例如 `name`、`description`、`metadata`

## 当前运行时流程

- 从配置中的 `skill_dirs` 发现目录
- 读取并解析 `SKILL.md`
- 使用 frontmatter 名称或目录名作为 skill 名
- 将启用的 skills 拼装进最终 system prompt

## 当前限制

- 目前只做指令加载，不做 skill 专属执行钩子
- `references/`、`scripts/`、`assets/` 目录结构可兼容，但还没有做更深层自动装配
