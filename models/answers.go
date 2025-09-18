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

// AnswerWithQuestion represents an answer with its corresponding question details
type AnswerWithQuestion struct {
	Answer
	QuestionTitle string `json:"question_title" db:"question_title"`
	QuestionBody  string `json:"question_body" db:"question_body"`
	Department    string `json:"department" db:"department"`
}

type SaveApplicationRequest struct {
	Answers []AnswerRequest `json:"answers" binding:"required"`
}

type AnswerRequest struct {
	QuestionID uuid.UUID `json:"question_id" binding:"required"`
	Body       string    `json:"body" binding:"required"`
}

type PostAnswerRequest struct {
	ApplicationID uuid.UUID `json:"application_id" binding:"required"`
	QuestionID    uuid.UUID `json:"question_id" binding:"required"`
	Body          string    `json:"body" binding:"required"`
}
