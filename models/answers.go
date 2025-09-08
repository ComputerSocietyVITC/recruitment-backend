package models

import (
	"time"

	"github.com/google/uuid"
)

type Answer struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"application_id"`
	QuestionID    uuid.UUID `json:"question_id"`
	Body          string    `json:"body"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SaveApplicationRequest struct {
	Answers []AnswerRequest `json:"answers" binding:"required"`
}

type AnswerRequest struct {
	QuestionID uuid.UUID `json:"question_id" binding:"required"`
	Body       string    `json:"body" binding:"required"`
}
