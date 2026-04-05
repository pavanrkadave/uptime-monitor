package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pavanrkadave/uptime-monitor/internal/domain"
	"github.com/stretchr/testify/assert"
)

type MockStore struct{}

func (m MockStore) ListAll(ctx context.Context) ([]*domain.Monitor, error)         { return nil, nil }
func (m MockStore) GetByID(ctx context.Context, ID int64) (*domain.Monitor, error) { return nil, nil }
func (m MockStore) Create(ctx context.Context, monitor domain.Monitor) (*domain.Monitor, error) {
	return nil, errors.New("database Create called unexpectedly")
}
func (m MockStore) Update(ctx context.Context, monitor domain.Monitor) (*domain.Monitor, error) {
	return nil, nil
}
func (m MockStore) Delete(ctx context.Context, ID int64) error { return nil }

type MockStoreSuccess struct {
	MockStore
}

func (m MockStoreSuccess) Create(ctx context.Context, monitor domain.Monitor) (*domain.Monitor, error) {
	now := time.Now()
	monitor.ID = 99
	monitor.CreatedAt = &now
	monitor.UpdatedAt = &now
	return &monitor, nil
}

func TestMonitorService_Create_ValidationFails(t *testing.T) {
	// Arrange
	mockStore := MockStore{}
	svc := NewMonitorService(mockStore)

	// Act
	result, err := svc.Create(t.Context(), "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrEmptyURL)
}

func TestMonitorService_Create_Success(t *testing.T) {
	// Arrange
	mockStore := MockStoreSuccess{}
	svc := NewMonitorService(mockStore)

	// Act
	result, err := svc.Create(t.Context(), "https://golang.org")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(99), result.ID)
	assert.Equal(t, "https://golang.org", result.URL)
	assert.NotNil(t, result.CreatedAt)
}
