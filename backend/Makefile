.PHONY: help build run test clean docker-up docker-down migrate seed

# Variables
APP_NAME=financial-risk-monitor
MAIN_PATH=cmd/api/main.go
DOCKER_COMPOSE=docker-compose

# Colors for output
GREEN=\033[0;32m
NC=\033[0m # No Color

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  ${GREEN}%-15s${NC} %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete!"

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	@go run $(MAIN_PATH)

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -cover ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean
	@echo "Clean complete!"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):latest .

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	@$(DOCKER_COMPOSE) up -d
	@echo "Services started!"

docker-down: ## Stop all services
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) down
	@echo "Services stopped!"

docker-logs: ## View logs from all services
	@$(DOCKER_COMPOSE) logs -f

docker-reset: ## Reset all Docker volumes and containers
	@echo "Resetting Docker environment..."
	@$(DOCKER_COMPOSE) down -v
	@echo "Reset complete!"

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@migrate -path migrations -database "postgresql://riskmonitor:securepassword123@localhost:5432/financial_risk_db?sslmode=disable" up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@migrate -path migrations -database "postgresql://riskmonitor:securepassword123@localhost:5432/financial_risk_db?sslmode=disable" down 1

seed: ## Seed the database with sample data
	@echo "Seeding database..."
	@go run scripts/seed.go

dev: docker-up ## Start development environment
	@echo "Starting development environment..."
	@sleep 5
	@make migrate-up
	@make seed
	@make run

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted!"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated!"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed!"