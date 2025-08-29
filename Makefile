# Enable automatic deletion of targets on error
.DELETE_ON_ERROR:

include .env
export

.PHONY: test lint build clean help docker-build docker-test docker-integration docker-down format vet

# Build variables
BINARY_NAME = vpnbot
BUILD_DIR = bin

# Environment variables
TELEGRAM_BOT_TOKEN ?= test_token_123456789
DATABASE_URL ?= postgres://testuser:testpass@localhost:5432/arcanus_vpn_test?sslmode=disable
LOG_LEVEL ?= info

# Default target
all: test lint build

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linters
lint:
	golangci-lint run

# Build the application
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/bot/main.go

# Format code
format:
	go fmt ./...
	goimports -w .

# Run go vet
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run the application
run:
	go run cmd/bot/main.go

# Docker targets
docker-build:
	docker compose build

docker-run:
	docker compose up --build

docker-down:
	docker compose down
# Show help
help:
	@echo "Available targets:"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linters"
	@echo "  format        - Format code"
	@echo "  vet           - Run go vet"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  run           - Run the application"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-test   - Run tests in Docker"
	@echo "  docker-integration - Run integration tests in Docker"
	@echo "  docker-down   - Stop Docker containers"
	@echo "  all           - Run test, lint, and build"
