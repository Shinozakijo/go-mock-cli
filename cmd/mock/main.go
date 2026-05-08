package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/shinozakijo/go-mock-cli/config"
	"github.com/shinozakijo/go-mock-cli/internal/db"
	"github.com/shinozakijo/go-mock-cli/internal/repository"
	"github.com/shinozakijo/go-mock-cli/internal/server"
	"github.com/shinozakijo/go-mock-cli/internal/service"
)

func main() {
	fmt.Println("🚀 go-mock-cli starting...")

	cfg := config.Load()

	if err := db.RunMigrations(cfg); err != nil {
		log.Fatalf("❌ Migration error: %v", err)
	}

	pool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("❌ DB connection error: %v", err)
	}
	defer pool.Close()

	fmt.Println("✅ Database connected!")

	ctx := context.Background()

	routeRepo := repository.NewRouteRepository(pool)
	responseRepo := repository.NewResponseRepository(pool)

	if err := seedData(ctx, routeRepo, responseRepo); err != nil {
		log.Fatalf("❌ seed error: %v", err)
	}

	mockService := service.NewMockService(routeRepo, responseRepo)
	handler := server.NewHandler(mockService)
	srv := server.New(cfg.ServerPort, handler)

	fmt.Printf("🌐 Mock server running on http://localhost:%s\n", cfg.ServerPort)
	log.Fatal(srv.Start())
}

func seedData(
	ctx context.Context,
	routeRepo *repository.RouteRepository,
	responseRepo *repository.ResponseRepository,
) error {
	route, err := routeRepo.GetByMethodAndPath(ctx, "GET", "/api/payment")
	if err != nil {
		route, err = routeRepo.Create(ctx, "GET", "/api/payment", "Payment API mock")
		if err != nil {
			return err
		}

		successBody := json.RawMessage(`{"status":"success","amount":20000}`)
		failBody := json.RawMessage(`{"status":"fail","message":"insufficient funds"}`)
		headers := json.RawMessage(`{"Content-Type":"application/json"}`)

		resSuccess, err := responseRepo.Create(ctx, route.ID, "success_20000", 200, successBody, headers, 0)
		if err != nil {
			return err
		}

		_, err = responseRepo.Create(ctx, route.ID, "fail_response", 400, failBody, headers, 0)
		if err != nil {
			return err
		}

		if err := responseRepo.SetActive(ctx, route.ID, resSuccess.ID); err != nil {
			return err
		}

		fmt.Println("✅ Seed data created")
		return nil
	}

	fmt.Println("📦 Seed skipped: route already exists")

	_ = route
	return nil
}