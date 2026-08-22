package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/mohdrashid9678/xlink/migrations"
)

func RunMigrations(dbURL string) error {
	m, err := newMigrateInstance(dbURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	slog.Info("Database migrations up to date")
	return nil
}

func RollbackMigrations(dbURL string, steps int) error {
	m, err := newMigrateInstance(dbURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback migrations: %w", err)
	}

	slog.Info("Database rollback completed", slog.Int("steps", steps))
	return nil
}

func newMigrateInstance(dbURL string) (*migrate.Migrate, error) {
	driver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("create migration source driver: %w", err)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("open database for migration: %w", err)
	}

	instance, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create postgres migration driver instance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", driver, "postgres", instance)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create migrate instance: %w", err)
	}

	return m, nil
}
