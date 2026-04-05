package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/pavanrkadave/uptime-monitor/internal/api/handlers"
	"github.com/pavanrkadave/uptime-monitor/internal/service"
	"github.com/pavanrkadave/uptime-monitor/internal/storage/postgres"
)

func main() {
	dsn := "postgres://uptime_user:uptime_password@localhost:5432/uptime_db?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}
	log.Println("Connected to PostgreSQL successfully!")

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS monitors (
		id BIGSERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createTableQuery); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	log.Println("Database schema is ready.")

	monitorRepo := postgres.NewMonitorRepository(db)
	monitorService := service.NewMonitorService(monitorRepo)
	monitorHandler := handlers.NewMonitorHandler(monitorService)

	// Setup the router
	mux := http.NewServeMux()

	mux.HandleFunc("POST /monitors", monitorHandler.HandleCreate)
	mux.HandleFunc("GET /monitors", monitorHandler.HandleList)
	mux.HandleFunc("GET /monitors/{id}", monitorHandler.HandleGetByID)
	mux.HandleFunc("PUT /monitors/{id}", monitorHandler.HandleUpdate)
	mux.HandleFunc("DELETE /monitors/{id}", monitorHandler.HandleDelete)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
