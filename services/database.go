package services

import (
	"context"
	"fmt"
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var DB *pgxpool.Pool

// InitDB initializes the database connection
func InitDB(logger *zap.Logger) error {
	// For now, using environment variables. In production, use proper config management
	dbHost := utils.GetEnvWithDefault("DB_HOST", "localhost")
	dbPort := utils.GetEnvWithDefault("DB_PORT", "5432")
	dbUser := utils.GetEnvWithDefault("DB_USER", "postgres")
	dbPassword := utils.GetEnvWithDefault("DB_PASSWORD", "password")
	dbName := utils.GetEnvWithDefault("DB_NAME", "recruitment_db")

	// Construct connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Create connection pool
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set connection pool settings
	maxConns := utils.GetEnvAsInt("DB_MAX_CONNS", 10)
	minConns := utils.GetEnvAsInt("DB_MIN_CONNS", 2)

	config.MaxConns = int32(maxConns)
	config.MinConns = int32(minConns)

	// Connect to database
	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	err = DB.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to database")
	return nil

}

// CloseDB closes the database connection
func CloseDB(logger *zap.Logger) {
	if DB != nil {
		DB.Close()
		logger.Info("Database connection closed")
	}
}
