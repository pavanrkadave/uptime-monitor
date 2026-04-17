package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pavanrkadave/uptime-monitor/internal/api/response"
	"github.com/pavanrkadave/uptime-monitor/internal/domain"
)

type claimsKey struct{}

var ContextKeyClaims = claimsKey{}

func AuthMiddleware(jwtSecret string, log *slog.Logger) func(http.Handler) http.Handler {
	secretBytes := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return secretBytes, nil
			})

			if err != nil || !token.Valid {
				log.Warn("invalid or expired token", slog.Any("error", err), slog.String("client_ip", r.RemoteAddr))
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(allowedRoles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ContextKeyClaims).(jwt.MapClaims)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "unauthorized: missing claims")
				return
			}

			roleInterface, ok := claims["role"]
			if !ok {
				response.Error(w, http.StatusForbidden, "forbidden: role not found in token")
				return
			}

			userRoleStr, ok := roleInterface.(string)
			if !ok {
				response.Error(w, http.StatusForbidden, "forbidden: invalid role format")
				return
			}

			userRole := domain.Role(userRoleStr)

			isAllowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				response.Error(w, http.StatusForbidden, "forbidden: insufficient privileges")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
