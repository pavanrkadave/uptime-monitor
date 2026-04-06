package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/pavanrkadave/uptime-monitor/internal/config"
	"github.com/pavanrkadave/uptime-monitor/internal/logger"
	"github.com/pavanrkadave/uptime-monitor/migrations"

	"github.com/pavanrkadave/uptime-monitor/internal/api/handlers"
	"github.com/pavanrkadave/uptime-monitor/internal/service"
	"github.com/pavanrkadave/uptime-monitor/internal/storage/postgres"
	"github.com/pavanrkadave/uptime-monitor/internal/worker"
)

func main() {
	// Load Config
	cfg := config.Load()

	// Initialize logger
	log := logger.Init(cfg.Environment)
	log.Info("Starting Uptime Monitor API", slog.String("environment", cfg.Environment), slog.String("port", cfg.Port))

	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://uptime_user:uptime_password@localhost:5432/uptime_db?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error("Failed to open database connection ", slog.Any("error", err))
		os.Exit(1)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	if err := db.Ping(); err != nil {
		log.Error("Database is unreachable", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("Connected to PostgreSQL successfully!")

	if err := runDBMigrations(db); err != nil {
		log.Error("Failed to run database migrations", slog.Any("error", err))
		os.Exit(1)
	}

	monitorRepo := postgres.NewMonitorRepository(db)
	monitorService := service.NewMonitorService(monitorRepo)
	monitorHandler := handlers.NewMonitorHandler(monitorService)

	// Setup Ping Scheduler
	pingScheduler := worker.NewScheduler(monitorService)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		pingScheduler.Start(ctx, 5*time.Second)
	}()

	// Set up the router
	mux := http.NewServeMux()

	mux.HandleFunc("POST /monitors", monitorHandler.HandleCreate)
	mux.HandleFunc("GET /monitors", monitorHandler.HandleList)
	mux.HandleFunc("GET /monitors/{id}", monitorHandler.HandleGetByID)
	mux.HandleFunc("PUT /monitors/{id}", monitorHandler.HandleUpdate)
	mux.HandleFunc("DELETE /monitors/{id}", monitorHandler.HandleDelete)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Info("Starting API server on port 8080...")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Failed to start server", slog.Any("error", err))
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down signal received! Initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Warn("HTTP Server forced to shutdownL: %v", err)
	}

	log.Info("Waiting for background workers to finish...")
	wg.Wait()

	log.Info("Application shut down successfully!")
}

// runDBMigrations executes our embedded SQL migration files
func runDBMigrations(db *sql.DB) error {
	// 1. Tell golang-migrate to read from our embedded file system
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	// 2. Tell golang-migrate how to talk to our specific database
	dbDriver, err := pgMigrate.WithInstance(db, &pgMigrate.Config{})
	if err != nil {
		return err
	}

	// 3. Create the migration instance
	m, err := migrate.NewWithInstance(
		"iofs", sourceDriver,
		"postgres", dbDriver,
	)
	if err != nil {
		return err
	}

	// 4. Run it!
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("Database schema is up to date. No migrations applied.")
			return nil
		}
		return err
	}
	slog.Info("Database migrations applied successfully!")
	return nil
}
