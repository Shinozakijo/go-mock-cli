package main

import (
	"fmt"
	"log"

	"github.com/shinozakijo/go-mock-cli/config"
	"github.com/shinozakijo/go-mock-cli/internal/db"
)

func main() {
	fmt.Println("🚀 go-mock-cli starting...")

	// Load config
	cfg := config.Load()

	// Run migrations
	if err := db.RunMigrations(cfg); err != nil {
		log.Fatalf("❌ Migration error: %v", err)
	}

	// Connect to DB
	pool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("❌ DB connection error: %v", err)
	}
	defer pool.Close()

	fmt.Println("✅ Database connected!")
	fmt.Printf("   Server ready on port: %s\n", cfg.ServerPort)
}
