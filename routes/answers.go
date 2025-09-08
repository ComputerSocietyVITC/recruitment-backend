package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/ComputerSocietyVITC/recruitment-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PostAnswer handles POST /answers - creates or updates an answer
func PostAnswer(c *gin.Context) {
	var req models.PostAnswerRequest
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
	err := services.DB.QueryRow(ctx, queries.CheckApplicationOwnershipQuery, req.ApplicationID).Scan(&appUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}
	if appUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Validate that question department matches application department
	var appDepartment, questionDepartment string
	err = services.DB.QueryRow(ctx, queries.ValidateQuestionApplicationDepartmentQuery, req.ApplicationID, req.QuestionID).Scan(&appDepartment, &questionDepartment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application or question not found"})
		return
	}
	if appDepartment != questionDepartment {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Question department does not match application department",
			"details": map[string]string{
				"application_department": appDepartment,
				"question_department":    questionDepartment,
			},
		})
		return
	}

	// Upsert the answer
	var answer models.Answer
	err = services.DB.QueryRow(ctx, queries.UpsertAnswerQuery,
		uuid.New(),
		req.ApplicationID,
		userID,
		req.QuestionID,
		req.Body,
		time.Now(),
		time.Now(),
	).Scan(
		&answer.ID, &answer.ApplicationID, &userID, &answer.QuestionID,
		&answer.Body, &answer.CreatedAt, &answer.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save answer",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer saved successfully",
		"answer":  answer,
	})
}

// DeleteAnswer handles DELETE /answers/:id - deletes an answer
func DeleteAnswer(c *gin.Context) {
	// Get answer ID from URL
	answerIDStr := c.Param("id")
	answerID, err := uuid.Parse(answerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid answer ID"})
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

	// First, get the answer to verify ownership
	var answer models.Answer
	var answerUserID uuid.UUID
	err = services.DB.QueryRow(ctx, queries.GetAnswerByIDQuery, answerID).Scan(
		&answer.ID, &answer.ApplicationID, &answerUserID, &answer.QuestionID,
		&answer.Body, &answer.CreatedAt, &answer.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Answer not found"})
		return
	}

	// Verify ownership
	if answerUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete the answer
	result, err := services.DB.Exec(ctx, queries.DeleteAnswerQuery, answerID, answer.ApplicationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete answer",
			"details": err.Error(),
		})
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Answer not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer deleted successfully",
	})
}

// GetUserAnswersForApplication handles GET /answers/application/:id - gets current user's answers for an application
func GetUserAnswersForApplication(c *gin.Context) {
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

	// Get answers for the application (with ownership verification)
	rows, err := services.DB.Query(ctx, queries.GetUserAnswersForApplicationQuery, applicationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch answers",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var answers []models.Answer
	for rows.Next() {
		var answer models.Answer
		var answerUserID uuid.UUID

		err := rows.Scan(
			&answer.ID, &answer.ApplicationID, &answerUserID, &answer.QuestionID,
			&answer.Body, &answer.CreatedAt, &answer.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to scan answer data",
				"details": err.Error(),
			})
			return
		}

		answers = append(answers, answer)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error occurred while reading answers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Answers fetched successfully",
		"answers":        answers,
		"count":          len(answers),
		"application_id": applicationID,
	})
}

// GetAnswersByUser handles GET /answers/user/:id - gets all answers written by a specific user (admin/evaluator only)
func GetAnswersByUser(c *gin.Context) {
	// Get user ID from URL
	targetUserIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx := context.Background()

	// Get all answers by the specified user
	rows, err := services.DB.Query(ctx, queries.GetAnswersByUserQuery, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch answers",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var answers []models.Answer
	for rows.Next() {
		var answer models.Answer

		err := rows.Scan(
			&answer.ID, &answer.ApplicationID, &targetUserID, &answer.QuestionID,
			&answer.Body, &answer.CreatedAt, &answer.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to scan answer data",
				"details": err.Error(),
			})
			return
		}

		answers = append(answers, answer)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error occurred while reading answers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User answers fetched successfully",
		"answers": answers,
		"count":   len(answers),
		"user_id": targetUserID,
	})
}
