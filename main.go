package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

// performHealthCheck performs a health check and exits with appropriate code
func performHealthCheck() {
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
	if err := services.InitDB(logger); err != nil {
		fmt.Fprintf(os.Stderr, "Database health check failed: %v\n", err)
		os.Exit(1)
	}
	defer services.CloseDB(logger)

	// Test database connectivity with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := services.DB.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Database ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Health check passed")
	os.Exit(0)
}

func main() {
	// Parse command line flags
	healthCheck := flag.Bool("health-check", false, "Perform health check and exit")
	flag.Parse()

	// If health check flag is provided, perform health check and exit
	if *healthCheck {
		performHealthCheck()
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
		logger, err = zap.NewProduction()
	} else {
		gin.SetMode(gin.DebugMode)
		logger, err = zap.NewDevelopment()
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

	// Initialize email sender as a goroutine
	go services.InitMailer(logger)
	defer services.CloseMailer(logger)

	router := gin.New()

	router.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}
			// log request ID
			fields = append(fields, zap.String("request_id", requestid.Get(c)))

			// log trace and span ID
			if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
				fields = append(fields, zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()))
				fields = append(fields, zap.String("span_id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()))
			}

			// log client IP
			fields = append(fields, zap.String("client_ip", c.ClientIP()))

			// log request body
			var body []byte
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ = io.ReadAll(tee)
			c.Request.Body = io.NopCloser(&buf)
			fields = append(fields, zap.String("body", string(body)))

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

	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		auth.Use(middleware.StrictRateLimiter())
		{
			auth.POST("/register", routes.Register)                                      // POST /api/v1/auth/register
			auth.POST("/verify-otp", routes.VerifyOTP)                                   // POST /api/v1/auth/verify-otp
			auth.POST("/resend-otp", routes.ResendVerificationOTP)                       // POST /api/v1/auth/resend-otp
			auth.POST("/login", routes.Login)                                            // POST /api/v1/auth/login
			auth.POST("/refresh", routes.RefreshToken)                                   // POST /api/v1/auth/refresh
			auth.POST("/forgot-password", routes.ForgotPassword)                         // POST /api/v1/auth/forgot-password
			auth.POST("/reset-password", routes.ResetPassword)                           // POST /api/v1/auth/reset-password
			auth.GET("/profile", middleware.JWTAuthMiddleware(), routes.GetProfile)      // GET /api/v1/auth/profile
			auth.POST("/chicken-out", middleware.JWTAuthMiddleware(), routes.ChickenOut) // POST /api/v1/auth/chicken-out
		}

		applications := v1.Group("/applications")
		applications.Use(middleware.JWTAuthMiddleware()) // All application routes require authentication
		{
			applications.GET("", middleware.AdminOrAboveMiddleware(), routes.GetAllApplications) // GET /api/v1/applications (get all apps)
			applications.POST("", routes.CreateApplication)                                      // POST /api/v1/applications (create new app)
			applications.GET("/me", routes.GetMyApplications)                                    // GET /api/v1/applications/me (get user's apps)
			applications.PATCH("/:id/save", routes.SaveApplication)                              // PATCH /api/v1/applications/:id/save (save answers)
			applications.POST("/:id/submit", routes.SubmitApplication)                           // POST /api/v1/applications/:id/submit (submit app)
			applications.DELETE("/:id", routes.DeleteApplication)                                // DELETE /api/v1/applications/:id (delete app)
		}

		// Answers routes (protected)
		answers := v1.Group("/answers")
		answers.Use(middleware.JWTAuthMiddleware()) // All answer routes require authentication
		{
			answers.POST("", routes.PostAnswer)                                                        // POST /api/v1/answers (create/update answer)
			answers.DELETE("/:id", routes.DeleteAnswer)                                                // DELETE /api/v1/answers/:id (delete answer)
			answers.GET("/application/:id", routes.GetUserAnswersForApplication)                       // GET /api/v1/answers/application/:id (get user's answers for app)
			answers.GET("/user/:id", middleware.EvaluatorOrAboveMiddleware(), routes.GetAnswersByUser) // GET /api/v1/answers/user/:id (get all answers by user - evaluator+)
		}

		// Questions routes (public)
		questions := v1.Group("/questions")
		questions.Use(middleware.DefaultRateLimiter())
		questions.Use(middleware.JWTAuthMiddleware())
		{
			questions.GET("", routes.GetQuestions) // GET /api/v1/questions?dept=tech'
			questions.POST("", middleware.AdminOrAboveMiddleware(), routes.CreateQuestion)
			questions.DELETE("/:id", middleware.AdminOrAboveMiddleware(), routes.DeleteQuestion)
			questions.GET("/all", middleware.EvaluatorOrAboveMiddleware(), routes.GetAllQuestions)
			questions.GET("/:id", routes.GetQuestionByID)

		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.StrictRateLimiter())
		users.Use(middleware.JWTAuthMiddleware())
		users.Use(middleware.EvaluatorOrAboveMiddleware())
		{
			users.POST("", middleware.AdminOrAboveMiddleware(), routes.CreateUser)       // POST /api/v1/users (admin only)
			users.GET("", routes.GetAllUsers)                                            // GET /api/v1/users (evaluator+)
			users.GET("/:id", routes.GetUserByID)                                        // GET /api/v1/users/:id
			users.GET("/email/:email", routes.GetUserByEmail)                            // GET /api/v1/users/email/:email
			users.DELETE("/:id", middleware.AdminOrAboveMiddleware(), routes.DeleteUser) // DELETE /api/v1/users/:id (admin+)
		}

		// Super Admin routes (super admin only)
		superAdmin := v1.Group("/admin")
		superAdmin.Use(middleware.StrictRateLimiter())
		superAdmin.Use(middleware.JWTAuthMiddleware())
		superAdmin.Use(middleware.SuperAdminOnlyMiddleware())
		{
			// Reserved for super admin specific routes
			superAdmin.PUT("/users/:id/role", routes.UpdateUserRole) // PUT /api/v1/super-admin/users/:id/role
			superAdmin.PUT("/users/:id/verify", routes.VerifyUser)   // PUT /api/v1/super-admin/users/:id/verify
		}
	}

	port := utils.GetEnvWithDefault("PORT", "8080")
	if port[0] != ':' {
		port = ":" + port
	}
	log.Printf("Server starting on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
