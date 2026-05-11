package cli

import "encoding/json"

type EditableResponse struct {
	Name       string          `json:"name"`
	StatusCode int             `json:"status_code"`
	DelayMs    int             `json:"delay_ms"`
	Headers    json.RawMessage `json:"headers"`
	Body       json.RawMessage `json:"body"`
}
