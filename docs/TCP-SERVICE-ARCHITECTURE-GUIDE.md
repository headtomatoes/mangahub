# TCP Service - Architecture & Design Patterns

**Visual Guide to Go Idioms, Patterns, and Architecture**

---

## 📊 System Architecture

### High-Level Component Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      MANGAHUB ECOSYSTEM                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │  HTTP API    │  │  TCP Server  │  │  gRPC Server │        │
│  │  (Port 8080) │  │  (Port 8081) │  │  (Port 8083) │        │
│  └──────┬───────┘  └──────┬───────┘  └──────────────┘        │
│         │                  │                                    │
│         │                  │                                    │
│         └──────────┬───────┘                                    │
│                    │                                            │
│         ┌──────────┴──────────┐                                │
│         │                     │                                │
│    ┌────▼────┐          ┌────▼────┐                           │
│    │  Redis  │          │ Postgres│                           │
│    │ (Cache) │          │   (DB)  │                           │
│    └─────────┘          └─────────┘                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🏗️ TCP Service Internal Architecture

### Layer Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      APPLICATION LAYER                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              TCPServer (server.go)                      │   │
│  │  • Lifecycle Management                                 │   │
│  │  • Connection Acceptance                                │   │
│  │  • Graceful Shutdown                                    │   │
│  │  • Authentication Flow                                  │   │
│  └───────────────────────┬─────────────────────────────────┘   │
└────────────────────────────┼────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                      CONNECTION LAYER                           │
│  ┌────────────────────────▼──────────────────────────────────┐ │
│  │        ConnectionManager (manager.go)                     │ │
│  │  • Client Registry (map[string]*ClientConnection)        │ │
│  │  • Broadcast Orchestration                               │ │
│  │  • Concurrent Access Control (RWMutex)                   │ │
│  └────────────────────────┬──────────────────────────────────┘ │
│                           │                                     │
│  ┌────────────────────────▼──────────────────────────────────┐ │
│  │        ClientConnection (connection.go)                   │ │
│  │  • Per-Connection Goroutine                              │ │
│  │  • Buffered I/O (Reader/Writer)                          │ │
│  │  • Rate Limiting (50 msg/s)                              │ │
│  │  • Message Parsing & Routing                             │ │
│  └────────────────────────┬──────────────────────────────────┘ │
└────────────────────────────┼────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                      BUSINESS LOGIC LAYER                       │
│  ┌────────────────────────▼──────────────────────────────────┐ │
│  │        Message Handlers (handler.go)                      │ │
│  │  • HandleProgressMessage()                               │ │
│  │  • Validation & Authorization                            │ │
│  │  • Business Logic                                        │ │
│  └────────────────────────┬──────────────────────────────────┘ │
└────────────────────────────┼────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                      DATA ACCESS LAYER                          │
│  ┌────────────────────────▼──────────────────────────────────┐ │
│  │    ProgressRepository Interface (manager.go)             │ │
│  │  • SaveProgress()                                        │ │
│  │  • GetProgress()                                         │ │
│  │  • GetUserProgress()                                     │ │
│  │  • DeleteProgress()                                      │ │
│  └─────────┬──────────────────────────────────────┬─────────┘ │
│            │                                        │           │
│  ┌─────────▼──────────┐              ┌────────────▼─────────┐ │
│  │  ProgressRedisRepo │              │ HybridProgressRepo   │ │
│  │  (Redis Only)      │              │ (Redis + Postgres)   │ │
│  └────────────────────┘              └──────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Data Flow Patterns

### 1. Connection Lifecycle

```
┌─────────┐
│ Client  │
└────┬────┘
     │ 1. TCP Connect
     ▼
┌──────────────┐
│ net.Listener │
└────┬─────────┘
     │ 2. Accept()
     ▼
┌──────────────────────┐
│ handleConnection()   │ ◄─── New Goroutine Spawned
└────┬─────────────────┘
     │ 3. Authenticate (if AuthService exists)
     ▼
┌──────────────────────┐
│ authenticateClient() │
└────┬─────────────────┘
     │ 4a. Valid JWT
     ▼
┌──────────────────────┐
│ Manager.AddConnection│
└────┬─────────────────┘
     │ 5. Start Listening
     ▼
┌──────────────────────┐
│ client.Listen()      │ ◄─── Message Loop
└────┬─────────────────┘
     │ 6. On Disconnect
     ▼
┌──────────────────────┐
│Manager.RemoveConn    │
└──────────────────────┘
```

### 2. Message Processing Flow

