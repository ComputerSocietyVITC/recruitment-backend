package models

import (
	"time"

	"github.com/google/uuid"
)

// Application struct maps to your actual database columns
type Application struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Department   string    `json:"department"`
	Submitted    bool      `json:"submitted"`
	ChickenedOut bool      `json:"chickened_out"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateApplicationRequest is what we receive from client
type CreateApplicationRequest struct {
	Department string `json:"department" binding:"required"`
}
