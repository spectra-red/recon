# Spectra-Red Makefile
# Convenient commands for development and deployment

.PHONY: help up down restart logs build clean test db-setup db-reset health

# Default target
.DEFAULT_GOAL := help

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Spectra-Red Development Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# ============================================================================
# Docker Compose Commands
# ============================================================================

up: ## Start all services
	@echo "$(BLUE)Starting Spectra-Red services...$(NC)"
	cd deployments && docker-compose up -d
	@echo "$(GREEN)Services started!$(NC)"
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@$(MAKE) health

down: ## Stop all services
	@echo "$(BLUE)Stopping Spectra-Red services...$(NC)"
	cd deployments && docker-compose down
	@echo "$(GREEN)Services stopped!$(NC)"

restart: down up ## Restart all services

logs: ## View logs for all services
	cd deployments && docker-compose logs -f

logs-api: ## View API logs
	cd deployments && docker-compose logs -f api

logs-db: ## View SurrealDB logs
	cd deployments && docker-compose logs -f surrealdb

logs-restate: ## View Restate logs
	cd deployments && docker-compose logs -f restate

build: ## Build Docker images
	@echo "$(BLUE)Building Docker images...$(NC)"
	cd deployments && docker-compose build
	@echo "$(GREEN)Build complete!$(NC)"

build-api: ## Build only API image
	@echo "$(BLUE)Building API image...$(NC)"
	cd deployments && docker-compose build api
	@echo "$(GREEN)API build complete!$(NC)"

rebuild: clean build up ## Clean, rebuild, and restart

ps: ## Show running services
	cd deployments && docker-compose ps

# ============================================================================
# Database Commands
# ============================================================================

db-setup: ## Initialize database with schema and seed data
	@echo "$(BLUE)Setting up database...$(NC)"
	./scripts/setup-db.sh
	@echo "$(GREEN)Database setup complete!$(NC)"

db-reset: down ## Reset database (WARNING: deletes all data)
	@echo "$(YELLOW)WARNING: This will delete all data!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		cd deployments && docker-compose down -v; \
		$(MAKE) up; \
		$(MAKE) db-setup; \
	fi

# ============================================================================
# Health & Status
# ============================================================================

health: ## Check health of all services
	@echo "$(BLUE)Checking service health...$(NC)"
	@echo -n "SurrealDB: "
	@curl -sf http://localhost:8000/health > /dev/null && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(YELLOW)✗ Unhealthy$(NC)"
	@echo -n "Restate:   "
	@curl -sf http://localhost:9070/health > /dev/null && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(YELLOW)✗ Unhealthy$(NC)"
	@echo -n "API:       "
	@curl -sf http://localhost:3000/health > /dev/null && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(YELLOW)✗ Unhealthy$(NC)"

open-db: ## Open SurrealDB in browser
	@open http://localhost:8000

open-restate: ## Open Restate Admin UI
	@open http://localhost:9070

open-api: ## Open API docs
	@open http://localhost:3000/docs

# ============================================================================
# Development
# ============================================================================

dev: ## Start development environment
	@echo "$(BLUE)Starting development environment...$(NC)"
	@$(MAKE) up
	@$(MAKE) db-setup
	@echo "$(GREEN)Development environment ready!$(NC)"
	@echo ""
	@echo "Services:"
	@echo "  - API:        http://localhost:3000"
	@echo "  - SurrealDB:  http://localhost:8000"
	@echo "  - Restate UI: http://localhost:9070"
	@echo ""
	@echo "Run '$(GREEN)make logs$(NC)' to view logs"

clean: down ## Clean up containers, networks, and volumes
	@echo "$(BLUE)Cleaning up...$(NC)"
	cd deployments && docker-compose down -v --remove-orphans
	@echo "$(GREEN)Cleanup complete!$(NC)"

# ============================================================================
# Testing
# ============================================================================

test: ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v ./...

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	@$(MAKE) up
	@sleep 5
	go test -v -tags=integration ./tests/integration/...

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# ============================================================================
# Build & Run Locally (without Docker)
# ============================================================================

run-api: ## Run API server locally (requires local Go)
	@echo "$(BLUE)Starting API server...$(NC)"
	go run cmd/api/main.go

build-local: ## Build binaries locally
	@echo "$(BLUE)Building binaries...$(NC)"
	go build -o bin/api cmd/api/main.go
	go build -o bin/cli cmd/cli/main.go
	go build -o bin/workflows cmd/workflows/main.go
	@echo "$(GREEN)Binaries built in bin/$(NC)"

# ============================================================================
# Code Quality
# ============================================================================

fmt: ## Format Go code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Formatting complete!$(NC)"

lint: ## Lint Go code
	@echo "$(BLUE)Linting code...$(NC)"
	golangci-lint run
	@echo "$(GREEN)Linting complete!$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)Vet complete!$(NC)"

# ============================================================================
# Dependencies
# ============================================================================

deps: ## Download Go dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	go mod download
	@echo "$(GREEN)Dependencies downloaded!$(NC)"

tidy: ## Tidy Go modules
	@echo "$(BLUE)Tidying modules...$(NC)"
	go mod tidy
	@echo "$(GREEN)Modules tidied!$(NC)"

verify: ## Verify Go modules
	@echo "$(BLUE)Verifying modules...$(NC)"
	go mod verify
	@echo "$(GREEN)Modules verified!$(NC)"

# ============================================================================
# Utilities
# ============================================================================

shell-api: ## Open shell in API container
	cd deployments && docker-compose exec api sh

shell-db: ## Open SurrealDB SQL shell
	cd deployments && docker-compose exec surrealdb surreal sql \
		--conn http://localhost:8000 \
		--user root --pass root \
		--ns spectra --db intel

watch: ## Watch logs for all services
	cd deployments && docker-compose logs -f --tail=50

stats: ## Show container resource usage
	cd deployments && docker-compose stats
