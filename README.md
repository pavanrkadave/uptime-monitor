# Uptime Monitor API

A modern, production-ready Go service for monitoring website uptime, calculating latency statistics, and securely managing monitor configurations. Built strictly adhering to Clean Architecture principles.

## 🚀 Features

* **Background Workers:** Asynchronous, non-blocking scheduler that pings URLs at configurable intervals.
* **Native Database Aggregations:** High-performance calculation of Uptime %, Average Latency, and P95 Latency using PostgreSQL native functions (`percentile_cont`, `FILTER`).
* **Secure Identity & RBAC:** JWT-based authentication using Bcrypt password hashing. Auto-seeds a default Admin user on startup.
* **Clean Architecture:** Heavily decoupled Domain, Repository, Service, and Transport (Handler) layers.
* **Modern Telemetry:** Go 1.22+ structured logging (`slog`) with a custom Chi middleware bridge and local `tint` colorization.
* **Auto-Generated Docs:** Fully integrated Swagger UI mapping all endpoints and expected JSON payloads.
* **Fail-Fast Configuration:** Startup configuration validation via `.env` and struct tags.

## 🛠 Tech Stack

* **Language:** Go 1.22+
* **Database:** PostgreSQL
* **Router:** `go-chi/chi/v5`
* **Migrations:** `golang-migrate/migrate`
* **Configuration:** `caarlos0/env/v11` & `joho/godotenv`
* **Documentation:** `swaggo/swag`

## 📁 Project Structure

```text
├── cmd/api/                 # Application entrypoint (main.go)
├── docs/                    # Auto-generated Swagger documentation
├── internal/
│   ├── api/                 # HTTP Transport layer (Handlers, Middleware, Response formatting)
│   ├── config/              # Environment variable loading & validation
│   ├── domain/              # Core business entities and custom errors
│   ├── logger/              # Slog configuration and formatting
│   ├── service/             # Business logic layer (Use Cases)
│   ├── storage/postgres/    # Database implementations (Repositories)
│   └── worker/              # Background polling and scheduling
├── migrations/              # PostgreSQL schema definitions (.sql)
```

## ⚙️ Getting Started

### Prerequisites
* Go 1.22 or higher
* PostgreSQL 14+ running locally or via Docker

### 1. Configuration
Create a `.env` file in the root of the project using the following template:

```env
ENV=development
PORT=8080
DATABASE_URL=postgres://uptime_user:uptime_password@localhost:5432/uptime_db?sslmode=disable
JWT_SECRET=super_secret_local_key
ADMIN_PASSWORD=admin
CHECK_INTERVAL=10s
```

### 2. Running the Application
The application will automatically run pending database migrations and seed the default admin user on startup.

```bash
go run cmd/api/main.go
```

### 3. Accessing the API & Documentation
Once the server is running, you can view the fully interactive Swagger Documentation at:
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### 4. Authentication
To access protected routes (`POST`, `PUT`, `DELETE`), use the `/login` endpoint with the default admin credentials generated on startup:
* **Email:** `admin@example.com`
* **Password:** *(Takes the value of ADMIN_PASSWORD from your .env file)*

Copy the resulting JWT token, click the "Authorize" button in Swagger, and paste the token.

