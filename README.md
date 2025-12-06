# MangaHub

A comprehensive manga tracking and management platform built with Go, featuring a microservices architecture with multiple communication protocols (HTTP REST, gRPC, TCP, UDP, WebSocket).

![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)

---

## Features

- **Manga Search & Discovery** — Search and browse manga from integrated sources (MangaDex)
- **Reading Progress Sync** — Track your reading progress across devices
- **Personal Library** — Organize your manga collection with custom lists
- **Real-time Notifications** — Get notified about new chapters via UDP/WebSocket
- **Community Chat** — Join chat rooms for manga discussions
- **Secure Authentication** — JWT-based authentication with refresh tokens
- **Multi-Protocol Support** — HTTP REST, gRPC, TCP, UDP, and WebSocket

---

## Architecture

MangaHub follows a microservices architecture with the following components:

```text
┌─────────────────────────────────────────────────────────────────────┐
│                           MangaHub Platform                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │  API Server  │  │  TCP Server  │  │  UDP Server  │               │
│  │   (HTTP)     │  │   (Sync)     │  │  (Notify)    │               │
│  │   :8084      │  │   :8081      │  │   :8082      │               │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘               │
│         │                 │                 │                       │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐               │
│  │ gRPC Server  │  │  WebSocket   │  │ MangaDex     │               │
│  │   :8083      │  │  (Chat/WS)   │  │    Sync      │               │
│  └──────┬───────┘  └──────────────┘  └──────────────┘               │
│         │                                                           │
│  ┌──────┴────────────────────────────────────────────┐              │
│  │                  Shared Services                  │              │
│  │  (Auth, Manga, User, Library, Progress, Ratings)  │              │
│  └──────┬────────────────────────────────────────────┘              │
│         │                                                           │
│  ┌──────┴───────────────────────────────────────────┐               │
│  │          PostgreSQL          │       Redis       │               │
│  │          (Database)          │      (Cache)      │               │
│  └──────────────────────────────────────────────────┘               │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🛠️ Tech Stack

| Category | Technology |
|----------|------------|
| **Language** | Go 1.25 |
| **HTTP Framework** | [Gin](https://github.com/gin-gonic/gin) |
| **WebSocket** | [Gorilla WebSocket](https://github.com/gorilla/websocket) |
| **gRPC** | [gRPC-Go](https://grpc.io/docs/languages/go/) + Protocol Buffers |
| **Database** | PostgreSQL 15 |
| **ORM** | [GORM](https://gorm.io/) |
| **Cache** | Redis 7 |
| **Authentication** | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) |
| **CLI Framework** | [Cobra](https://github.com/spf13/cobra) |
| **Testing** | [Testify](https://github.com/stretchr/testify) |
| **Containerization** | Docker & Docker Compose |
| **Hot Reload** | [Air](https://github.com/air-verse/air) |
| **Monitoring** | Prometheus & Grafana (optional) |

---

## 📁 Project Structure

```text
mangahub/
├── cmd/                          # Application entry points
│   ├── api-server/               # HTTP REST API server
│   ├── tcp-server/               # TCP sync server
│   ├── udp-server/               # UDP notification server
│   ├── grpc-server/              # gRPC server
│   ├── grpc-client/              # gRPC client example
│   ├── mangadex-sync/            # MangaDex data synchronization service
│   └── cli/                      # Command-line interface
│       ├── command/              # CLI commands (auth, manga, library, etc.)
│       ├── authentication/       # CLI auth helpers
│       └── dto/                  # Data transfer objects
├── internal/                     # Private application code
│   ├── config/                   # Configuration management
│   ├── microservices/            # Service implementations
│   │   ├── http-api/             # HTTP handlers, services, repositories
│   │   ├── grpc/                 # gRPC service implementation
│   │   ├── tcp/                  # TCP protocol handling
│   │   ├── udp/                  # UDP protocol handling
│   │   └── websocket/            # WebSocket handlers
│   ├── search/                   # Search functionality
│   ├── ingestion/                # Data ingestion utilities
│   └── shared/                   # Shared utilities
├── pkg/                          # Public libraries
├── proto/                        # Protocol Buffer definitions
│   ├── manga.proto               # Manga service definitions
│   └── pb/                       # Generated Go code
├── database/                     # Database utilities
│   ├── migrations/               # SQL migration files
│   └── db.go                     # Database connection
├── data/                         # Local data storage
├── docs/                         # Documentation
├── test/                         # Integration tests
├── docker-compose.yml            # Docker services configuration
├── Dockerfile                    # Multi-stage Docker build
├── Makefile                      # Build automation
└── go.mod                        # Go module definition
```

---

## ▄︻デ══━一💥Quick Start

### Prerequisites

- **Go 1.25+** — [Download](https://go.dev/dl/)
- **Docker & Docker Compose** — [Download](https://www.docker.com/products/docker-desktop/)
- **Make** (optional) — For using Makefile commands

### 1. Clone the Repository

```bash
git clone https://github.com/headtomatoes/mangahub.git
cd mangahub
```

### 2. Environment Setup

Create a `.env` file in the project root:

```env
# Database
DB_PASS=mangahub_secret

# JWT Secret (change in production!)
JWT_SECRET=your-super-secret-jwt-key-at-least-32-chars