```
Client sends JSON message
    │
    ▼
reader.ReadBytes('\n')
    │
    ▼
Size Check (< 1MB)
    │
    ▼
Rate Limit Check (50 msg/s)
    │
    ▼
json.Unmarshal(msg)
    │
    ▼
Switch on msg.Type
    │
    ├─ "progress_update" ──► HandleProgressMessage()
    │                              │
    │                              ├─ Validate Data
    │                              ├─ Check Authorization
    │                              ├─ SaveProgress()
    │                              └─ Broadcast()
    │
    └─ default ──► Broadcast to all clients
```

### 3. Write Path (Hybrid Storage)

```
HandleProgressMessage()
    │
    ▼
progressRepo.SaveProgress()
    │
    ├────────────────────────────────────────┐
    │                                        │
    ▼                                        ▼
Redis.SaveProgress()              Queue to Channel
(< 1ms, synchronous)               (buffered, 10K)
    │                                        │
    │                                        ▼
    │                              Background Batch Writer
    │                                        │
    │                                 ┌──────┴──────┐
    │                                 │             │
    │                            Every 5 min    OR  1000 records
    │                                 │             │
    │                                 └──────┬──────┘
    │                                        │
    │                                        ▼
    │                              Postgres.BatchInsert()
    │                              (50ms per 1000 records)
    │                                        │
    └────────────────────────────────────────┘
                     │
                     ▼
            Broadcast to clients
```

### 4. Read Path (Cache-Aside)

```
GetProgress(userID, mangaID)
    │
    ▼
Try Redis.GetProgress()
    │
    ├─ Hit (cached) ──► Return immediately
    │
    └─ Miss (not in cache)
        │
        ▼
    Try Postgres.GetProgress()
        │
        ├─ Found ──► Warm Redis cache (async)
        │            │
        │            └─► Return data
        │
        └─ Not found ──► Return nil
```

---

## 🎯 Design Patterns Used

### 1. **Repository Pattern**

**Purpose:** Abstract data access logic

```go
// Interface defines contract
type ProgressRepository interface {
    SaveProgress(data *ProgressData) error
    GetProgress(userID string, mangaID int64) (*ProgressData, error)
    GetUserProgress(userID string) ([]*ProgressData, error)
    DeleteProgress(userID string, mangaID int64) error
}

// Multiple implementations
type ProgressRedisRepo struct { ... }      // Redis only
type HybridProgressRepository struct { ... } // Redis + Postgres

// Usage
manager := NewConnectionManager(progressRepo) // Works with any implementation
```

**Benefits:**
- Testability (easy mocking)
- Flexibility (swap implementations)
- Separation of concerns

---

### 2. **Factory Pattern**

**Purpose:** Create objects with different configurations

```go
// Redis-only mode
func NewServer(addrTCP, addrRedis string) *TCPServer

// Hybrid mode with auth
func NewServerWithHybridStorage(addrTCP, addrRedis string, db *sql.DB, jwtSecret string) *TCPServer

// Mock mode for testing
func NewServerWithMockRedis(addrTCP string) *TCPServer
```

**Benefits:**
- Clear initialization
- Configuration encapsulation
- Multiple creation strategies

---

### 3. **Observer Pattern (Broadcast)**

**Purpose:** Notify multiple subscribers of events

```go
// Publisher
func (m *ConnectionManager) Broadcast(msg []byte, senderID string) {
    m.mu.RLock()
    clients := make([]*ClientConnection, 0, len(m.clients))
    for _, c := range m.clients {
        clients = append(clients, c)
    }
    m.mu.RUnlock()
    
    // Notify all subscribers concurrently
    var wg sync.WaitGroup
    for _, c := range clients {
        wg.Add(1)
        go func(client *ClientConnection) {
            defer wg.Done()
            client.Send(msg) // Observer receives notification
        }(c)
    }
    wg.Wait()
}
```

