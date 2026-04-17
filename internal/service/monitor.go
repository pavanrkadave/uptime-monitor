package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/pavanrkadave/uptime-monitor/internal/domain"
)

type MonitorStore interface {
	ListAll(ctx context.Context) ([]*domain.Monitor, error)
	GetByID(ctx context.Context, ID int64) (*domain.Monitor, error)
	Create(ctx context.Context, monitor domain.Monitor) (*domain.Monitor, error)
	Update(ctx context.Context, id int64, url, expectedKeyword string) (*domain.Monitor, error)
	Delete(ctx context.Context, ID int64) error
	SavePingResult(ctx context.Context, monitorID int64, isUp bool, statusCode int, duration time.Duration, errMsg string) error
	GetStats(ctx context.Context, monitorID int64) (*domain.MonitorStats, error)
}

type MonitorService struct {
	store MonitorStore
	log   *slog.Logger
}

func NewMonitorService(store MonitorStore, log *slog.Logger) *MonitorService {
	return &MonitorService{
		store: store,
		log:   log.With(slog.String("component", "monitor-service")),
	}
}

func (m *MonitorService) ListAll(ctx context.Context) ([]*domain.Monitor, error) {
	m.log.Info("getting all monitors")
	return m.store.ListAll(ctx)
}

func (m *MonitorService) GetByID(ctx context.Context, id int64) (*domain.Monitor, error) {
	m.log.Info("getting monitor for id", slog.Int64("id", id))
	return m.store.GetByID(ctx, id)
}

func (m *MonitorService) Create(ctx context.Context, url, expectedKeyword string) (*domain.Monitor, error) {
	m.log.Info("creating monitor", slog.String("url", url))
	monitorToSave := domain.Monitor{
		URL:             url,
		ExpectedKeyword: expectedKeyword,
	}

	if err := monitorToSave.Validate(); err != nil {
		m.log.Debug("validation failed for monitor", slog.String("url", url))
		return nil, err
	}

	return m.store.Create(ctx, monitorToSave)
}

func (m *MonitorService) Update(ctx context.Context, id int64, url, expectedKeyword string) (*domain.Monitor, error) {
	m.log.Info("updating monitor", slog.Int64("id", id))
	monitorToUpdate := domain.Monitor{
		ID:              id,
		URL:             url,
		ExpectedKeyword: expectedKeyword,
	}

	if err := monitorToUpdate.Validate(); err != nil {
		m.log.Debug("validation failed for monitor", slog.String("url", url))
		return nil, err
	}

	return m.store.Update(ctx, monitorToUpdate.ID, monitorToUpdate.URL, monitorToUpdate.ExpectedKeyword)
}

func (m *MonitorService) Delete(ctx context.Context, ID int64) error {
	m.log.Info("deleting monitor", slog.Int64("id", ID))
	return m.store.Delete(ctx, ID)
}

func (m *MonitorService) SavePingResult(ctx context.Context, monitorID int64, isUp bool, statusCode int, duration time.Duration, errMsg string) error {
	m.log.Info("saving ping result", slog.Int64("monitorID", monitorID))
	return m.store.SavePingResult(ctx, monitorID, isUp, statusCode, duration, errMsg)
}

func (m *MonitorService) GetStats(ctx context.Context, monitorID int64) (*domain.MonitorStats, error) {
	m.log.Info("getting stats for monitor", slog.Int64("monitorID", monitorID))
	return m.store.GetStats(ctx, monitorID)
}
