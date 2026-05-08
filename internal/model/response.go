package model

import (
	"encoding/json"
	"time"
)

type Response struct {
	ID         string          `json:"id"`
	RouteID    string          `json:"route_id"`
	Name       string          `json:"name"`
	StatusCode int             `json:"status_code"`
	Body       json.RawMessage `json:"body"`
	Headers    json.RawMessage `json:"headers"`
	DelayMs    int             `json:"delay_ms"`
	IsActive   bool            `json:"is_active"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
