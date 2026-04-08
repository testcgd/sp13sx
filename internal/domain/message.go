package domain

import "time"

type ContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Message struct {
	ID        string        `json:"id"`
	SessionID string        `json:"session_id"`
	Role      string        `json:"role"`
	Content   []ContentPart `json:"content"`
	CreatedAt time.Time     `json:"created_at"`
}
