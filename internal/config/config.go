package config

type Config struct {
	App      AppConfig          `yaml:"app"`
	UI       UIConfig           `yaml:"ui"`
	Defaults DefaultsConfig     `yaml:"defaults"`
	Backends map[string]Backend `yaml:"backends"`
	Tools    ToolsConfig        `yaml:"tools"`
	MCP      MCPConfig          `yaml:"mcp"`
	Skills   SkillsConfig       `yaml:"skills"`
	Sessions SessionsConfig     `yaml:"sessions"`
	Storage  StorageConfig      `yaml:"storage"`
}

type AppConfig struct {
	Name     string `yaml:"name"`
	DataDir  string `yaml:"data_dir"`
	LogLevel string `yaml:"log_level"`
}

type UIConfig struct {
	Theme          string `yaml:"theme"`
	Layout         string `yaml:"layout"`
	ShowToolCalls  bool   `yaml:"show_tool_calls"`
	ShowTimestamps bool   `yaml:"show_timestamps"`
	RightPaneWidth int    `yaml:"right_pane_width"`
}

type DefaultsConfig struct {
	Backend          string   `yaml:"backend"`
	Model            string   `yaml:"model"`
	SkillDirs        []string `yaml:"skill_dirs"`
	EnabledSkills    []string `yaml:"enabled_skills"`
	SessionTitleMode string   `yaml:"session_title_mode"`
}

type Backend struct {
	Type            string        `yaml:"type"`
	BaseURL         string        `yaml:"base_url"`
	APIKeyEnv       string        `yaml:"api_key_env"`
	OrganizationEnv string        `yaml:"organization_env"`
	ProjectEnv      string        `yaml:"project_env"`
	Request         RequestConfig `yaml:"request"`
}

type RequestConfig struct {
	Temperature     float64 `yaml:"temperature"`
	MaxOutputTokens int     `yaml:"max_output_tokens"`
	TimeoutSeconds  int     `yaml:"timeout_seconds"`
	ReasoningEffort string  `yaml:"reasoning_effort"`
}

type ToolsConfig struct {
	Builtin BuiltinToolsConfig `yaml:"builtin"`
}

type BuiltinToolsConfig struct {
	ReadFile   ToolToggle       `yaml:"read_file"`
	ListFiles  ToolToggle       `yaml:"list_files"`
	RunCommand RunCommandConfig `yaml:"run_command"`
}

type ToolToggle struct {
	Enabled bool `yaml:"enabled"`
}

type RunCommandConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Allowlist      []string `yaml:"allowlist"`
	TimeoutSeconds int      `yaml:"timeout_seconds"`
	MaxOutputBytes int      `yaml:"max_output_bytes"`
}

type MCPConfig struct {
	Servers []MCPServerConfig `yaml:"servers"`
}

type MCPServerConfig struct {
	Name                  string            `yaml:"name"`
	Enabled               bool              `yaml:"enabled"`
	Transport             string            `yaml:"transport"`
	Command               string            `yaml:"command"`
	Args                  []string          `yaml:"args"`
	Env                   map[string]string `yaml:"env"`
	CWD                   string            `yaml:"cwd"`
	StartupTimeoutSeconds int               `yaml:"startup_timeout_seconds"`
}

type SkillsConfig struct {
	AutoDiscover       bool `yaml:"auto_discover"`
	PreferFrontmatter  bool `yaml:"prefer_frontmatter_name"`
	AllowProjectSkills bool `yaml:"allow_project_skills"`
}

type SessionsConfig struct {
	AutoResumeLast     bool `yaml:"auto_resume_last"`
	MaxHistoryMessages int  `yaml:"max_history_messages"`
	SummarizeThreshold int  `yaml:"summarize_threshold"`
}

type StorageConfig struct {
	ConversationsDir string `yaml:"conversations_dir"`
	StateDir         string `yaml:"state_dir"`
}
