package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/pavanrkadave/uptime-monitor/internal/api/handlers"
	"github.com/pavanrkadave/uptime-monitor/internal/api/server"
	"github.com/pavanrkadave/uptime-monitor/internal/config"
	"github.com/pavanrkadave/uptime-monitor/internal/logger"
	"github.com/pavanrkadave/uptime-monitor/internal/service"
	"github.com/pavanrkadave/uptime-monitor/internal/storage/postgres"
	"github.com/pavanrkadave/uptime-monitor/internal/worker"
)

const (
	ApplicationSuccess = iota
	ApplicationError
)

// @title 			Uptime Monitor API
// @version 		1.0
// @description 	A simple uptime monitoring service API based on Clean Architecture
//
// @host 			localhost:8080
// @BasePath 		/
//
// @securityDefinitions.apiKey BearerAuth
// @in 				header
// @name 			Authorization
// @description		Type "Bearer" followed by your JWT Token.
func main() {
	// Load Application Config & Initialize Logger
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Configuration ERROR: %v", err))
	}

	log := logger.Init(cfg.Environment)

	// Run Application
	if err := runApp(cfg, log); err != nil {
		log.Error("Application error", slog.Any("error", err))
		os.Exit(ApplicationError)
	}
}

func runApp(cfg *config.Config, log *slog.Logger) error {
	log.Info("Starting Uptime Monitor API", slog.String("environment", cfg.Environment), slog.String("port", cfg.Port))

	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- Database Setup ---
	db, err := postgres.NewDB(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Error closing DB gracefully", slog.Any("error", err))
		}
	}()
	log.Info("Connected to PostgreSQL successfully!")

	if err := postgres.RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run database migrations %v", err)
	}

	// --- Dependency Injection ---
	monitorRepo := postgres.NewMonitorRepository(db)
	monitorService := service.NewMonitorService(monitorRepo, log)
	monitorHandler := handlers.NewMonitorHandler(monitorService, log)

	authUseCase := service.NewAuthService(cfg.AdminPassword, cfg.JWTSecret, log)
	authHandler := handlers.NewAuthHandler(authUseCase, log)

	// --- Setup Workers ---
	pingScheduler := worker.New(monitorService, log)
	apiServer := server.New(cfg, log, monitorHandler, authHandler)

	// -- Create WaitGroup for background workers --
	var wg sync.WaitGroup

	// --- Start PingScheduler ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		pingScheduler.Start(ctx, cfg.CheckInterval)
	}()

	// --- Start HTTP Server ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		apiServer.Start(ctx)
	}()

	<-ctx.Done()
	log.Info("Shutdown signal received, initiating graceful shutdown...")

	// --- Teardown ---
	log.Info("Waiting for background workers to finish...")
	wg.Wait()
	log.Info("Application shut down successfully!")
	return nil
}
