package builtin

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"sp13sx/internal/config"
)

type RunCommand struct {
	cfg config.RunCommandConfig
}

func NewRunCommand(cfg config.RunCommandConfig) RunCommand {
	return RunCommand{cfg: cfg}
}

func (RunCommand) Name() string { return "run_command" }

func (RunCommand) Description() string { return "Run an allowlisted shell command." }

func (RunCommand) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{"type": "string"},
		},
		"required": []string{"command"},
	}
}

func (r RunCommand) Run(ctx context.Context, args map[string]any) (map[string]any, error) {
	command, _ := args["command"].(string)
	if !r.allowed(command) {
		return nil, fmt.Errorf("command %q is not allowlisted", command)
	}
	timeout := time.Duration(r.cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	cmd := exec.CommandContext(runCtx, parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if r.cfg.MaxOutputBytes > 0 && len(out) > r.cfg.MaxOutputBytes {
		out = out[:r.cfg.MaxOutputBytes]
	}
	result := map[string]any{"output": string(out)}
	if err != nil {
		result["exit_error"] = err.Error()
	}
	return result, nil
}

func (r RunCommand) allowed(command string) bool {
	for _, allowed := range r.cfg.Allowlist {
		if command == allowed {
			return true
		}
	}
	return false
}
