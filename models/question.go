package models

import (
	"time"

	"github.com/google/uuid"
)

// Department represents the possible departments in the system
type Department string

const (
	Technical  Department = "technical"
	Design     Department = "design"
	Management Department = "management"
	Social     Department = "social"
)

type Question struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	Department Department `json:"department" db:"department"`
	Title      string     `json:"title" db:"title"`
	Body       string     `json:"body" db:"body"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// CreateQuestionRequest represents the request body for creating a new question
type CreateQuestionRequest struct {
	Department string `json:"department" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Body       string `json:"body" binding:"required"`
}
