package cli

import (
	"encoding/json"
)

// ReadJSONFile exported สำหรับใช้จาก tui package
func ReadJSONFile(path string) (json.RawMessage, error) {
	return readJSONFile(path)
}
