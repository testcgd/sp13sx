---
id: "005-support-interruptible-multi-turn-guidance"
title: "Support interruptible multi-turn guidance with pending input queue"
status: "passed"
review_attempts: 0
max_review_attempts: 5
labels: ["feature", "ux", "streaming", "tools"]
dependencies: ["004-support-tool-call-during-reasoning"]
---

# Support interruptible multi-turn guidance with pending input queue

## Description

当 thinking 模型正在执行多步 tool 调用时，用户可能希望：
1. **队列模式**: 等当前 tool 结束后，将用户输入附加到下一轮推理
2. **打断模式**: 取消当前 tool 执行，立即处理用户输入

### 场景示例

```
模型正在执行：
  reasoning → tool_call_1 → tool_result_1 → tool_call_2 → ...

用户输入："请优先处理错误情况"

队列模式（默认）：
  tool_call_2 → tool_result_2 → [用户输入] → 继续推理

打断模式（按 ESC 切换）：
  [取消 tool_call_2] → [用户输入] → 重新规划推理
```

### 当前问题

1. **TUI 层**: `submitInput()` 直接调用 `runtime.Send()`，不检查当前状态
2. **Runtime 层**: `runTurnLoop()` 是阻塞循环，没有队列机制
3. **无状态管理**: 不知道当前是否在 tool 执行中
4. **无打断机制**: 无法取消正在执行的 tool 或请求

## Acceptance Criteria

### 核心功能

- [ ] 用户在 tool 执行期间输入内容时，自动进入队列模式
- [ ] 排队的文本显示在对话框上方，标记"正在排队中"
- [ ] 支持多次输入，全部排队（FIFO 顺序）
- [ ] 按 ESC 可切换为打断模式，取消当前 tool 执行
- [ ] Tool 执行结束后，检查队列，依次处理用户输入
- [ ] 打断模式下，立即停止当前请求，处理用户输入

### UI/UX

- [ ] 排队输入显示在对话框上方（非消息流中）
- [ ] 显示排队状态：`⏳ 排队中: "用户输入内容..."`
- [ ] 打断模式显示：`⚡ 打断模式: 按 ESC 切换回队列模式`
- [ ] 右侧面板显示排队输入数量
- [ ] 输入框提示随状态变化：
  - 空闲: `Ask the agent or run /help`
  - tool 执行中: `输入将排队等待... (ESC 打断)`

## Technical Design

### 1. 数据结构

#### 1.1 Runtime 状态

```go
type Runtime struct {
    // 现有字段...
    
    // 新增
    mu              sync.RWMutex
    isRunningTool   bool
    currentCancel   context.CancelFunc
    pendingInputs   []string
}

type RuntimeStatus struct {
    IsRunningTool bool     `json:"is_running_tool"`
    PendingInputs []string `json:"pending_inputs"`
}
```

#### 1.2 TUI Model

```go
type Model struct {
    // 现有字段...
    
    // 新增
    pendingInputs   []string
    interruptMode   bool  // true = 打断模式, false = 队列模式
}
```

### 2. Runtime 实现

#### 2.1 Send 方法修改

```go
func (r *Runtime) Send(ctx context.Context, text string) (<-chan llm.StreamEvent, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if r.isRunningTool {
        // 在 tool 执行中，加入队列
        r.pendingInputs = append(r.pendingInputs, text)
        
        // 返回一个 channel 发送排队通知
        ch := make(chan llm.StreamEvent, 1)
        ch <- llm.StreamEvent{Type: "input_queued", Content: text}
        close(ch)
        return ch, nil
    }
    
    // 正常发送
    // ...
}
```

#### 2.2 Cancel 方法

```go
func (r *Runtime) Cancel() {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if r.currentCancel != nil {
        r.currentCancel()
    }
}
```

#### 2.3 runTurnLoop 修改

```go
func (r *Runtime) runTurnLoop(ctx context.Context, req llm.GenerateRequest, out chan<- llm.StreamEvent) {
    // 创建可取消的 context
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    r.mu.Lock()
    r.currentCancel = cancel
    r.mu.Unlock()
    
    for {
        // ... 执行 tool
        
        // tool 执行结束，检查队列
        r.mu.Lock()
        r.isRunningTool = false
        pending := r.pendingInputs
        r.pendingInputs = nil
        r.mu.Unlock()
        
        if len(toolCalls) == 0 {
            if len(pending) > 0 {
                // 有排队输入，附加到当前请求继续
                for _, input := range pending {
                    current.Input = append(current.Input, llm.UserTextInput(input))
                }
                continue
            }
            return
        }
        
        // 执行 tools...
        r.mu.Lock()
        r.isRunningTool = true
        r.mu.Unlock()
    }
}
```

