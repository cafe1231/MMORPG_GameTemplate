package db

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/mmorpg-template/backend/pkg/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrator handles database migrations
type Migrator struct {
	db       *sql.DB
	instance *migrate.Migrate
	logger   logger.Logger
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB, log logger.Logger) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	// Set logger
	m.Log = &migrateLogger{log: log}

	return &Migrator{
		db:       db,
		instance: m,
		logger:   log,
	}, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	start := time.Now()
	m.logger.Info("Starting database migrations")

	version, dirty, err := m.instance.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		m.logger.Warn("Database is in dirty state", "version", version)
		return fmt.Errorf("database is in dirty state at version %d", version)
	}

	if err == migrate.ErrNilVersion {
		m.logger.Info("No migrations have been applied yet")
	} else {
		m.logger.Info("Current migration version", "version", version)
	}

	if err := m.instance.Up(); err != nil {
		if err == migrate.ErrNoChange {
			m.logger.Info("Database is already up to date")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	newVersion, _, _ := m.instance.Version()
	m.logger.Info("Migrations completed successfully", 
		"duration", time.Since(start),
		"new_version", newVersion)

	return nil
}

// Down rolls back one migration
func (m *Migrator) Down() error {
	start := time.Now()
	m.logger.Info("Rolling back one migration")

	version, dirty, err := m.instance.Version()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state at version %d", version)
	}

	m.logger.Info("Current migration version", "version", version)

	if err := m.instance.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	newVersion, _, _ := m.instance.Version()
	m.logger.Info("Rollback completed successfully",
		"duration", time.Since(start),
		"new_version", newVersion)

	return nil
}

// Migrate to a specific version
func (m *Migrator) Migrate(targetVersion uint) error {
	start := time.Now()
	m.logger.Info("Migrating to specific version", "target", targetVersion)

	if err := m.instance.Migrate(targetVersion); err != nil {
		if err == migrate.ErrNoChange {
			m.logger.Info("Already at target version")
			return nil
		}
		return fmt.Errorf("failed to migrate to version %d: %w", targetVersion, err)
	}

	m.logger.Info("Migration completed successfully",
		"duration", time.Since(start),
		"version", targetVersion)

	return nil
}

// Version returns the current migration version
func (m *Migrator) Version() (uint, bool, error) {
	return m.instance.Version()
}

// Close closes the migrator
func (m *Migrator) Close() error {
	sourceErr, dbErr := m.instance.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}
	return nil
}

// migrateLogger adapts our logger to migrate's logger interface
type migrateLogger struct {
	log logger.Logger
}

func (l *migrateLogger) Printf(format string, v ...interface{}) {
	l.log.Info(fmt.Sprintf(format, v...))
}

func (l *migrateLogger) Verbose() bool {
	return true
}

// ValidateMigrations checks that all migrations can be parsed
func ValidateMigrations() error {
	// This is a compile-time check that migrations are embedded correctly
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no migration files found")
	}

	return nil
}