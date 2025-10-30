# syntax=docker/dockerfile:1

# ============================================
# Stage 1: Base builder with dependencies
# ============================================
FROM golang:1.25-alpine AS base

# Install build dependencies
RUN apk add --no-cache git make protobuf protobuf-dev gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# ============================================
# Stage 2: Development (with Air)
# ============================================
FROM base AS development

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

# Expose all service ports
EXPOSE 8081 8082 8083 8084

# Default command (overridden in docker-compose)
CMD ["air", "-c", ".air.api.toml"]

# ============================================
# Stage 3: Build API Server
# ============================================
FROM base AS build-api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/api-server ./cmd/api-server

# ============================================
# Stage 4: Build TCP Server
# ============================================
FROM base AS build-tcp
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/tcp-server ./cmd/tcp-server

# ============================================
# Stage 5: Build UDP Server
# ============================================
FROM base AS build-udp
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/udp-server ./cmd/udp-server

# ============================================
# Stage 6: Build gRPC Server
# ============================================
FROM base AS build-grpc
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/grpc-server ./cmd/grpc-server

# ============================================
# Stage 7: Test stage
# ============================================
FROM base AS test
RUN go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# ============================================
# Final Stages: Production Runtime (minimal)
# ============================================

# API Server runtime
FROM alpine:latest AS production-api
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=build-api /bin/api-server .
EXPOSE 8084
CMD ["./api-server"]

# TCP Server runtime
FROM alpine:latest AS production-tcp
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=build-tcp /bin/tcp-server .
EXPOSE 8081
CMD ["./tcp-server"]

# UDP Server runtime
FROM alpine:latest AS production-udp
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=build-udp /bin/udp-server .
EXPOSE 8082
CMD ["./udp-server"]

# gRPC Server runtime
FROM alpine:latest AS production-grpc
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=build-grpc /bin/grpc-server .
EXPOSE 8083
CMD ["./grpc-server"]
