package config

import "os"

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
	Test     TestConfig         `yaml:"test"`
}

// TestMode 测试模式类型
type TestMode string

const (
	TestModeLive     TestMode = "live"     // 正常模式，使用真实 LLM
	TestModeMock     TestMode = "mock"     // Mock 模式，使用场景脚本
	TestModePlayback TestMode = "playback" // Playback 模式，回放录制
	TestModeRecord   TestMode = "record"   // Record 模式，录制交互
)

// TestConfig 测试模式配置
type TestConfig struct {
	Mode            TestMode           `yaml:"mode"`
	ScenarioDir     string             `yaml:"scenario_dir"`
	RecordingDir    string             `yaml:"recording_dir"`
	DefaultScenario string             `yaml:"default_scenario"`
	Mock            MockTestConfig     `yaml:"mock"`
	Playback        PlaybackTestConfig `yaml:"playback"`
	Record          RecordTestConfig   `yaml:"record"`
}

// MockTestConfig Mock 模式配置
type MockTestConfig struct {
	Scenario    string `yaml:"scenario"`
	RandomDelay bool   `yaml:"random_delay"`
}

// PlaybackTestConfig Playback 模式配置
type PlaybackTestConfig struct {
	Recording   string `yaml:"recording"`
	StrictMatch bool   `yaml:"strict_match"`
}

// RecordTestConfig Record 模式配置
type RecordTestConfig struct {
	Backend  string `yaml:"backend"`
	AutoSave bool   `yaml:"auto_save"`
	Output   string `yaml:"output"`
}

// EffectiveMode 返回实际使用的测试模式
// 优先使用环境变量 SP13SX_TEST_MODE，否则使用配置文件中的值
func (c *TestConfig) EffectiveMode() TestMode {
	if mode := os.Getenv("SP13SX_TEST_MODE"); mode != "" {
		return TestMode(mode)
	}
	if c.Mode == "" {
		return TestModeLive
	}
	return c.Mode
}

// EffectiveScenario 返回实际使用的场景名称
// 优先使用环境变量 SP13SX_SCENARIO，否则使用配置文件中的值
func (c *TestConfig) EffectiveScenario() string {
	if scenario := os.Getenv("SP13SX_SCENARIO"); scenario != "" {
		return scenario
	}
	if c.Mock.Scenario != "" {
		return c.Mock.Scenario
	}
	return c.DefaultScenario
}

// EffectiveRecording 返回实际使用的录制文件路径
// 优先使用环境变量 SP13SX_RECORDING，否则使用配置文件中的值
func (c *TestConfig) EffectiveRecording() string {
	if recording := os.Getenv("SP13SX_RECORDING"); recording != "" {
		return recording
	}
	return c.Playback.Recording
}

// EffectiveRecordingOutput 返回录制输出路径
// 优先使用环境变量 SP13SX_RECORDING_OUTPUT，否则使用配置文件中的值
func (c *TestConfig) EffectiveRecordingOutput() string {
	if output := os.Getenv("SP13SX_RECORDING_OUTPUT"); output != "" {
		return output
	}
	return c.Record.Output
}

// IsTestMode 检查是否处于测试模式（非 Live 模式）
func (c *TestConfig) IsTestMode() bool {
	return c.EffectiveMode() != TestModeLive
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
