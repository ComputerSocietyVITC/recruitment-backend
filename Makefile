# Makefile for Recruitment Backend

# Default target
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  run             Start the server"
	@echo "  build           Build the application"
	@echo "  jwt-secret      Generate a random JWT secret"
	@echo "  hash-password   Generate a password hash (usage: make hash-password PWD=yourpassword)"
	@echo "  build-cli       Build standalone CLI utility"
	@echo "  test            Run tests"
	@echo "  clean           Clean build artifacts"

# Server commands
.PHONY: run
run:
	go run main.go

.PHONY: build
build:
	go build -o bin/recruitment-backend main.go

# CLI utilities
.PHONY: jwt-secret
jwt-secret:
	go run cmds/*.go jwt-secret

.PHONY: hash-password
hash-password:
	@if [ -z "$(PWD)" ]; then \
		echo "Usage: make hash-password PWD=yourpassword"; \
		exit 1; \
	fi
	go run cmds/*.go hash-password $(PWD)

.PHONY: build-cli
build-cli:
	mkdir -p bin
	go build -o bin/recruitment-cli cmds/*.go

# Testing
.PHONY: test
test:
	go test ./...

# Cleanup
.PHONY: clean
clean:
	rm -rf bin/
	go clean

# Development helpers
.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: format
format:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run
