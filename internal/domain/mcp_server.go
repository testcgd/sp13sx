package domain

type MCPServer struct {
	Name      string `json:"name"`
	Transport string `json:"transport"`
	Command   string `json:"command,omitempty"`
	Status    string `json:"status"`
}
