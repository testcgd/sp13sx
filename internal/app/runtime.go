package app

import (
	"context"
	"encoding/json"
	"fmt"

	"sp13sx/internal/config"
	"sp13sx/internal/domain"
	"sp13sx/internal/llm"
	openaiBackend "sp13sx/internal/llm/openai"
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
}

func NewRuntime(cfg config.Config) (*Runtime, error) {
	paths := store.NewPaths(cfg.Storage.ConversationsDir)
	sessionManager := NewSessionManager(paths)
	session, err := sessionManager.EnsureSession(cfg.Defaults.Backend, cfg.Defaults.Model)
	if err != nil {
		return nil, err
	}

	backendCfg := cfg.Backends[cfg.Defaults.Backend]
	backend, err := openaiBackend.NewBackend(backendCfg)
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

func (r *Runtime) Send(ctx context.Context, text string) (<-chan llm.StreamEvent, error) {
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

	current := req
	executor := tools.NewExecutor(r.Tools)

	for {
		stream, err := r.Backend.Generate(ctx, current)
		if err != nil {
			out <- llm.StreamEvent{Type: "error", Error: err}
			return
		}

		var responseID string
		var toolCalls []*llm.ToolCall
		for event := range stream {
			if event.Type == "response_id" && event.ResponseID != "" {
				responseID = event.ResponseID
			}
			if event.Type == "tool_call" && event.ToolCall != nil {
				toolCalls = append(toolCalls, event.ToolCall)
			}
			out <- event
		}

		if len(toolCalls) == 0 {
			return
		}
		if responseID == "" {
			out <- llm.StreamEvent{Type: "error", Error: fmt.Errorf("missing response id for tool continuation")}
			return
		}

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

			nextInput = append(nextInput, llm.FunctionOutputInput(call.CallID, outputText))
		}

		current = llm.GenerateRequest{
			Model:              r.Session.Model,
			Instructions:       r.EffectiveInstructions(),
			Input:              nextInput,
			Tools:              buildToolDefinitions(r.Tools),
			PreviousResponseID: responseID,
		}
	}
}

func persistAssistantStream(path string, sessionID string, in <-chan llm.StreamEvent) <-chan llm.StreamEvent {
	out := make(chan llm.StreamEvent, 8)
	go func() {
		defer close(out)
		var final string
		for event := range in {
			if event.Type == "message" {
				final += event.Content
			}
			out <- event
		}
		if final == "" {
			return
		}
		now := util.NowUTC()
		_ = store.SaveMessage(path, store.MessageRecord{
			ID:        util.NewID("evt"),
			Type:      "message",
			CreatedAt: now.Format(timeLayout),
			Payload: domain.Message{
				ID:        util.NewID("msg"),
				SessionID: sessionID,
				Role:      "assistant",
				Content:   []domain.ContentPart{{Type: "text", Text: final}},
				CreatedAt: now,
			},
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
