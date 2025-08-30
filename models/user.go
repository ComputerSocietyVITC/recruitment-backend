package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the possible roles in the system
type UserRole string

const (
	RoleApplicant  UserRole = "applicant"
	RoleEvaluator  UserRole = "evaluator"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "super_admin"
)

// User represents a user in the system
type User struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	FullName            string    `json:"full_name" db:"full_name"`
	Email               string    `json:"email" db:"email"`
	PhoneNumber         string    `json:"phone_number" db:"phone_number"`
	HashedPassword      string    `json:"-" db:"hashed_password"` // JSON tag "-" to exclude from JSON serialization
	Verified            bool      `json:"verified" db:"verified"`
	ResetToken          string    `json:"reset_token" db:"reset_token"`
	ResetTokenExpiresAt time.Time `json:"reset_token_expires_at" db:"reset_token_expires_at"`
	Role                UserRole  `json:"role" db:"role"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	FullName    string   `json:"full_name" binding:"required"`
	Email       string   `json:"email" binding:"required,email"`
	PhoneNumber string   `json:"phone_number" binding:"required"`
	Password    string   `json:"password" binding:"required,min=6"` // Plain password from request
	Role        UserRole `json:"role,omitempty"`                    // Optional role field for admin creation
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the response for authentication endpoints
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Role        UserRole  `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts a User to UserResponse (excludes sensitive data)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		FullName:    u.FullName,
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
