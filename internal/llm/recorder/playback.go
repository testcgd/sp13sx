package recorder

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"sp13sx/internal/llm"
)

// PlaybackBackend 回放录制的交互
type PlaybackBackend struct {
	records    []EventRecord
	requests   []RequestRecord
	currentIdx int
	options    PlaybackOptions
}

// PlaybackOptions 回放配置选项
type PlaybackOptions struct {
	// StrictMatch 是否严格匹配输入
	StrictMatch bool

	// PreserveDelay 是否保持原始延迟
	PreserveDelay bool

	// DefaultDelay 默认延迟
	DefaultDelay time.Duration
}

// NewPlaybackBackend 从录制文件创建回放 Backend
func NewPlaybackBackend(recordingPath string, opts ...PlaybackOptions) (*PlaybackBackend, error) {
	file, err := os.Open(recordingPath)
	if err != nil {
		return nil, fmt.Errorf("open recording file: %w", err)
	}
	defer file.Close()

	options := PlaybackOptions{
		DefaultDelay: 50 * time.Millisecond,
	}
	if len(opts) > 0 {
		options = opts[0]
	}

	var records []EventRecord
	var requests []RequestRecord

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		var base struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(line, &base); err != nil {
			continue
		}

		switch base.Type {
		case "request":
			var req RequestRecord
			if err := json.Unmarshal(line, &req); err != nil {
				return nil, fmt.Errorf("parse request at line %d: %w", lineNum, err)
			}
			requests = append(requests, req)
		case "complete":
			// 忽略完成标记
		default:
			var event EventRecord
			if err := json.Unmarshal(line, &event); err != nil {
				return nil, fmt.Errorf("parse event at line %d: %w", lineNum, err)
			}
			records = append(records, event)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read recording file: %w", err)
	}

	return &PlaybackBackend{
		records:  records,
		requests: requests,
		options:  options,
	}, nil
}

// Name 返回 Backend 名称
func (p *PlaybackBackend) Name() string {
	return "playback"
}

// Generate 实现 llm.Backend 接口
func (p *PlaybackBackend) Generate(ctx context.Context, req llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 16)

	go func() {
		defer close(ch)

		// 找到对应的请求索引
		startIdx := p.findStartIndex(req)
		if startIdx < 0 {
			ch <- llm.StreamEvent{
				Type:  "error",
				Error: fmt.Errorf("no matching recording found"),
			}
			return
		}

		var lastTimestamp time.Time

		for i := startIdx; i < len(p.records); i++ {
			record := p.records[i]

			// 检查是否到达下一个请求（新的一轮对话）
			if record.Type == "request" {
				break
			}

			// 处理延迟
			if p.options.PreserveDelay && !lastTimestamp.IsZero() {
				currentTs, err := time.Parse(time.RFC3339Nano, record.Timestamp)
				if err == nil {
					delay := currentTs.Sub(lastTimestamp)
					if delay > 0 {
						select {
						case <-ctx.Done():
							ch <- llm.StreamEvent{Type: "error", Error: ctx.Err()}
							return
						case <-time.After(delay):
						}
					}
				}
				lastTimestamp = currentTs
			} else if p.options.DefaultDelay > 0 {
				select {
				case <-ctx.Done():
					ch <- llm.StreamEvent{Type: "error", Error: ctx.Err()}
					return
				case <-time.After(p.options.DefaultDelay):
				}
			}

			event := p.recordToEvent(record)
			ch <- event

			// 遇到 response_id 后的工具调用结束本轮
			if record.Type == "message" && i+1 < len(p.records) && p.records[i+1].Type != "tool_call" {
				break
			}
		}

		p.currentIdx = startIdx
	}()

	return ch, nil
}

// findStartIndex 找到匹配请求的开始索引
func (p *PlaybackBackend) findStartIndex(req llm.GenerateRequest) int {
	if !p.options.StrictMatch {
		// 非严格匹配，顺序播放
		return 0
	}

	// 提取当前请求的用户输入
	userInput := extractUserInputFromRequest(req)

	// 查找匹配的请求
	for _, requestRecord := range p.requests {
		requestInput := extractUserInputFromInputItems(requestRecord.Input)
		if requestInput == userInput {
			// 找到匹配请求后，返回其对应的事件起始索引
			return p.findEventIndexAfterTimestamp(requestRecord.Timestamp)
		}
	}

	return -1
}

// findEventIndexAfterTimestamp 找到指定时间戳之后的第一个事件索引
func (p *PlaybackBackend) findEventIndexAfterTimestamp(timestamp string) int {
	reqTs, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return 0
	}

	for i, record := range p.records {
		recordTs, err := time.Parse(time.RFC3339Nano, record.Timestamp)
		if err != nil {
			continue
		}
		if recordTs.After(reqTs) || recordTs.Equal(reqTs) {
			return i
		}
	}

	return 0
}

// recordToEvent 将记录转换为流事件
func (p *PlaybackBackend) recordToEvent(record EventRecord) llm.StreamEvent {
	event := llm.StreamEvent{
		Type:       record.Type,
		Content:    record.Content,
		ResponseID: record.ResponseID,
	}

	if record.ToolCall != nil {
		event.ToolCall = &llm.ToolCall{
			ID:        record.ToolCall.ID,
			CallID:    record.ToolCall.CallID,
			Name:      record.ToolCall.Name,
			Arguments: record.ToolCall.Arguments,
		}
	}

	if record.Error != "" {
		event.Error = fmt.Errorf("%s", record.Error)
	}

	return event
}

// extractUserInputFromRequest 从请求中提取用户输入
func extractUserInputFromRequest(req llm.GenerateRequest) string {
	return extractUserInputFromInputItems(req.Input)
}

func extractUserInputFromInputItems(items []llm.InputItem) string {
	for _, item := range items {
		if item.Type == "user_message" {
			return item.Content
		}
	}
	return ""
}

// Reset 重置回放状态
func (p *PlaybackBackend) Reset() {
	p.currentIdx = 0
}

// TotalEvents 返回总事件数
func (p *PlaybackBackend) TotalEvents() int {
	return len(p.records)
}

// TotalRequests 返回总请求数
func (p *PlaybackBackend) TotalRequests() int {
	return len(p.requests)
}
