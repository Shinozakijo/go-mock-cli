package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadEditableResponseFile exported สำหรับใช้จาก tui package ด้วย
func ReadEditableResponseFile(path string) (*EditableResponse, error) {
	return readEditableResponseFile(path)
}

func readEditableResponseFile(path string) (*EditableResponse, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var editable EditableResponse
	if err := json.Unmarshal(data, &editable); err != nil {
		return nil, fmt.Errorf("invalid editable response json: %w", err)
	}

	if editable.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if editable.StatusCode <= 0 {
		return nil, fmt.Errorf("status_code must be > 0")
	}
	if len(editable.Headers) == 0 {
		editable.Headers = json.RawMessage(`{}`)
	}
	if len(editable.Body) == 0 {
		editable.Body = json.RawMessage(`{}`)
	}
	if !json.Valid(editable.Headers) {
		return nil, fmt.Errorf("headers must be valid JSON")
	}
	if !json.Valid(editable.Body) {
		return nil, fmt.Errorf("body must be valid JSON")
	}

	return &editable, nil
}

func readJSONFile(path string) (json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("file does not contain valid JSON")
	}
	return json.RawMessage(data), nil
}
