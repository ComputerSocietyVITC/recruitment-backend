package middleware

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
)

var logger *zap.Logger

// InitLogger initializes the middleware logger
func InitLogger(l *zap.Logger) {
	logger = l
}

// JWTAuthMiddleware validates JWT tokens and sets user information in context
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if logger != nil {
			logger.Info("Authorization Header Received", zap.String("header", authHeader))
		}
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]
		if logger != nil {
			logger.Info("Extracted Token", zap.String("token", tokenString))
		}
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		userRole := models.UserRole(claims.Role)

		c.Set("userID", userID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", userRole)
		c.Set("jwtClaims", claims)
		c.Next()
	}
}

// RoleBasedAuthMiddleware checks if the user has the required role(s)
func RoleBasedAuthMiddleware(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found in context",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(models.UserRole)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user role format",
			})
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		if slices.Contains(allowedRoles, role) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
		c.Abort()
	}
}

// SuperAdminOnlyMiddleware allows only super admin access
func SuperAdminOnlyMiddleware() gin.HandlerFunc {
	return RoleBasedAuthMiddleware(models.RoleSuperAdmin)
}

// AdminOrAboveMiddleware allows admin and super admin access
func AdminOrAboveMiddleware() gin.HandlerFunc {
	return RoleBasedAuthMiddleware(models.RoleAdmin, models.RoleSuperAdmin)
}

// EvaluatorOrAboveMiddleware allows evaluator, admin, and super admin access
func EvaluatorOrAboveMiddleware() gin.HandlerFunc {
	return RoleBasedAuthMiddleware(models.RoleEvaluator, models.RoleAdmin, models.RoleSuperAdmin)
}

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	Rate limiter.Rate
}

// RateLimiterMiddleware creates a rate limiting middleware
func RateLimiterMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	// Create an in-memory store for rate limiting
	store := memory.NewStore()

	// Create the limiter instance
	instance := limiter.New(store, config.Rate)

	return func(c *gin.Context) {
		// Get the real client IP
		clientIP := c.ClientIP()

		// Check rate limit
		context, err := instance.Get(c.Request.Context(), clientIP)
		if err != nil {
			if logger != nil {
				logger.Error("Rate limiter error", zap.Error(err))
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		// Check if rate limit is exceeded
		if context.Reached {
			// Calculate retry after duration
			resetTime := time.Unix(context.Reset, 0)
			retryAfter := time.Until(resetTime).Seconds()
			if retryAfter < 0 {
				retryAfter = 0
			}

			c.Header("Retry-After", strconv.FormatFloat(retryAfter, 'f', 0, 64))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// DefaultRateLimiter creates a rate limiter with sensible defaults
// 100 requests per minute per IP
func DefaultRateLimiter() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: time.Minute,
		Limit:  100,
	}

	config := RateLimiterConfig{
		Rate: rate,
	}

	return RateLimiterMiddleware(config)
}

// StrictRateLimiter creates a more restrictive rate limiter
// 20 requests per minute per IP (useful for auth endpoints)
func StrictRateLimiter() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: time.Minute,
		Limit:  20,
	}

	config := RateLimiterConfig{
		Rate: rate,
	}

	return RateLimiterMiddleware(config)
}

// CustomRateLimiter allows a rate limiter with custom limits
func CustomRateLimiter(requestsPerPeriod int64, period time.Duration, trustedProxies []string) gin.HandlerFunc {
	rate := limiter.Rate{
		Period: period,
		Limit:  requestsPerPeriod,
	}

	config := RateLimiterConfig{
		Rate: rate,
	}

	return RateLimiterMiddleware(config)
}
