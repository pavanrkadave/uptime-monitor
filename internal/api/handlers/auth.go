package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/pavanrkadave/uptime-monitor/internal/api/response"
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
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /login [post]
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debug("failed to decode login request", slog.Any("error", err))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := h.useCase.Login(r.Context(), req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		h.log.Error("login error", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, LoginResponse{Token: token})
}
