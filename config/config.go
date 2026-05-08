package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

func Load() *Config {
	// โหลด .env ถ้ามี (ถ้าไม่มีก็ไม่ error)
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "gomock"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}

	log.Printf("✅ Config loaded | DB: %s:%s/%s | Server: :%s",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.ServerPort)

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
