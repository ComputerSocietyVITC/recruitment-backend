package middleware

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// JWTAuthMiddleware validates JWT tokens and sets user information in context
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		log.Println("Authorization Header Received:", authHeader)
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
		log.Println("Extracted Token:", tokenString)
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
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
		c.Abort()
	}
}

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	Rate limiter.Rate
}

// getClientIP extracts the real client IP, considering reverse proxy headers
func getClientIP(c *gin.Context) string {
	// Check for real IP headers in order of preference
	headers := []string{
		"X-Real-IP",
		"X-Forwarded-For",
		"CF-Connecting-IP", // Cloudflare
		"True-Client-IP",   // Akamai and Cloudflare
	}

	for _, header := range headers {
		ip := c.GetHeader(header)
		if ip != "" {
			// For X-Forwarded-For, take the first IP (original client)
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				ip = strings.TrimSpace(ips[0])
			}

			// Validate IP format
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// RateLimiterMiddleware creates a rate limiting middleware
func RateLimiterMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	// Create an in-memory store for rate limiting
	store := memory.NewStore()

	// Create the limiter instance
	instance := limiter.New(store, config.Rate)

	return func(c *gin.Context) {
		// Get the real client IP
		clientIP := getClientIP(c)

		// Check rate limit
		context, err := instance.Get(c.Request.Context(), clientIP)
		if err != nil {
			log.Printf("Rate limiter error: %v", err)
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
// 250 requests per minute per IP
func DefaultRateLimiter() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: time.Minute,
		Limit:  250,
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

// ReviewerAuthMiddleware ensures the user is a reviewer and fetches their department
func ReviewerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found in context",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(models.UserRole)
		if !ok || role != models.RoleReviewer {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: Reviewer role required",
			})
			c.Abort()
			return
		}

		// Get user ID from context
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in context",
			})
			c.Abort()
			return
		}

		// Fetch user's department from database
		// Note: We'll need to import the services package for DB access
		// For now, we'll just pass the user ID and let the route handlers fetch the department
		c.Set("reviewerID", userID)
		c.Next()
	}
}
