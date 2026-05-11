package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/shinozakijo/go-mock-cli/internal/repository"
)

type Seeder struct {
	routeRepo    *repository.RouteRepository
	responseRepo *repository.ResponseRepository
}

func NewSeeder(
	routeRepo *repository.RouteRepository,
	responseRepo *repository.ResponseRepository,
) *Seeder {
	return &Seeder{
		routeRepo:    routeRepo,
		responseRepo: responseRepo,
	}
}

func (s *Seeder) Run(cfg *MockConfig) error {
	ctx := context.Background()

	log.Printf("🌱 seeding %d routes from config...", len(cfg.Routes))

	for _, routeCfg := range cfg.Routes {
		method := strings.ToUpper(routeCfg.Method)

		if err := s.seedRoute(ctx, method, routeCfg); err != nil {
			return fmt.Errorf("seed route %s %s: %w", method, routeCfg.Path, err)
		}
	}

	log.Println("✅ seeding complete")
	return nil
}

func (s *Seeder) seedRoute(ctx context.Context, method string, routeCfg RouteConfig) error {
	// เช็กว่า route มีอยู่แล้วไหม
	existingRoute, err := s.routeRepo.GetByMethodAndPath(ctx, method, routeCfg.Path)
	if err != nil {
		// ยังไม่มี → สร้างใหม่
		newRoute, err := s.routeRepo.Create(ctx, method, routeCfg.Path, routeCfg.Description)
		if err != nil {
			return fmt.Errorf("create route: %w", err)
		}
		log.Printf("  ➕ created route: %s %s", method, routeCfg.Path)

		return s.seedResponses(ctx, newRoute.ID, routeCfg.Responses)
	}

	// มีอยู่แล้ว → skip (ไม่ overwrite ของเดิม)
	log.Printf("  ⏭️  skipped route: %s %s (already exists)", method, routeCfg.Path)
	_ = existingRoute
	return nil
}

func (s *Seeder) seedResponses(ctx context.Context, routeID string, responses []ResponseConfig) error {
	var activeID string

	for _, resCfg := range responses {
		headers, err := headersToJSON(resCfg.Headers)
		if err != nil {
			return err
		}

		res, err := s.responseRepo.Create(
			ctx,
			routeID,
			resCfg.Name,
			resCfg.StatusCode,
			resCfg.Body,
			headers,
			resCfg.DelayMs,
		)
		if err != nil {
			return fmt.Errorf("create response %q: %w", resCfg.Name, err)
		}

		log.Printf("    ➕ created response: %s (status=%d, active=%v)",
			res.Name, res.StatusCode, resCfg.IsActive)

		if resCfg.IsActive {
			activeID = res.ID
		}
	}

	// set active response
	if activeID != "" {
		if err := s.responseRepo.SetActive(ctx, routeID, activeID); err != nil {
			return fmt.Errorf("set active response: %w", err)
		}
		log.Printf("    ✔ active response set")
	}

	return nil
}

func headersToJSON(headers map[string]string) (json.RawMessage, error) {
	if len(headers) == 0 {
		return json.RawMessage(`{"Content-Type":"application/json"}`), nil
	}

	data, err := json.Marshal(headers)
	if err != nil {
		return nil, fmt.Errorf("marshal headers: %w", err)
	}

	return json.RawMessage(data), nil
}
