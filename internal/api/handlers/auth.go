package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/pavanrkadave/uptime-monitor/internal/service"
)

type AuthUseCase interface {
	Login(ctx context.Context, password string) (string, error)
}

type AuthHandler struct {
	useCase AuthUseCase
	log     *slog.Logger
}

type LoginRequest struct {
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func NewAuthHandler(useCase AuthUseCase, log *slog.Logger) *AuthHandler {
	return &AuthHandler{
		useCase: useCase,
		log:     log.With(slog.String("component", "auth-handler")),
	}
}

// HandleLogin decodes a JSON body {"password": "..."} and returns a JWT
//
// @Summary      Admin Login
// @Description  Authenticate using an admin password to receive a JWT.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login Credentials"
// @Success      200 {object} LoginResponse
// @Failure      400 {string} string "invalid request body"
// @Failure      401 {string} string "unauthorized"
// @Failure      500 {string} string "internal server error"
// @Router       /login [post]
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debug("failed to decode login request", slog.Any("error", err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.useCase.Login(r.Context(), req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		h.log.Error("login error", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(LoginResponse{Token: token}); err != nil {
		h.log.Error("failed to encode response", slog.Any("error", err))
	}
}
