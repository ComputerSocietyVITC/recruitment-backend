package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

// CreateAdminUserIfNotExists creates an admin user if it doesn't exist
// This function checks for ADMIN_EMAIL and ADMIN_PASSWORD environment variables
// and creates an admin user only if both are provided and no user with that email exists
func CreateAdminUserIfNotExists(logger *zap.Logger) error {
	adminEmail := utils.GetEnvWithDefault("ADMIN_EMAIL", "")
	adminPassword := utils.GetEnvWithDefault("ADMIN_PASSWORD", "")

	// If either admin email or password is not set, skip admin creation
	if adminEmail == "" || adminPassword == "" {
		logger.Info("Admin user creation skipped - ADMIN_EMAIL or ADMIN_PASSWORD not provided")
		return nil
	}

	logger.Info("Checking if admin user exists", zap.String("email", adminEmail))

	ctx := context.Background()

	// Check if user with admin email already exists
	var existingUser models.User
	err := DB.QueryRow(ctx, queries.GetUserByEmailQuery, adminEmail).Scan(
		&existingUser.ID, &existingUser.FullName, &existingUser.Email, &existingUser.RegNum,
		&existingUser.PhoneNumber, &existingUser.Verified, &existingUser.ResetToken,
		&existingUser.ResetTokenExpiresAt, &existingUser.HashedPassword,
		&existingUser.Role, &existingUser.ChickenedOut, &existingUser.CreatedAt,
		&existingUser.UpdatedAt,
	)

	if err == nil {
		// User exists, check if they are already an admin
		if existingUser.Role == models.RoleAdmin || existingUser.Role == models.RoleSuperAdmin {
			logger.Info("Admin user already exists with admin privileges", zap.String("email", adminEmail))
			return nil
		} else {
			logger.Info("User exists but is not an admin - skipping admin creation", zap.String("email", adminEmail))
			return nil
		}
	} else if err != pgx.ErrNoRows {
		// An actual error occurred (not just "no rows found")
		return fmt.Errorf("failed to check for existing admin user: %w", err)
	}

	// User doesn't exist, create admin user
	logger.Info("Creating admin user", zap.String("email", adminEmail))

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Get admin name from environment or use default
	adminName := utils.GetEnvWithDefault("ADMIN_NAME", "System Administrator")
	adminPhone := utils.GetEnvWithDefault("ADMIN_PHONE", "+911000000000")

	adminUser := models.User{
		ID:                  uuid.New(),
		FullName:            adminName,
		Email:               adminEmail,
		RegNum:              "ADMIN001",
		PhoneNumber:         adminPhone,
		HashedPassword:      string(hashedPassword),
		Verified:            true, // Admin users are verified by default
		ChickenedOut:        false,
		ResetToken:          nil,
		ResetTokenExpiresAt: nil,
		Role:                models.RoleAdmin,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Create the admin user
	err = DB.QueryRow(ctx, queries.CreateUserQuery,
		adminUser.FullName, adminUser.Email, adminUser.RegNum, adminUser.PhoneNumber,
		adminUser.Verified, adminUser.ResetToken, adminUser.ResetTokenExpiresAt,
		adminUser.HashedPassword, adminUser.Role,
	).Scan(
		&adminUser.ID, &adminUser.FullName, &adminUser.Email, &adminUser.RegNum,
		&adminUser.PhoneNumber, &adminUser.Verified, &adminUser.Role,
		&adminUser.ChickenedOut, &adminUser.CreatedAt, &adminUser.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	logger.Info("Admin user created successfully",
		zap.String("email", adminUser.Email),
		zap.String("id", adminUser.ID.String()),
		zap.String("role", string(adminUser.Role)))

	return nil
}

// CloseDB closes the database connection
func CloseDB(logger *zap.Logger) {
	if DB != nil {
		DB.Close()
		logger.Info("Database connection closed")
	}
}
