# sp13sx

`sp13sx` 是一个基于 Go 的终端 AI 编程助手，使用 Bubble Tea 构建 TUI 界面。

## 功能特性

### LLM 后端

- **OpenAI Chat API 支持**：支持 OpenAI 及兼容 API（如腾讯云、DeepSeek 等）
- **流式响应**：实时显示 LLM 输出
- **推理支持**：支持 reasoning_content，可折叠显示推理过程
- **多后端配置**：支持配置多个 LLM 后端，运行时切换

### 工具调用

- **内置工具**：
  - `read_file`：读取文件内容
  - `list_files`：列出目录文件
  - `run_command`：执行命令（支持 allowlist 权限控制）
- **MCP 工具**：支持 Model Context Protocol 远程工具
- **工具回环**：自动处理工具调用结果，支持多轮工具链

### 会话管理

- **JSONL 持久化**：会话自动保存，支持历史恢复
- **多会话支持**：管理多个独立会话
- **上下文压缩**：支持长对话自动摘要

### Skill 系统

- **自动发现**：从目录自动加载 Skill 定义
- **YAML Frontmatter**：支持 Markdown 格式的 Skill 定义
- **项目级 Skills**：支持项目目录下的 `.agent/skills/`

### TUI 界面

- **双栏布局**：左侧对话，右侧工具状态
- **实时事件**：显示工具调用、状态变化
- **输入队列**：支持排队多轮输入
- **中断模式**：Esc 切换中断/排队模式

### 测试框架

- **Mock 后端**：使用场景脚本测试，无需真实 LLM
- **录制回放**：录制真实交互，后续回放测试
- **VHS 自动化**：支持 vhs 终端 UI 自动化测试
- **场景脚本**：YAML 格式定义测试场景

## 快速开始

### 安装

```bash
go build -o sp13sx ./cmd/sp13sx
```

### 配置

创建 `config.local.yml`（参考 `examples/config.example.yml`）：

```yaml
defaults:
  backend: openai
  model: gpt-4

backends:
  openai:
    type: openai-chat
    base_url: https://api.openai.com/v1
    api_key_env: OPENAI_API_KEY

tools:
  builtin:
    read_file:
      enabled: true
    run_command:
      enabled: true
      allowlist:
        - git status
        - go test ./...
```

### 运行

```bash
# 正常模式
./sp13sx -f ./config.local.yml

# Mock 模式测试
./sp13sx -f ./config.local.yml test -scenario basic_chat

# 录制真实交互
SP13SX_TEST_MODE=record ./sp13sx -f ./config.local.yml
```

## 测试

### 单元测试

```bash
go test ./...
```

### 集成测试

```bash
SP13SX_INTEGRATION_TEST=1 go test ./test/integration/... -v
```

### VHS TUI 测试

```bash
# 安装依赖
sudo apt install ttyd
go install github.com/charmbracelet/vhs@latest

# 运行测试
./test/vhs/run.sh all
```

## 项目结构

```
sp13sx/
├── cmd/sp13sx/           # 程序入口
├── internal/
│   ├── app/              # 运行时组装、会话流程
│   ├── tui/              # Bubble Tea TUI 模型
│   ├── llm/              # LLM 后端抽象
│   │   ├── openai/       # OpenAI Chat 实现
│   │   ├── mock/         # Mock 场景后端
│   │   └── recorder/     # 录制回放后端
│   ├── tools/            # 工具注册与执行
│   ├── mcp/              # MCP 客户端管理
│   ├── skills/           # Skill 发现与加载
│   ├── config/           # YAML 配置解析
│   └── store/            # JSONL 持久化
├── test/
│   ├── e2e/              # 端到端测试
│   ├── integration/      # 集成测试
│   ├── scenarios/        # Mock 场景脚本
│   └── vhs/              # VHS TUI 测试
├── examples/             # 示例配置和 Skills
```

## 开发命令

```bash
# 运行程序
go run ./cmd/sp13sx

# 格式化
gofmt -w $(find . -name '*.go' -print)

# 运行测试
go test ./...
```

## 配置说明

### 环境变量

| 变量 | 说明 |
|------|------|
| `SP13SX_CONFIG` | 配置文件路径 |
| `SP13SX_TEST_MODE` | 测试模式：live/mock/playback/record |
| `SP13SX_SCENARIO` | Mock 场景名称 |
| `SP13SX_RECORDING` | 回放录制文件路径 |
| `SP13SX_RECORDING_OUTPUT` | 录制输出路径 |
| `OPENAI_API_KEY` | OpenAI API 密钥 |

### 命令行参数

```bash
sp13sx [-f 配置文件] [command]

Commands:
  test      使用 Mock 模式测试 TUI
  record    录制真实 LLM 交互
  validate  验证场景脚本文件
  help      显示帮助
```

## 状态

当前为早期开发版本，核心功能已实现：
- TUI 界面与交互
- LLM 对话与工具调用
- 会话持久化
- Mock 测试框架
- VHS 自动化测试

待完善：
- MCP SDK 集成
- 更多内置工具
- 会话管理 UI
