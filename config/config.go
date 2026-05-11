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
	SeedFile   string // path ไปยัง mock-config.json
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "gomock"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		SeedFile:   getEnv("SEED_FILE", ""),
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
