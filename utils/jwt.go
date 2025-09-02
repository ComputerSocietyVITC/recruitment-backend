package utils

import (
	"errors"
	"log"
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/models"
	"github.com/golang-jwt/jwt/v5"
	
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`  // ← string
	Email  string `json:"email"`    // ← string  
	Role   string `json:"role"`     // ← string
	jwt.RegisteredClaims
}

var jwtSecret []byte

// InitJWT initializes the JWT secret from environment variables
func InitJWT() {
	secret := GetEnvWithDefault("JWT_SECRET", "your-secret-key-change-in-production")
	jwtSecret = []byte(secret)
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(user *models.User) (string, error) {
	if len(jwtSecret) == 0 {
		InitJWT()
	}

	expirationTime := time.Now().Add(GetJWTExpiryDuration())

	claims := &JWTClaims{
		UserID: user.ID.String(),   // ← Convert UUID to string
		Email:  user.Email,         // ← Already string
		Role:   string(user.Role),  // ← Convert UserRole to string
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "recruitment-backend",
			Subject:   user.ID.String(), // ← Convert to string
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	if len(jwtSecret) == 0 {
		InitJWT()
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		log.Println("JWT Parsing Error:", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshJWT generates a new JWT token with extended expiration time
func RefreshJWT(tokenString string) (string, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}

	// Create new token with extended expiration
	expirationTime := time.Now().Add(GetJWTExpiryDuration())
	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	claims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}