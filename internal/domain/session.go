package domain

import "time"

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Backend   string    `json:"backend"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
