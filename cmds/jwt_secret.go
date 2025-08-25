package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

// generateJWTSecret generates a cryptographically secure random JWT secret
func generateJWTSecret() {
	// Generate 32 bytes (256 bits) of random data
	secretBytes := make([]byte, 32)
	_, err := rand.Read(secretBytes)
	if err != nil {
		log.Fatalf("Failed to generate random secret: %v", err)
	}

	// Encode to base64 for easy storage
	secret := base64.URLEncoding.EncodeToString(secretBytes)

	fmt.Println("Generated JWT Secret:")
	fmt.Println("=====================")
	fmt.Printf("Base64 encoded: %s\n", secret)
	fmt.Printf("Hex encoded: %x\n", secretBytes)
	fmt.Println("")
	fmt.Println("Add this to your .env file:")
	fmt.Printf("JWT_SECRET=%s\n", secret)
	fmt.Println("")
	fmt.Println("Note: Keep this secret secure and never commit it to version control!")
}
