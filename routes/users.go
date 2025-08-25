package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/database"
	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process password",
		})
		return
	}

	user := models.User{
		ID:             uuid.New(),
		FullName:       req.FullName,
		Email:          req.Email,
		PhoneNumber:    req.PhoneNumber,
		HashedPassword: string(hashedPassword),
		Role:           req.Role,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	ctx := context.Background()
	err = database.DB.QueryRow(ctx, queries.CreateUserQuery,
		user.ID, user.FullName, user.Email, user.PhoneNumber,
		user.HashedPassword, user.Role, user.CreatedAt, user.UpdatedAt,
	).Scan(
		&user.ID, &user.FullName, &user.Email, &user.PhoneNumber,
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
	rows, err := database.DB.Query(ctx, queries.GetAllUsersQuery)
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
			&user.ID, &user.FullName, &user.Email, &user.PhoneNumber,
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
	err = database.DB.QueryRow(ctx, queries.GetUserByIDQuery, userID).Scan(
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
