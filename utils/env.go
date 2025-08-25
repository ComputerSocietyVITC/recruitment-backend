package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
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

// GetJWTExpiryDuration returns JWT expiry duration from environment variable
// Default is 24 hours. Accepts values like "1h", "30m", "24h", "7d"
func GetJWTExpiryDuration() time.Duration {
	defaultDuration := 24 * time.Hour
	expiryStr := GetEnvWithDefault("JWT_EXPIRY_DURATION", "24h")

	duration, err := time.ParseDuration(expiryStr)
	if err != nil {
		log.Printf("Warning: Invalid duration format for JWT_EXPIRY_DURATION: %s, using default: %v", expiryStr, defaultDuration)
		return defaultDuration
	}

	// Ensure minimum expiry of 1 minute for security
	if duration < time.Minute {
		log.Printf("Warning: JWT_EXPIRY_DURATION too short (%v), using minimum: 1m", duration)
		return time.Minute
	}

	return duration
}
