package routes

import (
	"context"
	"net/http"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/ComputerSocietyVITC/recruitment-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UpdateUserRoleRequest represents the request body for updating user role
type UpdateUserRoleRequest struct {
	Role models.UserRole `json:"role" binding:"required"`
}

// VerifyUserRequest represents the request body for verifying a user
type VerifyUserRequest struct {
	Verified bool `json:"verified" binding:"required"`
}

// UpdateUserRole handles PUT /admin/users/:id/role - updates a user's role
func UpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate role
	if req.Role != models.RoleApplicant && req.Role != models.RoleEvaluator &&
		req.Role != models.RoleAdmin && req.Role != models.RoleSuperAdmin {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role specified",
		})
		return
	}

	ctx := context.Background()
	var user models.User

	err = services.DB.QueryRow(ctx, queries.UpdateUserRoleQuery, userID, req.Role).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber,
		&user.Verified, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user role",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"user":    user.ToResponse(),
	})
}

// VerifyUser handles PUT /admin/users/:id/verify - updates a user's verification status
func VerifyUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req VerifyUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := context.Background()
	var user models.User

	err = services.DB.QueryRow(ctx, queries.UpdateUserVerificationStatusQuery, userID, req.Verified).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.PhoneNumber,
		&user.Verified, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user verification status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User verification status updated successfully",
		"user":    user.ToResponse(),
	})
}
