.PHONY: help build migrate-up migrate-down migrate-status server test docker-up docker-down clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binaries
	@echo "Building migration CLI..."
	@go build -o bin/migrate cmd/migrate/main.go
	@echo "Building HTTP server..."
	@go build -o bin/server cmd/http/main.go
	@echo "Build complete!"

migrate-up: ## Run database migrations
	@./bin/migrate up

migrate-down: ## Rollback last migration
	@./bin/migrate down

migrate-status: ## Show migration status
	@./bin/migrate status

server: ## Start HTTP server
	@./bin/server start

test: ## Run API tests
	@./test_api.sh

docker-up: ## Start services with Docker Compose
	@docker compose up -d

docker-down: ## Stop services with Docker Compose
	@docker compose down

docker-logs: ## Show Docker logs
	@docker compose logs -f

clean: ## Clean build artifacts
	@rm -rf bin/
	@echo "Cleaned build artifacts"

deps: ## Download dependencies
	@go mod download
	@go mod tidy

install: deps build ## Install dependencies and build

run: build migrate-up server ## Build, migrate, and run server
