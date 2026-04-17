package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	_ "github.com/lib/pq"
	"github.com/pavanrkadave/uptime-monitor/internal/api/handlers"
	"github.com/pavanrkadave/uptime-monitor/internal/api/server"
	"github.com/pavanrkadave/uptime-monitor/internal/config"
	"github.com/pavanrkadave/uptime-monitor/internal/domain"
	"github.com/pavanrkadave/uptime-monitor/internal/logger"
	"github.com/pavanrkadave/uptime-monitor/internal/service"
	"github.com/pavanrkadave/uptime-monitor/internal/storage/postgres"
	"github.com/pavanrkadave/uptime-monitor/internal/worker"
)

const (
	ApplicationSuccess = iota
	ApplicationError
)

var (
	dbSizeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "uptime_monitor_db_size_bytes",
		Help: "The current size of the PostgreSQL database in bytes",
	})
)

// @title 			Uptime Monitor API
// @version 		1.0
// @description 	A simple uptime monitoring service API based on Clean Architecture
//
// @host 			api.localhost
// @schems          https http
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

	userRepo := postgres.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, log)
	authHandler := handlers.NewAuthHandler(authService, log)

	healthHandler := handlers.NewHealthHandler(&dbChecker{db: db})

	defaultAdminEmail := "admin@example.com"
	_, err = userRepo.GetByEmail(ctx, defaultAdminEmail)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Info("Default admin user not found, generating seed user..")
			_, err := authService.Register(ctx, defaultAdminEmail, cfg.AdminPassword, domain.RoleAdmin)
			if err != nil {
				return fmt.Errorf("failed to generate default admin user %v", err)
			}
			log.Info("Successfully generated default admin user", slog.String("email", defaultAdminEmail))
		} else {
			return fmt.Errorf("failed to check for default admin user %v", err)
		}
	}

	// --- Setup Workers ---
	pingScheduler := worker.New(monitorService, log)
	apiServer := server.New(cfg, log, monitorHandler, authHandler, healthHandler)

	// -- Create WaitGroup for background workers --
	var wg sync.WaitGroup

	// --- Start DB Size Metrics Collector ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		collectDBMetrics(ctx, db, log)
	}()

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

type dbChecker struct {
	db *sql.DB
}

func (d *dbChecker) Check(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func collectDBMetrics(ctx context.Context, db *sql.DB, log *slog.Logger) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// Wait for the first tick
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var size int64
			err := db.QueryRowContext(ctx, "SELECT pg_database_size(current_database())").Scan(&size)
			if err != nil {
				log.Warn("Failed to retrieve database size metric", slog.Any("error", err))
				continue
			}
			dbSizeGauge.Set(float64(size))
		}
	}
}
