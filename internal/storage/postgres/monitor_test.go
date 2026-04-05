package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/pavanrkadave/uptime-monitor/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("uptime_db_test"),
		postgres.WithUsername("uptime_user"),
		postgres.WithPassword("uptime_password"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate PostgreSQL container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS monitors (
		id BIGSERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		created_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ
	);`)
	require.NoError(t, err)

	return db
}

func TestCreate(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	store := NewMonitorRepository(db)
	ctx := t.Context()

	monitorToCreate := domain.Monitor{
		URL:       "http://google.com",
		UpdatedAt: new(time.Now()),
	}

	// Act
	createdMonitor, err := store.Create(ctx, monitorToCreate)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, createdMonitor)
	assert.NotZero(t, createdMonitor.ID)
	assert.Equal(t, monitorToCreate.URL, createdMonitor.URL)
}

func TestGetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DROP TABLE IF EXISTS monitors;")
	repo := NewMonitorRepository(db)

	_, err := repo.GetByID(t.Context(), 999)
	assert.ErrorIs(t, err, domain.ErrMonitorNotFound)

	created, err := repo.Create(t.Context(), domain.Monitor{URL: "https://test.com"})
	assert.NoError(t, err)

	fetched, err := repo.GetByID(t.Context(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "https://test.com", fetched.URL)
	assert.Equal(t, created.ID, fetched.ID)
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DROP TABLE IF EXISTS monitors;")
	repo := NewMonitorRepository(db)

	err := repo.Delete(t.Context(), 999)
	assert.ErrorIs(t, err, domain.ErrMonitorNotFound)

	created, _ := repo.Create(t.Context(), domain.Monitor{URL: "https://delete.com"})

	err = repo.Delete(t.Context(), created.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(t.Context(), created.ID)
	assert.ErrorIs(t, err, domain.ErrMonitorNotFound)
}
