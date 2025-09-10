package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"go.uber.org/zap"
)

// performHealthCheck performs a health check and exits with appropriate code
func PerformHealthCheck() {
	// Load environment variables for database connection
	if err := utils.LoadEnvironment(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load environment: %v\n", err)
		os.Exit(1)
	}

	// Initialize a simple logger for health check
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Initialize database connection
	if err := InitDB(logger); err != nil {
		fmt.Fprintf(os.Stderr, "Database health check failed: %v\n", err)
		os.Exit(1)
	}
	defer CloseDB(logger)

	// Test database connectivity with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := DB.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Database ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Health check passed")
	os.Exit(0)
}
