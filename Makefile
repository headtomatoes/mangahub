.PHONY: dev build test clean logs proto

# Development commands
dev:
	docker-compose up --build

dev-api:
	docker-compose up api-server redis --build

dev-tcp:
	docker-compose up tcp-server --build

dev-udp:
	docker-compose up udp-server --build

dev-grpc:
	docker-compose up grpc-server --build

# Build production images
build-prod:
	docker build --target api-production -t mangahub-api:latest .
	docker build --target tcp-production -t mangahub-tcp:latest .
	docker build --target udp-production -t mangahub-udp:latest .
	docker build --target grpc-production -t mangahub-grpc:latest .

# Generate protobuf files
proto:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

# Testing
test:
	docker-compose exec api-server go test ./...

# Monitoring
monitoring:
	docker-compose --profile monitoring up -d

# Logs
logs:
	docker-compose logs -f

logs-api:
	docker-compose logs -f api-server

logs-tcp:
	docker-compose logs -f tcp-server

# Cleanup
clean:
	docker-compose down -v
	docker system prune -f

# Database operations
db-reset:
	docker-compose down -v
	docker volume rm mangahub_sqlite_data
	docker-compose up -d
