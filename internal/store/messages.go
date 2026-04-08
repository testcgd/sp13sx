package store

import (
	"encoding/json"

	"sp13sx/internal/domain"
)

type MessageRecord struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	CreatedAt string         `json:"created_at"`
	Payload   domain.Message `json:"payload"`
}

func SaveMessage(path string, rec MessageRecord) error {
	return Append(path, rec)
}

func LoadMessages(path string, sessionID string) ([]domain.Message, error) {
	rows, err := ReadAll(path)
	if err != nil {
		return nil, err
	}
	var out []domain.Message
	for _, row := range rows {
		var rec MessageRecord
		if err := json.Unmarshal(row, &rec); err != nil {
			continue
		}
		if rec.Payload.SessionID == sessionID {
			out = append(out, rec.Payload)
		}
	}
	return out, nil
}
