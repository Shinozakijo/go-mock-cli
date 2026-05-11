package cli

import (
	"fmt"
	"strings"
)

var allowedMethods = map[string]bool{
	"GET":    true,
	"POST":   true,
	"PUT":    true,
	"PATCH":  true,
	"DELETE": true,
}

func validateMethod(method string) error {
	upper := strings.ToUpper(method)
	if !allowedMethods[upper] {
		return fmt.Errorf("invalid method: %s (allowed: GET, POST, PUT, PATCH, DELETE)", method)
	}
	return nil
}

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with /")
	}
	return nil
}

func validateStatusCode(code int) error {
	if code < 100 || code > 599 {
		return fmt.Errorf("invalid status code: %d (must be 100-599)", code)
	}
	return nil
}

func validateDelay(delayMs int) error {
	if delayMs < 0 {
		return fmt.Errorf("delay cannot be negative")
	}
	if delayMs > 60000 {
		return fmt.Errorf("delay too large: max 60000ms (60s)")
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if strings.Contains(name, " ") {
		return fmt.Errorf("name cannot contain spaces (use underscore)")
	}
	return nil
}
