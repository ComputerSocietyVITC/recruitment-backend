package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "jwt-secret":
		generateJWTSecret()
	case "hash-password":
		if len(os.Args) < 3 {
			fmt.Println("Error: Password is required")
			fmt.Println("Usage: go run cmds/main.go hash-password <password>")
			os.Exit(1)
		}
		password := os.Args[2]
		hashPassword(password)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Available commands:")
	fmt.Println("  jwt-secret          Generate a random JWT secret")
	fmt.Println("  hash-password <pwd> Generate a bcrypt hash for the given password")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run cmds/main.go jwt-secret")
	fmt.Println("  go run cmds/main.go hash-password mypassword123")
}
