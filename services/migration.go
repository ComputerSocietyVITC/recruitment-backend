package services

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// RunMigrations executes database migrations
func RunMigrations(logger *zap.Logger) error {
	// Get the database connection from our pool
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Create a standard database/sql connection for the migrate library
	connStr := DB.Config().ConnString()
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open SQL connection for migrations: %w", err)
	}
	defer sqlDB.Close()

	// Create postgres driver instance
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get the migrations directory path
	migrationsPath, err := filepath.Abs("models/migrations")
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}

	logger.Info("Using migrations directory", zap.String("path", migrationsPath))

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	logger.Info("Running database migrations...")

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info("No new migrations to apply")
	} else {
		logger.Info("Database migrations completed successfully")
	}

	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(logger *zap.Logger) (uint, bool, error) {
	if DB == nil {
		return 0, false, fmt.Errorf("database connection not initialized")
	}

	connStr := DB.Config().ConnString()
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return 0, false, fmt.Errorf("failed to open SQL connection: %w", err)
	}
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationsPath, err := filepath.Abs("models/migrations")
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migrations path: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
