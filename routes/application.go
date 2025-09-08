package routes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/ComputerSocietyVITC/recruitment-backend/services"
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// GetAllApplications fetches all applications
func GetAllApplications(c *gin.Context) {
	ctx := context.Background()

	rows, err := services.DB.Query(ctx, queries.GetAllApplicationsQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch applications",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var applications []models.Application
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
		applications = append(applications, app)
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

	// Check if user has reached the maximum number of applications
	maxApplications := utils.GetEnvAsInt("MAXIMUM_APPLICATIONS_PER_USER", 2) // Default to 3 if not set
	var currentCount int
	err := services.DB.QueryRow(ctx, queries.CountUserApplicationsQuery, userID).Scan(&currentCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check application count",
			"details": err.Error(),
		})
		return
	}

	if currentCount >= maxApplications {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Maximum number of applications reached",
			"details": map[string]interface{}{
				"current_applications": currentCount,
				"maximum_allowed":      maxApplications,
				"message":              "You have reached the maximum number of applications allowed per user",
			},
		})
		return
	}

	err = services.DB.QueryRow(ctx, queries.CreateApplicationQuery,
		application.ID, application.UserID, application.Department,
		application.Submitted, application.CreatedAt, application.UpdatedAt,
	).Scan(
		&application.ID, &application.UserID, &application.Department,
		&application.Submitted, &application.CreatedAt, &application.UpdatedAt,
	)

	if err != nil {
		// Check if this is a unique constraint violation
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			// Check if the constraint name contains user_id and department
			if strings.Contains(pgErr.ConstraintName, "user_id") && strings.Contains(pgErr.ConstraintName, "department") {
				c.JSON(http.StatusConflict, gin.H{
					"error": "You have already created an application for this department",
					"details": map[string]interface{}{
						"department": req.Department,
						"message":    "Only one application per department is allowed per user",
					},
				})
				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create application",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Application created successfully",
		"application": application,
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
	rows, err := services.DB.Query(ctx, queries.GetUserApplicationsQuery, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch your applications",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var applications []models.Application
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
		applications = append(applications, app)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Your applications fetched successfully",
		"applications": applications,
		"count":        len(applications),
	})
}

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
	err = services.DB.QueryRow(ctx, queries.CheckApplicationOwnershipQuery, applicationID).Scan(&appUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}
	if appUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Upsert each answer
	for _, answerReq := range req.Answers {
		// Validate that question department matches application department
		var appDepartment, questionDepartment string
		fmt.Println("Validating question", answerReq.QuestionID, "for application", applicationID)
		err = services.DB.QueryRow(ctx, queries.ValidateQuestionApplicationDepartmentQuery, applicationID, answerReq.QuestionID).Scan(&appDepartment, &questionDepartment)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Application or question not found",
				"details": map[string]any{
					"question_id": answerReq.QuestionID,
					"error":       err.Error(),
				},
			})
			return
		}
		if appDepartment != questionDepartment {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Question department does not match application department",
				"details": map[string]any{
					"question_id":            answerReq.QuestionID,
					"application_department": appDepartment,
					"question_department":    questionDepartment,
				},
			})
			return
		}

		answer := models.Answer{
			ID:            uuid.New(),
			ApplicationID: applicationID,
			QuestionID:    answerReq.QuestionID,
			Body:          answerReq.Body,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		_, err := services.DB.Exec(ctx, queries.UpsertAnswerQuery,
			answer.ID,
			answer.ApplicationID,
			userID,
			answer.QuestionID,
			answer.Body,
			answer.CreatedAt,
			answer.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save answers",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answers saved successfully",
	})
}

// SubmitApplication handles POST /applications/:id/submit - submits an application
func SubmitApplication(c *gin.Context) {
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

	ctx := context.Background()

	// Submit the application
	var application models.Application
	err = services.DB.QueryRow(ctx, queries.SubmitApplicationQuery,
		applicationID, time.Now(), userID).Scan(
		&application.ID, &application.UserID, &application.Department,
		&application.Submitted, &application.CreatedAt, &application.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Application not found or access denied",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Application submitted successfully",
		"application": application,
	})
}

// DeleteApplication handles DELETE /applications/:id - deletes an application
func DeleteApplication(c *gin.Context) {
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

	ctx := context.Background()

	// Delete the application (this will cascade delete all associated answers)
	result, err := services.DB.Exec(ctx, queries.DeleteApplicationQuery, applicationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete application",
			"details": err.Error(),
		})
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Application deleted successfully",
	})
}
