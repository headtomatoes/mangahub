# UDP Notification Service - Comprehensive Analysis

## 📋 Overview

**Purpose:** Real-time push notification system for MangaHub using UDP protocol

**Architecture Pattern:** Pub/Sub (Publisher-Subscriber) with persistent storage fallback

**Key Features:**
- ✅ Real-time notifications via UDP
- ✅ Subscriber management with activity tracking
- ✅ Offline user support (notification persistence)
- ✅ Selective broadcasting (library-based targeting)
- ✅ Automatic cleanup of inactive subscribers
- ✅ Missed notification synchronization

---

## 🏗️ Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        UDP Server                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Server     │  │ Subscriber   │  │ Broadcaster  │      │
│  │              │──│  Manager     │──│              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         │                  │                  │              │
│         │                  │                  │              │
│         ▼                  ▼                  ▼              │
│  ┌──────────────────────────────────────────────────┐       │
│  │         PostgreSQL (Notification Storage)         │       │
│  └──────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────┘
         ▲                                          │
         │ UDP Messages                             │ UDP Broadcasts
         │                                          ▼
    ┌─────────┐                              ┌─────────┐
    │ Client  │                              │ Client  │
    │ (User1) │                              │ (User2) │
    └─────────┘                              └─────────┘
```

### Core Components

#### 1. **Server** (`server.go`)
**Responsibility:** Main entry point, orchestrates all components

**Key Features:**
- UDP connection management
- Message routing (SUBSCRIBE/UNSUBSCRIBE/PING)
- Signal handling for graceful shutdown
- Missed notification synchronization

**Pattern:** Coordinator/Facade

#### 2. **SubscriberManager** (`subscriber.go`)
**Responsibility:** Manages active client subscriptions

**Key Features:**
- Thread-safe subscriber storage (RWMutex)
- Activity tracking (heartbeat/ping)
- Automatic cleanup of inactive subscribers
- Selective retrieval (by user ID, bulk)

**Pattern:** Registry/Manager with TTL

#### 3. **Broadcaster** (`broadcaster.go`)
**Responsibility:** Sends notifications to subscribers

**Key Features:**
- Concurrent message delivery (goroutines + WaitGroup)
- Dual storage (UDP + Database)
- Selective broadcasting (library-based)
- Automatic read-receipt marking

**Pattern:** Fan-out with persistence fallback

#### 4. **Notification** (`notification.go`)
**Responsibility:** Message protocol definition

**Key Features:**
- Strongly-typed notification types
- JSON serialization
- Factory methods for common notifications

**Pattern:** Value Object / DTO

---

## 🎯 Use Cases

### 1. New Manga Added
```go
server.NotifyNewManga(mangaID, title)
```
- Broadcasts to **all users**
- Stores notification for **all users** (online + offline)
- Online users receive via UDP immediately
- Offline users get it on next connection (sync)

### 2. New Chapter Released
```go
server.NotifyNewChapter(ctx, mangaID, title, chapterNum)
```
- Broadcasts to **library users only** (users who follow this manga)
- Stores notification for all library users
- Selective targeting reduces spam

### 3. User Connection
```
Client → SUBSCRIBE → Server
Server → Confirmation → Client
Server → Syncs missed notifications → Client
```

### 4. Heartbeat/Keep-Alive
```
Client → PING → Server
Server → PONG → Client
Server updates LastSeen timestamp
```

---

## 💡 Go Idioms & Best Practices

### ✅ Excellent Go Patterns Used

#### 1. **Concurrent Safe Collections**
```go
type SubscriberManager struct {
    mu          sync.RWMutex
    subscribers map[string]*Subscriber
}

// Read lock for read operations
func (sm *SubscriberManager) GetAll() []*Subscriber {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    // ... read operation
}

// Write lock for modifications
func (sm *SubscriberManager) Add(userID string, addr *net.UDPAddr) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    // ... write operation
}
```
**Why Good:**
- RWMutex allows multiple concurrent readers
- `defer` ensures unlock even on panic
- Clean separation of read/write operations

#### 2. **Context Propagation**
```go
func (s *Server) NotifyNewChapter(ctx context.Context, mangaID int64, ...) error {
    // Context passed through the call chain
}

func (s *Server) syncMissedNotifications(userID string, addr *net.UDPAddr) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    // Timeout protection
}
```
**Why Good:**
- Enables cancellation and timeout control
- Follows Go context conventions
- Prevents goroutine leaks

#### 3. **Graceful Shutdown**
```go
func (s *Server) Start() error {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
    
    log.Println("Shutting down UDP server...")
    return s.Shutdown()
}

func (s *Server) Shutdown() error {
    close(s.done)  // Signal all goroutines
    return s.conn.Close()
}
```
**Why Good:**
- Handles SIGTERM/SIGINT properly
- Coordinates shutdown across components
- Prevents abrupt termination

#### 4. **Goroutine Coordination**
```go
func (b *Broadcaster) BroadcastToAll(notification *Notification) error {
    var wg sync.WaitGroup
    
    for _, sub := range subscribers {
        wg.Add(1)
        go func(s *Subscriber) {
            defer wg.Done()
            b.sendToSubscriber(s, data)
        }(s)
    }
    
    wg.Wait()  // Wait for all sends to complete
    return nil
}
```
**Why Good:**
- Concurrent broadcast for performance
- WaitGroup ensures all sends complete
- Proper variable capture in closure

#### 5. **Factory Methods**
```go
func NewMangaNotification(mangaID int64, title string) *Notification {
    return &Notification{
        Type:      NotificationNewManga,
        MangaID:   mangaID,
        Title:     title,
        Message:   "New manga added: " + title,
        Timestamp: time.Now(),
    }
}
```
**Why Good:**
- Encapsulates creation logic
- Ensures consistent initialization
- Self-documenting API

#### 6. **Periodic Cleanup**
```go
func (sm *SubscriberManager) StartCleanupRoutine(interval time.Duration, done <-chan struct{}) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            sm.CleanupInactive()
        case <-done:
            return
        }
    }
}
```
**Why Good:**
- Ticker for periodic tasks
- Proper cleanup with defer
- Cancellable via done channel

#### 7. **Interface Segregation**
```go
type Broadcaster struct {
    libraryRepo      repository.LibraryRepository
    notificationRepo repository.NotificationRepository
    userRepo         repository.UserRepository
}
```
**Why Good:**
- Depends on interfaces, not concrete types
- Enables easy testing with mocks
- Follows Dependency Inversion Principle

#### 8. **Error Wrapping**
```go
func NewServer(port string, ...) (*Server, error) {
    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        return nil, fmt.Errorf("failed to listen on UDP: %w", err)
    }
    // ...
}
```
**Why Good:**
- Adds context to errors
- Preserves original error for unwrapping
- Clear error messages
