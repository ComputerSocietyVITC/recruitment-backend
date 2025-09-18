package models

import (
	"time"

	"github.com/google/uuid"
)

// Review represents a review of an application by a reviewer
type Review struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ApplicationID uuid.UUID `json:"application_id" db:"application_id"`
	ReviewerID    uuid.UUID `json:"reviewer_id" db:"reviewer_id"`
	Department    string    `json:"department" db:"department"`
	Shortlisted   bool      `json:"shortlisted" db:"shortlisted"`
	Comments      *string   `json:"comments" db:"comments"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// CreateReviewRequest represents the request body for creating/updating a review
type CreateReviewRequest struct {
	ApplicationID uuid.UUID `json:"application_id" binding:"required"`
	Shortlisted   bool      `json:"shortlisted"`
	Comments      *string   `json:"comments"`
}

// UpdateReviewRequest represents the request body for updating a review
type UpdateReviewRequest struct {
	Shortlisted bool    `json:"shortlisted"`
	Comments    *string `json:"comments"`
}

// ReviewWithDetails represents a review with additional application and reviewer details
type ReviewWithDetails struct {
	Review
	ApplicationUserName  string `json:"application_user_name" db:"application_user_name"`
	ApplicationUserEmail string `json:"application_user_email" db:"application_user_email"`
	ReviewerName         string `json:"reviewer_name" db:"reviewer_name"`
	ReviewerEmail        string `json:"reviewer_email" db:"reviewer_email"`
}

// ApplicationWithReview represents an application with its review status
type ApplicationWithReview struct {
	Application
	ReviewID       *uuid.UUID `json:"review_id" db:"review_id"`
	Shortlisted    *bool      `json:"shortlisted" db:"shortlisted"`
	ReviewComments *string    `json:"review_comments" db:"review_comments"`
	ReviewerID     *uuid.UUID `json:"reviewer_id" db:"reviewer_id"`
	ReviewerName   *string    `json:"reviewer_name" db:"reviewer_name"`
	ReviewedAt     *time.Time `json:"reviewed_at" db:"reviewed_at"`
	// User details
	UserName  string `json:"user_name" db:"user_name"`
	UserEmail string `json:"user_email" db:"user_email"`
}

// ReviewStats represents statistics for a reviewer's department
type ReviewStats struct {
	Department        string `json:"department"`
	TotalApplications int    `json:"total_applications"`
	ReviewedCount     int    `json:"reviewed_count"`
	ShortlistedCount  int    `json:"shortlisted_count"`
	RejectedCount     int    `json:"rejected_count"`
	PendingCount      int    `json:"pending_count"`
}
