package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"sp13sx/internal/config"
	"sp13sx/internal/domain"
	"sp13sx/internal/llm"
	"sp13sx/internal/mcp"
	"sp13sx/internal/skills"
	"sp13sx/internal/store"
	"sp13sx/internal/tools"
	"sp13sx/internal/tools/builtin"
	"sp13sx/internal/util"
)

const timeLayout = "2006-01-02T15:04:05Z07:00"

type Runtime struct {
	Config         config.Config
	StorePaths     store.Paths
	Backend        llm.Backend
	Tools          *tools.Registry
	SessionManager *SessionManager
	MCP            *mcp.Manager
	Skills         *skills.Registry
	Session        domain.Session

	mu            sync.RWMutex
	isRunningTool bool
	currentCancel context.CancelFunc
	pendingInputs []string
}

func NewRuntime(cfg config.Config) (*Runtime, error) {
	paths := store.NewPaths(cfg.Storage.ConversationsDir)
	sessionManager := NewSessionManager(paths)
	session, err := sessionManager.EnsureSession(cfg.Defaults.Backend, cfg.Defaults.Model)
	if err != nil {
		return nil, err
	}

	// 使用 BackendFactory 根据测试模式创建 Backend
	factory := NewBackendFactory(cfg)
	backend, err := factory.Create()
	if err != nil {
		return nil, err
	}

	registry := tools.NewRegistry()
	if cfg.Tools.Builtin.ReadFile.Enabled {
		registry.Register(builtin.ReadFile{})
	}
	if cfg.Tools.Builtin.ListFiles.Enabled {
		registry.Register(builtin.ListFiles{})
	}
	if cfg.Tools.Builtin.RunCommand.Enabled {
		registry.Register(builtin.NewRunCommand(cfg.Tools.Builtin.RunCommand))
	}

	mcpManager := mcp.NewManager(cfg.MCP.Servers)
	_ = mcpManager.Connect(context.Background())
	for _, tool := range mcpManager.Tools(context.Background()) {
		registry.Register(tool)
	}

	discovered, err := skills.Discover(cfg.Defaults.SkillDirs)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		Config:         cfg,
		StorePaths:     paths,
		Backend:        backend,
		Tools:          registry,
		SessionManager: sessionManager,
		MCP:            mcpManager,
		Skills:         skills.NewRegistry(discovered),
		Session:        session,
	}, nil
}

func (r *Runtime) BaseInstructions() string {
	return "You are sp13sx, a terminal-based coding and tooling agent."
}

func (r *Runtime) BackendName() string {
	return r.Backend.Name()
}

func (r *Runtime) ModelName() string {
	return r.Session.Model
}

func (r *Runtime) SessionTitle() string {
	return r.Session.Title
}

func (r *Runtime) EnabledSkillNames() []string {
	return append([]string(nil), r.Config.Defaults.EnabledSkills...)
}

func (r *Runtime) DiscoveredSkillNames() []string {
	all := r.Skills.All()
	out := make([]string, 0, len(all))
	for _, skill := range all {
		out = append(out, skill.Name)
	}
	return out
}

func (r *Runtime) MCPStatusLines() []string {
	states := r.MCP.States()
	out := make([]string, 0, len(states))
	for _, state := range states {
		line := state.Name + " [" + state.Status + "]"
		if state.Error != "" {
			line += " " + state.Error
		}
		out = append(out, line)
	}
	return out
}

func (r *Runtime) RightPaneWidth() int {
	return r.Config.UI.RightPaneWidth
}

func (r *Runtime) EffectiveInstructions() string {
	return skills.Compose(r.BaseInstructions(), r.Config.Defaults.EnabledSkills, r.Skills)
}

func (r *Runtime) Close() error {
	var errs []error
	if closer, ok := r.Backend.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if r.MCP != nil {
		if err := r.MCP.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (r *Runtime) Send(ctx context.Context, text string) (<-chan llm.StreamEvent, error) {
	r.mu.Lock()
	if r.isRunningTool {
		r.pendingInputs = append(r.pendingInputs, text)
		r.mu.Unlock()
		ch := make(chan llm.StreamEvent, 1)
		ch <- llm.StreamEvent{Type: "input_queued", Content: text}
		close(ch)
		return ch, nil
	}
	r.mu.Unlock()

	now := util.NowUTC()
	userMsg := domain.Message{
		ID:        util.NewID("msg"),
		SessionID: r.Session.ID,
		Role:      "user",
		Content:   []domain.ContentPart{{Type: "text", Text: text}},
		CreatedAt: now,
	}
	if err := store.SaveMessage(r.StorePaths.MessagesPath, store.MessageRecord{
		ID:        util.NewID("evt"),
		Type:      "message",
		CreatedAt: now.Format(timeLayout),
		Payload:   userMsg,
	}); err != nil {
		return nil, err
	}

	req := llm.GenerateRequest{
		Model:        r.Session.Model,
		Instructions: r.EffectiveInstructions(),
		Input:        []llm.InputItem{llm.UserTextInput(text)},
		Tools:        buildToolDefinitions(r.Tools),
	}

	stream := make(chan llm.StreamEvent, 32)
	go r.runTurnLoop(ctx, req, stream)
	return persistAssistantStream(r.StorePaths.MessagesPath, r.Session.ID, stream), nil
}

func (r *Runtime) Cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.currentCancel != nil {
		r.currentCancel()
	}
}

func (r *Runtime) Status() llm.RuntimeStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pending := make([]string, len(r.pendingInputs))
	copy(pending, r.pendingInputs)
	return llm.RuntimeStatus{
		IsRunningTool: r.isRunningTool,
		PendingInputs: pending,
	}
}

