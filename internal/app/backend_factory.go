package app

import (
	"fmt"
	"path/filepath"

	"sp13sx/internal/config"
	"sp13sx/internal/llm"
	"sp13sx/internal/llm/mock"
	openaiBackend "sp13sx/internal/llm/openai"
	"sp13sx/internal/llm/recorder"
	"sp13sx/internal/store"
	"sp13sx/internal/tools"
	"sp13sx/internal/tools/builtin"
)

// BackendFactory 创建 LLM Backend 的工厂
type BackendFactory struct {
	cfg config.Config
}

// NewBackendFactory 创建 Backend 工厂
func NewBackendFactory(cfg config.Config) *BackendFactory {
	return &BackendFactory{cfg: cfg}
}

// Create 根据测试模式创建 Backend
func (f *BackendFactory) Create() (llm.Backend, error) {
	mode := f.cfg.Test.EffectiveMode()

	switch mode {
	case config.TestModeLive:
		return f.createLiveBackend()

	case config.TestModeMock:
		return f.createMockBackend()

	case config.TestModePlayback:
		return f.createPlaybackBackend()

	case config.TestModeRecord:
		return f.createRecordBackend()

	default:
		return f.createLiveBackend()
	}
}

func (f *BackendFactory) createLiveBackend() (llm.Backend, error) {
	backendCfg, ok := f.cfg.Backends[f.cfg.Defaults.Backend]
	if !ok {
		return nil, fmt.Errorf("backend %q not found in config", f.cfg.Defaults.Backend)
	}

	return openaiBackend.NewChatBackend(backendCfg)
}

// createMockBackend 创建 Mock 模式 Backend
func (f *BackendFactory) createMockBackend() (llm.Backend, error) {
	scenarioName := f.cfg.Test.EffectiveScenario()
	if scenarioName == "" {
		return nil, fmt.Errorf("scenario name is required for mock mode")
	}

	scenarioDir := f.cfg.Test.ScenarioDir
	if scenarioDir == "" {
		scenarioDir = "./test/scenarios"
	}

	scenarioPath := filepath.Join(scenarioDir, scenarioName+".yaml")
	scenario, err := mock.LoadScenario(scenarioPath)
	if err != nil {
		return nil, fmt.Errorf("load scenario %q: %w", scenarioName, err)
	}

	if err := scenario.Validate(); err != nil {
		return nil, fmt.Errorf("validate scenario %q: %w", scenarioName, err)
	}

	opts := mock.ScenarioOptions{
		RandomDelay: f.cfg.Test.Mock.RandomDelay,
	}

	return mock.NewScenarioBackend(scenario, opts), nil
}

// createPlaybackBackend 创建 Playback 模式 Backend
func (f *BackendFactory) createPlaybackBackend() (llm.Backend, error) {
	recordingPath := f.cfg.Test.EffectiveRecording()
	if recordingPath == "" {
		return nil, fmt.Errorf("recording path is required for playback mode")
	}

	opts := recorder.PlaybackOptions{
		StrictMatch:   f.cfg.Test.Playback.StrictMatch,
		PreserveDelay: false,
	}

	return recorder.NewPlaybackBackend(recordingPath, opts)
}

// createRecordBackend 创建 Record 模式 Backend
func (f *BackendFactory) createRecordBackend() (llm.Backend, error) {
	// 创建真实 Backend
	realBackend, err := f.createLiveBackend()
	if err != nil {
		return nil, fmt.Errorf("create real backend: %w", err)
	}

	// 确定输出路径
	outputPath := f.cfg.Test.EffectiveRecordingOutput()
	if outputPath == "" {
		recordingDir := f.cfg.Test.RecordingDir
		if recordingDir == "" {
			recordingDir = "./test/recordings"
		}
		timestamp := "recording"
		outputPath = filepath.Join(recordingDir, timestamp+".jsonl")
	}

	return recorder.NewRecorderBackend(realBackend, outputPath)
}

// CreateTestRuntime 创建用于测试的 Runtime（简化版）
func CreateTestRuntime(backend llm.Backend, cfg config.Config) (*Runtime, error) {
	paths := store.NewPaths(cfg.Storage.ConversationsDir)
	sessionManager := NewSessionManager(paths)
	session, err := sessionManager.EnsureSession(cfg.Defaults.Backend, cfg.Defaults.Model)
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

	return &Runtime{
		Config:         cfg,
		StorePaths:     paths,
		Backend:        backend,
		Tools:          registry,
		SessionManager: sessionManager,
		Session:        session,
	}, nil
}
