package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/database"
	"github.com/ComputerSocietyVITC/recruitment-backend/middleware"
	"github.com/ComputerSocietyVITC/recruitment-backend/routes"
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
	if err := database.InitDB(); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.CloseDB()

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

	// Set trusted proxies (update during deployment)
	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "pong",
			"request_id": requestid.Get(c),
		})
	})

	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", routes.Register)                                 // POST /api/v1/auth/register
			auth.POST("/login", routes.Login)                                       // POST /api/v1/auth/login
			auth.POST("/refresh", routes.RefreshToken)                              // POST /api/v1/auth/refresh
			auth.GET("/profile", middleware.JWTAuthMiddleware(), routes.GetProfile) // GET /api/v1/auth/profile
		}

		// Questions routes (public)
		questions := v1.Group("/questions")
		{
			questions.GET("", routes.GetQuestions) // GET /api/v1/questions?dept=tech
		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.JWTAuthMiddleware()) // All user routes require authentication
		{
			users.POST("", middleware.AdminOrAboveMiddleware(), routes.CreateUser)     // POST /api/v1/users (admin only)
			users.GET("", middleware.EvaluatorOrAboveMiddleware(), routes.GetAllUsers) // GET /api/v1/users (evaluator+)
			users.GET("/:id", routes.GetUserByID)                                      // GET /api/v1/users/:id
		}

		// Admin routes (admin and super admin only)
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuthMiddleware())
		admin.Use(middleware.AdminOrAboveMiddleware())
		{
			// Other admin routes can be added here
		}

		// Super Admin routes (super admin only)
		superAdmin := v1.Group("/super-admin")
		superAdmin.Use(middleware.JWTAuthMiddleware())
		superAdmin.Use(middleware.SuperAdminOnlyMiddleware())
		{
			// Reserved for super admin specific routes
			superAdmin.POST("/admin-users", routes.Register)         // POST /api/v1/super-admin/admin-users
			superAdmin.PUT("/users/:id/role", routes.UpdateUserRole) // PUT /api/v1/super-admin/users/:id/role
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