func buildToolDefinitions(registry *tools.Registry) []llm.ToolDefinition {
	all := registry.Definitions()
	out := make([]llm.ToolDefinition, 0, len(all))
	for _, tool := range all {
		out = append(out, llm.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.Schema(),
		})
	}
	return out
}

func (r *Runtime) runTurnLoop(ctx context.Context, req llm.GenerateRequest, out chan<- llm.StreamEvent) {
	defer close(out)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r.mu.Lock()
	r.currentCancel = cancel
	r.mu.Unlock()

	current := req
	executor := tools.NewExecutor(r.Tools)

	for {
		stream, err := r.Backend.Generate(ctx, current)
		if err != nil {
			out <- llm.StreamEvent{Type: "error", Error: err}
			return
		}

		var toolCalls []*llm.ToolCall
		for event := range stream {
			if event.Type == "tool_call" && event.ToolCall != nil {
				toolCalls = append(toolCalls, event.ToolCall)
			}
			out <- event
		}

		if len(toolCalls) == 0 {
			r.mu.Lock()
			r.isRunningTool = false
			pending := r.pendingInputs
			r.pendingInputs = nil
			r.currentCancel = nil
			r.mu.Unlock()

			if len(pending) > 0 {
				for _, input := range pending {
					current.Input = append(current.Input, llm.UserTextInput(input))
				}
				continue
			}
			return
		}

		r.mu.Lock()
		r.isRunningTool = true
		r.mu.Unlock()

		nextInput := make([]llm.InputItem, 0, len(toolCalls))
		for _, call := range toolCalls {
			invocationID := util.NewID("tool")
			now := util.NowUTC()
			_ = store.SaveToolCall(r.StorePaths.ToolCallsPath, store.ToolCallRecord{
				ID:        util.NewID("evt"),
				Type:      "tool.queued",
				CreatedAt: now.Format(timeLayout),
				Payload: domain.ToolInvocation{
					ID:        invocationID,
					SessionID: r.Session.ID,
					ToolName:  call.Name,
					Status:    "queued",
					Arguments: call.Arguments,
					CreatedAt: now,
				},
			})
			out <- llm.StreamEvent{Type: "status", Content: "tool queued: " + call.Name}

			result, err := executor.Run(ctx, call.Name, call.Arguments)
			status := "completed"
			outputText := ""
			if err != nil {
				status = "error"
				outputText = err.Error()
				out <- llm.StreamEvent{Type: "status", Content: "tool error: " + call.Name + ": " + err.Error()}
			} else {
				data, marshalErr := json.Marshal(result)
				if marshalErr != nil {
					outputText = fmt.Sprintf("%v", result)
				} else {
					outputText = string(data)
				}
				out <- llm.StreamEvent{Type: "status", Content: "tool completed: " + call.Name}
			}

			_ = store.SaveToolCall(r.StorePaths.ToolCallsPath, store.ToolCallRecord{
				ID:        util.NewID("evt"),
				Type:      "tool." + status,
				CreatedAt: util.NowUTC().Format(timeLayout),
				Payload: domain.ToolInvocation{
					ID:        invocationID,
					SessionID: r.Session.ID,
					ToolName:  call.Name,
					Status:    status,
					Arguments: call.Arguments,
					Output:    result,
					Error:     errString(err),
					CreatedAt: now,
				},
			})

			nextInput = append(nextInput, llm.ToolResultInput(call.CallID, call.Name, outputText))
		}

		r.mu.Lock()
		r.isRunningTool = false
		r.mu.Unlock()

		current = llm.GenerateRequest{
			Model:        r.Session.Model,
			Instructions: r.EffectiveInstructions(),
			Input:        nextInput,
			Tools:        buildToolDefinitions(r.Tools),
		}
	}
}

func persistAssistantStream(path string, sessionID string, in <-chan llm.StreamEvent) <-chan llm.StreamEvent {
	out := make(chan llm.StreamEvent, 8)
	go func() {
		defer close(out)
		var final string
		var reasoning string
		for event := range in {
			if event.Type == "message" {
				final += event.Content
			}
			if event.Type == "reasoning" {
				reasoning += event.ReasoningContent
			}
			out <- event
		}
		if final == "" && reasoning == "" {
			return
		}
		now := util.NowUTC()
		msg := domain.Message{
			ID:        util.NewID("msg"),
			SessionID: sessionID,
			Role:      "assistant",
			Content:   []domain.ContentPart{{Type: "text", Text: final}},
			CreatedAt: now,
		}
		if reasoning != "" {
			msg.Reasoning = []domain.ContentPart{{Type: "text", Text: reasoning}}
		}
		_ = store.SaveMessage(path, store.MessageRecord{
			ID:        util.NewID("evt"),
			Type:      "message",
			CreatedAt: now.Format(timeLayout),
			Payload:   msg,
		})
	}()
	return out
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
