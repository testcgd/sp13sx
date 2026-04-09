//go:build smoke

package smoke

import (
	"context"
	"os"
	"testing"
	"time"

	"sp13sx/internal/config"
	"sp13sx/internal/llm"
)

// SmokeTest 定义一个冒烟测试
type SmokeTest struct {
	Name        string
	Description string
	Timeout     time.Duration
	Run         func(ctx context.Context, backend llm.Backend) error
}

// SmokeTestSuite 冒烟测试套件
type SmokeTestSuite struct {
	config  config.Config
	backend llm.Backend
	tests   []SmokeTest
}

// NewSmokeTestSuite 创建冒烟测试套件
func NewSmokeTestSuite(cfg config.Config, backend llm.Backend) *SmokeTestSuite {
	return &SmokeTestSuite{
		config:  cfg,
		backend: backend,
		tests:   DefaultSmokeTests(),
	}
}

// AddTest 添加测试
func (s *SmokeTestSuite) AddTest(test SmokeTest) {
	s.tests = append(s.tests, test)
}

// Run 运行所有冒烟测试
func (s *SmokeTestSuite) Run(t *testing.T) {
	if os.Getenv("SP13SX_SMOKE_TEST") == "" {
		t.Skip("设置 SP13SX_SMOKE_TEST=1 运行真实环境冒烟测试")
	}

	for _, test := range s.tests {
		t.Run(test.Name, func(t *testing.T) {
			timeout := test.Timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if err := test.Run(ctx, s.backend); err != nil {
				t.Errorf("%s: %v", test.Description, err)
			}
		})
	}
}

// DefaultSmokeTests 返回默认的冒烟测试列表
func DefaultSmokeTests() []SmokeTest {
	return []SmokeTest{
		{
			Name:        "basic_connection",
			Description: "测试基本 LLM 连接",
			Timeout:     30 * time.Second,
			Run: func(ctx context.Context, backend llm.Backend) error {
				req := llm.GenerateRequest{
					Model: "test-model",
					Input: []llm.InputItem{
						{Type: "message", Role: "user", Content: "hello"},
					},
				}

				stream, err := backend.Generate(ctx, req)
				if err != nil {
					return err
				}

				var received bool
				for event := range stream {
					if event.Type == "message" && event.Content != "" {
						received = true
					}
					if event.Error != nil {
						return event.Error
					}
				}

				if !received {
					return &SmokeTestError{Message: "未收到有效响应"}
				}
				return nil
			},
		},
		{
			Name:        "stream_integrity",
			Description: "测试流式响应完整性",
			Timeout:     30 * time.Second,
			Run: func(ctx context.Context, backend llm.Backend) error {
				req := llm.GenerateRequest{
					Model: "test-model",
					Input: []llm.InputItem{
						{Type: "message", Role: "user", Content: "请回复一段完整的话"},
					},
				}

				stream, err := backend.Generate(ctx, req)
				if err != nil {
					return err
				}

				var messageParts []string
				var hasResponseID bool

				for event := range stream {
					if event.Type == "message" {
						messageParts = append(messageParts, event.Content)
					}
					if event.Type == "response_id" && event.ResponseID != "" {
						hasResponseID = true
					}
					if event.Error != nil {
						return event.Error
					}
				}

				if len(messageParts) == 0 {
					return &SmokeTestError{Message: "未收到任何消息片段"}
				}

				if !hasResponseID {
					return &SmokeTestError{Message: "未收到 response_id"}
				}

				return nil
			},
		},
		{
			Name:        "timeout_handling",
			Description: "测试超时处理",
			Timeout:     5 * time.Second,
			Run: func(ctx context.Context, backend llm.Backend) error {
				// 使用很短的超时来测试取消行为
				ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()

				req := llm.GenerateRequest{
					Model: "test-model",
					Input: []llm.InputItem{
						{Type: "message", Role: "user", Content: "hello"},
					},
				}

				stream, err := backend.Generate(ctx, req)
				if err != nil {
					// 连接错误是可以接受的
					return nil
				}

				for range stream {
					// 等待上下文取消
				}

				return nil
			},
		},
	}
}

// SmokeTestError 冒烟测试错误
type SmokeTestError struct {
	Message string
}

func (e *SmokeTestError) Error() string {
	return e.Message
}
