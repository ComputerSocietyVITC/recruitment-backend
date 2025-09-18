package routes

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/ComputerSocietyVITC/recruitment-backend/models/queries"
	"github.com/ComputerSocietyVITC/recruitment-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetApplicationsForReview handles GET /reviewer/applications - gets applications for reviewer's department
func GetApplicationsForReview(c *gin.Context) {
	reviewerID, exists := c.Get("reviewerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Reviewer ID not found in context",
		})
		return
	}

	reviewerUUID, ok := reviewerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid reviewer ID format",
		})
		return
	}

	// First, get the reviewer's department
	ctx := context.Background()
	var department string
	err := services.DB.QueryRow(ctx, "SELECT department FROM users WHERE id = $1 AND role = 'reviewer'", reviewerUUID).Scan(&department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get reviewer department",
			"details": err.Error(),
		})
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 50
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	// Get applications with pagination
	rows, err := services.DB.Query(ctx, queries.GetApplicationsForReviewerWithPaginationQuery, department, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch applications",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var applications []models.ApplicationWithReview
	for rows.Next() {
		var app models.ApplicationWithReview
		err := rows.Scan(
			&app.ID, &app.UserID, &app.Department, &app.Submitted, &app.ChickenedOut,
			&app.CreatedAt, &app.UpdatedAt, &app.UserName, &app.UserEmail,
			&app.ReviewID, &app.Shortlisted, &app.ReviewComments,
			&app.ReviewerID, &app.ReviewerName, &app.ReviewedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to scan application",
				"details": err.Error(),
			})
			return
		}
		applications = append(applications, app)
	}

	// Get total count for pagination
	var totalCount int
	err = services.DB.QueryRow(ctx, queries.CountApplicationsForReviewerQuery, department).Scan(&totalCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to count applications",
			"details": err.Error(),
		})
		return
	}

	totalPages := (totalCount + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"applications": applications,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  totalCount,
			"limit":        limit,
		},
		"department": department,
	})
}

// CreateOrUpdateReview handles POST /reviewer/reviews - creates or updates a review
func CreateOrUpdateReview(c *gin.Context) {
	reviewerID, exists := c.Get("reviewerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Reviewer ID not found in context",
		})
		return
	}

	reviewerUUID, ok := reviewerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid reviewer ID format",
		})
		return
	}

	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := context.Background()

	// First, get the reviewer's department and verify the application belongs to that department
	var reviewerDept, appDept string
	var appSubmitted bool
	err := services.DB.QueryRow(ctx, `
		SELECT u.department, a.department, a.submitted 
		FROM users u, applications a 
		WHERE u.id = $1 AND u.role = 'reviewer' AND a.id = $2`,
		reviewerUUID, req.ApplicationID).Scan(&reviewerDept, &appDept, &appSubmitted)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to validate reviewer and application",
			"details": err.Error(),
		})
		return
	}

	if reviewerDept != appDept {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Cannot review applications from other departments",
		})
		return
	}

	if !appSubmitted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot review unsubmitted applications",
		})
		return
	}

	// Create or update the review
	var review models.Review
	err = services.DB.QueryRow(ctx, queries.CreateReviewQuery,
		req.ApplicationID, reviewerUUID, reviewerDept, req.Shortlisted, req.Comments).Scan(
		&review.ID, &review.ApplicationID, &review.ReviewerID, &review.Department,
		&review.Shortlisted, &review.Comments, &review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create/update review",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Review saved successfully",
		"review":  review,
	})
}

// GetReviewStats handles GET /reviewer/stats - gets review statistics for reviewer's department
func GetReviewStats(c *gin.Context) {
	reviewerID, exists := c.Get("reviewerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Reviewer ID not found in context",
		})
		return
	}

	reviewerUUID, ok := reviewerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid reviewer ID format",
		})
		return
	}

	// Get reviewer's department
	ctx := context.Background()
	var department string
	err := services.DB.QueryRow(ctx, "SELECT department FROM users WHERE id = $1 AND role = 'reviewer'", reviewerUUID).Scan(&department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get reviewer department",
			"details": err.Error(),
		})
		return
	}

	// Get review statistics
	var stats models.ReviewStats
	err = services.DB.QueryRow(ctx, queries.GetReviewStatsQuery, department).Scan(
		&stats.Department, &stats.TotalApplications, &stats.ReviewedCount,
		&stats.ShortlistedCount, &stats.RejectedCount, &stats.PendingCount,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get review statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetApplicationDetails handles GET /reviewer/applications/:id - gets detailed application info for review
func GetApplicationDetails(c *gin.Context) {
	reviewerID, exists := c.Get("reviewerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Reviewer ID not found in context",
		})
		return
	}

	reviewerUUID, ok := reviewerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid reviewer ID format",
		})
		return
	}

	applicationIDStr := c.Param("id")
	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid application ID format",
		})
		return
	}

	ctx := context.Background()

	// Get reviewer's department
	var reviewerDept string
	err = services.DB.QueryRow(ctx, "SELECT department FROM users WHERE id = $1 AND role = 'reviewer'", reviewerUUID).Scan(&reviewerDept)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get reviewer department",
			"details": err.Error(),
		})
		return
	}

	// Get application with review details
	var app models.ApplicationWithReview
	err = services.DB.QueryRow(ctx, queries.GetApplicationWithReviewQuery, applicationID).Scan(
		&app.ID, &app.UserID, &app.Department, &app.Submitted, &app.ChickenedOut,
		&app.CreatedAt, &app.UpdatedAt, &app.UserName, &app.UserEmail,
		&app.ReviewID, &app.Shortlisted, &app.ReviewComments,
		&app.ReviewerID, &app.ReviewerName, &app.ReviewedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get application details",
			"details": err.Error(),
		})
		return
	}

	// Check if application belongs to reviewer's department
	if app.Department != reviewerDept {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Cannot access applications from other departments",
		})
		return
	}

	// Get application answers
	answersRows, err := services.DB.Query(ctx, queries.GetAnswersWithQuestionsForApplicationQuery, applicationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get application answers",
			"details": err.Error(),
		})
		return
	}
	defer answersRows.Close()

	var answers []models.AnswerWithQuestion
	for answersRows.Next() {
		var answer models.AnswerWithQuestion
		var answerUserID uuid.UUID
		err := answersRows.Scan(
			&answer.ID, &answer.ApplicationID, &answerUserID, &answer.QuestionID,
			&answer.Body, &answer.CreatedAt, &answer.UpdatedAt,
			&answer.QuestionTitle, &answer.QuestionBody, &answer.Department,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to scan answer",
				"details": err.Error(),
			})
			return
		}
		answers = append(answers, answer)
	}

	c.JSON(http.StatusOK, gin.H{
		"application": app,
		"answers":     answers,
	})
}
