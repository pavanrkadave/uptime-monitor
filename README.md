# Uptime Monitor API

A robust, production-ready REST API for monitoring website uptime, latency, and TLS certificate expirations. Built with Go, following Clean Architecture principles, and packed with a full-scale observability stack running behind a secure Caddy reverse proxy.

## Architecture

This project strictly follows **Clean Architecture**:
- **Domain:** Core business models (`User`, `Monitor`, `MonitorStats`).
- **Storage (Repository):** PostgreSQL implementation using `sql.DB` and `golang-migrate` for database migrations.
- **Service:** Business logic layer encapsulating core operations.
- **Transport (Handlers):** HTTP layer powered by `go-chi/chi/v5` and Swagger docs.
- **Worker:** Background scheduling loop that actively tests HTTP endpoints, checks TLS certificates, and records results.

## Tech Stack

- **Language:** Go 1.26
- **Database:** PostgreSQL 16
- **Router:** `go-chi/chi/v5`
- **Reverse Proxy:** Caddy (Auto HTTPS)
- **Logging/Metrics:** `log/slog`, Prometheus Client
- **Docs:** `swaggo/swag`
- **Observability Stack:** Prometheus, Grafana, Loki, Promtail
- **Deployment:** Docker & Docker Compose (Multi-stage lightweight builds)

## Features

- **RBAC Authentication:** Secure endpoints using JWTs with `admin` and `viewer` roles. Publicly accessible `/stats` endpoints.
- **Auto-Seeding:** Instantly generates a default admin user on startup.
- **Monitor CRUD:** Add, update, and manage websites to monitor.
- **Background Worker & TLS Monitoring:** Non-blocking concurrent pinger that tests monitor endpoints, tracks response times, and parses SSL/TLS certificate expiry dates.
- **Statistics Aggregation:** Real-time calculation of Uptime %, Average Latency, and P95 Latency using PostgreSQL aggregates.
- **Observability Dashboards:** Fully auto-provisioned Grafana dashboards tracking Ping Latency, API Traffic, TLS Expiry Countdowns, Goroutines, and container logs (via Loki).
- **Persistent Storage:** Docker named volumes ensure your Grafana dashboards, Prometheus metrics, Loki logs, and PostgreSQL data survive container restarts.
- **Secure Traffic (HTTPS):** All services are hidden behind a securely configured Caddy reverse proxy on `localhost`.

---

## 🚀 Getting Started (Docker Compose)

The easiest way to run the entire backend, worker, database, and observability stack is using Docker Compose. The stack is securely locked down—direct port access is disabled, and all traffic routes through Caddy.

### 1. Configure Environment
Create a `.env` file in the root of the project:
```env
DATABASE_URL=postgres://uptime_user:uptime_password@db:5432/uptime_db?sslmode=disable
JWT_SECRET=super_secret_key
ADMIN_PASSWORD=admin
```

### 2. Generate Local SSL Certificates
To ensure your browser trusts the local domains, we use `mkcert` to generate certificates for Caddy. 
Run these commands on your host machine:
```bash
brew install mkcert nss
mkcert -install
mkcert api.localhost grafana.localhost
```
*(This will generate `api.localhost+1.pem` and `api.localhost+1-key.pem` in your project root, which Docker Compose will mount into Caddy).*

### 3. Boot the Full Stack
Bring up the Go API, PostgreSQL, Prometheus, Grafana, Loki, Promtail, and Caddy all at once:
```bash
docker compose up --build -d
```

### 4. Access the Services
Since everything runs through Caddy, access your apps via these secure local domains:
- **API Health:** [https://api.localhost/healthz](https://api.localhost/healthz)
- **API Swagger Docs:** [https://api.localhost/swagger/index.html](https://api.localhost/swagger/index.html)
- **Grafana Dashboards:** [https://grafana.localhost](https://grafana.localhost) (Login: `admin` / `admin`)
- **Direct Database Access:** `localhost:5432` (Only the DB port is exposed publicly for local inspection tools like DBeaver)

---

## Observability & Dashboards

By navigating to **Grafana (https://grafana.localhost)** -> **Dashboards**, you will see an automatically provisioned `Uptime Monitor & API Health` dashboard. This JSON-defined dashboard streams:
1. **Target Up/Down Status:** Real-time Prometheus Gauges showing if your links are active.
2. **TLS Certificate Expiry:** A live countdown of the days remaining until a monitored website's SSL certificate expires.
3. **Ping Latency:** Histograms tracking HTTP check speeds.
4. **API Traffic Rates:** Total requests hitting the chi router.
5. **Go Internal Stats:** Goroutines and memory allocations.
6. **Container Logs:** Live streaming logs from the API using Loki and Promtail.

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

