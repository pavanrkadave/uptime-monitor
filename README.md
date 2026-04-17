# Uptime Monitor API

A robust, production-ready REST API for monitoring website uptime, latency, and generating statistics. Built with Go, following Clean Architecture principles.

## Architecture

This project strictly follows **Clean Architecture**:
- **Domain:** Core business models (`User`, `Monitor`, `MonitorStats`).
- **Storage (Repository):** PostgreSQL implementation using `sql.DB` and `golang-migrate` for database migrations.
- **Service:** Business logic layer encapsulating core operations.
- **Transport (Handlers):** HTTP layer powered by `go-chi/chi/v5` and Swagger docs.

## Tech Stack

- **Language:** Go 1.21+
- **Database:** PostgreSQL
- **Router:** `go-chi/chi/v5`
- **Logging:** `log/slog` (with `lmittmann/tint` for local development)
- **Migrations:** `golang-migrate/migrate`
- **Documentation:** `swaggo/swag`
- **Auth:** JWT (JSON Web Tokens) & `bcrypt` for password hashing

## Features

- **RBAC Authentication:** Secure endpoints using JWTs with Admin/Viewer roles.
- **Monitor CRUD:** Add, update, and manage websites to monitor.
- **Background Worker:** Non-blocking pinger that tests monitor endpoints and logs results.
- **Statistics Aggregation:** Real-time calculation of Uptime %, Average Latency, and P95 Latency using advanced PostgreSQL aggregates.
- **Swagger Documentation:** Auto-generated interactive API docs (`/swagger/index.html`).
- **Health Checks:** `/healthz` and `/readyz` endpoints for infrastructure checks.

## Getting Started

### Prerequisites
- Go 1.21+
- Make (optional, for automation)
- Docker & Docker Compose (for Postgres)

### 1. Start the Database
```bash
docker-compose up -d
```

### 2. Configure Environment
Create a `.env` file referencing your database credentials and `JWT_SECRET`:
```env
DB_URL=postgres://postgres:postgres@localhost:5432/uptime?sslmode=disable
JWT_SECRET=supersecret
```

### 3. Run the API
```bash
go run cmd/api/main.go
```
*Note: Depending on your migrations setup, they will apply automatically on startup via the `postgres` package.*

## API Documentation

Once the app is running, interactive API docs are available at:
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### Updating SWAG Docs
If you change the handlers or documentation comments, regenerate the Swagger specifications:
```bash
swag init -d cmd/api,internal/api/handlers,internal/domain,internal/api/response --parseDependency --parseInternal
```

## Project Structure

```
.
├── cmd/
│   └── api/                # Application entrypoint
├── docs/                   # Auto-generated Swagger documentation
├── internal/
│   ├── api/                # HTTP handlers, middleware, and response formatting
│   ├── config/             # Environment & configuration management
│   ├── domain/             # Core structs and interfaces
│   ├── logger/             # slog initialization
│   ├── service/            # Core business logic
│   ├── storage/            # PostgreSQL adapters and DB interactions
│   └── worker/             # Background ping scheduler and URL checker
├── migrations/             # SQL Up/Down migration files
└──  docker-compose.yaml    # Infrastructure orchestration
```
