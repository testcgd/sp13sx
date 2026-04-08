# 包结构

- `cmd/sp13sx`：程序入口
- `internal/app`：启动与运行时组装
- `internal/tui`：Bubble Tea 模型、更新循环、视图渲染
- `internal/llm`：后端无关的请求、事件、工具调用类型
- `internal/llm/openai`：OpenAI Responses 实现
- `internal/tools`：统一工具接口、注册表、执行器
- `internal/mcp`：MCP client 管理与工具发现
- `internal/skills`：标准 skill 发现与 prompt 拼装
- `internal/config`：YAML 配置加载、校验、保存
- `internal/store`：JSONL 追加写入与回放
- `internal/domain`：共享领域模型
