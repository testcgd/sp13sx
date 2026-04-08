package store

import (
	"encoding/json"

	"sp13sx/internal/domain"
)

type SessionRecord struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	CreatedAt string         `json:"created_at"`
	Payload   domain.Session `json:"payload"`
}

func SaveSession(path string, rec SessionRecord) error {
	return Append(path, rec)
}

func LoadSessions(path string) ([]domain.Session, error) {
	rows, err := ReadAll(path)
	if err != nil {
		return nil, err
	}
	latest := map[string]domain.Session{}
	for _, row := range rows {
		var rec SessionRecord
		if err := json.Unmarshal(row, &rec); err != nil {
			continue
		}
		latest[rec.Payload.ID] = rec.Payload
	}
	out := make([]domain.Session, 0, len(latest))
	for _, session := range latest {
		out = append(out, session)
	}
	return out, nil
}
