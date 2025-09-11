package routes

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/ComputerSocietyVITC/recruitment-backend/services"
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
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

	// Roles other than applicant are not allowed to register
	if req.Role != models.RoleApplicant {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role specified",
		})
		return
	}

	// Split email to get domain
	emailParts := strings.Split(req.Email, "@")
	if len(emailParts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}
	domain := emailParts[1]

	// Check if email domain is allowed
	allowedDomains := utils.GetEnvAsSlice("ALLOWED_EMAIL_DOMAINS", ",", []string{"vit.ac.in", "vitstudent.ac.in"})
	domainAllowed := false
	for _, d := range allowedDomains {
		if strings.EqualFold(strings.TrimSpace(d), domain) {
			domainAllowed = true
			break
		}
	}

	if !domainAllowed {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email domain is not allowed",
		})
		return
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

	emailVerifyDuration := utils.GetEnvAsDuration("EMAIL_VERIFICATION_OTP_DURATION", 10*time.Minute)
	tokenExpiresAt := time.Now().Add(emailVerifyDuration)

	user := models.User{
		FullName:            req.FullName,
		Email:               req.Email,
		RegNum:              req.RegNum,
		PhoneNumber:         req.PhoneNumber,
		HashedPassword:      string(hashedPassword),
		Role:                req.Role,
		Verified:            false,
		ResetToken:          &otp,
		ResetTokenExpiresAt: &tokenExpiresAt,
	}

	ctx := context.Background()
	err = services.DB.QueryRow(ctx, queries.CreateUserQuery,
		user.FullName, user.Email, user.RegNum, user.PhoneNumber, user.Verified, user.ResetToken,
		user.ResetTokenExpiresAt, user.HashedPassword, user.Role,
	).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified,
		&user.Role, &user.ChickenedOut, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	emailTemplate := utils.GetEmailVerificationTemplate(otp, emailVerifyDuration)
	m := gomail.NewMessage()
	m.SetHeader("From", utils.GetEnvWithDefault("EMAIL_FROM", "recruitments@no-reply.ieeecsvitc.com"))
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", emailTemplate.Subject)
	m.SetBody("text/html", emailTemplate.Body)

	services.Mailer <- m

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully. Please check email for the verification code.",
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
	err := services.DB.QueryRow(ctx, queries.GetUserByEmailQuery, req.Email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified, &user.ResetToken,
		&user.ResetTokenExpiresAt, &user.HashedPassword, &user.Role, &user.ChickenedOut, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	if user.ResetToken == nil || user.ResetTokenExpiresAt == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User or OTP not found",
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
		err := services.DB.QueryRow(ctx, queries.UpdateUserVerificationStatusQuery, user.ID, true).Scan(
			&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified, &user.Role, &user.CreatedAt, &user.UpdatedAt,
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

	c.JSON(http.StatusOK, gin.H{
		"message": "User verified successfully",
	})
}

// ResendVerificationOTP handles POST /auth/resend-otp - resends verification OTP for unverified users
func ResendVerificationOTP(c *gin.Context) {
	var req models.ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := context.Background()
	var user models.User
	err := services.DB.QueryRow(ctx, queries.GetUserByEmailQuery, req.Email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified, &user.ResetToken,
		&user.ResetTokenExpiresAt, &user.HashedPassword, &user.Role, &user.ChickenedOut, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Check if user is already verified
	if user.Verified {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User is already verified",
		})
		return
	}

	// Generate a new OTP
	otp, err := utils.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate OTP",
		})
		return
	}

	emailVerifyDuration := utils.GetEnvAsDuration("EMAIL_VERIFICATION_OTP_DURATION", 10*time.Minute)
	tokenExpiresAt := time.Now().Add(emailVerifyDuration)

	// Update user's reset token and expiration time
	err = services.DB.QueryRow(ctx, queries.UpdateUserResetTokenQuery, user.ID, otp, tokenExpiresAt).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified, &user.ResetToken,
		&user.ResetTokenExpiresAt, &user.HashedPassword, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update verification token",
			"details": err.Error(),
		})
		return
	}

	// Send verification email
	emailTemplate := utils.GetResendVerificationTemplate(otp, emailVerifyDuration)
	m := gomail.NewMessage()
	m.SetHeader("From", utils.GetEnvWithDefault("EMAIL_FROM", "recruitments@no-reply.ieeecsvitc.com"))
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", emailTemplate.Subject)
	m.SetBody("text/html", emailTemplate.Body)

	services.Mailer <- m

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification OTP has been resent. Please check your email.",
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

	err := services.DB.QueryRow(ctx, queries.GetUserByEmailQuery, req.Email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified, &user.ResetToken,
		&user.ResetTokenExpiresAt, &user.HashedPassword, &user.Role, &user.ChickenedOut, &user.CreatedAt, &user.UpdatedAt,
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in token",
		})
		return
	}

	ctx := context.Background()
	var user models.User

	err := services.DB.QueryRow(ctx, queries.GetUserByIDQuery, userID).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified,
		&user.Role, &user.ChickenedOut, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// ForgotPassword handles POST /auth/forgot-password - sends password reset email
func ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := context.Background()
	var user models.User
	err := services.DB.QueryRow(ctx, queries.GetUserByEmailQuery, req.Email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified, &user.ResetToken,
		&user.ResetTokenExpiresAt, &user.HashedPassword, &user.Role, &user.ChickenedOut, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		// For security, don't reveal if email exists or not
		c.JSON(http.StatusOK, gin.H{
			"message": "If the email exists, a password reset code has been sent.",
		})
		return
	}

	// Check if user is verified
	if !user.Verified {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User email is not verified. Please verify your email first.",
		})
		return
	}

	// Generate a reset token (reusing OTP generation for consistency)
	resetToken, err := utils.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate reset token",
		})
		return
	}

	passwordResetDuration := utils.GetEnvAsDuration("PASSWORD_RESET_OTP_DURATION", 30*time.Minute)
	tokenExpiresAt := time.Now().Add(passwordResetDuration)

	// Update user's reset token and expiration time
	err = services.DB.QueryRow(ctx, queries.UpdateUserResetTokenQuery, user.ID, resetToken, tokenExpiresAt).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified, &user.ResetToken,
		&user.ResetTokenExpiresAt, &user.HashedPassword, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate reset token",
			"details": err.Error(),
		})
		return
	}

	// Send password reset email
	emailTemplate := utils.GetPasswordResetTemplate(resetToken, passwordResetDuration)
	m := gomail.NewMessage()
	m.SetHeader("From", utils.GetEnvWithDefault("EMAIL_FROM", "recruitments@no-reply.ieeecsvitc.com"))
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", emailTemplate.Subject)
	m.SetBody("text/html", emailTemplate.Body)

	services.Mailer <- m

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset code has been sent.",
	})
}

// ResetPassword handles POST /auth/reset-password - resets password using token
func ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := context.Background()
	var user models.User
	err := services.DB.QueryRow(ctx, queries.GetUserByEmailQuery, req.Email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified, &user.ResetToken,
		&user.ResetTokenExpiresAt, &user.HashedPassword, &user.Role, &user.ChickenedOut, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Check if user is verified
	if !user.Verified {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User email is not verified. Please verify your email first.",
		})
		return
	}

	// Validate reset token
	if user.ResetToken == nil || user.ResetTokenExpiresAt == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No password reset request found",
		})
		return
	}

	if user.ResetTokenExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Reset token has expired",
		})
		return
	}

	if *user.ResetToken != req.ResetToken {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reset token",
		})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process new password",
		})
		return
	}

	// Update password and clear reset token
	err = services.DB.QueryRow(ctx, queries.UpdateUserPasswordQuery, user.ID, string(hashedPassword)).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update password",
			"details": err.Error(),
		})
		return
	}

	// Send confirmation email
	emailTemplate := utils.GetPasswordResetSuccessTemplate()
	m := gomail.NewMessage()
	m.SetHeader("From", utils.GetEnvWithDefault("EMAIL_FROM", "recruitments@no-reply.ieeecsvitc.com"))
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", emailTemplate.Subject)
	m.SetBody("text/html", emailTemplate.Body)

	services.Mailer <- m

	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}

// ChickenOut handles POST /auth/chicken-out - marks the user as chickened out
func ChickenOut(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in token",
		})
		return
	}

	ctx := context.Background()
	var user models.User

	err := services.DB.QueryRow(ctx, queries.UpdateUserChickenedOutStatusQuery, userID, true).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber, &user.Verified,
		&user.Role, &user.ChickenedOut, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "You have successfully chickened out.",
		"user":    user.ToResponse(),
	})
}
