package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

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