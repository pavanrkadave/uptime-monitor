# =======================================
# STAGE 1: Build the application
# =======================================
FROM golang:1.26-alpine AS builder

RUN adduser -D -g '' appuser

WORKDIR /app

# Copy go.mod and go.sum to the workspace
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code to the workspace
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o uptime-api ./cmd/api

# =======================================
# STAGE 2: Create a minimal runtime image
# =======================================
FROM alpine:latest

COPY --from=builder /etc/passwd /etc/passwd

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/uptime-api .

USER appuser

# Expose the port the application runs on
EXPOSE 8080

# Command to run the application
CMD ["./uptime-api"]

