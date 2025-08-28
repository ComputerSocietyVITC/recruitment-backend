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
	@echo ""
	@echo "Database commands:"
	@echo "  db-up           Run database migrations (up)"
	@echo "  db-down         Rollback database migrations (down)"
	@echo "  db-docker       Start PostgreSQL in Docker container"
	@echo "  db-docker-stop  Stop PostgreSQL Docker container"

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

# Database commands
.PHONY: db-up
db-up:
	@echo "Running database migrations (up)..."
	@if command -v psql >/dev/null 2>&1; then \
		psql -h ${DB_HOST:-localhost} -U ${DB_USER:-postgres} -d ${DB_NAME:-recruitment_db} -f models/migrations/001_initial_up.sql; \
	else \
		echo "psql not found. Please install PostgreSQL client or use Docker:"; \
		echo "docker exec -i recruitment_postgres psql -U postgres -d recruitment_db < models/migrations/001_initial_up.sql"; \
	fi

.PHONY: db-down
db-down:
	@echo "Rolling back database migrations (down)..."
	@if command -v psql >/dev/null 2>&1; then \
		psql -h ${DB_HOST:-localhost} -U ${DB_USER:-postgres} -d ${DB_NAME:-recruitment_db} -f models/migrations/001_initial_down.sql; \
	else \
		echo "psql not found. Please install PostgreSQL client or use Docker:"; \
		echo "docker exec -i recruitment_postgres psql -U postgres -d recruitment_db < models/migrations/001_initial_down.sql"; \
	fi

.PHONY: db-docker
db-docker:
	@echo "Starting PostgreSQL in Docker container..."
	@if command -v docker >/dev/null 2>&1; then \
		docker run -d \
			--name recruitment_postgres \
			-e POSTGRES_DB=recruitment_db \
			-e POSTGRES_USER=postgres \
			-e POSTGRES_PASSWORD=password \
			-p 5432:5432 \
			-v recruitment_postgres_data:/var/lib/postgresql/data \
			postgres:15 || echo "Container may already exist. Use 'docker start recruitment_postgres' if stopped."; \
		echo "Waiting for database to be ready..."; \
		sleep 5; \
		make db-up; \
	else \
		echo "Docker not found. Please install Docker or set up PostgreSQL manually."; \
	fi

.PHONY: db-docker-stop
db-docker-stop:
	@echo "Stopping PostgreSQL Docker container..."
	@if command -v docker >/dev/null 2>&1; then \
		docker stop recruitment_postgres; \
	else \
		echo "Docker not found."; \
	fi
