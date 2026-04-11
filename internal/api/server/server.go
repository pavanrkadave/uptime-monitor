package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/pavanrkadave/uptime-monitor/internal/api/handlers"
	"github.com/pavanrkadave/uptime-monitor/internal/config"
)

type Server struct {
	httpServer *http.Server
	log        *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger, monitorHandler *handlers.MonitorHandler, authHandler *handlers.AuthHandler) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /login", authHandler.HandleLogin)

	mux.HandleFunc("POST /monitors", monitorHandler.HandleCreate)
	mux.HandleFunc("GET /monitors", monitorHandler.HandleList)
	mux.HandleFunc("GET /monitors/{id}", monitorHandler.HandleGetByID)
	mux.HandleFunc("PUT /monitors/{id}", monitorHandler.HandleUpdate)
	mux.HandleFunc("DELETE /monitors/{id}", monitorHandler.HandleDelete)

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
