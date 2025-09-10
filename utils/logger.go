package utils

import (
	"os"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level       string `json:"level"`
	Environment string `json:"environment"`
	Service     string `json:"service"`
	Version     string `json:"version"`
}

// NewProductionLogger creates a production-ready logger optimized for Docker/container environments
func NewProductionLogger(config LoggerConfig) (*zap.Logger, error) {
	// Create a custom encoder config for structured JSON logging
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "function",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create JSON encoder for structured logging
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Parse log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create core that outputs to stdout (Docker/container best practice)
	core := zapcore.NewCore(
		encoder,
		zapcore.Lock(os.Stdout), // Lock for concurrent safety
		level,
	)

	// Add global fields for service identification
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// Add service metadata as global fields
	logger = logger.With(
		zap.String("service", config.Service),
		zap.String("version", config.Version),
		zap.String("environment", config.Environment),
		zap.String("hostname", getHostname()),
		zap.Int("pid", os.Getpid()),
	)

	return logger, nil
}

// NewDevelopmentLogger creates a development logger with pretty printing
func NewDevelopmentLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, err
	}

	return logger.With(
		zap.String("service", "recruitment-backend"),
		zap.String("environment", "development"),
		zap.String("hostname", getHostname()),
	), nil
}

// LogLevel represents available log levels
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// GetLogLevel returns the log level from environment or default
func GetLogLevel() LogLevel {
	level := strings.ToLower(GetEnvWithDefault("LOG_LEVEL", "info"))
	switch level {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	case "fatal":
		return LogLevelFatal
	default:
		return LogLevelInfo
	}
}

// getHostname returns the hostname or container ID for identification
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// ContextLogger provides common logging patterns with structured fields
type ContextLogger struct {
	logger *zap.Logger
}

// NewContextLogger creates a new context logger
func NewContextLogger(logger *zap.Logger) *ContextLogger {
	return &ContextLogger{logger: logger}
}

// WithRequestContext adds request-specific fields
func (cl *ContextLogger) WithRequestContext(requestID, userID, clientIP, method, path string) *zap.Logger {
	return cl.logger.With(
		zap.String("request_id", requestID),
		zap.String("user_id", userID),
		zap.String("client_ip", clientIP),
		zap.String("http_method", method),
		zap.String("http_path", path),
		zap.Time("request_time", time.Now().UTC()),
	)
}

// WithDatabaseContext adds database operation fields
func (cl *ContextLogger) WithDatabaseContext(operation, table string, duration time.Duration) *zap.Logger {
	return cl.logger.With(
		zap.String("db_operation", operation),
		zap.String("db_table", table),
		zap.Duration("db_duration", duration),
	)
}

// WithEmailContext adds email operation fields
func (cl *ContextLogger) WithEmailContext(to, subject, template string) *zap.Logger {
	return cl.logger.With(
		zap.String("email_to", to),
		zap.String("email_subject", subject),
		zap.String("email_template", template),
	)
}

// WithAuthContext adds authentication operation fields
func (cl *ContextLogger) WithAuthContext(userID, action, role string) *zap.Logger {
	return cl.logger.With(
		zap.String("auth_user_id", userID),
		zap.String("auth_action", action),
		zap.String("auth_role", role),
	)
}

// Business logic logging helpers
func LogBusinessEvent(logger *zap.Logger, event, entity, entityID string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("event_type", "business"),
		zap.String("business_event", event),
		zap.String("entity_type", entity),
		zap.String("entity_id", entityID),
		zap.Time("event_time", time.Now().UTC()),
	}, fields...)

	logger.Info("Business event", allFields...)
}

// Security event logging
func LogSecurityEvent(logger *zap.Logger, event, userID, clientIP, reason string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("event_type", "security"),
		zap.String("security_event", event),
		zap.String("user_id", userID),
		zap.String("client_ip", clientIP),
		zap.String("reason", reason),
		zap.Time("event_time", time.Now().UTC()),
	}, fields...)

	logger.Warn("Security event", allFields...)
}

// Performance event logging
func LogPerformanceEvent(logger *zap.Logger, operation string, duration time.Duration, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("event_type", "performance"),
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Time("event_time", time.Now().UTC()),
	}, fields...)

	level := zapcore.InfoLevel
	if duration > 5*time.Second {
		level = zapcore.WarnLevel
	}
	if duration > 10*time.Second {
		level = zapcore.ErrorLevel
	}

	logger.Log(level, "Performance event", allFields...)
}

// sanitizeLogBody removes sensitive information from request body logs
func SanitizeLogBody(body string) string {
	// Define patterns for sensitive information
	sensitivePatterns := []string{
		`"password":\s*"[^"]*"`,
		`"token":\s*"[^"]*"`,
		`"secret":\s*"[^"]*"`,
		`"key":\s*"[^"]*"`,
		`"otp":\s*"[^"]*"`,
	}

	result := body
	for _, pattern := range sensitivePatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		result = re.ReplaceAllString(result, strings.Replace(pattern, `"[^"]*"`, `"[REDACTED]"`, 1))
	}

	return result
}
