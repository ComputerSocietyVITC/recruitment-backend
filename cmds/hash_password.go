package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword generates a bcrypt hash for the given password
func hashPassword(password string) {
	if password == "" {
		log.Fatal("Password cannot be empty")
	}

	// Generate hash with default cost (currently 10)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	fmt.Println("Password Hash Generated:")
	fmt.Println("=======================")
	fmt.Printf("Original password: %s\n", password)
	fmt.Printf("Bcrypt hash: %s\n", string(hashedPassword))
	fmt.Printf("Cost factor: %d\n", bcrypt.DefaultCost)
	fmt.Println("")
	fmt.Println("Note: This hash can be safely stored in your database.")
	fmt.Println("Warning: The original password is shown above for reference only.")
	fmt.Println("         In production, never log or display plain text passwords!")
}
