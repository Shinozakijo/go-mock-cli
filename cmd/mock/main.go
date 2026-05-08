package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/shinozakijo/go-mock-cli/config"
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

	fmt.Println("✅ Database connected!")

	ctx := context.Background()
	routeRepo := repository.NewRouteRepository(pool)
	responseRepo := repository.NewResponseRepository(pool)

	// --- ทดสอบ Create Route ---
	fmt.Println("\n📝 Creating test route...")
	route, err := routeRepo.Create(ctx, "GET", "/api/payment", "Payment API mock")
	if err != nil {
		log.Fatalf("❌ Create route error: %v", err)
	}
	fmt.Printf("✅ Route created: [%s] %s (id: %s)\n", route.Method, route.Path, route.ID)

	// --- ทดสอบ Create Response ---
	fmt.Println("\n📝 Creating responses...")

	successBody := json.RawMessage(`{"status": "success", "amount": 20000}`)
	failBody := json.RawMessage(`{"status": "fail", "message": "insufficient funds"}`)
	headers := json.RawMessage(`{"Content-Type": "application/json"}`)

	resSuccess, err := responseRepo.Create(ctx, route.ID, "success_20000", 200, successBody, headers, 0)
	if err != nil {
		log.Fatalf("❌ Create response error: %v", err)
	}
	fmt.Printf("✅ Response created: %s (status: %d)\n", resSuccess.Name, resSuccess.StatusCode)

	resFail, err := responseRepo.Create(ctx, route.ID, "fail_response", 400, failBody, headers, 0)
	if err != nil {
		log.Fatalf("❌ Create response error: %v", err)
	}
	fmt.Printf("✅ Response created: %s (status: %d)\n", resFail.Name, resFail.StatusCode)

	// --- ทดสอบ SetActive ---
	fmt.Println("\n🔄 Setting active response to 'success_20000'...")
	if err := responseRepo.SetActive(ctx, route.ID, resSuccess.ID); err != nil {
		log.Fatalf("❌ SetActive error: %v", err)
	}
	fmt.Println("✅ Active response set!")

	// --- ทดสอบ GetAll ---
	fmt.Println("\n📋 All routes:")
	routes, err := routeRepo.GetAll(ctx)
	if err != nil {
		log.Fatalf("❌ GetAll error: %v", err)
	}
	for _, rt := range routes {
		fmt.Printf("   [%s] %s — %s\n", rt.Method, rt.Path, rt.Description)
	}

	// --- ทดสอบ GetActiveByRouteID ---
	fmt.Println("\n🎯 Active response for /api/payment:")
	active, err := responseRepo.GetActiveByRouteID(ctx, route.ID)
	if err != nil {
		log.Fatalf("❌ GetActive error: %v", err)
	}
	fmt.Printf("   Name: %s | Status: %d | Body: %s\n", active.Name, active.StatusCode, active.Body)

	// --- Cleanup: ลบ test data ออก ---
	fmt.Println("\n🧹 Cleaning up test data...")
	if err := routeRepo.Delete(ctx, route.ID); err != nil {
		log.Fatalf("❌ Delete error: %v", err)
	}
	fmt.Println("✅ Test data cleaned up!")
}
