package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment   string        `env:"ENV" envDefault:"development"`
	Port          string        `env:"PORT" envDefault:"8080"`
	DatabaseURL   string        `env:"DATABASE_URL,required"`
	JWTSecret     string        `env:"JWT_SECRET,required"`
	AdminPassword string        `env:"ADMIN_PASSWORD,required"`
	CheckInterval time.Duration `env:"CHECK_INTERVAL" envDefault:"10s"`
}

// Load reads configuration from .env file and environment variables
// Throws error if required variables are missing or invalid
func Load() (*Config, error) {
	// Attempt to load .env file
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	return &cfg, nil
}
