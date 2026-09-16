package database

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func newMigrateInstance(db *sql.DB, migrationsDir string) (*migrate.Migrate, error) {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create mysql migration driver: %w", err)
	}

	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("could not resolve migrations directory: %w", err)
	}

	sourceURL := fmt.Sprintf("file://%s", absDir)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "mysql", driver)
	if err != nil {
		return nil, fmt.Errorf("could not create migrate instance: %w", err)
	}

	return m, nil
}

// RunMigrations applies all pending migrations from migrationsDir.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	m, err := newMigrateInstance(db, migrationsDir)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration up failed: %w", err)
	}

	return nil
}

// RollbackMigrations rolls back all migrations.
func RollbackMigrations(db *sql.DB, migrationsDir string) error {
	m, err := newMigrateInstance(db, migrationsDir)
	if err != nil {
		return err
	}

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration down failed: %w", err)
	}

	return nil
}
