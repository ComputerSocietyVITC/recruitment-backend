package main

import (
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
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables from .env file (only in development)
	if utils.GetEnvWithDefault("ENV", "development") == "development" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found or error loading it, using system environment variables")
		}
	}

	if err := utils.ValidateRequiredEnvVars(); err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}

	var logger *zap.Logger
	var err error

	if utils.GetEnvWithDefault("GIN_MODE", "debug") == "release" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize JWT
	utils.InitJWT()

	// Initialize database connection
	if err := services.InitDB(logger); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer services.CloseDB(logger)

	// Initialize email sender as a goroutine
	go services.InitMailer(logger)
	defer services.CloseMailer(logger)

	// Set Gin to release mode for production logging
	if utils.GetEnvWithDefault("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))
	router.Use(requestid.New())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Configure properly for production
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Set trusted proxies for production
	if utils.GetEnvWithDefault("ENV", "development") != "development" {
		// Development: Trust local networks
		router.SetTrustedProxies([]string{
			"127.0.0.1",
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
		})
	} else {
		// Production: Only trust specific load balancers/proxies
		trustedProxies := utils.GetEnvWithDefault("TRUSTED_PROXIES", "")
		if trustedProxies != "" {
			proxies := strings.Split(trustedProxies, ",")
			router.SetTrustedProxies(proxies)
		} else {
			// Disable trusted proxies if not configured
			router.SetTrustedProxies(nil)
		}
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		auth.Use(middleware.StrictRateLimiter())
		{
			auth.POST("/register", routes.Register)                // POST /api/v1/auth/register
			auth.POST("/verify-otp", routes.VerifyOTP)             // POST /api/v1/auth/verify-otp
			auth.POST("/resend-otp", routes.ResendVerificationOTP) // POST /api/v1/auth/resend-otp
			auth.POST("/login", routes.Login)                      // POST /api/v1/auth/login
			auth.POST("/refresh", routes.RefreshToken)             // POST /api/v1/auth/refresh
			auth.POST("/forgot-password", routes.ForgotPassword)   // POST /api/v1/auth/forgot-password
			auth.POST("/reset-password", routes.ResetPassword)     // POST /api/v1/auth/reset-password
		}

		applications := v1.Group("/applications")
		applications.Use(middleware.DefaultRateLimiter())
		applications.Use(middleware.JWTAuthMiddleware()) // All application routes require authentication
		{
			applications.POST("", routes.CreateApplication)                                         // POST /api/v1/applications (create new app)
			applications.GET("/me", routes.GetMyApplications)                                       // GET /api/v1/applications/me (get user's apps)
			applications.GET("/:id", routes.GetApplicationByID)                                     // GET /api/v1/applications/:id (get app by ID)
			applications.PATCH("/:id/save", middleware.StrictRateLimiter(), routes.SaveApplication) // PATCH /api/v1/applications/:id/save (save answers)
			applications.POST("/:id/submit", routes.SubmitApplication)                              // POST /api/v1/applications/:id/submit (submit app)
			applications.DELETE("/:id", routes.DeleteApplication)                                   // DELETE /api/v1/applications/:id (delete app)
			applications.POST("/:id/chicken-out", routes.ChickenOut)                                // POST /api/v1/applications/:id/chicken-out
		}

		// Answers routes (protected)
		answers := v1.Group("/answers")
		applications.Use(middleware.DefaultRateLimiter())
		answers.Use(middleware.JWTAuthMiddleware()) // All answer routes require authentication
		{
			answers.POST("", routes.PostAnswer)                                  // POST /api/v1/answers (create/update answer)
			answers.GET("/application/:id", routes.GetUserAnswersForApplication) // GET /api/v1/answers/application/:id (get user's answers for app)
		}

		// Questions routes (protected)
		questions := v1.Group("/questions")
		questions.Use(middleware.DefaultRateLimiter())
		questions.Use(middleware.JWTAuthMiddleware())
		{
			questions.GET("/application/:id", routes.GetQuestionByApplicationID)
			questions.GET("/:id", routes.GetQuestionByID)
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