# MangaDex API (optional)
MANGADEX_API_KEY=your-api-key

# Grafana (optional)
GRAFANA_PASSWORD=admin
```

### 3. Start Services with Docker

```bash
# Start all services
make docker-up

# Or using docker compose directly
docker compose up -d
```

### 4. Verify Services

```bash
# Check database health
make db-check

# View logs
make docker-logs
```

---

## Service Ports

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| API Server | `8084` | HTTP | REST API endpoints |
| TCP Server | `8081` | TCP | Sync protocol |
| UDP Server | `8082` | UDP | Notifications |
| UDP HTTP | `8085` | HTTP | UDP service HTTP API |
| gRPC Server | `8083` | gRPC | gRPC service |
| PostgreSQL | `5432` | TCP | Database |
| Redis | `6379` | TCP | Cache |
| pgAdmin | `5050` | HTTP | Database admin UI |
| Prometheus | `9090` | HTTP | Metrics (optional) |
| Grafana | `3000` | HTTP | Dashboards (optional) |

---

## CLI Usage

MangaHub includes a powerful CLI for interacting with the platform:

```bash
# Build the CLI
go build -o mangahub ./cmd/cli

# Or run directly
go run ./cmd/cli/main.go
```

### Available Commands

```bash
mangahubCLI auth login          # Login to your account
mangahubCLI auth register       # Create a new account
mangahubCLI auth logout         # Logout from current session

mangahubCLI manga search <query>  # Search for manga
mangahubCLI manga get <id>        # Get manga details

mangahubCLI library list          # List your library
mangahubCLI library add <id>      # Add manga to library

mangahubCLI progress sync         # Sync reading progress
mangahubCLI progress update       # Update chapter progress

mangahubCLI chat join <room>      # Join a chat room

mangahubCLI grpc <command>        # gRPC operations
mangahubCLI udp <command>         # UDP operations
```

---

## Development

### Install Development Tools

```bash
make install-tools
```

This installs:

- **Air** — Hot reload for Go
- **protoc-gen-go** — Protocol Buffer compiler plugin
- **protoc-gen-go-grpc** — gRPC code generator

### Build Binaries

```bash
# Build all services
make build

# Build individual services
make build-api
make build-tcp
make build-udp
make build-grpc
```

### Run Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration
```

### Code Formatting

```bash
make fmt
```

### Database Operations

```bash
# Quick health check
make db-check

# Comprehensive health report
make db-health

# Open PostgreSQL shell
make db-shell

# Reset database (WARNING: deletes all data)
make db-reset
```

### Data Import from MangaDex

```bash
# Scrape manga data
make scrape

# Import to database
make import

# Or do both
make scrape-and-import

# Verify imported data
make verify-data
```

---

## Docker

### Multi-Stage Build Targets

The `Dockerfile` supports multiple build targets:

```bash
# Development (with hot reload)
docker build --target development -t mangahub:dev .

# Production builds
docker build --target production-api -t mangahub-api:prod .
docker build --target production-tcp -t mangahub-tcp:prod .
docker build --target production-udp -t mangahub-udp:prod .
docker build --target production-grpc -t mangahub-grpc:prod .

# Run tests
docker build --target test -t mangahub:test .
```

### Docker Compose Commands

```bash
# Start all services
docker compose up -d

# Start with monitoring (Prometheus + Grafana)
docker compose --profile monitoring up -d

# View logs
docker compose logs -f

# Rebuild containers
docker compose up -d --build

# Stop all services
docker compose down

# Stop and remove volumes
docker compose down -v
```

---

## API Examples

### REST API

```bash
# Register a new user
curl -X POST http://localhost:8084/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "email": "user@example.com", "password": "password123"}'

# Login
curl -X POST http://localhost:8084/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "password123"}'

# Search manga (with auth token)
curl http://localhost:8084/api/v1/manga/search?q=naruto \
  -H "Authorization: Bearer <your-token>"
```

### gRPC

```bash
# Using grpcurl
grpcurl -plaintext localhost:8083 pb.MangaService/SearchManga

# Using the provided client
go run ./cmd/grpc-client/main.go
```

---

## Documentation

Detailed documentation is available in the `docs/` directory:

- [CLI Testing Guide](docs/CLI-TESTING-GUIDE.md)
- [TCP Service Documentation](docs/TCP-SERVICE-DOCUMENTATION-INDEX.md)
- [WebSocket Implementation](docs/WEBSOCKET-IMPLEMENTATION-SUMMARY.md)
- [UDP Auth Implementation](docs/UDP-AUTH-IMPLEMENTATION-SUMMARY.md)
- [Database Setup](docs/DATABASE-SETUP-REVIEW.md)
- [Use Cases](docs/USE-CASES-DOCUMENTATION.md)

---

## Testing

```bash
# Run all tests with coverage
make test

# Run specific test file
go test -v ./test/tcp_integration_test.go

# Run with race detection
go test -race ./...
```

---

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/a-feature`)
3. Commit your changes (`git commit -m 'Add some a feature'`)
4. Push to the branch (`git push origin feature/a-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [MangaDex](https://mangadex.org/) for manga metadata
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [GORM](https://gorm.io/)

---
