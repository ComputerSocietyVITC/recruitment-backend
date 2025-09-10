package routes

import (
	"github.com/ComputerSocietyVITC/recruitment-backend/middleware"
	"github.com/gin-gonic/gin"
)

// SetupV1Routes configures all API v1 routes
func SetupV1Routes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		auth.Use(middleware.StrictRateLimiter())
		{
			auth.POST("/register", Register)                                      // POST /api/v1/auth/register
			auth.POST("/verify-otp", VerifyOTP)                                   // POST /api/v1/auth/verify-otp
			auth.POST("/resend-otp", ResendVerificationOTP)                       // POST /api/v1/auth/resend-otp
			auth.POST("/login", Login)                                            // POST /api/v1/auth/login
			auth.POST("/refresh", RefreshToken)                                   // POST /api/v1/auth/refresh
			auth.POST("/forgot-password", ForgotPassword)                         // POST /api/v1/auth/forgot-password
			auth.POST("/reset-password", ResetPassword)                           // POST /api/v1/auth/reset-password
			auth.GET("/profile", middleware.JWTAuthMiddleware(), GetProfile)      // GET /api/v1/auth/profile
			auth.POST("/chicken-out", middleware.JWTAuthMiddleware(), ChickenOut) // POST /api/v1/auth/chicken-out
		}

		applications := v1.Group("/applications")
		applications.Use(middleware.JWTAuthMiddleware()) // All application routes require authentication
		{
			applications.GET("", middleware.AdminOrAboveMiddleware(), GetAllApplications) // GET /api/v1/applications (get all apps)
			applications.POST("", CreateApplication)                                      // POST /api/v1/applications (create new app)
			applications.GET("/me", GetMyApplications)                                    // GET /api/v1/applications/me (get user's apps)
			applications.PATCH("/:id/save", SaveApplication)                              // PATCH /api/v1/applications/:id/save (save answers)
			applications.POST("/:id/submit", SubmitApplication)                           // POST /api/v1/applications/:id/submit (submit app)
			applications.DELETE("/:id", DeleteApplication)                                // DELETE /api/v1/applications/:id (delete app)
		}

		// Answers routes (protected)
		answers := v1.Group("/answers")
		answers.Use(middleware.JWTAuthMiddleware()) // All answer routes require authentication
		{
			answers.POST("", PostAnswer)                                                        // POST /api/v1/answers (create/update answer)
			answers.DELETE("/:id", DeleteAnswer)                                                // DELETE /api/v1/answers/:id (delete answer)
			answers.GET("/application/:id", GetUserAnswersForApplication)                       // GET /api/v1/answers/application/:id (get user's answers for app)
			answers.GET("/user/:id", middleware.EvaluatorOrAboveMiddleware(), GetAnswersByUser) // GET /api/v1/answers/user/:id (get all answers by user - evaluator+)
		}

		// Questions routes (public)
		questions := v1.Group("/questions")
		questions.Use(middleware.DefaultRateLimiter())
		questions.Use(middleware.JWTAuthMiddleware())
		{
			questions.GET("", GetQuestions) // GET /api/v1/questions?dept=tech'
			questions.POST("", middleware.AdminOrAboveMiddleware(), CreateQuestion)
			questions.DELETE("/:id", middleware.AdminOrAboveMiddleware(), DeleteQuestion)
			questions.GET("/all", middleware.EvaluatorOrAboveMiddleware(), GetAllQuestions)
			questions.GET("/:id", GetQuestionByID)
		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.StrictRateLimiter())
		users.Use(middleware.JWTAuthMiddleware())
		users.Use(middleware.EvaluatorOrAboveMiddleware())
		{
			users.POST("", middleware.AdminOrAboveMiddleware(), CreateUser)       // POST /api/v1/users (admin only)
			users.GET("", GetAllUsers)                                            // GET /api/v1/users (evaluator+)
			users.GET("/:id", GetUserByID)                                        // GET /api/v1/users/:id
			users.GET("/email/:email", GetUserByEmail)                            // GET /api/v1/users/email/:email
			users.DELETE("/:id", middleware.AdminOrAboveMiddleware(), DeleteUser) // DELETE /api/v1/users/:id (admin+)
		}

		// Super Admin routes (super admin only)
		superAdmin := v1.Group("/admin")
		superAdmin.Use(middleware.StrictRateLimiter())
		superAdmin.Use(middleware.JWTAuthMiddleware())
		superAdmin.Use(middleware.SuperAdminOnlyMiddleware())
		{
			// Reserved for super admin specific routes
			superAdmin.PUT("/users/:id/role", UpdateUserRole) // PUT /api/v1/super-admin/users/:id/role
			superAdmin.PUT("/users/:id/verify", VerifyUser)   // PUT /api/v1/super-admin/users/:id/verify
		}
	}
}
