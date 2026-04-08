package app

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"sp13sx/internal/config"
	"sp13sx/internal/tui"
)

func Run() error {
	cfgPath := resolveConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config %s: %w", cfgPath, err)
	}

	runtime, err := NewRuntime(cfg)
	if err != nil {
		return err
	}

	model := tui.NewModel(runtime)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func resolveConfigPath() string {
	if path := os.Getenv("SP13SX_CONFIG"); path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "examples/config.example.yml"
	}
	defaultPath := filepath.Join(home, ".sp13sx", "config.yml")
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	return "examples/config.example.yml"
}