**Benefits:**
- Decoupling (publishers don't know subscribers)
- Real-time updates
- Scalable notification

---

### 4. **Strategy Pattern**

**Purpose:** Swap algorithms at runtime

```go
// Different storage strategies
type ProgressRepository interface { ... }

// Fast but temporary
type ProgressRedisRepo struct { ... }

// Durable but slower
type HybridProgressRepository struct { ... }

// Client code doesn't know which strategy is used
manager.progressRepo.SaveProgress(data)
```

**Benefits:**
- Algorithm flexibility
- Runtime selection
- Clean abstraction

---

### 5. **Registry Pattern**

**Purpose:** Central registry for objects

```go
type ConnectionManager struct {
    clients map[string]*ClientConnection // Registry
    mu      sync.RWMutex                  // Thread-safe access
}

// Register
func (m *ConnectionManager) AddConnection(client *ClientConnection) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.clients[client.ID] = client
}

// Lookup
func (m *ConnectionManager) GetConnection(id string) *ClientConnection {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.clients[id]
}
```

**Benefits:**
- Central access point
- O(1) lookup
- Easy iteration

---

### 6. **Command Pattern (Message Types)**

**Purpose:** Encapsulate requests as objects

```go
type Message struct {
    Type string         `json:"type"` // Command
    Data map[string]any `json:"data"` // Parameters
}

// Router
switch msg.Type {
case "progress_update":
    c.HandleProgressMessage(msg.Data) // Execute command
case "get_progress":
    c.HandleGetProgress(msg.Data)
default:
    c.Manager.Broadcast(payload, c.ID)
}
```

**Benefits:**
- Extensible (add new commands easily)
- Decoupled (handler doesn't know caller)
- Flexible routing

---

### 7. **Write-Through Cache**

**Purpose:** Keep cache and DB in sync

```go
func (r *HybridProgressRepository) SaveProgress(data *ProgressData) error {
    // 1. Write to cache immediately
    if err := r.redis.SaveProgress(data); err != nil {
        return err
    }
    
    // 2. Queue for DB write
    select {
    case r.writeChan <- data:
        // Successfully queued
    default:
        // Fallback to direct write
        ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
        defer cancel()
        r.postgres.SaveProgress(ctx, data)
    }
    
    return nil
}
```

**Benefits:**
- Fast reads (from cache)
- Data durability (DB backup)
- Eventual consistency

---

### 8. **Cache-Aside**

**Purpose:** Lazy-load cache on read

```go
func (r *HybridProgressRepository) GetProgress(userID string, mangaID int64) (*ProgressData, error) {
    // Try cache first
    data, err := r.redis.GetProgress(userID, mangaID)
    if err == nil && data != nil {
        return data, nil // Cache hit
    }
    
    // Cache miss - load from DB
    data, err = r.postgres.GetProgress(ctx, userID, mangaID)
    if err != nil {
        return nil, err
    }
    
    // Warm cache (async)
    if data != nil {
        go r.redis.SaveProgress(data)
    }
    
    return data, nil
}
```

**Benefits:**
- Reduced DB load
- Only caches what's needed
- Self-healing cache

---

## 🔒 Concurrency Patterns

### 1. **Goroutine per Connection**

```go
for {
    conn, err := listener.Accept()
    if err != nil {
        continue
    }
    
    s.wg.Add(1)
    go func(conn net.Conn) { // Dedicated goroutine
        defer s.wg.Done()
        s.handleConnection(conn)
    }(conn)
}
```

**Benefits:**
- Scales with connections (lightweight)
- Isolation (one client doesn't block others)
- Simple error handling

---

### 2. **RWMutex for Reader/Writer Lock**

```go
type ConnectionManager struct {
    clients map[string]*ClientConnection
    mu      sync.RWMutex // Read-write lock
}

// Many readers can hold lock simultaneously
func (m *ConnectionManager) Broadcast(msg []byte, senderID string) {
    m.mu.RLock() // Read lock (shared)
    defer m.mu.RUnlock()
    // ... iterate clients
}

// Only one writer at a time
func (m *ConnectionManager) AddConnection(client *ClientConnection) {
    m.mu.Lock() // Write lock (exclusive)
    defer m.mu.Unlock()
    m.clients[client.ID] = client
}
```

**Benefits:**
- High concurrency for reads
- Safe writes
- Prevents data races

---

### 3. **Channel-Based Pipeline**

```go
// Producer
func (r *HybridProgressRepository) SaveProgress(data *ProgressData) error {
    select {
    case r.writeChan <- data: // Send to channel
        return nil
    default:
        // Channel full - fallback
    }
}

// Consumer
func (r *HybridProgressRepository) StartBatchWriter(ctx context.Context) {
    for {
        select {
        case data := <-r.writeChan: // Receive from channel
            batch = append(batch, data)
            if len(batch) >= 1000 {
                r.flushBatch(batch)
            }
        case <-ticker.C:
            r.flushBatch(batch)
        }
    }
}
```

**Benefits:**
- Decoupled producer/consumer
- Backpressure handling
- Async processing

---

### 4. **WaitGroup for Synchronization**

```go
var wg sync.WaitGroup

for _, client := range clients {
    wg.Add(1)
    go func(c *ClientConnection) {
        defer wg.Done() // Signal completion
        c.Send(msg)
    }(client)
}

wg.Wait() // Wait for all goroutines
```

**Benefits:**
- Coordinated concurrent work
- Ensures all complete before proceeding
- Simple API

---

### 5. **Context for Cancellation**

```go
// Create cancellable context
ctx, cancel := context.WithCancel(context.Background())

// Start background worker
go hybridRepo.StartBatchWriter(ctx)

// Later: signal cancellation
cancel() // Worker stops gracefully

// Worker respects cancellation
func (r *HybridProgressRepository) StartBatchWriter(ctx context.Context) {
    for {
        select {
        case <-ctx.Done(): // Cancellation signal
            r.flushBatch(batch)
            return
        // ... other cases
        }
    }
}
```

**Benefits:**
- Graceful cancellation
- Timeout handling
- Request-scoped lifecycle

---

## 🛡️ Go Idioms Used

### 1. **Error Wrapping with %w**

```go
if err := r.redis.SaveProgress(data); err != nil {
    return fmt.Errorf("redis write failed: %w", err)
}
```

**Why:** Preserves error chain for `errors.Is()` and `errors.As()`

---

### 2. **Defer for Cleanup**

```go
func (m *ConnectionManager) AddConnection(client *ClientConnection) {
    m.mu.Lock()
    defer m.mu.Unlock() // Always released, even on panic
    m.clients[client.ID] = client
}
```

**Why:** Ensures resources are always released

---

### 3. **Type Assertions with ok Idiom**

```go
claims, ok := token.Claims.(jwt.MapClaims)
if !ok {
    return errors.New("invalid token claims")
}
```

**Why:** Safe type conversion without panic

---

### 4. **Buffered I/O**

```go
reader := bufio.NewReader(conn)  // Buffer reads
writer := bufio.NewWriter(conn)  // Buffer writes
writer.Flush()                   // Explicit flush
```

**Why:** Reduces system calls, improves performance

---

### 5. **Select for Multiplexing**

```go
select {
case data := <-r.writeChan:
    // Process data
case <-ticker.C:
    // Periodic task
case <-ctx.Done():
    // Cancellation
}
```

**Why:** Listen to multiple channels simultaneously

---

### 6. **Pointer Receivers**

```go
// Mutating method - use pointer
func (m *ConnectionManager) AddConnection(client *ClientConnection) {
    m.clients[client.ID] = client // Modifies m
}

// Non-mutating - could use value (but pointer for consistency)
func (m *ConnectionManager) GetCount() int {
    return len(m.clients)
}
```

**Why:** Efficient for large structs, enables mutation

---

## 📈 Performance Characteristics

### Latency Profile

| Operation | Typical Latency | Notes |
|-----------|----------------|-------|
| JWT Validation | < 5ms | CPU-bound |
| Redis Write | < 1ms | In-memory |
| Postgres Batch (1000) | ~50ms | Network + disk |
| Message Parse (JSON) | < 0.5ms | CPU-bound |
| Broadcast (1000 clients) | < 10ms | Parallel goroutines |
| Total per Message | < 5ms | Excluding broadcast |

### Throughput Limits

| Resource | Limit | Bottleneck |
|----------|-------|-----------|
| Connections | 10K+ | File descriptors |
| Messages/sec/conn | 50 | Rate limiter |
| Total msgs/sec | 500K+ | CPU/Network |
| Redis ops/sec | 100K | Single-threaded |
| Postgres writes/min | 12 batches | Batch interval |

---

## 🎓 Key Takeaways

### What Makes This Service Production-Grade?

1. **Graceful Shutdown**
   - Signals all goroutines
   - Waits for completion
   - Flushes pending data

2. **Resource Management**
   - Rate limiting per connection
   - Message size limits
   - Connection pooling
   - TTL on cached data

3. **Error Handling**
   - Proper error wrapping
   - Structured logging
   - Client error responses

4. **Concurrency Safety**
   - RWMutex for shared data
   - Channel for async processing
   - Context for cancellation

5. **Storage Strategy**
   - Fast cache for reads
   - Durable DB for writes
   - Batch processing for efficiency

---

**Next Steps:** Read [Critical Issues & Quick Fixes](./TCP-SERVICE-ISSUES-AND-FIXES.md)

**Author:** MentorDev  
**Date:** November 7, 2025
