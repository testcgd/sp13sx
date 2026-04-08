package recorder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"sp13sx/internal/llm"
)

// RecorderBackend 包装真实 Backend 并录制所有交互
type RecorderBackend struct {
	real    llm.Backend
	output  *os.File
	encoder *json.Encoder
	mu      sync.Mutex
}

// NewRecorderBackend 创建录制 Backend
func NewRecorderBackend(real llm.Backend, outputPath string) (*RecorderBackend, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("create recording directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create recording file: %w", err)
	}

	return &RecorderBackend{
		real:    real,
		output:  file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Name 返回 Backend 名称
func (r *RecorderBackend) Name() string {
	return "recorder:" + r.real.Name()
}

// Generate 实现 llm.Backend 接口
func (r *RecorderBackend) Generate(ctx context.Context, req llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	r.recordRequest(req)

	stream, err := r.real.Generate(ctx, req)
	if err != nil {
		r.recordError(err)
		return nil, err
	}

	out := make(chan llm.StreamEvent, 32)
	go func() {
		defer close(out)
		for event := range stream {
			r.recordEvent(event)
			out <- event
		}
		r.recordComplete()
	}()

	return out, nil
}

// recordRequest 记录请求
func (r *RecorderBackend) recordRequest(req llm.GenerateRequest) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record := RequestRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "request",
		Model:     req.Model,
		Input:     req.Input,
	}
	r.encoder.Encode(record)
}

// recordEvent 记录事件
func (r *RecorderBackend) recordEvent(event llm.StreamEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record := EventRecord{
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Type:       event.Type,
		Content:    event.Content,
		ResponseID: event.ResponseID,
	}

	if event.ToolCall != nil {
		record.ToolCall = &ToolCallRecord{
			ID:        event.ToolCall.ID,
			CallID:    event.ToolCall.CallID,
			Name:      event.ToolCall.Name,
			Arguments: event.ToolCall.Arguments,
		}
	}

	if event.Error != nil {
		record.Error = event.Error.Error()
	}

	r.encoder.Encode(record)
}

// recordError 记录错误
func (r *RecorderBackend) recordError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.encoder.Encode(EventRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "error",
		Error:     err.Error(),
	})
}

// recordComplete 记录完成
func (r *RecorderBackend) recordComplete() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.encoder.Encode(EventRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      "complete",
	})
}

// Close 关闭录制器
func (r *RecorderBackend) Close() error {
	return r.output.Close()
}

type RequestRecord struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Model     string          `json:"model,omitempty"`
	Input     []llm.InputItem `json:"input,omitempty"`
}

// EventRecord 事件记录
type EventRecord struct {
	Timestamp  string          `json:"timestamp"`
	Type       string          `json:"type"`
	Content    string          `json:"content,omitempty"`
	ResponseID string          `json:"response_id,omitempty"`
	ToolCall   *ToolCallRecord `json:"tool_call,omitempty"`
	Error      string          `json:"error,omitempty"`
}

// ToolCallRecord 工具调用记录
type ToolCallRecord struct {
	ID        string         `json:"id"`
	CallID    string         `json:"call_id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}
