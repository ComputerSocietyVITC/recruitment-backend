package utils

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// LoadEnvironment loads environment variables following best practices
// It checks for environment-specific .env files first, then falls back to .env
func LoadEnvironment() error {
	// Get the current environment
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}
	if env == "" {
		env = "development" // Default to development
	}

	// Try to load environment-specific file first
	envFiles := []string{
		fmt.Sprintf(".env.%s.local", env),
		fmt.Sprintf(".env.%s", env),
		".env.local",
		".env",
	}

	var loadedFile string
	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			if err := godotenv.Load(file); err != nil {
				log.Printf("Warning: Error loading %s: %v", file, err)
				continue
			}
			loadedFile = file
			break
		}
	}

	// In production, don't load .env files unless explicitly told to
	if env == "production" && loadedFile != "" {
		log.Printf("Info: Loaded environment file %s in production mode", loadedFile)
	} else if env != "production" && loadedFile != "" {
		log.Printf("Info: Loaded environment file %s", loadedFile)
	} else if env != "production" && loadedFile == "" {
		log.Println("Info: No .env file found, using system environment variables")
	}

	// Set ENV if it wasn't already set
	if os.Getenv("ENV") == "" {
		os.Setenv("ENV", env)
	}

	return nil
}

// ValidateRequiredEnvVars checks if all required environment variables are set
// Different requirements based on environment
func ValidateRequiredEnvVars() error {
	env := GetEnvWithDefault("ENV", "development")

	// Core required variables for all environments
	coreRequired := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_NAME",
		"JWT_SECRET",
	}

	// Additional required variables for production
	prodRequired := []string{
		"DB_PASSWORD", // Should always be set, but we'll be more strict in prod
		"EMAIL_FROM",
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_USER",
		"SMTP_PASSWORD",
	}

	required := coreRequired
	if env == "production" {
		required = append(required, prodRequired...)

		// Additional production validations
		if len(GetEnvWithDefault("JWT_SECRET", "")) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
		}

		// Validate CORS configuration in production
		corsOrigins := GetEnvAsSlice("CORS_ALLOWED_ORIGINS", ",", []string{})
		if len(corsOrigins) == 0 {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS must be set in production")
		}
		if slices.Contains(corsOrigins, "*") {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS should not contain '*' in production for security")
		}
	} else {
		// In development, warn about missing production variables but don't fail
		for _, env := range prodRequired {
			if os.Getenv(env) == "" {
				log.Printf("Warning: %s is not set (required for production)", env)
			}
		}
	}

	var missing []string
	for _, envVar := range required {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

// GetEnvWithDefault returns environment variable value or default if not set
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsInt returns environment variable as integer or default if not set/invalid
func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// GetEnvAsBool returns environment variable as boolean or default if not set/invalid
// Accepts: true, false, 1, 0, yes, no (case insensitive)
func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "TRUE", "1", "yes", "YES":
			return true
		case "false", "FALSE", "0", "no", "NO":
			return false
		default:
			log.Printf("Warning: Invalid boolean value for %s: %s, using default: %v", key, value, defaultValue)
		}
	}
	return defaultValue
}

func GetEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Warning: Invalid duration format for %s: %s, using default: %v", key, value, defaultValue)
	}
	return defaultValue
}

// GetEnvironment returns the current environment (development, testing, production)
func GetEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}
	if env == "" {
		env = "development"
	}
	return env
}

// IsProduction returns true if running in production environment
func IsProduction() bool {
	return GetEnvironment() == "production"
}

// IsTesting returns true if running in testing environment
func IsTesting() bool {
	return GetEnvironment() == "testing"
}

// IsDevelopment returns true if running in development environment
func IsDevelopment() bool {
	env := GetEnvironment()
	return env == "development" || env == "dev"
}

// GetEnvAsSlice returns environment variable as a slice of strings split by delimiter
// Trims whitespace from each element and filters out empty strings
func GetEnvAsSlice(key, delimiter string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parts := strings.Split(value, delimiter)
	var result []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultValue
	}

	return result
}
