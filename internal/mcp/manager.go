package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"sp13sx/internal/config"
	toolmcp "sp13sx/internal/tools/mcp"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type ServerState struct {
	Name    string
	Enabled bool
	Status  string
	Error   string
}

type Manager struct {
	servers []config.MCPServerConfig
	clients map[string]*serverClient
	states  map[string]ServerState
}

type serverClient struct {
	name    string
	session *sdkmcp.ClientSession
	tools   []remoteTool
}

type remoteTool struct {
	Name        string
	RawName     string
	Description string
	Schema      any
}

func NewManager(servers []config.MCPServerConfig) *Manager {
	mgr := &Manager{
		servers: servers,
		clients: map[string]*serverClient{},
		states:  map[string]ServerState{},
	}
	for _, server := range servers {
		status := "disabled"
		if server.Enabled {
			status = "configured"
		}
		mgr.states[server.Name] = ServerState{
			Name:    server.Name,
			Enabled: server.Enabled,
			Status:  status,
		}
	}
	return mgr
}

func (m *Manager) Connect(ctx context.Context) error {
	for _, server := range m.servers {
		if !server.Enabled {
			continue
		}
		if err := m.connectServer(ctx, server); err != nil {
			state := m.states[server.Name]
			state.Status = "error"
			state.Error = err.Error()
			m.states[server.Name] = state
		}
	}
	return nil
}

func (m *Manager) States() []ServerState {
	out := make([]ServerState, 0, len(m.states))
	for _, state := range m.states {
		out = append(out, state)
	}
	return out
}

func (m *Manager) Tools(_ context.Context) []toolmcp.RemoteTool {
	var out []toolmcp.RemoteTool
	for _, client := range m.clients {
		for _, remote := range client.tools {
			client := client
			remote := remote
			out = append(out, toolmcp.RemoteTool{
				NameValue:        remote.Name,
				DescriptionValue: remote.Description,
				SchemaValue:      toSchemaMap(remote.Schema),
				Invoke: func(ctx context.Context, args map[string]any) (map[string]any, error) {
					return client.CallTool(ctx, remote.RawName, args)
				},
			})
		}
	}
	return out
}

func (m *Manager) Close() error {
	var errs []string
	for _, client := range m.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(errs, "; "))
}

func (m *Manager) connectServer(ctx context.Context, server config.MCPServerConfig) error {
	cmd := exec.CommandContext(ctx, server.Command, server.Args...)
	cmd.Dir = server.CWD
	if len(server.Env) > 0 {
		cmd.Env = append(cmd.Environ(), flattenEnv(server.Env)...)
	}
	client := sdkmcp.NewClient(&sdkmcp.Implementation{
		Name:    "sp13sx",
		Version: "0.1.0",
	}, nil)
	transport := &sdkmcp.CommandTransport{
		Command:           cmd,
		TerminateDuration: time.Duration(server.StartupTimeoutSeconds) * time.Second,
	}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return err
	}

	list, err := session.ListTools(ctx, nil)
	if err != nil {
		_ = session.Close()
		return err
	}

	remoteTools := make([]remoteTool, 0, len(list.Tools))
	for _, tool := range list.Tools {
		remoteTools = append(remoteTools, remoteTool{
			Name:        server.Name + "." + tool.Name,
			RawName:     tool.Name,
			Description: tool.Description,
			Schema:      tool.InputSchema,
		})
	}
	m.clients[server.Name] = &serverClient{
		name:    server.Name,
		session: session,
		tools:   remoteTools,
	}
	m.states[server.Name] = ServerState{
		Name:    server.Name,
		Enabled: server.Enabled,
		Status:  "connected",
	}
	return nil
}

func (c *serverClient) CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	result, err := c.session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return nil, err
	}
	return callToolResultToMap(result)
}

func (c *serverClient) Close() error {
	if c.session == nil {
		return nil
	}
	return c.session.Close()
}

func flattenEnv(input map[string]string) []string {
	out := make([]string, 0, len(input))
	for key, value := range input {
		out = append(out, key+"="+value)
	}
	return out
}

func toSchemaMap(schema any) map[string]any {
	if schema == nil {
		return map[string]any{"type": "object"}
	}
	switch v := schema.(type) {
	case map[string]any:
		return v
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return map[string]any{"type": "object"}
		}
		var out map[string]any
		if err := json.Unmarshal(data, &out); err != nil {
			return map[string]any{"type": "object"}
		}
		return out
	}
}

func callToolResultToMap(result *sdkmcp.CallToolResult) (map[string]any, error) {
	out := map[string]any{
		"is_error": result.IsError,
	}
	if result.StructuredContent != nil {
		out["structured"] = result.StructuredContent
	}
	var texts []string
	for _, content := range result.Content {
		switch part := content.(type) {
		case *sdkmcp.TextContent:
			texts = append(texts, part.Text)
		default:
			data, err := json.Marshal(part)
			if err == nil {
				texts = append(texts, string(data))
			}
		}
	}
	if len(texts) > 0 {
		out["content"] = strings.Join(texts, "\n")
	}
	return out, nil
}
