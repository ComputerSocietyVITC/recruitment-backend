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

		// Scan application fields
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

	application := models.Application{
		ID:         uuid.New(),
		UserID:     userID,
		Department: req.Department,
		Submitted:  false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	ctx := context.Background()

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
}

// SaveApplication handles PATCH /applications/:id/save - saves application answers
// SaveApplication handles PATCH /applications/:id/save - saves application answers
// SaveApplication handles PATCH /applications/:id/save - saves application answers
// SaveApplication handles PATCH /applications/:id/save - saves application answers
// SaveApplication handles PATCH /applications/:id/save - saves application answers
// SaveApplication handles PATCH /applications/:id/save - saves application answers
func SaveApplication(c *gin.Context) {
	ctx := context.Background()

	// Get application ID from URL
	applicationIDStr := c.Param("id")
	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	// Get user ID from JWT token
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDInterface.(uuid.UUID)

	// ✅ FIX: Use correct request model
	var req models.SaveApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// ✅ CRITICAL: Verify user owns this application
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

	// ✅ Use transaction for atomic operations
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	// Save each answer
	for _, answerReq := range req.Answers {
		var existingID uuid.UUID
		err := tx.QueryRow(ctx,
			"SELECT id FROM answers WHERE application_id = $1 AND question_id = $2",
			applicationID, answerReq.QuestionID,
		).Scan(&existingID)

		if err == nil {
			// UPDATE existing answer
			_, err = tx.Exec(ctx,
				"UPDATE answers SET body = $1, updated_at = $2 WHERE id = $3",
				answerReq.Body, time.Now(), existingID,
			)
		} else {
			// INSERT new answer
			_, err = tx.Exec(ctx, queries.SaveAnswersQuery,
				uuid.New(),
				applicationID,
				userID,
				answerReq.QuestionID,
				answerReq.Body,
				time.Now(),
				time.Now(),
			)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save answers",
				"details": err.Error(),
			})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save changes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "draft_saved",
		"saved_at":      time.Now().Format(time.RFC3339),
		"saved_answers": len(req.Answers),
	})
}
