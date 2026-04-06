package config

import (
	"os"
	"time"
)

type Config struct {
	Environment   string
	Port          string
	DatabaseURL   string
	JWTSecret     string
	CheckInterval time.Duration
}

func Load() *Config {
	return &Config{
		Environment:   getEnv("ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://uptime_user:uptime_password@localhost:5432/uptime_db?sslmode=disable"),
		CheckInterval: 10 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
