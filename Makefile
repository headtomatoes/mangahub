# MangaHub Makefile
.PHONY: help build run test clean docker-up docker-down migrate seed install-tools

# Variables
BINARY_DIR := bin
API_BINARY := $(BINARY_DIR)/api-server
TCP_BINARY := $(BINARY_DIR)/tcp-server
UDP_BINARY := $(BINARY_DIR)/udp-server
GRPC_BINARY := $(BINARY_DIR)/grpc-server

.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "MangaHub - Available Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build                Build all server binaries"
	@echo "  build-api            Build API server"
	@echo "  build-tcp            Build TCP server"
	@echo "  build-udp            Build UDP server"
	@echo "  build-grpc           Build gRPC server"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-up            Start all Docker containers"
	@echo "  docker-down          Stop all Docker containers"
	@echo "  docker-logs          View Docker container logs"
	@echo "  docker-rebuild       Rebuild and restart all containers"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-check             Quick database health check"
	@echo "  db-health            Comprehensive database health report"
	@echo "  db-shell             Open PostgreSQL shell"
	@echo "  db-reset             Reset database (WARNING: deletes all data)"
	@echo ""
	@echo "Test Commands:"
	@echo "  test                 Run all tests"
	@echo "  test-unit            Run unit tests only"
	@echo "  test-integration     Run integration tests with test database"
	@echo "  test-db-setup        Setup test database"
	@echo ""
	@echo "Development:"
	@echo "  dev                  Start all services in development mode"
	@echo "  fmt                  Format Go code"
	@echo "  clean                Clean build artifacts"
	@echo "  install-tools        Install development tools"
	@echo ""

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Tools installed successfully!"

## build: Build all server binaries
build: build-api build-tcp build-udp build-grpc

## build-api: Build API server
build-api:
	@echo "Building API server..."
	@if not exist $(BINARY_DIR) mkdir $(BINARY_DIR)
	go build -o $(API_BINARY) ./cmd/api-server
	@echo "API server built: $(API_BINARY)"

## build-tcp: Build TCP server
build-tcp:
	@echo "Building TCP server..."
	@if not exist $(BINARY_DIR) mkdir $(BINARY_DIR)
	go build -o $(TCP_BINARY) ./cmd/tcp-server
	@echo "TCP server built: $(TCP_BINARY)"

## build-udp: Build UDP server
build-udp:
	@echo "Building UDP server..."
	@if not exist $(BINARY_DIR) mkdir $(BINARY_DIR)
	go build -o $(UDP_BINARY) ./cmd/udp-server
	@echo "UDP server built: $(UDP_BINARY)"

## build-grpc: Build gRPC server
build-grpc:
	@echo "Building gRPC server..."
	@if not exist $(BINARY_DIR) mkdir $(BINARY_DIR)
	go build -o $(GRPC_BINARY) ./cmd/grpc-server
	@echo "gRPC server built: $(GRPC_BINARY)"

## docker-up: Start all Docker containers
docker-up:
	@echo "Starting Docker containers..."
	docker compose up -d
	@echo "Docker containers started!"

## docker-down: Stop all Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	docker compose down
	@echo "Docker containers stopped!"

## docker-logs: View Docker container logs
docker-logs:
	docker compose logs -f

## docker-rebuild: Rebuild and restart all containers
docker-rebuild:
	@echo "Rebuilding Docker containers..."
	docker compose up -d --build
	@echo "Docker containers rebuilt!"

## test: Run all tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Tests completed!"

## test-unit: Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...
	@echo "Unit tests completed!"

## test-db-setup: Create and migrate test database
test-db-setup:
	@echo "Setting up test database..."
	@powershell -ExecutionPolicy Bypass -File ./database/setup-test-db.ps1

## test-integration: Run integration tests with test database
test-integration: test-db-setup
	@echo "Running integration tests..."
	@powershell -Command "$$env:DATABASE_URL='postgres://mangahub:mangahub_secret@localhost:5432/mangahub_test?sslmode=disable'; $$env:JWT_SECRET='test-secret-key-must-be-at-least-32-chars-long'; go test -v -race ./test/..."

## db-check: Quick database health check
db-check:
	@echo "Checking database health..."
	@docker exec mangahub_db psql -U mangahub -d mangahub -c "SELECT '✅ Database connected' AS status;"
	@docker exec mangahub_db psql -U mangahub -d mangahub -t -c "SELECT COUNT(*) || ' tables' FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';"
	@docker exec mangahub_db psql -U mangahub -d mangahub -t -c "SELECT CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role') THEN '✅ Role column exists' ELSE '❌ Role column MISSING' END;"

## db-health: Comprehensive database health report
db-health:
	@echo "Running comprehensive health check..."
	@docker exec -i mangahub_db psql -U mangahub -d mangahub < database/health-check.sql

## db-shell: Open PostgreSQL shell
db-shell:
	@echo "Opening PostgreSQL shell..."
	@docker exec -it mangahub_db psql -U mangahub -d mangahub

## db-reset: Reset database (WARNING: deletes all data)
db-reset:
	@echo "WARNING: This will delete all data!"
	@echo "Press Ctrl+C to cancel, or wait 5 seconds to continue..."
	@powershell -Command "Start-Sleep -Seconds 5"
	@echo "Resetting database..."
	docker compose down -v
	docker compose up -d db
	@echo "Waiting for database to be ready..."
	@powershell -Command "Start-Sleep -Seconds 10"
	@echo "Database reset complete!"
	@echo "Run 'make db-check' to verify"

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@if exist $(BINARY_DIR) rmdir /s /q $(BINARY_DIR)
	@if exist tmp rmdir /s /q tmp
	@if exist coverage.txt del /q coverage.txt
	@if exist coverage.html del /q coverage.html
	@echo "Clean completed!"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted!"

## dev: Start all services in development mode
dev:
	@echo "Starting all services in development mode..."
	docker compose up
