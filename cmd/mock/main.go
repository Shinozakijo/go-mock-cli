package main

import (
	"fmt"
	"log"
	"os"

	"github.com/shinozakijo/go-mock-cli/config"
	"github.com/shinozakijo/go-mock-cli/internal/cli"
	"github.com/shinozakijo/go-mock-cli/internal/db"
	"github.com/shinozakijo/go-mock-cli/internal/repository"
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

	app := cli.NewApp(routeRepo, responseRepo, cfg.ServerPort)

	if err := app.Run(os.Args[1:]); err != nil {
		log.Fatalf("❌ %v", err)
	}
}