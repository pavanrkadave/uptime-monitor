package service

import (
	"context"

	"github.com/pavanrkadave/uptime-monitor/internal/domain"
)

type MonitorStore interface {
	ListAll(ctx context.Context) ([]*domain.Monitor, error)
	GetByID(ctx context.Context, ID int64) (*domain.Monitor, error)
	Create(ctx context.Context, monitor domain.Monitor) (*domain.Monitor, error)
	Update(ctx context.Context, monitor domain.Monitor) (*domain.Monitor, error)
	Delete(ctx context.Context, ID int64) error
}

type MonitorService struct {
	store MonitorStore
}

func NewMonitorService(store MonitorStore) *MonitorService {
	return &MonitorService{
		store: store,
	}
}

func (m *MonitorService) ListAll(ctx context.Context) ([]*domain.Monitor, error) {
	return m.store.ListAll(ctx)
}

func (m *MonitorService) GetByID(ctx context.Context, id int64) (*domain.Monitor, error) {
	return m.store.GetByID(ctx, id)
}

func (m *MonitorService) Create(ctx context.Context, url string) (*domain.Monitor, error) {
	monitorToSave := domain.Monitor{
		URL: url,
	}

	if err := monitorToSave.Validate(); err != nil {
		return nil, err
	}

	return m.store.Create(ctx, monitorToSave)
}

func (m *MonitorService) Update(ctx context.Context, id int64, url string) (*domain.Monitor, error) {
	monitorToUpdate := domain.Monitor{
		ID:  id,
		URL: url,
	}

	if err := monitorToUpdate.Validate(); err != nil {
		return nil, err
	}

	return m.store.Update(ctx, monitorToUpdate)
}

func (m *MonitorService) Delete(ctx context.Context, ID int64) error {
	return m.store.Delete(ctx, ID)
}
