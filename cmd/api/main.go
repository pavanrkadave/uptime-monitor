package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
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
	"github.com/pavanrkadave/uptime-monitor/migrations"

	"github.com/pavanrkadave/uptime-monitor/internal/api/handlers"
	"github.com/pavanrkadave/uptime-monitor/internal/service"
	"github.com/pavanrkadave/uptime-monitor/internal/storage/postgres"
	"github.com/pavanrkadave/uptime-monitor/internal/worker"
)

func main() {
	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := "postgres://uptime_user:uptime_password@localhost:5432/uptime_db?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}
	log.Println("Connected to PostgreSQL successfully!")

	runDBMigrations(db)

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

	// Setup the router
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
		log.Println("Starting API server on port 8080...")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down signal received! Initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP Server forced to shutdownL: %v", err)
	}

	log.Println("Waiting for background workers to finish...")
	wg.Wait()

	log.Println("Application shut down successfully!")
}

// runDBMigrations executes our embedded SQL migration files
func runDBMigrations(db *sql.DB) {
	// 1. Tell golang-migrate to read from our embedded file system
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("Failed to load embedded migrations: %v", err)
	}

	// 2. Tell golang-migrate how to talk to our specific database
	dbDriver, err := pgMigrate.WithInstance(db, &pgMigrate.Config{})
	if err != nil {
		log.Fatalf("Failed to create postgres migration driver: %v", err)
	}

	// 3. Create the migration instance
	m, err := migrate.NewWithInstance(
		"iofs", sourceDriver,
		"postgres", dbDriver,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrate instance: %v", err)
	}

	// 4. Run it!
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Database schema is up to date. No migrations applied.")
		} else {
			log.Fatalf("Failed to run migrations: %v", err)
		}
	} else {
		log.Println("Database migrations applied successfully!")
	}
}
