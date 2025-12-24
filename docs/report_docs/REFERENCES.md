# MangaHub Project References

**Document Version:** 1.0  
**Last Updated:** December 24, 2025  
**Project:** MangaHub - Multi-Protocol Manga Management System

---

## Table of Contents

1. [External APIs](#external-apis)
2. [Go Libraries & Frameworks](#go-libraries--frameworks)
3. [Infrastructure & Tools](#infrastructure--tools)
4. [Database Systems](#database-systems)
5. [Network Protocols & Standards](#network-protocols--standards)
6. [Development Tools](#development-tools)
7. [Documentation & Learning Resources](#documentation--learning-resources)
8. [Third-Party Services](#third-party-services)

---

## External APIs

### Manga Data Sources

#### MangaDex API
- **URL:** https://api.mangadex.org
- **Documentation:** https://api.mangadex.org/docs/
- **Protocol:** REST API
- **Rate Limit:** 5 requests/second, burst of 10
- **Authentication:** Optional API key
- **Purpose:** Primary manga metadata and chapter data source
- **Endpoints Used:**
  - `GET /manga` - Manga discovery with filters
  - `GET /manga/{id}/feed` - Chapter listings
  - `GET /chapter/{id}` - Chapter details
- **Data Retrieved:** 
  - Manga metadata (title, author, description, genres, cover images)
  - Chapter information (number, volume, pages, published dates)
  - Relationships (authors, artists, cover art)

#### AniList GraphQL API
- **URL:** https://graphql.anilist.co
- **Documentation:** https://anilist.gitbook.io/anilist-apiv2-docs/
- **Protocol:** GraphQL
- **Rate Limit:** ~90 requests/minute (1.5 req/sec)
- **Authentication:** OAuth2 (optional for public data)
- **Purpose:** Complementary manga metadata with user ratings
- **Endpoints Used:**
  - GraphQL endpoint (single endpoint for all queries)
- **Queries:**
  - `Page` - Paginated manga listings
  - `Media` - Individual manga details
- **Data Retrieved:**
  - Manga metadata (title variants, status, chapters, volumes)
  - User scores and ratings (0-100 scale)
  - Staff information (authors, artists with roles)
  - Rich tagging system
  - Publication dates

#### Kitsu API
- **URL:** https://kitsu.io/api/edge
- **Documentation:** https://kitsu.docs.apiary.io/
- **Protocol:** JSON:API
- **Purpose:** External search functionality
- **Endpoints Used:**
  - `GET /manga` - Manga search
- **Data Retrieved:**
  - Basic manga information for search results

---

## Go Libraries & Frameworks

### Core Dependencies

#### Web Framework
```
github.com/gin-gonic/gin v1.11.0
```
- **Purpose:** HTTP REST API framework
- **Documentation:** https://gin-gonic.com/docs/
- **Usage:** Main HTTP server, routing, middleware
- **Features Used:**
  - Route handling
  - JSON binding and validation
  - Middleware support (CORS, authentication)
  - Context management

#### CORS Middleware
```
github.com/gin-contrib/cors v1.7.6
```
- **Purpose:** Cross-Origin Resource Sharing support
- **Documentation:** https://github.com/gin-contrib/cors
- **Usage:** Enable cross-origin requests from web clients

#### WebSocket
```
github.com/gorilla/websocket v1.5.3
```
- **Purpose:** Real-time bidirectional communication
- **Documentation:** https://pkg.go.dev/github.com/gorilla/websocket
- **Usage:** Chat functionality, real-time notifications
- **Features Used:**
  - WebSocket connection upgrade
  - Message broadcasting
  - Connection management

#### gRPC
```
google.golang.org/grpc v1.76.0
google.golang.org/protobuf v1.36.10
```
- **Purpose:** High-performance RPC framework
- **Documentation:** https://grpc.io/docs/languages/go/
- **Usage:** Service-to-service communication
- **Features Used:**
  - Protocol Buffer serialization
  - Bidirectional streaming
  - Service definitions

### Authentication & Security

#### JWT Authentication
```
github.com/golang-jwt/jwt/v5 v5.3.0
```
- **Purpose:** JSON Web Token generation and validation
- **Documentation:** https://pkg.go.dev/github.com/golang-jwt/jwt/v5
- **Usage:** User authentication, session management
- **Features Used:**
  - Token signing with HS256
  - Claims validation
  - Expiration handling

#### Password Hashing
```
golang.org/x/crypto v0.43.0
```
- **Purpose:** Cryptographic functions
- **Documentation:** https://pkg.go.dev/golang.org/x/crypto
- **Usage:** Password hashing with bcrypt
- **Algorithms:** bcrypt for password hashing

#### Secure Credential Storage
```
github.com/zalando/go-keyring v0.2.6
```
- **Purpose:** OS keychain integration
- **Documentation:** https://github.com/zalando/go-keyring
- **Usage:** CLI credential storage (macOS Keychain, Windows Credential Manager, Linux Secret Service)

### Database Libraries

#### PostgreSQL Driver (GORM)
```
gorm.io/gorm v1.31.0
gorm.io/driver/postgres v1.6.0
```
- **Purpose:** ORM and PostgreSQL driver
- **Documentation:** https://gorm.io/docs/
- **Usage:** Database queries, migrations, relationships
- **Features Used:**
  - Auto-migrations
  - Associations (has-many, many-to-many)
  - Transactions
  - Connection pooling

#### PostgreSQL Driver (pgx)
```
github.com/jackc/pgx/v5 v5.7.6
```
- **Purpose:** Low-level PostgreSQL driver
- **Documentation:** https://pkg.go.dev/github.com/jackc/pgx/v5
- **Usage:** Direct SQL queries, connection pooling
- **Features Used:**
  - Connection pool management
  - Prepared statements
  - Context support

#### PostgreSQL Driver (lib/pq)
```
github.com/lib/pq v1.10.9
```
- **Purpose:** Pure Go PostgreSQL driver
- **Documentation:** https://pkg.go.dev/github.com/lib/pq
- **Usage:** Standard database/sql interface

#### Redis Client
```
github.com/redis/go-redis/v9 v9.16.0
```
- **Purpose:** Redis client for caching
- **Documentation:** https://redis.uptrace.dev/
- **Usage:** Caching user progress, session data
- **Features Used:**
  - Key-value operations
  - TTL management
  - Pipelining
  - Pub/Sub (for future use)

### Rate Limiting & Concurrency

#### Rate Limiting
```
golang.org/x/time v0.14.0
```
- **Purpose:** Rate limiting with token bucket algorithm
- **Documentation:** https://pkg.go.dev/golang.org/x/time/rate
- **Usage:** API rate limiting for MangaDex and AniList
- **Algorithm:** Token bucket (5 req/sec for MangaDex, 1 req/sec for AniList)

### CLI Framework

#### Command-Line Interface
```
github.com/spf13/cobra v1.10.1
```
- **Purpose:** CLI framework
- **Documentation:** https://cobra.dev/
- **Usage:** MangaHub CLI tool commands
- **Features Used:**
  - Subcommands
  - Flag parsing
  - Help generation
  - Command aliases

### Utilities

#### UUID Generation
```
github.com/google/uuid v1.6.0
```
- **Purpose:** UUID generation
- **Documentation:** https://pkg.go.dev/github.com/google/uuid
- **Usage:** Unique identifiers for users, sessions

#### Environment Variables
```
github.com/joho/godotenv v1.5.1
```
- **Purpose:** Load environment variables from .env files
- **Documentation:** https://github.com/joho/godotenv
- **Usage:** Configuration management

#### Terminal Colors
```
github.com/fatih/color v1.18.0
```
- **Purpose:** Colored terminal output
- **Documentation:** https://github.com/fatih/color
- **Usage:** CLI output formatting

### Testing Libraries

#### Testing Framework
```
github.com/stretchr/testify v1.11.1
```
- **Purpose:** Testing assertions and mocking
- **Documentation:** https://pkg.go.dev/github.com/stretchr/testify
- **Packages Used:**
  - `assert` - Test assertions
  - `mock` - Mock object generation
  - `suite` - Test suites

#### Mock Generation
```
go.uber.org/mock v0.5.0
```
- **Purpose:** Mock code generation
- **Documentation:** https://github.com/uber-go/mock
- **Usage:** Generate mocks for interfaces (mockgen)

---

## Infrastructure & Tools

### Container & Orchestration

#### Docker
- **Website:** https://www.docker.com/
- **Documentation:** https://docs.docker.com/
- **Version:** Docker Engine 24.x+
- **Purpose:** Containerization of all services
- **Images Used:**
  - `golang:1.25-alpine` - Go application builder
  - `postgres:16-alpine` - PostgreSQL database
  - `redis:7-alpine` - Redis cache
  - `cosmtrek/air` - Live reload for development

#### Docker Compose
- **Documentation:** https://docs.docker.com/compose/
- **Version:** v2.x
- **Purpose:** Multi-container orchestration
- **Services Orchestrated:**
  - api-server (HTTP REST API)
  - tcp-server (TCP sync service)
  - udp-server (UDP notifications)
  - grpc-server (gRPC service)
  - mangadex-sync (Background sync service)
  - anilist-sync (Background sync service)
  - postgres (Database)
  - redis (Cache)

### Development Tools

#### Live Reload
```
cosmtrek/air
```
- **Documentation:** https://github.com/cosmtrek/air
- **Purpose:** Live reload for Go applications
- **Configuration Files:**
  - `.air.api.toml` - API server
  - `.air.tcp.toml` - TCP server
  - `.air.udp.toml` - UDP server
  - `.air.grpc.toml` - gRPC server

#### Makefile
- **Purpose:** Task automation
- **Common Commands:**
  - `make dev` - Start all services
  - `make build` - Build all binaries
  - `make test` - Run tests
  - `make clean` - Clean build artifacts

---

## Database Systems

### PostgreSQL
- **Version:** 16.x
- **Website:** https://www.postgresql.org/
- **Documentation:** https://www.postgresql.org/docs/16/
- **Purpose:** Primary relational database
- **Port:** 5432
- **Features Used:**
  - JSONB data type
  - Full-text search
  - Foreign key constraints
  - Indexes (B-tree, GIN)
  - Triggers
  - Transactions (ACID)
  - Connection pooling

**Schema:**
- users
- manga
- chapters
- genres
- manga_genres (many-to-many)
- comments
- ratings
- user_progress
- user_library
- notifications
- chat_messages
- refresh_tokens
- sync_state

### Redis
- **Version:** 7.x
- **Website:** https://redis.io/
- **Documentation:** https://redis.io/docs/
- **Purpose:** In-memory cache and session store
- **Port:** 6379
- **Features Used:**
  - Key-value storage
  - TTL (Time To Live)
  - String operations
  - Hash operations
- **Use Cases:**
  - User progress caching (1-hour TTL)
  - Session state
  - Rate limiting counters (future)

---

## Network Protocols & Standards

### HTTP/1.1 & HTTP/2
- **RFC:** RFC 7540 (HTTP/2), RFC 2616 (HTTP/1.1)
- **Port:** 8084
- **Purpose:** REST API communication
- **Methods Used:** GET, POST, PUT, DELETE
- **Features:**
  - JSON request/response format
  - Bearer token authentication
  - CORS support
  - Content negotiation

### WebSocket (RFC 6455)
- **RFC:** RFC 6455
- **Port:** 8084 (upgrade from HTTP)
- **Endpoint:** `/ws/chat/:room_id`
- **Purpose:** Real-time chat functionality
- **Features:**
  - Bidirectional communication
  - Low-latency message delivery
  - Room-based broadcasting

### TCP (Transmission Control Protocol)
- **RFC:** RFC 793
- **Port:** 8081
- **Purpose:** Custom binary protocol for progress sync
- **Features:**
  - Connection-oriented
  - Reliable delivery
  - Sequential message processing
  - Binary protocol with length prefixes

### UDP (User Datagram Protocol)
- **RFC:** RFC 768
- **Ports:** 8082 (client), 8085 (server)
- **Purpose:** Push notifications
- **Features:**
  - Connectionless
  - Low overhead
  - Fire-and-forget messaging
  - Broadcast capability

### gRPC (HTTP/2-based RPC)
- **Port:** 8083
- **Protocol:** Protocol Buffers over HTTP/2
- **Purpose:** Service-to-service communication
- **Features:**
  - Strongly-typed messages
  - Bidirectional streaming
  - Efficient binary serialization
  - Code generation

### GraphQL
- **Specification:** https://spec.graphql.org/
- **Used For:** AniList API queries
- **Features:**
  - Flexible query structure
  - Single endpoint
  - Type system
  - Nested data fetching

---

## Development Tools

### Version Control
- **Git:** https://git-scm.com/
- **GitHub:** https://github.com/
- **Repository:** https://github.com/headtomatoes/mangahub

### Programming Language
- **Go (Golang):** Version 1.25
- **Website:** https://go.dev/
- **Documentation:** https://go.dev/doc/

### Build Tools
- **Go Modules:** Dependency management
- **Make:** Task automation
- **Protocol Buffers Compiler:** `protoc` for gRPC code generation

### Code Quality
- **golangci-lint:** Linting tool
- **go fmt:** Code formatting
- **go vet:** Code analysis

---

## Documentation & Learning Resources

### Project Documentation

**Internal Documentation:**
- `/docs/MangaDex-sync REPORT.md` - MangaDex synchronization system
- `/docs/AniList_REPORT.md` - AniList synchronization system
- `/docs/UDP_REPORT.md` - UDP notification system
- `/docs/API_REPORT.md` - HTTP API documentation
- `/docs/api_docs/` - API endpoint documentation
- `/docs/cli_docs/` - CLI usage guides
- `/docs/database_docs/` - Database schema and migrations
- `/docs/grpc_docs/` - gRPC service documentation
- `/docs/tcp_docs/` - TCP protocol documentation
- `/docs/sync_mangadex/` - MangaDex sync detailed guides
- `/README.md` - Project overview

### External Learning Resources

#### Go Programming
- **Official Tour:** https://go.dev/tour/
- **Effective Go:** https://go.dev/doc/effective_go
- **Go by Example:** https://gobyexample.com/

#### Networking
- **HTTP/2 Explained:** https://http2-explained.haxx.se/
- **WebSocket RFC 6455:** https://tools.ietf.org/html/rfc6455
- **gRPC Documentation:** https://grpc.io/docs/

#### Database
- **PostgreSQL Tutorial:** https://www.postgresqltutorial.com/
- **GORM Guides:** https://gorm.io/docs/
- **Redis University:** https://university.redis.com/

#### Concurrency Patterns
- **Go Concurrency Patterns:** https://go.dev/blog/pipelines
- **Context Package:** https://go.dev/blog/context
- **Rate Limiting:** https://pkg.go.dev/golang.org/x/time/rate

---

## Third-Party Services

### Data Sources
- **MangaDex:** Community-driven manga database
- **AniList:** Anime and manga community database
- **Kitsu:** Anime/manga tracking platform

### Development Services
- **GitHub:** Source code hosting and version control
- **Docker Hub:** Container image registry

---

## Package Dependencies Summary

### Direct Dependencies (25)
```
github.com/fatih/color v1.18.0
github.com/gin-contrib/cors v1.7.6
github.com/gin-gonic/gin v1.11.0
github.com/golang-jwt/jwt/v5 v5.3.0
github.com/google/uuid v1.6.0
github.com/gorilla/websocket v1.5.3
github.com/jackc/pgx/v5 v5.7.6
github.com/joho/godotenv v1.5.1
github.com/lib/pq v1.10.9
github.com/redis/go-redis/v9 v9.16.0
github.com/spf13/cobra v1.10.1
github.com/stretchr/testify v1.11.1
github.com/zalando/go-keyring v0.2.6
golang.org/x/crypto v0.43.0
golang.org/x/time v0.14.0
google.golang.org/grpc v1.76.0
google.golang.org/protobuf v1.36.10
gorm.io/driver/postgres v1.6.0
gorm.io/gorm v1.31.0
```

### Indirect Dependencies (40+)
All transitive dependencies are managed through `go.mod` and automatically resolved.

---

## API Endpoints Reference

### MangaHub HTTP API (Port 8084)

**Base URL:** `http://localhost:8084/api`

#### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - User logout

#### Manga
- `GET /manga` - List all manga
- `GET /manga/:id` - Get manga details
- `POST /manga` - Create manga (admin)
- `PUT /manga/:id` - Update manga (admin)
- `DELETE /manga/:id` - Delete manga (admin)
- `GET /manga/search` - Search manga
- `GET /manga/:id/chapters` - Get chapters for manga

#### Library
- `GET /library` - Get user's library
- `POST /library/:manga_id` - Add to library
- `DELETE /library/:manga_id` - Remove from library

#### Progress
- `GET /progress/:manga_id` - Get reading progress
- `PUT /progress/:manga_id` - Update progress

#### Comments & Ratings
- `GET /manga/:id/comments` - Get comments
- `POST /manga/:id/comments` - Add comment
- `POST /manga/:id/rating` - Rate manga

#### External Search
- `GET /search/external` - Search across MangaDex, AniList, Kitsu

### WebSocket Endpoints

**Chat:** `ws://localhost:8084/ws/chat/:room_id`

### gRPC Services (Port 8083)

**Service:** `MangaService`
- `GetManga` - Retrieve manga by ID
- `ListManga` - List manga with pagination
- `SearchManga` - Search manga by query
- `StreamManga` - Streaming manga updates

### TCP Server (Port 8081)

**Protocol:** Custom binary protocol
- Progress sync messages
- Real-time updates

### UDP Notification Server (Port 8085)

**Endpoints:**
- `POST /notify/new-manga` - New manga notification
- `POST /notify/new-chapter` - New chapter notification
- `POST /notify/manga-update` - Manga update notification

---

## License Information

### Project License
- **MangaHub:** Educational/Academic Project

### Third-Party Licenses
All dependencies are open-source with permissive licenses:
- **MIT License:** Most Go libraries (Gin, JWT, UUID, etc.)
- **BSD License:** PostgreSQL, Redis clients
- **Apache 2.0:** gRPC, Protocol Buffers

### External API Terms
- **MangaDex API:** https://api.mangadex.org/docs/legal/
- **AniList API:** https://anilist.co/terms
- **Kitsu API:** https://kitsu.docs.apiary.io/

---

## Version Information

| Component | Version | Release Date |
|-----------|---------|--------------|
| Go | 1.25 | Latest stable |
| PostgreSQL | 16.x | 2024 |
| Redis | 7.x | Latest |
| Docker | 24.x+ | Latest |
| Gin Framework | v1.11.0 | Latest |
| GORM | v1.31.0 | Latest |
| gRPC | v1.76.0 | Latest |

---

## Contact & Support

**Project Repository:** https://github.com/headtomatoes/mangahub  
**Documentation:** See `/docs` directory  
**Issues:** GitHub Issues  

---

**Document End**  
*For detailed implementation guides, see individual documentation files in the `/docs` directory.*
