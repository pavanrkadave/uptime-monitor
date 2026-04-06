package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/pavanrkadave/uptime-monitor/internal/domain"
)

type MonitorRepository struct {
	db *sql.DB
}

func NewMonitorRepository(db *sql.DB) *MonitorRepository {
	return &MonitorRepository{
		db: db,
	}
}

func (m *MonitorRepository) ListAll(ctx context.Context) ([]*domain.Monitor, error) {
	query := `SELECT id, url, created_at, updated_at FROM monitors ORDER BY id`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monitors := make([]*domain.Monitor, 0)
	for rows.Next() {
		var monitor domain.Monitor
		if err := rows.Scan(&monitor.ID, &monitor.URL, &monitor.CreatedAt, &monitor.UpdatedAt); err != nil {
			return nil, err
		}
		monitors = append(monitors, &monitor)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return monitors, nil
}

func (m *MonitorRepository) GetByID(ctx context.Context, ID int64) (*domain.Monitor, error) {
	query := `SELECT id, url, created_at, updated_at FROM monitors WHERE id = $1`

	var monitor domain.Monitor
	err := m.db.QueryRowContext(ctx, query, ID).Scan(&monitor.ID, &monitor.URL, &monitor.CreatedAt, &monitor.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMonitorNotFound
		}
		return nil, err
	}
	return &monitor, nil
}

func (m *MonitorRepository) Create(ctx context.Context, monitor domain.Monitor) (*domain.Monitor, error) {
	query := `
             INSERT INTO monitors (url)
             VALUES ($1)
             RETURNING id, created_at , updated_at`

	err := m.db.QueryRowContext(ctx, query, monitor.URL).Scan(&monitor.ID, &monitor.CreatedAt, &monitor.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &monitor, nil
}

func (m *MonitorRepository) Update(ctx context.Context, monitor domain.Monitor) (*domain.Monitor, error) {
	query := `UPDATE monitors
			  SET url = $2, updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id,url, created_at, updated_at;`

	var updatedMonitor domain.Monitor
	err := m.db.QueryRowContext(ctx, query, monitor.URL, monitor.ID).Scan(
		&updatedMonitor.ID, &updatedMonitor.URL, &updatedMonitor.CreatedAt, &updatedMonitor.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMonitorNotFound
		}
		return nil, err
	}
	return &updatedMonitor, nil
}

func (m *MonitorRepository) Delete(ctx context.Context, ID int64) error {
	query := `DELETE FROM monitors WHERE id = $1;`

	result, err := m.db.ExecContext(ctx, query, ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrMonitorNotFound
	}

	return nil
}

func (m *MonitorRepository) SavePingResult(ctx context.Context, monitorID int64, isUp bool, statusCode int, duration time.Duration, errMsg string) error {
	query := `INSERT INTO ping_results (monitor_id, is_up, status_code, duration_ms, error_message) VALUES ($1, $2, $3, $4, $5);`

	_, err := m.db.ExecContext(ctx, query, monitorID, isUp, statusCode, duration.Milliseconds(), errMsg)
	return err
}
