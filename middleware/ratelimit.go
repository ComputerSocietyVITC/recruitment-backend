package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimit creates a rate limiter middleware with custom limits
func RateLimit(rate string) gin.HandlerFunc {
	// Parse rate string manually
	rateLimit, err := parseRate(rate)
	if err != nil {
		panic(fmt.Sprintf("Invalid rate format: %s. Use format like '5-M' for 5 per minute", rate))
	}

	// Create memory store
	store := memory.NewStore()

	// Create limiter instance
	instance := limiter.New(store, rateLimit)

	return func(c *gin.Context) {
		// Get unique key for rate limiting
		key := getKey(c)

		// Check rate limit
		context, err := instance.Get(c, key)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// Check if rate limit exceeded
		if context.Reached {
			retryAfter := time.Unix(context.Reset, 0).Format(time.RFC3339)
			c.JSON(429, gin.H{
				"error":       "Too many requests. Please slow down.",
				"retry_after": retryAfter,
				"limit":       context.Limit,
				"remaining":   context.Remaining,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// parseRate manually parses rate strings like "5-M" (5 requests per minute)
func parseRate(rate string) (limiter.Rate, error) {
	var period time.Duration
	var limit int64

	// Parse the rate string
	_, err := fmt.Sscanf(rate, "%d-%s", &limit, &rate)
	if err != nil {
		return limiter.Rate{}, err
	}

	// Determine period based on suffix
	switch rate {
	case "S":
		period = time.Second
	case "M":
		period = time.Minute
	case "H":
		period = time.Hour
	case "D":
		period = time.Hour * 24
	default:
		return limiter.Rate{}, fmt.Errorf("invalid period suffix: %s", rate)
	}

	return limiter.Rate{
		Period: period,
		Limit:  limit,
	}, nil
}

// getKey returns a unique identifier for rate limiting
func getKey(c *gin.Context) string {
	// Use user ID if authenticated
	if userID, exists := c.Get("userID"); exists {
		return "user:" + userID.(uuid.UUID).String()
	}
	// Fall back to IP address for non-authenticated requests
	return "ip:" + c.ClientIP()
}
