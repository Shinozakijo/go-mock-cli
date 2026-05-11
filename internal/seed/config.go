package seed

import (
	"encoding/json"
	"fmt"
	"os"
)

type ResponseConfig struct {
	Name       string            `json:"name"`
	StatusCode int               `json:"status_code"`
	IsActive   bool              `json:"is_active"`
	DelayMs    int               `json:"delay_ms"`
	Headers    map[string]string `json:"headers"`
	Body       json.RawMessage   `json:"body"`
}

type RouteConfig struct {
	Method      string           `json:"method"`
	Path        string           `json:"path"`
	Description string           `json:"description"`
	Responses   []ResponseConfig `json:"responses"`
}

type MockConfig struct {
	Routes []RouteConfig `json:"routes"`
}

func LoadConfig(path string) (*MockConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg MockConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func validateConfig(cfg *MockConfig) error {
	if len(cfg.Routes) == 0 {
		return fmt.Errorf("no routes defined")
	}

	allowedMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true,
		"PATCH": true, "DELETE": true,
	}

	for i, route := range cfg.Routes {
		if !allowedMethods[route.Method] {
			return fmt.Errorf("route[%d]: invalid method %q", i, route.Method)
		}
		if route.Path == "" {
			return fmt.Errorf("route[%d]: path is required", i)
		}
		if route.Path[0] != '/' {
			return fmt.Errorf("route[%d]: path must start with /", i)
		}
		if len(route.Responses) == 0 {
			return fmt.Errorf("route[%d] %s %s: must have at least 1 response",
				i, route.Method, route.Path)
		}

		// ต้องมี is_active = true อย่างน้อย 1 ตัว
		hasActive := false
		for _, res := range route.Responses {
			if res.IsActive {
				hasActive = true
				break
			}
		}
		if !hasActive {
			return fmt.Errorf("route %s %s: at least one response must have is_active=true",
				route.Method, route.Path)
		}
	}

	return nil
}
