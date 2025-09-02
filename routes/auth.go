package routes

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/database"
	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Register handles POST /auth/register - creates a new user account
func Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = models.RoleApplicant
	}

	// Validate role
	if req.Role != models.RoleApplicant && req.Role != models.RoleEvaluator &&
		req.Role != models.RoleAdmin && req.Role != models.RoleSuperAdmin {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role specified",
		})
		return
	}

	// Only super admins can create admin or super admin accounts
	if req.Role == models.RoleAdmin || req.Role == models.RoleSuperAdmin {
		userRole, exists := c.Get("user_role")
		if !exists || userRole != models.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only super administrators can create admin accounts",
			})
			return
		}
	}

	// Only administrators can create evaluator accounts
	if req.Role == models.RoleEvaluator {
		userRole, exists := c.Get("user_role")
		if !exists || (userRole != models.RoleAdmin && userRole != models.RoleSuperAdmin) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Only administrators can create evaluator accounts",
			})
			return
		}
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process password",
		})
		return
	}

	// Generate an OTP for verification
	otp, err := utils.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate OTP",
		})
		return
	}

	tokenExpiresAt := time.Now().Add(10 * time.Minute)

	user := models.User{
		ID:                  uuid.New(),
		FullName:            req.FullName,
		Email:               req.Email,
		PhoneNumber:         req.PhoneNumber,
		HashedPassword:      string(hashedPassword),
		Role:                req.Role,
		Verified:            false,
		ResetToken:          &otp,
		ResetTokenExpiresAt: &tokenExpiresAt,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	ctx := context.Background()
	err = database.DB.QueryRow(ctx, queries.CreateUserWithVerificationQuery,
		user.ID, user.FullName, user.Email, user.PhoneNumber, user.Verified, user.ResetToken,
		user.ResetTokenExpiresAt, user.HashedPassword, user.Role, user.CreatedAt, user.UpdatedAt,
	).Scan(
		&user.ID, &user.FullName, &user.Email, &user.PhoneNumber,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	subject := "Thank you for applying to IEEE CompSoc. Please verify your email address"
	body := "Your OTP is: <strong>" + otp + "</strong>. It is valid for 10 minutes."
	if err := utils.GetMailerInstance().Send([]string{user.Email}, subject, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send verification email",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully. Please check your email for the verification code.",
		"user":    user.ToResponse(),
	})
}

func VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := context.Background()
	var user models.User
	err := database.DB.QueryRow(ctx, queries.GetUserByEmailQuery, req.Email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.PhoneNumber, &user.Verified, &user.ResetToken,
		&user.ResetTokenExpiresAt, &user.HashedPassword, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	if user.ResetTokenExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "OTP has expired",
		})
		return
	}

	if *user.ResetToken == req.Code {
		err := database.DB.QueryRow(ctx, queries.UpdateUserVerificationStatusQuery, user.ID, true).Scan(
			&user.ID, &user.FullName, &user.Email, &user.PhoneNumber, &user.Verified, &user.Role, &user.CreatedAt, &user.UpdatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update user verification status",
				"details": err.Error(),
			})
			return
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid OTP",
		})
		return
	}

	token, err := utils.GenerateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate authentication token",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User verified successfully",
		"user":    user.ToResponse(),
		"token":   token,
	})
}

// Login handles POST /auth/login - authenticates a user
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := context.Background()
	var user models.User

	err := database.DB.QueryRow(ctx, queries.GetUserByEmailQuery, req.Email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.PhoneNumber, &user.Verified, &user.ResetToken,
		&user.ResetTokenExpiresAt, &user.HashedPassword, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	if !user.Verified {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "User email is not verified",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	token, err := utils.GenerateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication token",
		})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	})
}

// RefreshToken handles POST /auth/refresh - refreshes a JWT token
func RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization header is required",
		})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid authorization header format",
		})
		return
	}

	newToken, err := utils.RefreshJWT(tokenParts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
}

// GetProfile handles GET /auth/profile - gets current user profile
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in token",
		})
		return
	}

	ctx := context.Background()
	var user models.User

	err := database.DB.QueryRow(ctx, queries.GetUserByIDQuery, userID).Scan(
		&user.ID, &user.FullName, &user.Email, &user.PhoneNumber,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}
