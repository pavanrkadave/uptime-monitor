package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pavanrkadave/uptime-monitor/internal/domain"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pavanrkadave/uptime-monitor/internal/api/handlers"
	middlewares "github.com/pavanrkadave/uptime-monitor/internal/api/middleware"
	"github.com/pavanrkadave/uptime-monitor/internal/config"

	_ "github.com/pavanrkadave/uptime-monitor/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	httpServer *http.Server
	log        *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger, monitorHandler *handlers.MonitorHandler, authHandler *handlers.AuthHandler, healthHandler *handlers.HealthHandler) *Server {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.MetricsMiddleware())
	r.Use(middlewares.RequestLogger(log))

	authMW := middlewares.AuthMiddleware(cfg.JWTSecret, log)

	// --- Swagger UI ---
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	// --- Operations Endpoints ---
	r.Get("/healthz", healthHandler.HandleHealth)
	r.Get("/readyz", healthHandler.HandleReadiness)
	r.Handle("/metrics", promhttp.Handler())

	// --- Public Routes ---
	r.Post("/login", authHandler.HandleLogin)
	r.Get("/monitors", monitorHandler.HandleList)
	r.Get("/monitors/{id}", monitorHandler.HandleGetByID)
	r.Get("/monitors/{id}/stats", monitorHandler.HandleMonitorStats)

	// --- Protected Routes ---
	r.Group(func(r chi.Router) {
		r.Use(authMW)

		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole(domain.RoleAdmin))
			r.Post("/register", authHandler.HandleRegister)
			r.Delete("/monitors/{id}", monitorHandler.HandleDelete)
		})

		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole(domain.RoleAdmin, domain.RoleViewer))
			r.Post("/monitors", monitorHandler.HandleCreate)
			r.Put("/monitors/{id}", monitorHandler.HandleUpdate)
		})
	})

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: r,
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
