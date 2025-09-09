package routes

import (
	"context"
	"net/http"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/ComputerSocietyVITC/recruitment-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser handles POST /users - creates a new user
func CreateUser(c *gin.Context) {
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

	userRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User role not found in context",
		})
		c.Abort()
		return
	}
	_, ok := userRole.(models.UserRole)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user role format",
		})
		c.Abort()
		return
	}
	// Note: Only applicant role is supported now

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process password",
		})
		return
	}

	user := models.User{
		FullName:            req.FullName,
		Email:               req.Email,
		RegNum:              req.RegNum,
		Verified:            false,
		ResetToken:          nil,
		ResetTokenExpiresAt: nil,
		HashedPassword:      string(hashedPassword),
		Role:                req.Role,
	}

	ctx := context.Background()
	err = services.DB.QueryRow(ctx, queries.CreateUserQuery,
		user.FullName, user.Email, user.RegNum, user.Verified,
		user.ResetToken, user.ResetTokenExpiresAt,
		user.HashedPassword, user.Role,
	).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.Verified,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		// Check if it's a unique constraint violation (duplicate email)
		if err.Error() == "UNIQUE constraint failed" ||
			err.Error() == "duplicate key value violates unique constraint" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user.ToResponse(),
	})
}

// GetAllUsers handles GET /users - fetches all users
func GetAllUsers(c *gin.Context) {
	ctx := context.Background()
	rows, err := services.DB.Query(ctx, queries.GetAllUsersQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch users",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.Verified,
			&user.Role, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to scan user data",
				"details": err.Error(),
			})
			return
		}
		users = append(users, user.ToResponse())
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error occurred while reading users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Users fetched successfully",
		"users":   users,
		"count":   len(users),
	})
}

// GetUserByID handles GET /users/:id - fetches a single user by ID
func GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	ctx := context.Background()
	var user models.User
	err = services.DB.QueryRow(ctx, queries.GetUserByIDQuery, userID).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.Verified,
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
			"error":   "Failed to fetch user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User fetched successfully",
		"user":    user.ToResponse(),
	})
}

// DeleteUser handles DELETE /users/:id - deletes a user by ID
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// Check if the user exists first
	ctx := context.Background()
	var existingUser models.User
	err = services.DB.QueryRow(ctx, queries.GetUserByIDQuery, userID).Scan(
		&existingUser.ID, &existingUser.FullName, &existingUser.Email, &existingUser.RegNum,
		&existingUser.Verified, &existingUser.Role, &existingUser.CreatedAt, &existingUser.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch user",
			"details": err.Error(),
		})
		return
	}

	// Get the current user's role to check permissions
	userRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User role not found in context",
		})
		return
	}

	_, ok := userRole.(models.UserRole)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user role format",
		})
		return
	}

	// Note: Only applicant role exists now, no special deletion restrictions needed

	// Prevent users from deleting themselves
	userIDFromContext, exists := c.Get("userID")
	if exists {
		if currentUserID, ok := userIDFromContext.(uuid.UUID); ok && currentUserID == userID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot delete your own account",
			})
			return
		}
	}

	// Delete the user
	result, err := services.DB.Exec(ctx, queries.DeleteUserQuery, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user",
			"details": err.Error(),
		})
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// GetUserByEmail handles GET /users/email/:email - fetches a user by email
func GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email parameter is required",
		})
		return
	}

	ctx := context.Background()
	var user models.User
	err := services.DB.QueryRow(ctx, queries.GetUserByEmailPublicQuery, email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.RegNum, &user.Verified,
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
			"error":   "Failed to fetch user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User fetched successfully",
		"user":    user.ToResponse(),
	})
}
