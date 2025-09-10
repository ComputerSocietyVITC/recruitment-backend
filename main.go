package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/middleware"
	"github.com/ComputerSocietyVITC/recruitment-backend/routes"
	"github.com/ComputerSocietyVITC/recruitment-backend/services"
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Parse command line flags
	healthCheck := flag.Bool("health-check", false, "Perform health check and exit")
	flag.Parse()

	// If health check flag is provided, perform health check and exit
	if *healthCheck {
		services.PerformHealthCheck()
		return
	}

	if err := utils.LoadEnvironment(); err != nil {
		log.Fatalf("Failed to load environment: %v", err)
	}

	if err := utils.ValidateRequiredEnvVars(); err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}

	var logger *zap.Logger
	var err error

	if utils.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
		// Use enhanced production logger for better Docker/container logging
		config := utils.LoggerConfig{
			Level:       string(utils.GetLogLevel()),
			Environment: "production",
			Service:     "recruitment-backend",
			Version:     utils.GetEnvWithDefault("APP_VERSION", "dev"),
		}
		logger, err = utils.NewProductionLogger(config)
	} else {
		gin.SetMode(gin.DebugMode)
		logger, err = utils.NewDevelopmentLogger()
	}
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize JWT
	utils.InitJWT()

	// Initialize middleware logger
	middleware.InitLogger(logger)

	// Initialize database connection
	if err := services.InitDB(logger); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer services.CloseDB(logger)

	// Run database migrations
	if err := services.RunMigrations(logger); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Create admin user if not exists (based on environment variables)
	if err := services.CreateAdminUserIfNotExists(logger); err != nil {
		logger.Fatal("Failed to create admin user", zap.Error(err))
	}

	// Initialize email sender as a goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Mailer goroutine panicked", zap.Any("panic", r))
			}
		}()
		services.InitMailer(logger)
	}()
	defer services.CloseMailer(logger)

	router := gin.New()

	router.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}

			// Core request information
			fields = append(fields, zap.String("request_id", requestid.Get(c)))
			fields = append(fields, zap.String("client_ip", c.ClientIP()))
			fields = append(fields, zap.String("user_agent", c.GetHeader("User-Agent")))
			fields = append(fields, zap.String("referer", c.GetHeader("Referer")))
			fields = append(fields, zap.String("bearer", c.GetHeader("Authorization")))

			// User context if available
			if userID, exists := c.Get("user_id"); exists {
				fields = append(fields, zap.String("user_id", userID.(string)))
			}
			if userRole, exists := c.Get("user_role"); exists {
				fields = append(fields, zap.String("user_role", userRole.(string)))
			}

			// OpenTelemetry trace information
			if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
				fields = append(fields, zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()))
				fields = append(fields, zap.String("span_id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()))
			}

			// Request body for non-GET requests (be selective about logging bodies)
			if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
				contentType := c.GetHeader("Content-Type")
				if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "application/x-www-form-urlencoded") {
					var buf bytes.Buffer
					tee := io.TeeReader(c.Request.Body, &buf)
					body, _ := io.ReadAll(tee)
					c.Request.Body = io.NopCloser(&buf)

					// Sanitize sensitive information from logs
					bodyStr := string(body)
					if len(bodyStr) > 10000 { // Limit body size in logs
						bodyStr = bodyStr[:10000] + "... [truncated]"
					}
					// Remove sensitive fields (you can expand this list)
					bodyStr = utils.SanitizeLogBody(bodyStr)
					fields = append(fields, zap.String("request_body", bodyStr))
				}
			}

			// Request size
			if c.Request.ContentLength > 0 {
				fields = append(fields, zap.Int64("request_size", c.Request.ContentLength))
			}

			return fields
		}),
	}))
	router.Use(ginzap.RecoveryWithZap(logger, true))
	router.Use(requestid.New())
	defaultProxies := []string{
		"127.0.0.1",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     utils.GetEnvAsSlice("CORS_ALLOWED_ORIGINS", ",", []string{"*"}),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
	}))

	router.SetTrustedProxies(utils.GetEnvAsSlice("TRUSTED_PROXIES", ",", defaultProxies))

	// Health check endpoint for container orchestration
	router.GET("/health", func(c *gin.Context) {
		// Check database connectivity
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if err := services.DB.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":     "unhealthy",
				"error":      "database connection failed",
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"request_id": requestid.Get(c),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": requestid.Get(c),
			"checks": gin.H{
				"database": "ok",
			},
		})
	})

	// Setup API v1 routes
	routes.SetupV1Routes(router)

	port := utils.GetEnvWithDefault("PORT", "8080")
	if port[0] != ':' {
		port = ":" + port
	}
	log.Printf("Server starting on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
