package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/pavanrkadave/uptime-monitor/internal/api/response"
	"github.com/pavanrkadave/uptime-monitor/internal/domain"
	"github.com/pavanrkadave/uptime-monitor/internal/service"
)

type AuthUseCase interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, error)
}

type AuthHandler struct {
	useCase AuthUseCase
	log     *slog.Logger
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Role     domain.Role `json:"role"`
}

func NewAuthHandler(useCase AuthUseCase, log *slog.Logger) *AuthHandler {
	return &AuthHandler{
		useCase: useCase,
		log:     log.With(slog.String("component", "auth-handler")),
	}
}

// HandleLogin decodes a JSON body {"email": "...", "password": "..."} and returns a JWT
//
// @Summary      Login to application
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

	if req.Email == "" || req.Password == "" {
		h.log.Debug("invalid login request", slog.Any("error", errors.New("invalid login request")))
		response.Error(w, http.StatusBadRequest, "invalid login request")
		return
	}

	token, err := h.useCase.Login(r.Context(), req.Email, req.Password)
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

// HandleRegister creates a new user. Only Admins should be able to reach this.
//
// @Summary      Create a new user
// @Description  Creates a new admin or viewer.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "User Details"
// @Success      201 {object} response.SuccessResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /register [post]
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debug("failed to decode register request", slog.Any("error", err))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Role == "" {
		response.Error(w, http.StatusBadRequest, "email, password and role are required")
		return
	}

	if req.Role != domain.RoleAdmin && req.Role != domain.RoleViewer {
		response.Error(w, http.StatusBadRequest, "invalid role: must be admin or viewer")
	}

	user, err := h.useCase.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			response.Error(w, http.StatusBadRequest, "user with this email already exists")
			return
		}
		h.log.Error("register error", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user.PasswordHash = ""
	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Message: "User registered successfully",
		Data:    user,
	})
}
