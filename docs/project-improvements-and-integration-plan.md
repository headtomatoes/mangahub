# MangaHub Project Improvements & Service Integration Plan

## Executive Summary

This document outlines identified improvements for the MangaHub project and provides a comprehensive plan for integrating the TCP service with the HTTP-API service, including authentication, progress backup, and persistent storage.

---

## 1. Current Architecture Analysis

### 1.1 Existing Services
- **HTTP-API Service** (port 8080): REST API with JWT authentication, PostgreSQL database
- **TCP Service** (port 8081): Real-time progress tracking with Redis in-memory storage
- **Separate databases**: TCP uses Redis, HTTP-API uses PostgreSQL
- **No integration**: Services operate independently without communication

### 1.2 Identified Issues

#### Critical Issues
1. **No Authentication on TCP Service**: Anyone can connect and send progress updates
2. **Data Loss Risk**: Redis-only storage means data is lost on restart/failure
3. **Service Isolation**: No communication between TCP and HTTP-API services
4. **Incomplete Implementation**: `hybrid_progress.go` is commented out but contains the solution
5. **No Progress API**: HTTP-API has no endpoints to read/write progress data

#### Moderate Issues
1. **Code Duplication**: `ProgressData` struct only in TCP service, not shared
2. **Missing Models**: No PostgreSQL model for `user_progress` in HTTP-API
3. **No Service Discovery**: Services don't know about each other
4. **Limited Error Handling**: TCP service has basic error handling
5. **No Metrics/Monitoring**: No observability for TCP connections

#### Minor Issues
1. **Empty Files**: `handler.go`, `protocol.go`, `progress_postgress.go` are empty/stub files
2. **Inconsistent Naming**: "postgress" typo in `progress_postgress.go`
3. **No Graceful Degradation**: If Redis fails, TCP service fails completely
4. **No Rate Limiting per User**: Current rate limiting is per connection, not per user

---

## 2. Proposed Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Client Layer                         │
├─────────────────┬───────────────────────────────────────────┤
│   HTTP Client   │          TCP Client (WebSocket)           │
│   (REST API)    │        (Real-time Updates)                │
└────────┬────────┴──────────────────┬────────────────────────┘
         │                           │
         │ JWT Token                 │ JWT Token in handshake
         ▼                           ▼
┌────────────────────┐      ┌────────────────────┐
│   HTTP-API Service │◄────►│    TCP Service     │
│   (Port 8080)      │      │   (Port 8081)      │
│                    │      │                    │
│ - Auth/Register    │      │ - Auth via JWT     │
│ - CRUD Operations  │      │ - Real-time sync   │
│ - Progress API     │      │ - Progress updates │
└─────────┬──────────┘      └──────────┬─────────┘
          │                            │
          │ PostgreSQL                 │ Redis
          │ (Persistent)               │ (Cache)
          ▼                            ▼
┌────────────────────┐      ┌────────────────────┐
│   PostgreSQL DB    │◄─────│    Redis Cache     │
│                    │      │                    │
│ - Users            │      │ - Active Progress  │
│ - Manga            │      │ - Sessions         │
│ - Progress (Backup)│      │ - Real-time Data   │
└────────────────────┘      └────────────────────┘
         ▲
         │ Background Sync
         │ (Every 5 min)
         │
┌────────┴──────────┐
│  Hybrid Worker    │
│  - Batch writes   │
│  - Redis → PG     │
└───────────────────┘
```

---

## 3. Implementation Plan

### Phase 1: Foundation (Days 1-2)

#### 1.1 Create Shared Types Package
**File**: `internal/microservices/shared/types.go`
```go
package shared

import "time"

type ProgressData struct {
    UserID     string    `json:"user_id" db:"user_id"`
    MangaID    string    `json:"manga_id" db:"manga_id"`
    Chapter    int       `json:"chapter" db:"chapter"`
    Status     string    `json:"status" db:"status"`
    LastReadAt time.Time `json:"last_read_at" db:"last_read_at"`
}

type AuthClaims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
}
```

#### 1.2 Create Progress Model in HTTP-API
**File**: `internal/microservices/http-api/models/progress.go`
```go
package models

import "time"

