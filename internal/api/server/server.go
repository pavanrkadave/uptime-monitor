package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/pavanrkadave/uptime-monitor/internal/api/handlers"
	"github.com/pavanrkadave/uptime-monitor/internal/api/middleware"
	"github.com/pavanrkadave/uptime-monitor/internal/config"

	_ "github.com/pavanrkadave/uptime-monitor/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	httpServer *http.Server
	log        *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger, monitorHandler *handlers.MonitorHandler, authHandler *handlers.AuthHandler) *Server {
	mux := http.NewServeMux()

	authMW := middleware.AuthMiddleware(cfg.JWTSecret, log)

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json")))

	// Public Routes
	mux.HandleFunc("POST /login", authHandler.HandleLogin)
	mux.HandleFunc("GET /monitors", monitorHandler.HandleList)
	mux.HandleFunc("GET /monitors/{id}", monitorHandler.HandleGetByID)

	// Protected Routes
	mux.Handle("POST /monitors", authMW(http.HandlerFunc(monitorHandler.HandleCreate)))
	mux.Handle("PUT /monitors/{id}", authMW(http.HandlerFunc(monitorHandler.HandleUpdate)))
	mux.Handle("DELETE /monitors/{id}", authMW(http.HandlerFunc(monitorHandler.HandleDelete)))

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: mux,
		},
		log: log.With(slog.String("component", "api-server")),
	}
}

func (s *Server) Start(ctx context.Context) {
	errCh := make(chan error, 1)
	s.log.Info("Starting API server", slog.String("addr", s.httpServer.Addr))

	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		s.log.Info("HTTP Server received shutdown signal...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.log.Warn("HTTP Server forced to shutdown", slog.Any("error", err))
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("Failed to start or run server", slog.Any("error", err))
		}
	}
}
