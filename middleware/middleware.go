package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		c.Set("jwt_claims", claims)
		c.Next()
	}
}

// RoleBasedAuthMiddleware checks if the user has the required role(s)
func RoleBasedAuthMiddleware(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
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
