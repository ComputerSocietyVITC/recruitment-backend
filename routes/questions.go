package routes

import (
	"context"
	"net/http"

	"github.com/ComputerSocietyVITC/recruitment-backend/database"
	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/gin-gonic/gin"
)

func GetQuestions(c *gin.Context) {
	dept := c.Query("dept")
	if dept == "" {
		dept = "technical"
	}

	ctx := context.Background()
	rows, err := database.DB.Query(ctx, queries.GetQuestionsByDepartmentQuery, dept)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions", "details": err.Error()})
		return
	}
	defer rows.Close()

	var questions []map[string]interface{}
	for rows.Next() {
		var q models.Question
		err := rows.Scan(&q.ID, &q.Department, &q.Body, &q.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan question", "details": err.Error()})
			return
		}
		questions = append(questions, map[string]interface{}{
			"id":   q.ID,
			"body": q.Body,
			"type": "text", // always text for now
		})
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while reading questions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, questions)
}
