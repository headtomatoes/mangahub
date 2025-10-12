# Multi-stage Dockerfile for multiple Go services

# Base stage with all necessary tools for building and development
FROM golang:1.25-alpine AS base
# Install system packages
RUN apk add --no-cache git sqlite-dev build-base protobuf-dev
# Install Go-based protobuf tools
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
WORKDIR /app

# Development stage with live reload
FROM base AS development
# Install Air for hot reloading
RUN go install github.com/air-verse/air@latest
# Install Delve for debugging
RUN go install github.com/go-delve/delve/cmd/dlv@latest
# Install protobuf compiler for gRPC
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy proto files and generate gRPC code
COPY proto/ ./proto/
RUN protoc --go_out=. --go-grpc_out=. proto/*.proto

# Expose ports for all services
EXPOSE 8080 8081 8082 8083 2345

# Default development command (can be overridden)
CMD ["air", "-c", ".air.toml"]

# Builder stage - builds all binaries
FROM base AS builder
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf files
RUN protoc --go_out=. --go-grpc_out=. proto/*.proto

# Build all binaries with CGO enabled for SQLite
ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o bin/api-server ./cmd/api-server
RUN go build -ldflags="-s -w" -o bin/tcp-server ./cmd/tcp-server  
RUN go build -ldflags="-s -w" -o bin/udp-server ./cmd/udp-server
RUN go build -ldflags="-s -w" -o bin/grpc-server ./cmd/grpc-server

# Production stage for API server
FROM alpine:latest AS api-production
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /app
COPY --from=builder /app/bin/api-server ./
COPY --from=builder /app/data/ ./data/
EXPOSE 8080
CMD ["./api-server"]

# Production stage for TCP server
FROM alpine:latest AS tcp-production
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /app
COPY --from=builder /app/bin/tcp-server ./
COPY --from=builder /app/data/ ./data/
EXPOSE 8081
CMD ["./tcp-server"]

# Production stage for UDP server
FROM alpine:latest AS udp-production
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /app
COPY --from=builder /app/bin/udp-server ./
COPY --from=builder /app/data/ ./data/
EXPOSE 8082
CMD ["./udp-server"]

# Production stage for gRPC server
FROM alpine:latest AS grpc-production
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /app
COPY --from=builder /app/bin/grpc-server ./
COPY --from=builder /app/data/ ./data/
EXPOSE 8083
CMD ["./grpc-server"]
