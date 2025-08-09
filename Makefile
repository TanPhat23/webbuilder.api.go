# Variables
APP_NAME := webbuilder-api
MAIN_PATH := ./cmd/main.go
BUILD_DIR := ./bin
GO_VERSION := 1.24

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development
.PHONY: run
run: ## Run the application in development mode
	@echo "Starting development server..."
	go run ./cmd/main.go

.PHONY: dev
dev: ## Run with hot reload (requires air - go install github.com/air-verse/air@latest)
	@echo "Starting development server with hot reload..."
	air

.PHONY: watch
watch: ## Watch for changes and restart (alias for dev)
	@make dev

# Build
.PHONY: build
build: ## Build the application
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux $(MAIN_PATH)

.PHONY: build-windows
build-windows: ## Build for Windows
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows.exe $(MAIN_PATH)

.PHONY: build-mac
build-mac: ## Build for macOS
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-mac $(MAIN_PATH)

.PHONY: build-all
build-all: build-linux build-windows build-mac ## Build for all platforms

# Testing
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -race -v ./...

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Code Quality
.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint)
	@echo "Running linter..."
	golangci-lint run

.PHONY: check
check: fmt vet lint test ## Run all code quality checks

# Dependencies
.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

.PHONY: deps-tidy
deps-tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	go mod tidy

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	go mod verify

# Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(APP_NAME):latest

# Database
.PHONY: db-migrate
db-migrate: ## Run database migrations (customize as needed)
	@echo "Running database migrations..."
	# Add your migration command here

.PHONY: db-rollback
db-rollback: ## Rollback database migrations (customize as needed)
	@echo "Rolling back database migrations..."
	# Add your rollback command here

# Cleanup
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: clean-all
clean-all: clean ## Clean all generated files
	@echo "Cleaning all generated files..."
	go clean -cache
	go clean -modcache

# Installation
.PHONY: install
install: ## Install the application
	@echo "Installing application..."
	go install $(MAIN_PATH)

# Security
.PHONY: security
security: ## Run security checks (requires gosec)
	@echo "Running security checks..."
	gosec ./...

# Documentation
.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	godoc -http=:6060

# Environment
.PHONY: env-check
env-check: ## Check Go environment
	@echo "Go version: $(shell go version)"
	@echo "Go environment:"
	@go env

# Quick start
.PHONY: setup
setup: deps-tidy ## Setup project for development
	@echo "Setting up project..."
	@echo "Project setup complete!"

.PHONY: start
start: build ## Build and run the application
	@echo "Starting application..."
	./$(BUILD_DIR)/$(APP_NAME)