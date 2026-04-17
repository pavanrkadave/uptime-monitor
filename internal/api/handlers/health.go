package handlers

import (
	"context"
	"net/http"

	"github.com/pavanrkadave/uptime-monitor/internal/api/response"
)

type HealthChecker interface {
	Check(ctx context.Context) error
}
type HealthHandler struct {
	checker HealthChecker
}

func NewHealthHandler(checker HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// HandleHealth responds with HTTP 200 OK simply indicating the application process is running.
//
// @Summary      Liveness Probe
// @Description  Check if the application process is running.
// @Tags         Operations
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Router       /healthz [get]
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, response.SuccessResponse{Message: "OK"})
}

// HandleReadiness responds with HTTP 200 OK only if all underlying dependencies (like PostgreSQL) are connected.
//
// @Summary      Readiness Probe
// @Description  Check if the application is fully ready to accept traffic.
// @Tags         Operations
// @Produce      json
// @Success      200 {object} response.SuccessResponse
// @Failure      503 {object} response.ErrorResponse
// @Router       /readyz [get]
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	if h.checker != nil {
		if err := h.checker.Check(r.Context()); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "Database not ready: "+err.Error())
			return
		}
	}
	response.JSON(w, http.StatusOK, response.SuccessResponse{Message: "Ready"})
}
