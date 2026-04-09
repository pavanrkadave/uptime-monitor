package postgres

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	pgMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/pavanrkadave/uptime-monitor/migrations"
)

// RunMigrations executes our SQL migrations files against Postgres
func RunMigrations(db *sql.DB) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	dbDriver, err := pgMigrate.WithInstance(db, &pgMigrate.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(
		"iofs", sourceDriver,
		"postgres", dbDriver,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("Database schema is up to date. No migrations applied.")
			return nil
		}
		return err
	}
	slog.Info("Database migrations applied successfully!")
	return nil
}