type UserProgress struct {
    ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID        string    `gorm:"type:uuid;not null;index:idx_user_manga" json:"user_id"`
    MangaID       int64     `gorm:"not null;index:idx_user_manga" json:"manga_id"`
    CurrentChapter int      `gorm:"default:0" json:"current_chapter"`
    Status        string    `gorm:"type:text" json:"status"`
    UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}
```

#### 1.3 Create Progress Repository
**File**: `internal/microservices/http-api/repository/progress_repository.go`

#### 1.4 Create Progress Service
**File**: `internal/microservices/http-api/service/progress_service.go`

#### 1.5 Create Progress Handler
**File**: `internal/microservices/http-api/handler/progress_handler.go`

### Phase 2: HTTP-API Progress Endpoints (Day 3)

Add REST endpoints for progress management:
```
GET    /api/progress              - Get all user progress
GET    /api/progress/:manga_id     - Get specific manga progress
POST   /api/progress               - Update progress
DELETE /api/progress/:manga_id     - Delete progress
```

### Phase 3: TCP Authentication (Day 4)

#### 3.1 Add JWT Validation to TCP Service
Modify TCP handshake to:
1. Accept initial auth message with JWT token
2. Validate token using shared JWT secret
3. Extract user_id from token
4. Associate connection with user_id

#### 3.2 Implementation
**File**: `internal/microservices/tcp/auth.go`
```go
package tcp

import (
    "errors"
    "github.com/golang-jwt/jwt/v5"
)

type TCPAuthService struct {
    jwtSecret string
}

func (a *TCPAuthService) ValidateToken(tokenString string) (string, string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("invalid signing method")
        }
        return []byte(a.jwtSecret), nil
    })

    if err != nil || !token.Valid {
        return "", "", errors.New("invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return "", "", errors.New("invalid token claims")
    }

    userID := claims["user_id"].(string)
    username := claims["username"].(string)
    
    return userID, username, nil
}
```

### Phase 4: Implement Hybrid Storage (Days 5-7)

#### 4.1 Complete PostgreSQL Progress Repository for TCP
**File**: `internal/microservices/tcp/progress_postgres.go`
```go
package tcp

import (
    "database/sql"
    "fmt"
    "time"
)

type ProgressPostgresRepo struct {
    db *sql.DB
}

func NewProgressPostgresRepo(db *sql.DB) *ProgressPostgresRepo {
    return &ProgressPostgresRepo{db: db}
}

func (r *ProgressPostgresRepo) SaveProgress(data *ProgressData) error {
    query := `
        INSERT INTO user_progress (user_id, manga_id, current_chapter, status, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (user_id, manga_id) 
        DO UPDATE SET 
            current_chapter = EXCLUDED.current_chapter,
            status = EXCLUDED.status,
            updated_at = EXCLUDED.updated_at
    `
    _, err := r.db.Exec(query, data.UserID, data.MangaID, data.Chapter, data.Status, data.LastReadAt)
    return err
}

