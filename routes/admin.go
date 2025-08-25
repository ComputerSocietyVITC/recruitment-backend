package routes

import (
	"context"
	"net/http"

	"github.com/ComputerSocietyVITC/recruitment-backend/database"
	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UpdateUserRoleRequest represents the request body for updating user role
type UpdateUserRoleRequest struct {
	Role models.UserRole `json:"role" binding:"required"`
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

	err = database.DB.QueryRow(ctx, queries.UpdateUserRoleQuery, userID, req.Role).Scan(
		&user.ID, &user.FullName, &user.Email, &user.PhoneNumber,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
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
