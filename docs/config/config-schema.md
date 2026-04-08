# 配置结构

`config.yml` 用于保存稳定配置，不保存对话历史或流式运行事件。

## 顶层字段

- `app`：应用名、数据目录、日志级别
- `ui`：TUI 布局与显示选项
- `defaults`：默认 backend、model、skill 目录、默认启用 skills
- `backends`：各模型后端配置
- `tools`：内置工具开关与策略
- `mcp`：MCP server 定义
- `skills`：skill 自动发现策略
- `sessions`：会话默认行为
- `storage`：JSONL 存储目录

## 当前约束

- 密钥不直接写入 YAML
- 使用环境变量名引用密钥，例如 `OPENAI_API_KEY`
- MCP server 定义放在 YAML 中，由运行时启动
- tool allowlist 也放在 YAML 中，由运行时执行时校验
