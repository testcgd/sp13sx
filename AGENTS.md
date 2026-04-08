# Repository Guidelines

## 项目结构与模块划分

- `cmd/sp13sx`：程序入口。
- `internal/app`：运行时组装、会话流程、工具调用回环。
- `internal/tui`：Bubble Tea 界面模型、更新循环、视图渲染。
- `internal/llm`：模型后端抽象；`internal/llm/openai` 为 OpenAI Responses 实现；`internal/llm/mock` 为测试用 mock backend。
- `internal/mcp`：MCP client 管理、远程工具发现与适配。
- `internal/tools`：统一工具接口、注册表、执行器、内置工具。
- `internal/config`：YAML 配置加载、校验、保存。
- `internal/store`：JSONL 持久化。
- `docs`：中文设计文档。
- `examples`：示例配置与示例 skill。

## 构建、测试与开发命令

- `go test ./...`：运行全部测试。
- `gofmt -w $(find . -name '*.go' -print)`：格式化所有 Go 文件。
- `go run ./cmd/sp13sx`：直接启动程序。
- `./scripts/dev.sh`：使用示例配置本地运行。
- `./scripts/test.sh`：测试脚本封装。

如需指定配置文件，可使用：

```bash
SP13SX_CONFIG=/path/to/config.yml go run ./cmd/sp13sx
```

## 代码风格与命名约定

- 必须遵循标准 Go 风格，并在提交前运行 `gofmt`。
- 导出标识符使用 `CamelCase`，未导出标识符使用 Go 常规命名。
- 包应保持单一职责，优先放在 `internal/...`，不要过早暴露公共 API。
- 测试函数命名使用 `TestXxx`。
- MCP 工具命名遵循 `server_name.tool_name`。

## 测试规范

- 使用 Go 原生 `testing` 包。
- 优先写 mock 驱动测试，不依赖真实 OpenAI 或 MCP 服务。
- 重点覆盖：
  - runtime 工具回环
  - TUI 事件处理
  - JSONL 持久化
  - 配置解析与校验

## 提交与合并请求规范

- 提交信息使用简短祈使句，例如：`Add mock backend tests`。
- 一个提交尽量只包含一个逻辑变更。
- PR 需要包含：
  - 变更摘要
  - 测试结果，例如 `go test ./...`
  - 配置或接口影响说明
  - 若改动 TUI，附终端截图或输出说明

## 安全与配置建议

- 不要提交密钥或令牌。
- OpenAI 密钥通过环境变量注入，例如 `OPENAI_API_KEY`。
- 命令执行能力必须受 YAML allowlist 控制，不要随意放宽 `run_command` 权限。
