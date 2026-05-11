package main

import (
	"fmt"
	"log"
	"os"

	"github.com/shinozakijo/go-mock-cli/config"
	"github.com/shinozakijo/go-mock-cli/internal/cli"
	"github.com/shinozakijo/go-mock-cli/internal/db"
	"github.com/shinozakijo/go-mock-cli/internal/repository"
	"github.com/shinozakijo/go-mock-cli/internal/seed"
	"github.com/shinozakijo/go-mock-cli/internal/tui"
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

	routeRepo := repository.NewRouteRepository(pool)
	responseRepo := repository.NewResponseRepository(pool)

	// ─── Auto Seed ────────────────────────────────
	if cfg.SeedFile != "" {
		if err := runSeed(cfg.SeedFile, routeRepo, responseRepo); err != nil {
			log.Fatalf("❌ Seed error: %v", err)
		}
	}

	args := os.Args[1:]

	// ไม่มี args หรือ "ui" → เปิด TUI
	if len(args) == 0 || args[0] == "ui" {
		if err := tui.Run(routeRepo, responseRepo, cfg.ServerPort); err != nil {
			log.Fatalf("❌ TUI error: %v", err)
		}
		return
	}

	// CLI command
	app := cli.NewApp(routeRepo, responseRepo, cfg.ServerPort)
	if err := app.Run(args); err != nil {
		log.Fatalf("❌ %v", err)
	}
}

func runSeed(filePath string, routeRepo *repository.RouteRepository, responseRepo *repository.ResponseRepository) error {
	mockCfg, err := seed.LoadConfig(filePath)
	if err != nil {
		return err
	}

	seeder := seed.NewSeeder(routeRepo, responseRepo)
	return seeder.Run(mockCfg)
}
