package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/pavanrkadave/uptime-monitor/internal/domain"
)

type MonitorUseCase interface {
	ListAll(ctx context.Context) ([]*domain.Monitor, error)
	Create(ctx context.Context, url string) (*domain.Monitor, error)
	GetByID(ctx context.Context, id int64) (*domain.Monitor, error)
	Update(ctx context.Context, id int64, url string) (*domain.Monitor, error)
	Delete(ctx context.Context, id int64) error
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

func (h *MonitorHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	monitors, err := h.useCase.ListAll(r.Context())
	if err != nil {
		h.log.Error("failed list monitors", slog.Any("error", err))
		http.Error(w, "Failed to retrieve monitors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(monitors); err != nil {
		h.log.Error("failed encode response", slog.Any("error", err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *MonitorHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("failed parse id", slog.Any("error", err))
		http.Error(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}

	monitor, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrMonitorNotFound) {
			h.log.Error("failed find monitor", slog.Any("error", err))
			http.Error(w, "Monitor not found", http.StatusNotFound)
			return
		}
		h.log.Error("error fetching monitor", slog.Any("error", err))
		http.Error(w, "Failed to retrieve monitor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(monitor); err != nil {
		h.log.Error("failed encode response", slog.Any("error", err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
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
// @Failure      400 {string} string "invalid request payload"
// @Failure      401 {string} string "unauthorized"
// @Failure      500 {string} string "internal server error"
// @Security     BearerAuth
// @Router       /monitors [post]
func (h *MonitorHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var request CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// If the JSON is malformed, send a 400 Bad Request
		h.log.Error("failed decode create request", slog.Any("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	monitor, err := h.useCase.Create(r.Context(), request.URL)
	if err != nil {
		h.log.Error("failed create new monitor", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := CreateResponse{
		MonitorID: monitor.ID,
		URL:       monitor.URL,
		CreatedAt: monitor.CreatedAt,
		UpdatedAt: monitor.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("failed encode response", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *MonitorHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}

	var request UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.log.Error("failed decode update request", slog.Any("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedMonitor, err := h.useCase.Update(r.Context(), id, request.URL)
	if err != nil {
		if errors.Is(err, domain.ErrMonitorNotFound) {
			h.log.Error("failed find monitor", slog.Any("error", err))
			http.Error(w, "Monitor not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, domain.ErrEmptyURL) || errors.Is(err, domain.ErrInvalidURL) || errors.Is(err, domain.ErrMissingScheme) {
			h.log.Error("validation failed for monitor", slog.Any("error", err))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		h.log.Error("failed to update monitor", slog.Any("error", err))
		http.Error(w, "Failed to retrieve monitor", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedMonitor); err != nil {
		h.log.Error("failed encode response", slog.Any("error", err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *MonitorHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("failed parse id", slog.Any("error", err))
		http.Error(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}

	err = h.useCase.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrMonitorNotFound) {
			h.log.Error("failed find monitor", slog.Any("error", err))
			http.Error(w, "Monitor not found", http.StatusNotFound)
			return
		}
		h.log.Error("failed to delete monitor", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
