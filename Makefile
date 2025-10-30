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
	@echo "  build                Build all server binaries"
	@echo "  build-api            Build API server"
	@echo "  build-tcp            Build TCP server"
	@echo "  build-udp            Build UDP server"
	@echo "  build-grpc           Build gRPC server"
	@echo "  docker-up            Start all Docker containers"
	@echo "  docker-down          Stop all Docker containers"
	@echo "  docker-logs          View Docker container logs"
	@echo "  docker-rebuild       Rebuild and restart all containers"
	@echo "  test                 Run all tests"
	@echo "  clean                Clean build artifacts"
	@echo "  fmt                  Format Go code"
	@echo "  dev                  Start all services in development mode"
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
