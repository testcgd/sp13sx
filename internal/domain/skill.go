package domain

type Skill struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Path        string            `json:"path"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Body        string            `json:"body,omitempty"`
}
