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
)

// GetAllApplications handles GET /applications - fetches all applications
func GetAllApplications(c *gin.Context) {
	ctx := context.Background()
	
	rows, err := database.DB.Query(ctx, queries.GetAllApplicationsQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch applications",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var applications []models.ApplicationResponse
	for rows.Next() {
		var app models.Application
		
		// Updated scan to match actual database columns: id, user_id, department, submitted, created_at, updated_at
		err := rows.Scan(
			&app.ID, &app.UserID, &app.Department, &app.Submitted,
			&app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to scan application data",
				"details": err.Error(),
			})
			return
		}
		applications = append(applications, app.ToResponse())
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error occurred while reading applications",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Applications fetched successfully",
		"applications": applications,
		"count":        len(applications),
	})
}

// CreateApplication handles POST /applications - creates a new application
func CreateApplication(c *gin.Context) {
	var req models.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from JWT token (set by your auth middleware)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	application := models.Application{
		ID:         uuid.New(),
		UserID:     userID,
		Department: req.Department,
		Submitted:  false, // Default to false for new applications
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	ctx := context.Background()
	
	// Updated to match actual database columns: id, user_id, department, submitted, created_at, updated_at
	err := database.DB.QueryRow(ctx, queries.CreateApplicationQuery,
		application.ID, application.UserID, application.Department,
		application.Submitted, application.CreatedAt, application.UpdatedAt,
	).Scan(
		&application.ID, &application.UserID, &application.Department,
		&application.Submitted, &application.CreatedAt, &application.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create application",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Application created successfully",
		"application": application.ToResponse(),
	})
}

// GetMyApplications handles GET /applications/me - fetches current user's applications
// GetMyApplications handles GET /applications/me - fetches current user's applications
func GetMyApplications(c *gin.Context) {
	// Get user ID from JWT token
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	ctx := context.Background()
	rows, err := database.DB.Query(ctx, queries.GetUserApplicationsQuery, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch your applications",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var applications []models.ApplicationResponse
	for rows.Next() {
		var app models.Application
		
		// Updated scan to match actual database columns
		err := rows.Scan(
			&app.ID, &app.UserID, &app.Department, &app.Submitted,
			&app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to scan application data",
				"details": err.Error(),
			})
			return
		}
		applications = append(applications, app.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Your applications fetched successfully",
		"applications": applications,
		"count":        len(applications),
	})
} // ← CLOSE GetMyApplications function here

// SaveApplication handles PATCH /applications/:id/save - saves application answers
func SaveApplication(c *gin.Context) {
	// Get application ID from URL
	applicationIDStr := c.Param("id")
	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	// Parse request body
	var req models.SaveApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from JWT token
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDInterface.(uuid.UUID)

	ctx := context.Background()
	
	// Verify user owns this application
	var appUserID uuid.UUID
	err = database.DB.QueryRow(ctx, queries.CheckApplicationOwnershipQuery, applicationID).Scan(&appUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}
	if appUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Save each answer
	for _, answerReq := range req.Answers {
		answer := models.Answer{
			ID:            uuid.New(),
			ApplicationID: applicationID,
			QuestionID:    answerReq.QuestionID,
			Body:          answerReq.Body,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		_, err := database.DB.Exec(ctx, queries.SaveAnswersQuery,
			answer.ID, answer.ApplicationID,userID, answer.QuestionID,
			answer.Body, answer.CreatedAt, answer.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save answers",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "draft_saved"})
}