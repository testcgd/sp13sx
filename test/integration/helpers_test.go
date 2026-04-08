//go:build integration

package integration

import (
	"os"
	"testing"

	"sp13sx/internal/config"
)

func loadTestConfig(t *testing.T) config.Config {
	cfgPath := os.Getenv("SP13SX_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config.local.yml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Test.EffectiveMode() != config.TestModeLive {
		t.Skip("集成测试需要 TestMode=live")
	}

	return cfg
}

func requireEnv(t *testing.T, key string) string {
	val := os.Getenv(key)
	if val == "" {
		t.Fatalf("需要设置环境变量 %s", key)
	}
	return val
}

func skipIfNoIntegration(t *testing.T) {
	if os.Getenv("SP13SX_INTEGRATION_TEST") == "" {
		t.Skip("设置 SP13SX_INTEGRATION_TEST=1 运行集成测试")
	}
}
