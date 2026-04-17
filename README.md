# Uptime Monitor API

A robust, production-ready REST API for monitoring website uptime, latency, and generating statistics. Built with Go, following Clean Architecture principles, and packed with a full-scale observability stack.

## Architecture

This project strictly follows **Clean Architecture**:
- **Domain:** Core business models (`User`, `Monitor`, `MonitorStats`).
- **Storage (Repository):** PostgreSQL implementation using `sql.DB` and `golang-migrate` for database migrations.
- **Service:** Business logic layer encapsulating core operations.
- **Transport (Handlers):** HTTP layer powered by `go-chi/chi/v5` and Swagger docs.
- **Worker:** Background scheduling loop that actively tests HTTP endpoints and records results.

## Tech Stack

- **Language:** Go 1.26
- **Database:** PostgreSQL 16
- **Router:** `go-chi/chi/v5`
- **Logging/Metrics:** `log/slog`, Prometheus Client
- **Docs:** `swaggo/swag`
- **Observability Stack:** Prometheus, Grafana, Loki, Promtail
- **Deployment:** Docker & Docker Compose

## Features

- **RBAC Authentication:** Secure endpoints using JWTs with `admin` and `viewer` roles.
- **Auto-Seeding:** Instantly generates a default admin user on startup.
- **Monitor CRUD:** Add, update, and manage websites to monitor.
- **Background Worker:** Non-blocking concurrent pinger that tests monitor endpoints and logs results.
- **Statistics Aggregation:** Real-time calculation of Uptime %, Average Latency, and P95 Latency using PostgreSQL aggregates.
- **Observability Dashboards:** Fully auto-provisioned Grafana dashboards tracking Ping Latency, API Traffic, Goroutines, and container logs (via Loki).
- **Swagger Documentation:** Auto-generated interactive API docs (`/swagger/index.html`).

---

## 🚀 Getting Started (Docker Compose)

The easiest way to run the entire backend, worker, database, and observability stack is using Docker Compose.

### 1. Configure Environment
Create a `.env` file in the root of the project:
```env
DATABASE_URL=postgres://uptime_user:uptime_password@db:5432/uptime_db?sslmode=disable
JWT_SECRET=super_secret_key
ADMIN_PASSWORD=admin
```

### 2. Boot the Full Stack
Bring up the Go API, PostgreSQL, Prometheus, Grafana, Loki, and Promtail all at once:
```bash
docker compose up --build -d
```
*Note: The API container runs database migrations automatically on startup.*

### 3. Access the Services
Once running, the stack binds to the following local ports:
- **API & Swagger Docs:** [http://localhost:8000/swagger/index.html](http://localhost:8000/swagger/index.html)
- **Grafana Dashboards:** [http://localhost:3000](http://localhost:3000) (Login: `admin` / `admin`)
- **Prometheus Metrics:** [http://localhost:9090](http://localhost:9090)
- **Raw API Metrics Route:** [http://localhost:8000/metrics](http://localhost:8000/metrics)

---

## Observability & Dashboards

By navigating to **Grafana (Port 3000)** -> **Dashboards**, you will see an automatically provisioned `Uptime Monitor & API Health` dashboard. This JSON-defined dashboard streams:
1. **Target Up/Down Status:** Real-time Prometheus Gauges showing if your links are active.
2. **Ping Latency:** Histograms tracking HTTP check speeds.
3. **API Traffic Rates:** Total requests hitting the chi router.
4. **Go Internal Stats:** Goroutines and memory allocations.
5. **Container Logs:** Live streaming logs from the API using Loki and Promtail.

---

## Local Development (Without Docker)

If you only want to run the Go API locally, start just the Postgres container:
```bash
docker compose up db -d
```

Override the `DATABASE_URL` in your `.env` to point to `localhost`:
```env
DATABASE_URL=postgres://uptime_user:uptime_password@localhost:5432/uptime_db?sslmode=disable
```

Run the API:
```bash
go run cmd/api/main.go
```

### Updating Swagger Docs
If you change the handlers or documentation comments, regenerate the Swagger specifications:
```bash
swag init -d cmd/api,internal/api/handlers,internal/domain,internal/api/response --parseDependency --parseInternal
```

