package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pavanrkadave/uptime-monitor/internal/api/response"
	"github.com/pavanrkadave/uptime-monitor/internal/domain"
)

type MonitorUseCase interface {
	ListAll(ctx context.Context) ([]*domain.Monitor, error)
	Create(ctx context.Context, url string) (*domain.Monitor, error)
	GetByID(ctx context.Context, id int64) (*domain.Monitor, error)
	Update(ctx context.Context, id int64, url string) (*domain.Monitor, error)
	Delete(ctx context.Context, id int64) error
	GetStats(ctx context.Context, monitorID int64) (*domain.MonitorStats, error)
}

type MonitorHandler struct {
	useCase MonitorUseCase
	log     *slog.Logger
}

type CreateRequest struct {
	URL string `json:"url"`
}

type UpdateRequest struct {
	URL string `json:"url"`
}

type CreateResponse struct {
	MonitorID int64      `json:"monitor_id"`
	URL       string     `json:"url"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func NewMonitorHandler(useCase MonitorUseCase, log *slog.Logger) *MonitorHandler {
	return &MonitorHandler{
		useCase: useCase,
		log:     log.With(slog.String("component", "monitor-handler")),
	}
}

// HandleList returns all monitors in the system.
//
// @Summary      List Monitors
// @Description  Retrieve a list of all monitors currently being tracked.
// @Tags         Monitors
// @Produce      json
// @Success      200 {array} domain.Monitor
// @Failure      500 {object} response.ErrorResponse
// @Router       /monitors [get]
func (h *MonitorHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	monitors, err := h.useCase.ListAll(r.Context())
	if err != nil {
		h.log.Error("failed list monitors", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve monitors")
		return
	}

	response.JSON(w, http.StatusOK, monitors)
}

// HandleGetByID returns a specific monitor by its ID.
//
// @Summary      Get Monitor by ID
// @Description  Retrieve details of a specific monitor using its unique ID.
// @Tags         Monitors
// @Produce      json
// @Param        id   path      int  true  "Monitor ID"
// @Success      200 {object} domain.Monitor
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /monitors/{id} [get]
func (h *MonitorHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("failed parse id", slog.Any("error", err))
		response.Error(w, http.StatusBadRequest, "Invalid monitor ID")
		return
	}

	monitor, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrMonitorNotFound) {
			h.log.Error("failed find monitor", slog.Any("error", err))
			response.Error(w, http.StatusNotFound, "Monitor not found")
			return
		}
		h.log.Error("error fetching monitor", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve monitor")
		return
	}

	response.JSON(w, http.StatusOK, monitor)
}

// HandleCreate decodes a JSON body {"url": "..."} and creates a new monitor.
//
// @Summary      Create Monitor
// @Description  Add a new URL to the uptime monitoring system.
// @Tags         Monitors
// @Accept       json
// @Produce      json
// @Param        request body CreateRequest true "Monitor Details"
// @Success      201 {object} CreateResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /monitors [post]
func (h *MonitorHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var request CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// If the JSON is malformed, send a 400 Bad Request
		h.log.Error("failed decode create request", slog.Any("error", err))
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	monitor, err := h.useCase.Create(r.Context(), request.URL)
	if err != nil {
		h.log.Error("failed create new monitor", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	monitorResponse := CreateResponse{
		MonitorID: monitor.ID,
		URL:       monitor.URL,
		CreatedAt: monitor.CreatedAt,
		UpdatedAt: monitor.UpdatedAt,
	}
	response.JSON(w, http.StatusCreated, monitorResponse)
}

// HandleUpdate decodes a JSON body {"url": "..."} and updates a new monitor.
//
// @Summary      Update Monitor
// @Description  Update the URL of an existing monitor in the system.
// @Tags         Monitors
// @Accept       json
// @Produce      json
// @Param        id      path     int  true  "Monitor ID"
// @Param        request body UpdateRequest true "Updated Monitor Details"
// @Success      200 {object} domain.Monitor
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /monitors/{id} [put]
func (h *MonitorHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid monitor ID")
		return
	}

	var request UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.log.Error("failed decode update request", slog.Any("error", err))
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedMonitor, err := h.useCase.Update(r.Context(), id, request.URL)
	if err != nil {
		if errors.Is(err, domain.ErrMonitorNotFound) {
			h.log.Error("failed find monitor", slog.Any("error", err))
			response.Error(w, http.StatusNotFound, "Monitor not found")
			return
		}

		if errors.Is(err, domain.ErrEmptyURL) || errors.Is(err, domain.ErrInvalidURL) || errors.Is(err, domain.ErrMissingScheme) {
			h.log.Error("validation failed for monitor", slog.Any("error", err))
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		h.log.Error("failed to update monitor", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "Failed to update monitor")
		return
	}

	response.JSON(w, http.StatusOK, updatedMonitor)
}

// HandleDelete removes a monitor from the system entirely.
//
// @Summary      Delete Monitor
// @Description  Remove a monitor using its unique ID.
// @Tags         Monitors
// @Accept       json
// @Produce      json
// @Param        id      path     int  true  "Monitor ID"
// @Success      200 {object} domain.Monitor
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /monitors/{id} [delete]
func (h *MonitorHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("failed parse id", slog.Any("error", err))
		response.Error(w, http.StatusBadRequest, "Invalid monitor ID")
		return
	}

	err = h.useCase.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrMonitorNotFound) {
			h.log.Error("failed find monitor", slog.Any("error", err))
			response.Error(w, http.StatusNotFound, "Monitor not found")
			return
		}
		h.log.Error("failed to delete monitor", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusNoContent, nil)
}

// HandleMonitorStats Fetched stats for a monitored URL.
//
// @Summary      Fetch Monitor Stats
// @Description  Retrieve stats for a monitored URL using its unique ID.
// @Tags         Monitors
// @Produce      json
// @Param        id   path      int  true  "Monitor ID"
// @Success      200 {object} domain.MonitorStats
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /monitors/{id}/stats [get]
func (h *MonitorHandler) HandleMonitorStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("failed parse id", slog.Any("error", err))
		response.Error(w, http.StatusBadRequest, "Invalid monitor ID")
		return
	}

	monitor, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrMonitorNotFound) {
			h.log.Error("failed find monitor", slog.Any("error", err))
			response.Error(w, http.StatusNotFound, "Monitor not found")
			return
		}
		h.log.Error("failed to fetch monitor", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "Failed to fetch monitor")
		return
	}

	stats, err := h.useCase.GetStats(r.Context(), monitor.ID)
	if err != nil {
		h.log.Error("failed fetch monitor stats", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve monitor stats")
		return
	}

	response.JSON(w, http.StatusOK, stats)
}
