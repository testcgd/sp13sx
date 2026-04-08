package store

import (
	"encoding/json"

	"sp13sx/internal/domain"
)

type ToolCallRecord struct {
	ID        string                `json:"id"`
	Type      string                `json:"type"`
	CreatedAt string                `json:"created_at"`
	Payload   domain.ToolInvocation `json:"payload"`
}

func SaveToolCall(path string, rec ToolCallRecord) error {
	return Append(path, rec)
}

func LoadToolCalls(path string, sessionID string) ([]domain.ToolInvocation, error) {
	rows, err := ReadAll(path)
	if err != nil {
		return nil, err
	}
	var out []domain.ToolInvocation
	for _, row := range rows {
		var rec ToolCallRecord
		if err := json.Unmarshal(row, &rec); err != nil {
			continue
		}
		if rec.Payload.SessionID == sessionID {
			out = append(out, rec.Payload)
		}
	}
	return out, nil
}
