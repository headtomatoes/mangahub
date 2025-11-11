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
	@echo "  scrape               Scrape data from MangaDex API"
	@echo "  import               Import scraped data to database"
	@echo "  scrape-and-import    Scrape and import data in one command"
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
	@mkdir -p $(BINARY_DIR)
	go build -o $(API_BINARY) ./cmd/api-server
	@echo "API server built: $(API_BINARY)"

## build-tcp: Build TCP server
build-tcp:
	@echo "Building TCP server..."
	@mkdir -p $(BINARY_DIR)
	go build -o $(TCP_BINARY) ./cmd/tcp-server
	@echo "TCP server built: $(TCP_BINARY)"

## build-udp: Build UDP server
build-udp:
	@echo "Building UDP server..."
	@mkdir -p $(BINARY_DIR)
	go build -o $(UDP_BINARY) ./cmd/udp-server
	@echo "UDP server built: $(UDP_BINARY)"

## build-grpc: Build gRPC server
build-grpc:
	@echo "Building gRPC server..."
	@mkdir -p $(BINARY_DIR)
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
	@rm -rf $(BINARY_DIR)
	@rm -rf tmp
	@rm -f coverage.txt
	@rm -f coverage.html
	@rm -f database/migrations/scraped_data.json
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

## scrape: Scrape data from MangaDex API
scrape:
	@echo "Scraping data from MangaDex API..."
	@cd database/migrations/Scrape && go run mangadex_scraper.go types.go
	@echo "✓ Scraping completed! Data saved to database/migrations/scraped_data.json"

## import: Import scraped data to database
import:
	@echo "Importing data to database..."
	@cd database/migrations/import && go run import_to_db.go types.go
	@echo "✓ Import completed!"

## scrape-and-import: Scrape and import data in one command
scrape-and-import:
	@echo "Running complete scrape and import process..."
	@chmod +x database/migrations/scrape_and_import.sh
	@cd database/migrations && ./scrape_and_import.sh
	@echo "✓ Process completed!"

## verify-data: Verify imported data in database
verify-data:
	@echo "Verifying imported data..."
	@docker compose exec db psql -U mangahub -d mangahub -c "SELECT COUNT(*) as total_manga FROM manga;"
	@docker compose exec db psql -U mangahub -d mangahub -c "SELECT COUNT(*) as total_genres FROM genres;"
	@docker compose exec db psql -U mangahub -d mangahub -c "SELECT COUNT(*) as total_relationships FROM manga_genres;"
