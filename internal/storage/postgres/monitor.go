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

func (r *MonitorRepository) ListAll(ctx context.Context) ([]*domain.Monitor, error) {
	query := `SELECT id, url, expected_keyword, check_interval, created_at, updated_at FROM monitors ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []*domain.Monitor
	for rows.Next() {
		var monitor domain.Monitor
		if err := rows.Scan(&monitor.ID, &monitor.URL, &monitor.ExpectedKeyword, &monitor.CheckInterval, &monitor.CreatedAt, &monitor.UpdatedAt); err != nil {
			return nil, err
		}
		monitors = append(monitors, &monitor)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return monitors, nil
}

func (r *MonitorRepository) GetByID(ctx context.Context, id int64) (*domain.Monitor, error) {
	query := `SELECT id, url, expected_keyword, check_interval, created_at, updated_at FROM monitors WHERE id = $1`

	var monitor domain.Monitor
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&monitor.ID,
		&monitor.URL,
		&monitor.ExpectedKeyword,
		&monitor.CheckInterval,
		&monitor.CreatedAt,
		&monitor.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMonitorNotFound
		}
		return nil, err
	}
	return &monitor, nil
}

func (r *MonitorRepository) Create(ctx context.Context, monitor domain.Monitor) (*domain.Monitor, error) {
	query := `
		INSERT INTO monitors (url, expected_keyword, check_interval, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, url, expected_keyword, check_interval, created_at, updated_at
	`

	var m domain.Monitor
	err := r.db.QueryRowContext(ctx, query, monitor.URL, monitor.ExpectedKeyword, monitor.CheckInterval).Scan(
		&m.ID,
		&m.URL,
		&m.ExpectedKeyword,
		&m.CheckInterval,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MonitorRepository) Update(ctx context.Context, id int64, url, expectedKeyword string, checkInterval int) (*domain.Monitor, error) {
	query := `
		UPDATE monitors
		SET url = $1, expected_keyword = $2, check_interval = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, url, expected_keyword, check_interval, created_at, updated_at
	`

	var m domain.Monitor
	err := r.db.QueryRowContext(ctx, query, url, expectedKeyword, checkInterval, id).Scan(
		&m.ID,
		&m.URL,
		&m.ExpectedKeyword,
		&m.CheckInterval,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMonitorNotFound
		}
		return nil, err
	}
	return &m, nil
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

func (m *MonitorRepository) GetStats(ctx context.Context, monitorID int64) (*domain.MonitorStats, error) {
	query := `SELECT COUNT(*) as total_pings,
					 COALESCE(
					 	(COUNT(*) FILTER (WHERE is_up = true) * 100.0) / NULLIF(COUNT(*),0), 0
					 ) as uptime_percentage,
					 COALESCE(AVG(duration_ms) FILTER (WHERE is_up = true), 0) as avg_latency_ms,
					 COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms), 0) as p95_latency_ms
			  FROM ping_results
			  WHERE monitor_id = $1;`

	var stats domain.MonitorStats

	err := m.db.QueryRowContext(ctx, query, monitorID).Scan(
		&stats.TotalPings,
		&stats.UptimePercentage,
		&stats.AvgLatencyMs,
		&stats.P95LatencyMs,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}
