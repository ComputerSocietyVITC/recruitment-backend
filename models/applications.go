package models

import (
	"time"

	"github.com/google/uuid"
)

// Application struct maps to your actual database columns
type Application struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Department string    `json:"department"`
	Submitted  bool      `json:"submitted"` // This is the actual column name in your database
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ApplicationResponse is what we send to the client
type ApplicationResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Department string    `json:"department"`
	Submitted  bool      `json:"submitted"` // Matches the database column
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateApplicationRequest is what we receive from client
type CreateApplicationRequest struct {
	Department string `json:"department" binding:"required"`
}

// ToResponse converts Application database model to API response
func (a *Application) ToResponse() ApplicationResponse {
	return ApplicationResponse{
		ID:         a.ID.String(),
		UserID:     a.UserID.String(),
		Department: a.Department,
		Submitted:  a.Submitted, // Changed from Status to Submitted
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}
