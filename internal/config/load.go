package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	ApplyDefaults(&cfg)
	cfg.ExpandPaths()
	return cfg, Validate(cfg)
}

func ApplyDefaults(cfg *Config) {
	if cfg.App.Name == "" {
		cfg.App.Name = "sp13sx"
	}
	if cfg.App.DataDir == "" {
		cfg.App.DataDir = "~/.sp13sx"
	}
	if cfg.UI.Layout == "" {
		cfg.UI.Layout = "dual-pane"
	}
	if cfg.UI.RightPaneWidth == 0 {
		cfg.UI.RightPaneWidth = 40
	}
	if cfg.Defaults.Backend == "" {
		cfg.Defaults.Backend = "openai"
	}
	if cfg.Defaults.Model == "" {
		cfg.Defaults.Model = "gpt-5"
	}
	if cfg.Storage.ConversationsDir == "" {
		cfg.Storage.ConversationsDir = "~/.sp13sx/conversations"
	}
	if cfg.Storage.StateDir == "" {
		cfg.Storage.StateDir = "~/.sp13sx/state"
	}
}

func (c *Config) ExpandPaths() {
	c.App.DataDir = expandPath(c.App.DataDir)
	c.Storage.ConversationsDir = expandPath(c.Storage.ConversationsDir)
	c.Storage.StateDir = expandPath(c.Storage.StateDir)
	for i := range c.Defaults.SkillDirs {
		c.Defaults.SkillDirs[i] = expandPath(c.Defaults.SkillDirs[i])
	}
	for i := range c.MCP.Servers {
		c.MCP.Servers[i].CWD = expandPath(c.MCP.Servers[i].CWD)
	}
}

func expandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~/"))
}
