package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// ValidateRequiredEnvVars checks if all required environment variables are set
func ValidateRequiredEnvVars() error {
	required := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"JWT_SECRET",
	}

	var missing []string
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
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