func (r *ProgressPostgresRepo) BatchInsert(batch []*ProgressData) error {
    if len(batch) == 0 {
        return nil
    }

    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT INTO user_progress (user_id, manga_id, current_chapter, status, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (user_id, manga_id) 
        DO UPDATE SET 
            current_chapter = EXCLUDED.current_chapter,
            status = EXCLUDED.status,
            updated_at = EXCLUDED.updated_at
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, data := range batch {
        _, err = stmt.Exec(data.UserID, data.MangaID, data.Chapter, data.Status, data.LastReadAt)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

#### 4.2 Implement Hybrid Repository
**File**: `internal/microservices/tcp/hybrid_progress.go` (uncomment and complete)
```go
package tcp

import (
    "context"
    "log/slog"
    "time"
)

type HybridProgressRepository struct {
    redis     *ProgressRedisRepo
    postgres  *ProgressPostgresRepo
    writeChan chan *ProgressData
    stopChan  chan struct{}
    logger    *slog.Logger
}

func NewHybridProgressRepository(redis *ProgressRedisRepo, postgres *ProgressPostgresRepo) *HybridProgressRepository {
    return &HybridProgressRepository{
        redis:     redis,
        postgres:  postgres,
        writeChan: make(chan *ProgressData, 10000), // Buffer for 10k updates
        stopChan:  make(chan struct{}),
        logger:    slog.Default(),
    }
}

// SaveProgress writes to Redis immediately, queues for PostgreSQL
func (r *HybridProgressRepository) SaveProgress(data *ProgressData) error {
    // 1. Write to Redis immediately (fast)
    if err := r.redis.SaveProgress(data); err != nil {
        r.logger.Error("redis_save_failed", "error", err)
        // Continue to queue for PostgreSQL
    }

    // 2. Queue for PostgreSQL batch write (async)
    select {
    case r.writeChan <- data:
    default:
        r.logger.Warn("write_queue_full", "user_id", data.UserID)
        // Queue full, force direct write
        go r.postgres.SaveProgress(data)
    }

    return nil
}

// GetProgress tries Redis first, falls back to PostgreSQL
func (r *HybridProgressRepository) GetProgress(userID, mangaID string) (*ProgressData, error) {
    // Try Redis first (fast)
    data, err := r.redis.GetProgress(userID, mangaID)
    if err == nil && data != nil {
        return data, nil
    }

    // Fallback to PostgreSQL
    // TODO: Implement GetProgress in PostgreSQL repo
    r.logger.Warn("redis_miss_fallback_to_postgres", "user_id", userID, "manga_id", mangaID)
    return nil, nil
}

// StartBatchWriter runs background worker for batch writes
func (r *HybridProgressRepository) StartBatchWriter(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    batch := make([]*ProgressData, 0, 1000)

    for {
        select {
        case <-ctx.Done():
            // Flush remaining batch before shutdown
            if len(batch) > 0 {
                r.flushBatch(batch)
            }
            return

        case data := <-r.writeChan:
            batch = append(batch, data)

            // Flush when batch is full
            if len(batch) >= 1000 {
                r.flushBatch(batch)
                batch = batch[:0]
            }

        case <-ticker.C:
            // Periodic flush every 5 minutes
            if len(batch) > 0 {
                r.flushBatch(batch)
                batch = batch[:0]
            }
        }
    }
}

func (r *HybridProgressRepository) flushBatch(batch []*ProgressData) {
    if err := r.postgres.BatchInsert(batch); err != nil {
        r.logger.Error("batch_insert_failed", "count", len(batch), "error", err)
    } else {
        r.logger.Info("batch_insert_success", "count", len(batch))
    }
}

func (r *HybridProgressRepository) Close() error {
    close(r.stopChan)
    close(r.writeChan)
    return r.redis.Close()
}
```

### Phase 5: Integration & Configuration (Day 8)

#### 5.1 Update TCP Server to Use Hybrid Storage
**File**: `internal/microservices/tcp/server.go`

Modify constructor to accept PostgreSQL connection and create hybrid repo.

#### 5.2 Update Configuration
**File**: `internal/config/config.go`

Add configuration for:
- Shared JWT secret between services
- PostgreSQL connection for TCP service
- Batch write interval configuration

#### 5.3 Update Main Entry Points
**File**: `cmd/tcp-server/main.go`

Connect to PostgreSQL and initialize hybrid repository.

### Phase 6: Testing & Validation (Days 9-10)

#### 6.1 Unit Tests
- Test JWT validation in TCP service
- Test hybrid repository (Redis + PostgreSQL)
- Test batch write functionality

#### 6.2 Integration Tests
- Test auth flow: Login → Get JWT → Connect TCP
- Test progress sync: TCP write → Verify in PostgreSQL
- Test Redis failure: Ensure PostgreSQL fallback works

#### 6.3 Load Tests
- Test 100+ concurrent TCP connections
- Test batch write performance
- Test Redis eviction scenarios

---

## 4. Quick Wins (Immediate Improvements)

### 4.1 Fix File Naming
Rename `progress_postgress.go` → `progress_postgres.go`

### 4.2 Add Index to Database
Add unique constraint for user_progress:
```sql
CREATE UNIQUE INDEX idx_user_manga ON user_progress(user_id, manga_id);
```

### 4.3 Add Health Checks
Add TCP health endpoint that reports:
- Active connections
- Redis status
- PostgreSQL status

### 4.4 Add Graceful Shutdown
Improve TCP server shutdown:
- Flush all pending writes to PostgreSQL
- Close Redis connection properly
- Wait for batch writer to finish

---

## 5. Security Enhancements

### 5.1 TCP Authentication Flow
```
Client                  TCP Server              Auth Service
  |                         |                         |
  |--1. Connect------------>|                         |
  |<-2. Auth Required-------|                         |
  |--3. {"type":"auth",---->|                         |
  |     "token":"jwt..."}   |                         |
  |                         |--4. Validate----------->|
  |                         |<-5. UserID, Username----|
  |<-6. Auth Success--------|                         |
  |--7. Normal Messages---->|                         |
```

### 5.2 Rate Limiting
Implement per-user rate limiting:
```go
type UserRateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
}
```

### 5.3 Connection Timeout
Add idle timeout for inactive connections (e.g., 30 minutes).

---

## 6. Monitoring & Observability

### 6.1 Metrics to Track
- Active TCP connections per user
- Progress updates per second
- Redis hit/miss ratio
- PostgreSQL batch write latency
- Authentication failures
- Rate limit violations

### 6.2 Logging Standards
Use structured logging with:
- `user_id` in all progress-related logs
- `connection_id` for TCP events
- `operation` field (save, get, delete)
- Latency measurements

---

## 7. Migration Strategy

### 7.1 Phase 1: Dual Write (Week 1)
- Deploy HTTP-API with progress endpoints
- Keep TCP service as-is (Redis only)
- Manually sync data from Redis to PostgreSQL

### 7.2 Phase 2: Hybrid Mode (Week 2)
- Deploy TCP service with hybrid storage
- Monitor for errors/performance issues
- Keep old Redis-only as fallback

### 7.3 Phase 3: Full Integration (Week 3)
- Enable authentication on TCP
- Remove Redis-only fallback
- Deploy to production

---

## 8. Code Quality Improvements

### 8.1 Add Context Support
Pass `context.Context` through all repository methods for:
- Request cancellation
- Timeouts
- Tracing

### 8.2 Add Error Types
Define custom error types for better error handling:
```go
var (
    ErrUserNotAuthenticated = errors.New("user not authenticated")
    ErrProgressNotFound     = errors.New("progress not found")
    ErrRedisUnavailable     = errors.New("redis unavailable")
    ErrPostgresUnavailable  = errors.New("postgres unavailable")
)
```

### 8.3 Add DTOs
Create Data Transfer Objects for API requests/responses.

---

## 9. Documentation Updates

### 9.1 API Documentation
Document new progress endpoints with examples.

### 9.2 TCP Protocol Documentation
Document TCP message format, including authentication.

### 9.3 Architecture Diagrams
Update architecture diagrams with new integration.

---

## 10. Summary of Benefits

After implementation:

✅ **Security**: TCP service requires JWT authentication
✅ **Reliability**: Progress data backed up to PostgreSQL
✅ **Performance**: Redis provides fast reads, PostgreSQL ensures durability
✅ **Scalability**: Batch writes reduce database load
✅ **Integration**: Services communicate securely
✅ **Observability**: Better logging and metrics
✅ **Maintainability**: Shared types, better structure

---

## Appendix A: File Structure After Implementation

```
internal/
├── shared/
│   └── types.go                    # NEW: Shared types
├── microservices/
│   ├── http-api/
│   │   ├── handler/
│   │   │   ├── auth_handler.go
│   │   │   ├── progress_handler.go  # NEW
│   │   │   └── ...
│   │   ├── models/
│   │   │   ├── progress.go          # NEW
│   │   │   └── ...
│   │   ├── repository/
│   │   │   ├── progress_repository.go # NEW
│   │   │   └── ...
│   │   └── service/
│   │       ├── progress_service.go  # NEW
│   │       └── ...
│   └── tcp/
│       ├── auth.go                  # NEW
│       ├── connection.go
│       ├── hybrid_progress.go       # UPDATED
│       ├── manager.go
│       ├── progress_postgres.go     # UPDATED
│       ├── progress_redis.go
│       └── server.go                # UPDATED
```

---

## Next Steps

1. **Review this plan** with the team
2. **Prioritize phases** based on business needs
3. **Set up development environment** with both Redis and PostgreSQL
4. **Create feature branch**: `feature/tcp-http-integration`
5. **Start with Phase 1**: Foundation and shared types
6. **Iterate and test** each phase thoroughly

---

**Document Version**: 1.0
**Last Updated**: November 5, 2025
**Author**: GitHub Copilot
