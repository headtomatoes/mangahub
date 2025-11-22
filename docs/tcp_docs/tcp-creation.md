# Key Design Patterns

## SERVER.go

Constructor pattern - NewServer() ensures proper initialization

Graceful shutdown - quitChan + wg.Wait() combination

Concurrent handling - Each client connection runs in its own goroutine

Resource cleanup - defer statements and explicit removal from manager

Closure safety - Passing conn as parameter avoids race conditions

## MANAGER.go

Thread-safety with RWMutex:

Write operations (Add, Remove, CloseAll) use Lock()

Read operations (Broadcast) use RLock() for better concurrency

Defer pattern:

defer Unlock() ensures locks are always released, preventing deadlocks

Map-based registry:

Fast O(1) lookups, additions, and deletions by client ID

Separation of concerns:

BroadcastSystemMessage handles message formatting

Broadcast handles the actual distribution logic

Safe iteration:

Locking during iteration prevents concurrent modification issues

## CONNECTION.go

Buffered I/O:

bufio.Reader and bufio.Writer for performance optimization

Reduces system calls by batching network operations

Newline-delimited protocol:

Each message ends with \n

Simple, text-friendly, human-readable during debugging

JSON-based messaging:

Structured, flexible message format

Type field enables message routing via switch statement

Error resilience:

Malformed JSON doesn't disconnect the client (continue)

Connection errors properly clean up via defer

Broadcast pattern:

Client receives message → processes it → broadcasts to all clients

Real-time update distribution (like a chat or progress tracker)

Encapsulation:

conn is private; external code uses Send() and Close()

Ensures proper buffering and flushing

## MESSAGE.go

Struct Tag: `json:"type"`
What it does: This is a struct tag that tells Go's json package how to map this field when encoding/decoding JSON.

When unmarshaling (JSON → Go struct): looks for a JSON field named "type"

When marshaling (Go struct → JSON): creates a JSON field named "type"

Example JSON mapping:

```json
{
  "type": "progress_update",
  "data": { ... }
}
```

↕️

``` go
Message{
    Type: "progress_update",
    Data: map[string]interface{}{ ... }
}
```

Without the tag, Go would expect the JSON field to be capitalized as "Type" (matching the struct field name).
