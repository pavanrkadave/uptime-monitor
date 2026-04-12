package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				slog.Info("[API Request]",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", ww.Status()),
					slog.Duration("duration", time.Since(start)),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("request_id", middleware.GetReqID(r.Context())),
					slog.String("user_agent", r.UserAgent()),
				)
			}()
			next.ServeHTTP(ww, r)
		})
	}
}