### 3. TUI 实现

#### 3.1 Update 处理

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyEscape:
            // 切换打断模式
            if len(m.pendingInputs) > 0 {
                m.interruptMode = !m.interruptMode
                if m.interruptMode {
                    return m, m.cancelRuntime()
                }
            }
        case tea.KeyEnter:
            return m, m.submitInput()
        }
        
    case responseMsg:
        switch msg.event.Type {
        case "input_queued":
            m.pendingInputs = append(m.pendingInputs, msg.event.Content)
            m.interruptMode = false
        }
    }
    // ...
}
```

#### 3.2 View 渲染

```go
func (m Model) View() string {
    var b strings.Builder
    
    // 主视图
    b.WriteString(m.viewport.View())
    b.WriteString("\n")
    
    // 排队输入显示区域
    if len(m.pendingInputs) > 0 {
        b.WriteString(m.renderPendingInputs())
        b.WriteString("\n")
    }
    
    // 输入框
    b.WriteString(m.input.View())
    
    return b.String()
}

func (m Model) renderPendingInputs() string {
    var b strings.Builder
    b.WriteString("┌─ 排队输入 ─────────────────────────┐\n")
    for i, input := range m.pendingInputs {
        prefix := "⏳ "
        if m.interruptMode && i == 0 {
            prefix = "⚡ "
        }
        display := input
        if len(display) > 40 {
            display = display[:37] + "..."
        }
        b.WriteString(fmt.Sprintf("│ %s%s\n", prefix, display))
    }
    b.WriteString("└────────────────────────────────────┘\n")
    return b.String()
}
```

### 4. 流程图

```
用户输入 → 检查 isRunningTool
              │
              ├─ false → 正常 Send()
              │
              └─ true → 加入 pendingInputs
                         │
                         ├─ 队列模式（默认）→ 等待 tool 结束
                         │                    │
                         │                    └─ tool 结束 → 附加排队输入 → 继续推理
                         │
                         └─ 打断模式（ESC）→ Cancel() → 立即处理用户输入
```

### 5. UI 示例

队列模式：

```
┌─────────────────────────────────────────────┐
│ ASSISTANT                                   │
│ [reasoning ▸] 正在分析...                   │
│ Tool: read_file 执行中...                   │
├─────────────────────────────────────────────┤
│ ┌─ 排队输入 ────────────────────────────┐   │
│ │ ⏳ 请优先处理错误情况                  │   │
│ │ ⏳ 另外检查一下配置文件                │   │
│ └───────────────────────────────────────────┘   │
├─────────────────────────────────────────────┤
│ ⏳ 输入将排队等待... (ESC 打断)              │
│ > _                                          │
└─────────────────────────────────────────────┘
```

打断模式：

```
│ ┌─ 排队输入 ────────────────────────────┐   │
│ │ ⚡ 请优先处理错误情况                  │   │
│ │ ⏳ 另外检查一下配置文件                │   │
│ └───────────────────────────────────────────┘   │
├─────────────────────────────────────────────┤
│ ⚡ 打断模式 - 输入将立即处理 (ESC 切换)      │
│ > _                                          │
└─────────────────────────────────────────────┘
```

## Notes

### 相关文件

- `internal/app/runtime.go` - 添加状态管理、队列、取消机制
- `internal/tui/model.go` - 添加 pendingInputs、interruptMode
- `internal/tui/update.go` - 处理 ESC 事件、input_queued 事件
- `internal/tui/view.go` - 渲染排队输入区域
- `internal/llm/backend.go` - 新增 input_queued 事件类型

### 依赖

需要先完成 `004-support-tool-call-during-reasoning`，因为需要在 tool 循环中检查排队输入。

### 测试场景

1. 队列模式：tool 执行中输入 → 等待 → 自动附加
2. 打断模式：tool 执行中输入 → ESC → 立即处理
3. 多输入队列：连续输入多次 → 按顺序处理
4. 状态同步：右侧面板正确显示排队数量
5. 取消恢复：打断后队列内容不丢失

## Review Feedback

验收失败时，编排器会在这里追加反馈。
