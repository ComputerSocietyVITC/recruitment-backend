# Docker operations for recruitment backend

.PHONY: help build up down logs restart clean dev-setup

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env from template..."; \
		cp .env.docker .env; \
		echo "⚠️  Please edit .env file with your actual values before running 'make up'"; \
	else \
		echo ".env file already exists"; \
	fi

build: ## Build the Docker images
	docker-compose build

up: dev-setup ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

logs: ## Show logs from all services
	docker-compose logs -f

logs-app: ## Show logs from application only
	docker-compose logs -f app

logs-db: ## Show logs from database only
	docker-compose logs -f postgres

restart: ## Restart all services
	docker-compose restart

restart-app: ## Restart application only
	docker-compose restart app

clean: ## Stop services and remove containers, networks, and volumes
	docker-compose down -v --remove-orphans

clean-all: ## Remove everything including images
	docker-compose down -v --rmi all --remove-orphans

rebuild: ## Rebuild and restart the application
	docker-compose build app
	docker-compose up -d app

shell-app: ## Access application container shell
	docker-compose exec app sh

shell-db: ## Access database container shell
	docker-compose exec postgres psql -U postgres -d recruitment_db

status: ## Show status of all services
	docker-compose ps

health: ## Check health of services
	@echo "Checking service health..."
	@docker-compose ps
	@echo "\nTesting application endpoint..."
	@curl -s http://localhost:8080/ping || echo "Application not responding"

generate-jwt: ## Generate a secure JWT secret
	@echo "Generated JWT Secret:"
	@openssl rand -base64 32

# Development helpers
dev-logs: ## Follow development logs with timestamps
	docker-compose logs -f --timestamps

dev-rebuild: down build up ## Full rebuild for development

# Database operations
db-reset: ## Reset database (removes all data)
	docker-compose down postgres
	docker volume rm recruitment-backend_postgres_data || true
	docker-compose up -d postgres
	@echo "Database reset complete. Application will run migrations on next start."

db-backup: ## Backup database to backup.sql
	docker-compose exec postgres pg_dump -U postgres recruitment_db > backup.sql
	@echo "Database backed up to backup.sql"

db-restore: ## Restore database from backup.sql (requires backup.sql file)
	@if [ ! -f backup.sql ]; then \
		echo "backup.sql file not found"; \
		exit 1; \
	fi
	docker-compose exec -T postgres psql -U postgres recruitment_db < backup.sql
	@echo "Database restored from backup.sql"
