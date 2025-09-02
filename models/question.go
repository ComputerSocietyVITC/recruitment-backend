package models

import (
	"time"

	"github.com/google/uuid"
)

// Question represents a question for a department
// Type is always "text" for now
// Department is one of: technical, marketing, management, social_media

type Question struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Department string    `json:"department" db:"department"`
	Body       string    `json:"body" db:"body"`
	Type       string    `json:"type" db:"-"` // always "text" for now
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
