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

func GetQuestions(c *gin.Context) {
	dept := c.Query("dept")
	if dept == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department parameter is required"})
		return
	}

	ctx := context.Background()
	rows, err := services.DB.Query(ctx, queries.GetQuestionsByDepartmentQuery, dept)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions", "details": err.Error()})
		return
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		err := rows.Scan(&q.ID, &q.Department, &q.Title, &q.Body, &q.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan question", "details": err.Error()})
			return
		}

		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while reading questions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, questions)
}

// GetAllQuestions returns all questions from all departments
func GetAllQuestions(c *gin.Context) {
	ctx := context.Background()
	rows, err := services.DB.Query(ctx, queries.GetAllQuestionsQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions", "details": err.Error()})
		return
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		err := rows.Scan(&q.ID, &q.Department, &q.Title, &q.Body, &q.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan question", "details": err.Error()})
			return
		}

		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while reading questions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Questions fetched successfully",
		"questions": questions,
		"count":     len(questions),
	})
}

// GetQuestionByID returns a specific question by its ID
func GetQuestionByID(c *gin.Context) {
	idParam := c.Param("id")
	questionID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID format"})
		return
	}

	ctx := context.Background()
	row := services.DB.QueryRow(ctx, queries.GetQuestionByIDQuery, questionID)

	var q models.Question
	err = row.Scan(&q.ID, &q.Department, &q.Title, &q.Body, &q.CreatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch question", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, q)
}

// GetQuestionByApplicationID returns all questions for a specific application based on the application's department
func GetQuestionByApplicationID(c *gin.Context) {
	idParam := c.Param("id")
	applicationID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	ctx := context.Background()
	rows, err := services.DB.Query(ctx, queries.GetQuestionByApplicationIDQuery, applicationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions for application", "details": err.Error()})
		return
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		err := rows.Scan(&q.ID, &q.Department, &q.Title, &q.Body, &q.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan question", "details": err.Error()})
			return
		}

		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while reading questions", "details": err.Error()})
		return
	}

	if len(questions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No questions found for this application or application not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Questions fetched successfully",
		"questions": questions,
		"count":     len(questions),
	})
}

// CreateQuestion creates a new question
func CreateQuestion(c *gin.Context) {
	var req models.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate department
	validDepartments := map[string]bool{
		"technical":    true,
		"marketing":    true,
		"management":   true,
		"social_media": true,
	}
	if !validDepartments[req.Department] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department. Must be one of: technical, marketing, management, social_media"})
		return
	}

	questionID := uuid.New()
	createdAt := time.Now()

	ctx := context.Background()
	row := services.DB.QueryRow(ctx, queries.CreateQuestionQuery, questionID, req.Department, req.Title, req.Body, createdAt)

	var q models.Question
	err := row.Scan(&q.ID, &q.Department, &q.Title, &q.Body, &q.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, q)
}

// DeleteQuestion deletes a question by its ID
func DeleteQuestion(c *gin.Context) {
	idParam := c.Param("id")
	questionID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID format"})
		return
	}

	ctx := context.Background()
	result, err := services.DB.Exec(ctx, queries.DeleteQuestionByIDQuery, questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question", "details": err.Error()})
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question deleted successfully"})
}
