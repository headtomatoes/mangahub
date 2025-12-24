# MangaHub Use Cases Documentation

**Last Updated:** December 1, 2025  
**Version:** 1.2  
**Branch:** Websocket-server

## Table of Contents
1. [Authentication Use Cases](#authentication-use-cases)
2. [Manga Library Use Cases](#manga-library-use-cases)
3. [Progress Tracking Use Cases](#progress-tracking-use-cases)
4. [TCP Sync Use Cases](#tcp-sync-use-cases)
5. [gRPC Service Use Cases](#grpc-service-use-cases)
6. [WebSocket Real-Time Chat Use Cases](#websocket-real-time-chat-use-cases)
7. [Data Flow Architecture](#data-flow-architecture)
8. [API Endpoints Reference](#api-endpoints-reference)

---

## Authentication Use Cases

### UC-AUTH-001: User Registration
**Actor:** New User  
**Preconditions:** None  
**Flow:**
1. User provides username, email, and password
2. System validates input (username unique, email format, password strength)
3. System hashes password with bcrypt
4. System stores user in PostgreSQL database
5. System generates JWT access token (1h) and refresh token (7d)
6. System returns tokens and user profile

**CLI Command:**
```bash
mangahub auth register --username <username> --email <email> --password <password>
```

**HTTP Endpoint:**
```
POST /api/auth/register
Body: {
  "username": "string",
  "email": "string", 
  "password": "string"
}
Response: {
  "user_id": "uuid",
  "username": "string",
  "access_token": "jwt_token",
  "refresh_token": "jwt_token",
  "expires_at": "timestamp"
}
```

**Success Criteria:**
- User record created in database
- JWT tokens generated and returned
- Credentials stored in system keyring (CLI)

---

### UC-AUTH-002: User Login
**Actor:** Registered User  
**Preconditions:** User account exists  
**Flow:**
1. User provides username and password
2. System retrieves user from database
3. System verifies password with bcrypt
4. System generates new JWT tokens
5. System returns tokens and user profile
6. CLI stores credentials in system keyring

**CLI Command:**
```bash
mangahub auth login --username <username> --password <password>
```

**HTTP Endpoint:**
```
POST /api/auth/login
Body: {
  "username": "string",
  "password": "string"
}
Response: {
  "user_id": "uuid",
  "username": "string",
  "access_token": "jwt_token",
  "refresh_token": "jwt_token",
  "expires_at": "timestamp"
}
```

**Success Criteria:**
- Valid JWT tokens issued
- Session established
- Access token stored for subsequent requests

---

### UC-AUTH-003: Token Refresh
**Actor:** Authenticated User  
**Preconditions:** Valid refresh token exists  
**Flow:**
1. User's access token expires (1 hour)
2. CLI automatically detects expired token
3. CLI sends refresh token to server
4. Server validates refresh token
5. Server issues new access token
6. CLI updates stored credentials

**CLI Behavior:**
- Automatic refresh on token expiry
- Transparent to user (no manual intervention)

**HTTP Endpoint:**
```
POST /api/auth/refresh
Headers: {
  "Authorization": "Bearer <refresh_token>"
}
Response: {
  "access_token": "jwt_token",
  "expires_at": "timestamp"
}
```

**Success Criteria:**
- New access token issued without re-login
- Session continuity maintained

---

### UC-AUTH-004: User Logout
**Actor:** Authenticated User  
**Preconditions:** User is logged in  
**Flow:**
1. User initiates logout
2. CLI deletes credentials from system keyring
3. Session terminated locally

**CLI Command:**
```bash
mangahub auth logout
```

**Success Criteria:**
- Credentials removed from keyring
- User must login again for authenticated operations

---

### UC-AUTH-005: Check Authentication Status
**Actor:** Any User  
**Preconditions:** None  
**Flow:**
1. User checks current authentication status
2. CLI retrieves credentials from keyring
3. System displays current user and token status

**CLI Command:**
```bash
mangahub auth status
```

**Output:**
```
Authentication Status:
  Status: ✓ Logged in
  Username: testcli
  User ID: 71dfb4c1-44a5-4b3c-aa5e-863b9f62b9a8
  Token expires: 2025-11-25 13:42:18
```

**Success Criteria:**
- Current auth state displayed accurately

---

## Manga Library Use Cases

### UC-LIBRARY-001: Search Manga
**Actor:** Authenticated User  
**Preconditions:** User is logged in  
**Flow:**
1. User provides search query
2. System searches manga database by title/slug
3. System returns matching manga with details

**CLI Command:**
```bash
mangahub library search "<query>"
```

**HTTP Endpoint:**
```
GET /api/manga?search=<query>
Headers: {
  "Authorization": "Bearer <access_token>"
}
Response: [{
  "id": 1,
  "title": "Manga Title",
  "slug": "manga-title",
  "total_chapters": 100,
  "status": "ongoing",
  "created_at": "timestamp"
}]
```

**Success Criteria:**
- Relevant manga returned
- Partial matches supported

---

### UC-LIBRARY-002: Get Manga Details by ID
**Actor:** Authenticated User  
**Preconditions:** User is logged in, manga exists  
**Flow:**
1. User requests manga by ID
2. System retrieves manga from database
3. System returns complete manga details

**CLI Command:**
```bash
mangahub library get --manga-id <id>
```

**HTTP Endpoint:**
```
GET /api/manga/<id>
Headers: {
  "Authorization": "Bearer <access_token>"
}
Response: {
  "id": 1,
  "title": "Manga Title",
  "slug": "manga-title",
  "total_chapters": 100,
  "status": "ongoing",
  "created_at": "timestamp"
}
```

**Success Criteria:**
- Complete manga details returned
- 404 if manga not found

---

## Progress Tracking Use Cases

### UC-PROGRESS-001: Update Reading Progress (HTTP API)
**Actor:** Authenticated User  
**Preconditions:** User is logged in, manga exists  
**Flow:**
1. User specifies manga ID, chapter, and status
2. System validates input (chapter number, status enum)
3. System extracts user ID from JWT token
4. System checks for backward progress (unless --force)
5. System checks chapter doesn't exceed total chapters
6. System upserts progress in PostgreSQL
7. System returns updated progress

**CLI Command:**
```bash
mangahub progress update --manga-id <id> --chapter <num> --status <status>

# Options:
--status: reading|completed|plan_to_read|dropped (default: reading)
--force: Allow backward progress updates
```

**HTTP Endpoint:**
```
POST /api/progress/<manga_id>
Headers: {
  "Authorization": "Bearer <access_token>"
}
Body: {
  "chapter": 105,
  "status": "reading"
}
Response: {
  "user_id": "uuid",
  "manga_id": 1,
  "chapter": 105,
  "status": "reading",
  "updated_at": "2025-11-25T05:39:18Z"
}
```

**Validation Rules:**
- Chapter must be >= 0
- Chapter cannot exceed manga's total_chapters (unless forced)
- Status must be one of: reading, completed, plan_to_read, dropped
- Cannot update backward progress without --force flag
- User can only update their own progress (JWT user_id)

**Success Criteria:**
- Progress saved to PostgreSQL
- Progress returned with timestamp
- Foreign key constraints enforced (user_id, manga_id)

---

### UC-PROGRESS-002: Get Progress History
**Actor:** Authenticated User  
**Preconditions:** User is logged in  
**Flow:**
1. User requests progress history (all manga or specific manga)
2. System extracts user ID from JWT
3. System retrieves progress from PostgreSQL
4. System returns sorted progress list

**CLI Command:**
```bash
# All progress
mangahub progress history

# Specific manga
mangahub progress history --manga-id <id>
```

**HTTP Endpoint:**
```
GET /api/progress
Headers: {
  "Authorization": "Bearer <access_token>"
}
Response: {
  "history": [{
    "user_id": "uuid",
    "manga_id": 1,
    "chapter": 105,
    "status": "reading",
    "updated_at": "2025-11-25T05:39:18Z"
  }],
  "total": 2
}
```

**CLI Output:**
```
Progress History (All Manga)
[1] Chapter 105 - 2025-11-25T05:39:18Z
    Status: reading

[2] Chapter 10 - 2025-11-25T05:22:48Z
    Status: dropped

Total entries: 2
```

**Success Criteria:**
- All user's progress returned
- Sorted by updated_at DESC
- Empty array if no progress

---

### UC-PROGRESS-003: Get Single Manga Progress
**Actor:** Authenticated User  
**Preconditions:** User is logged in, progress exists  
**Flow:**
1. User requests progress for specific manga
2. System extracts user ID from JWT
3. System retrieves progress from PostgreSQL
4. System returns progress or 404

**HTTP Endpoint:**
```
GET /api/progress/<manga_id>
Headers: {
  "Authorization": "Bearer <access_token>"
}
Response: {
  "user_id": "uuid",
  "manga_id": 1,
  "chapter": 105,
  "status": "reading",
  "updated_at": "2025-11-25T05:39:18Z"
}
```

**Success Criteria:**
- Progress returned if exists
- 404 if no progress found

---

## TCP Sync Use Cases

### UC-TCP-001: Connect to TCP Sync Server
**Actor:** Authenticated User  
**Preconditions:** User is logged in, TCP server running  
**Flow:**
1. User initiates TCP connection
2. CLI establishes TCP connection to localhost:8081
3. CLI sends authentication message with JWT token
4. Server validates JWT token
5. Server extracts user_id and username from token
6. Server assigns session ID
7. Server sends auth_success response
8. Connection established for real-time sync

**CLI Command:**
```bash
mangahub sync connect

# Optional: specify server
mangahub sync connect --server <host:port>
```

**TCP Protocol:**
```json
// Client sends:
{
  "type": "auth",
  "data": {
    "username": "testcli",
    "token": "jwt_access_token"
  },
  "timestamp": "2025-11-25T05:38:59Z"
}

// Server responds:
{
  "type": "auth_success",
  "data": {
    "session_id": "b69be845-9114-4e95-abeb-a0d11ca70d0a",
    "user_id": "71dfb4c1-44a5-4b3c-aa5e-863b9f62b9a8",
    "username": "testcli"
  }
}
```

**CLI Output:**
```
✓ Connected successfully!

Connection Details:
  Server: localhost:8081
  User: testcli
  Session ID: b69be845-9114-4e95-abeb-a0d11ca70d0a
  Connected at: 2025-11-25 12:38:59 +07

Sync Status:
  Auto-sync: enabled
  Conflict resolution: last_write_wins
  Real-time sync is now active.
```

**Success Criteria:**
- TCP connection established
- JWT validated
- Session ID assigned
- Client authenticated on server

---

### UC-TCP-002: Sync Progress via TCP
**Actor:** Authenticated User  
**Preconditions:** User is logged in  
**Flow:**
1. User initiates progress sync command
2. CLI establishes temporary TCP connection
3. CLI authenticates with JWT
4. CLI sends progress_update message
5. Server validates user authorization
6. Server saves progress to Redis (immediate)
7. Server queues progress for PostgreSQL batch write
8. Server responds with success
9. CLI disconnects

**CLI Command:**
```bash
mangahub progress sync --manga-id <id> --chapter <num> --status <status>
```

**TCP Protocol:**
```json
// Client sends:
{
  "type": "progress_update",
  "data": {
    "user_id": "71dfb4c1-44a5-4b3c-aa5e-863b9f62b9a8",
    "manga_id": 1,
    "chapter": 110,
    "status": "reading"
  },
  "timestamp": "2025-11-25T05:44:51Z"
}

// Server logs:
{
  "level": "INFO",
  "msg": "progress_saved",
  "client_id": "c1c866d8-6c71-4c72-a624-27a163eabdee",
  "user_id": "71dfb4c1-44a5-4b3c-aa5e-863b9f62b9a8",
  "manga_id": 1,
  "chapter": 110
}
```

**Data Flow:**
```
CLI → TCP Server → Redis (immediate) → PostgreSQL (batched every 5 min)
```

**Storage Architecture:**
- **Redis**: Fast cache, immediate updates, 90-day TTL
- **PostgreSQL**: Persistent storage, batch writes (1000 records or 5 minutes)

**Redis Key Format:**
```
progress:user:<user_id>:manga:<manga_id>
```

**Redis Hash Fields:**
```
user_id: "71dfb4c1-44a5-4b3c-aa5e-863b9f62b9a8"
manga_id: "1"
current_chapter: "110"
status: "reading"
updated_at: "2025-11-25T05:44:51.537286567Z"
```

**Success Criteria:**
- Progress saved to Redis immediately
- Progress queued for PostgreSQL
- Foreign key constraints enforced on batch write
- User can only update their own progress

---

### UC-TCP-003: Check TCP Sync Status
**Actor:** Any User  
**Preconditions:** None  
**Flow:**
1. User checks TCP connection status
2. CLI displays connection state and statistics

**CLI Command:**
```bash
mangahub sync status
```

**Output (Not Connected):**
```
TCP Sync Status:
  Connection: ✗ Not connected
  Server: localhost:8081

To connect:
  mangahub sync connect
```

**Output (Connected - via persistent connection):**
```
TCP Sync Status:
  Connection: ✓ Active
  Server: localhost:8081
  Uptime: 5m 32s
  Last heartbeat: 15s ago

Session Info:
  User: testcli
  Session ID: b69be845-9114-4e95-abeb-a0d11ca70d0a

Sync Statistics:
  Messages sent: 3
  Messages received: 5
  Last sync: 2m 10s ago
  Sync conflicts: 0

Network Quality: Excellent (RTT: 15ms)
```

**Success Criteria:**
- Current connection state displayed
- Statistics shown if connected

---

### UC-TCP-004: Disconnect from TCP Server
**Actor:** Authenticated User (connected)  
**Preconditions:** TCP connection active  
**Flow:**
1. User initiates disconnect
2. CLI sends disconnect message
3. CLI closes TCP connection
4. Server removes client from connection manager

**CLI Command:**
```bash
mangahub sync disconnect
```

**Success Criteria:**
- Connection closed gracefully
- Server cleaned up client session

---

### UC-TCP-005: Monitor Real-time Sync Updates
**Actor:** Authenticated User (connected)  
**Preconditions:** TCP connection active  
**Flow:**
1. User starts monitoring mode
2. CLI listens for sync messages
3. System displays real-time updates from all devices
4. User stops monitoring with Ctrl+C

**CLI Command:**
```bash
mangahub sync monitor
```

**Output:**
```
Monitoring real-time sync updates... (Press Ctrl+C to exit)

[15:04:05] ← Device 'Mobile' updated: One Piece → Chapter 1050
[15:04:12] → Broadcasting update: Naruto → Chapter 700
[15:04:23] ⚠ Conflict resolved: Bleach → Chapter 500 (local) vs 498 (remote)
```

**Success Criteria:**
- Real-time updates displayed
- Direction indicators (incoming/outgoing/conflict)
- Graceful exit on Ctrl+C

---

## gRPC Service Use Cases

### UC-GRPC-001: Query Manga by ID via gRPC
**Actor:** Authenticated User  
**Preconditions:** gRPC server running, manga exists  
**Flow:**
1. User requests manga details by ID via gRPC
2. CLI establishes gRPC connection to server
3. CLI sends GetManga request with manga ID
4. Server retrieves manga from database (via repository)
5. Server returns manga details with genres and metadata
6. CLI displays formatted manga information
7. CLI closes gRPC connection

**CLI Command:**
```bash
mangahub grpc manga get --id <manga-id>

# Optional: specify gRPC server address
mangahub grpc manga get --id <manga-id> --grpc-addr <host:port>
```

**Example:**
```bash
mangahub grpc manga get --id 10

# Output:
📖 Manga Details (via gRPC)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ID:          10
Title:       Bleach
Description: Ichigo Kurosaki obtains the powers of a Soul Reaper...
Authors:     [Kubo Tite]
Genres:      [Action Adventure Supernatural Shounen]
Cover URL:   https://example.com/bleach.jpg
Chapters:    686
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**gRPC Protocol:**
```protobuf
// Request
message GetMangaRequest {
    int64 manga_id = 1;
}

// Response
message GetMangaResponse {
    Manga manga = 1;
}

message Manga {
    int64 id = 1;
    string title = 2;
    string description = 3;
    repeated string authors = 4;
    repeated string genres = 5;
    string cover_url = 6;
    int32 chapters_count = 7;
}
```

**Success Criteria:**
- Manga details retrieved successfully
- All fields populated correctly
- Genres and authors included
- Connection automatically closed

---

### UC-GRPC-002: Search Manga via gRPC
**Actor:** Authenticated User  
**Preconditions:** gRPC server running  
**Flow:**
1. User provides search query string
2. CLI establishes gRPC connection
3. CLI sends SearchManga request with query and pagination
4. Server searches manga database by title
5. Server applies pagination (limit/offset)
6. Server returns matching manga list with total count
7. CLI displays formatted search results
8. CLI closes gRPC connection

**CLI Command:**
```bash
mangahub grpc manga search --query "<search-term>"

# With pagination
mangahub grpc manga search --query "<search-term>" --limit <N> --offset <N>

# Optional: specify gRPC server
mangahub grpc manga search --query "<search-term>" --grpc-addr <host:port>
```

**Example:**
```bash
mangahub grpc manga search --query "one piece"

# Output:
🔍 Search Results (via gRPC): "one piece"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total matches: 1
Showing: 1 results (offset: 0)

1. [ID: 1] One Piece
   The story follows Monkey D. Luffy, a young man whose body...
   Genres: [Action Adventure Comedy Fantasy Shounen]
   Chapters: 1095

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**gRPC Protocol:**
```protobuf
// Request
message SearchRequest {
    string query = 1;
    int32 limit = 2;    // Default: 20
    int32 offset = 3;   // Default: 0
}

// Response
message SearchResponse {
    repeated Manga mangas = 1;
    int64 total_count = 2;
}
```

**Pagination:**
- Default limit: 20 results
- Default offset: 0
- Server enforces maximum limit
- Total count returned for pagination UI

**Success Criteria:**
- Relevant manga returned
- Partial title matching supported
- Pagination works correctly
- Total count accurate

---

### UC-GRPC-003: Update Progress via gRPC
**Actor:** Authenticated User  
**Preconditions:** User is logged in, manga exists  
**Flow:**
1. User specifies manga ID, chapter, and optional status
2. CLI retrieves user ID from stored credentials
3. CLI establishes gRPC connection
4. CLI sends UpdateProgress request with progress data
5. Server validates user authorization
6. Server updates/creates progress record in database
7. Server returns success response with message
8. CLI displays confirmation
9. CLI closes gRPC connection

**CLI Command:**
```bash
mangahub grpc progress update --manga-id <id> --chapter <number>

# With status
mangahub grpc progress update --manga-id <id> --chapter <number> --status <status>

# With title
mangahub grpc progress update --manga-id <id> --chapter <number> --title "<manga-title>"

# Full example
mangahub grpc progress update --manga-id 10 --chapter 686 --status completed --title "Bleach"
```

**Status Options:**
- `reading` (default)
- `completed`
- `on_hold`

**Example:**
```bash
mangahub grpc progress update --manga-id 10 --chapter 100 --status reading --title "Bleach"

# Output:
✅ Progress updated successfully (via gRPC)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Manga ID: 10
Title:    Bleach
Chapter:  100
Status:   reading
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**gRPC Protocol:**
```protobuf
// Request
message UpdateProgressRequest {
    string user_id = 1;
    int64 manga_id = 2;
    string manga_title = 3;  // Optional
    int32 chapter = 4;
    string status = 5;       // reading, completed, on_hold
}

// Response
message UpdateProgressResponse {
    bool success = 1;
    string message = 2;
}
```

**Validation Rules:**
- User must be logged in (user_id required)
- Chapter must be > 0
- Status must be valid enum value
- Manga must exist (foreign key constraint)
- User can only update their own progress

**Database Behavior:**
- Upsert operation (update if exists, create if not)
- Uses composite primary key (user_id, manga_id)
- Timestamp automatically updated

**Success Criteria:**
- Progress saved to database
- Success confirmation returned
- User sees formatted output
- Idempotent operation (can run multiple times)

---

### UC-GRPC-004: gRPC Connection Management
**Actor:** CLI Application  
**Preconditions:** gRPC server accessible  
**Flow:**
1. CLI creates gRPC client with configured address
2. Client establishes connection with 5-second timeout
3. Client validates connection is ready
4. Client performs RPC operations
5. Client closes connection after operation completes

**Connection Configuration:**
- **Default Address:** `localhost:8083`
- **Environment Variable:** `MANGAHUB_GRPC_ADDR`
- **Flag Override:** `--grpc-addr`
- **Connection Timeout:** 5 seconds
- **Operation Timeout:** 10 seconds per RPC
- **Transport:** Insecure credentials (development)

**Connection Priority:**
1. Command-line flag `--grpc-addr`
2. Environment variable `MANGAHUB_GRPC_ADDR`
3. Default `localhost:8083`

**Example Configuration:**
```bash
# Use default address
mangahub grpc manga get --id 1

# Use custom address via flag
mangahub grpc manga get --id 1 --grpc-addr localhost:9090

# Use custom address via environment
export MANGAHUB_GRPC_ADDR=production.example.com:8083
mangahub grpc manga get --id 1
```

**Error Handling:**
- Connection timeout: Clear error message
- Server unavailable: Connection failed error
- Invalid response: RPC error with details
- Network issues: Automatic timeout and cleanup

**Success Criteria:**
- Connection established within timeout
- RPC completes successfully
- Connection properly closed
- No resource leaks

---

## WebSocket Real-Time Chat Use Cases

### UC-WS-001: Join Chat Room via WebSocket
**Actor:** Authenticated User  
**Preconditions:** User logged in, API server running with WebSocket support  
**Flow:**
1. User requests to join chat room for specific manga
2. CLI retrieves stored authentication token from system keyring
3. CLI establishes WebSocket connection to `/ws` endpoint with JWT in Authorization header
4. Server validates JWT via AuthMiddleware
5. Server upgrades HTTP connection to WebSocket
6. Server extracts user_id and username from JWT claims
7. Server creates Client object and registers it in Hub
8. Server spawns ReadPump and WritePump goroutines for the client
9. Client sends join message with target room_id (manga_id)
10. Hub processes join request and creates room if it doesn't exist
11. Hub adds client to room's client map
12. Hub sends system message to client confirming room join
13. Hub broadcasts join notification to other room members (excludes joining user)
14. Client enters interactive message loop for real-time chat

**CLI Command:**
```bash
mangahub chat join --room <manga-id>

# Alternative: Specify token manually
mangahub chat join --room <manga-id> --token <jwt_token>
```

**Example:**
```bash
# Login first (stores credentials in system keyring)
mangahub auth login --username alice --password Pass123!

# Join chat room for manga ID 1 (One Piece)
mangahub chat join --room 1

# Output:
🔌 Connecting to chat room 1...
✅ Connected! Type your messages (or /quit to exit)

🔔 You have joined the chat room for manga ID 1. Users online: 1
```

**WebSocket Connection:**
```http
GET /ws HTTP/1.1
Host: localhost:8084
Upgrade: websocket
Connection: Upgrade
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Sec-WebSocket-Version: 13
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
```

**WebSocket Protocol:**
```json
// 1. Join Message (Client → Server)
{
    "type": "join",
    "room_id": 1
}

// 2. System Response to Joining User (Server → Client)
{
    "type": "system",
    "room_id": 1,
    "user_id": "system",
    "user_name": "System_Admin",
    "content": "You have joined the chat room for manga ID 1. Users online: 1",
    "timestamp": "2025-12-01T14:33:16Z"
}

// 3. Join Broadcast to Other Room Members (Server → Other Clients)
// Note: This is NOT sent to the joining user
{
    "type": "system",
    "room_id": 1,
    "user_id": "system",
    "user_name": "System_Admin",
    "content": "[alice] has joined the chat.",
    "timestamp": "2025-12-01T14:33:16Z"
}
```

**Multi-User Example:**
```bash
# Terminal 1 (Alice)
mangahub chat join --room 1
# Output:
🔔 You have joined the chat room for manga ID 1. Users online: 1

# Terminal 2 (Bob joins same room)
mangahub chat join --room 1

# Alice sees:
🔔 [bob] has joined the chat.

# Bob sees:
🔔 You have joined the chat room for manga ID 1. Users online: 2
```

**Server Architecture:**
- **Hub**: Central manager running in separate goroutine
  - Manages all client connections via map[clientID]*Client
  - Manages all rooms via map[roomID]*Room
  - Processes register/unregister/join/leave/broadcast channels
  
- **Client**: Per-connection handler
  - ReadPump goroutine: Reads messages from WebSocket
  - WritePump goroutine: Writes messages to WebSocket
  - SendChannel: Buffered channel for async message delivery
  - Ping/Pong heartbeat to keep connection alive

- **Room**: Per-manga chat room
  - Auto-created on first join
  - Maintains map of connected clients
  - Broadcasts messages to all room members

**Connection Parameters:**
- **Endpoint**: `ws://localhost:8084/ws`
- **Authentication**: JWT token in Authorization header
- **Ping Interval**: Every 54 seconds (90% of PongWait)
- **Pong Timeout**: 60 seconds
- **Write Timeout**: 10 seconds
- **Max Message Size**: 512 bytes
- **Send Channel Buffer**: 256 messages

**Success Criteria:**
- ✅ WebSocket connection established successfully
- ✅ JWT authentication passed
- ✅ Client registered in Hub
- ✅ Room created or joined successfully
- ✅ Client added to room's client map
- ✅ System messages received and displayed
- ✅ User count accurate
- ✅ Ready to send/receive chat messages

**Error Handling:**
- **401 Unauthorized**: Missing or invalid JWT token
  ```
  Error: connection failed: websocket: bad handshake
  ```
- **Connection Refused**: Server not running
  ```
  Error: connection failed: dial tcp [::1]:8084: connectex: No connection could be made...
  ```
- **Expired Token**: Re-login required
  ```
  Error: not logged in, please run 'mangahub auth login' first
  ```

---

### UC-WS-002: Send and Receive Chat Messages
**Actor:** Authenticated User in Chat Room  
**Preconditions:** User connected to chat room via WebSocket  
**Flow:**
1. User types message text in CLI terminal
2. CLI reads input from stdin (bufio.Scanner)
3. CLI constructs chat message JSON with type "chat"
4. CLI sends message via WebSocket connection
5. Server's ReadPump receives message from WebSocket
6. Server parses JSON and validates message type
7. Server enriches message with user_id, user_name from client
8. Server uses client's current room_id
9. Server sends message to Hub's Broadcast channel
10. Hub looks up room by room_id
11. Hub calls room.Broadcast() to send to all clients
12. Room iterates through clients and sends to each SendChannel
13. Each client's WritePump writes message to WebSocket
14. All connected CLIs receive and display formatted message

**Message Input:**
```
# User types in terminal (no command prefix for chat)
Hello everyone! Has anyone read chapter 1095 yet?
```

**WebSocket Protocol:**
```json
// Chat Message (Client → Server)
{
    "type": "chat",
    "content": "Hello everyone! Has anyone read chapter 1095 yet?"
}

// Note: Client does NOT send room_id, user_id, or user_name
// Server fills these in from the connection context

// Broadcast Message (Server → All Clients in Room)
{
    "type": "chat",
    "room_id": 1,
    "user_id": "2c8a250e-b03a-4853-9643-89e6f2f17e7c",
    "user_name": "alice",
    "content": "Hello everyone! Has anyone read chapter 1095 yet?",
    "timestamp": "2025-12-01T14:35:42Z"
}
```

**Client Display (Color-Coded):**
```
[alice] Hello everyone! Has anyone read chapter 1095 yet?
[bob] Not yet! Waiting for the official release
[charlie] It was amazing! 🔥
```

**Message Flow Sequence:**
```
┌──────────┐                ┌─────────┐              ┌──────┐             ┌──────┐
│   CLI    │                │ Client  │              │ Hub  │             │ Room │
│  (User)  │                │ReadPump │              │      │             │      │
└────┬─────┘                └────┬────┘              └───┬──┘             └───┬──┘
     │                           │                       │                    │
     │ Type: "hello"             │                       │                    │
     ├──────────────────────────>│                       │                    │
     │                           │                       │                    │
     │                           │ Parse JSON            │                    │
     │                           │ Add user_id, name     │                    │
     │                           │ Set room_id           │                    │
     │                           │                       │                    │
     │                           │ Hub.Broadcast <- msg  │                    │
     │                           ├──────────────────────>│                    │
     │                           │                       │                    │
     │                           │                       │ room.Broadcast(msg)│
     │                           │                       ├───────────────────>│
     │                           │                       │                    │
     │                           │                       │                    │ For each client:
     │                           │                       │                    │   client.SendMessage(msg)
     │                           │<──────────────────────┴────────────────────┤
     │                           │ WritePump receives    │                    │
     │<──────────────────────────┤ from SendChannel      │                    │
     │ Display: [alice] hello    │                       │                    │
     │                           │                       │                    │
```

**Real-Time Characteristics:**
- **Latency**: < 100ms for local network
- **Delivery**: All room members receive message simultaneously
- **Order**: Messages displayed in chronological order
- **Persistence**: Messages are NOT stored (real-time only)

**Success Criteria:**
- ✅ Message sent successfully without errors
- ✅ All room members receive message
- ✅ Username, room_id, and timestamp correctly populated
- ✅ Messages displayed in chronological order
- ✅ No message loss or duplication
- ✅ Color-coded display for better UX

**Special Commands:**
```bash
/quit   # Exit chat and close WebSocket connection gracefully
```

**Message Validation:**
- Maximum message size: 512 bytes
- Empty messages: Sent but may appear as blank line
- Special characters: Supported (UTF-8 encoding)

**Error Scenarios:**
- **Send Channel Full**: Message dropped, warning logged
  ```
  WARN Send channel full, unable to send message client_id=...
  ```
- **Client Not in Room**: Message ignored (shouldn't happen with proper join)
- **Connection Lost**: ReadPump detects close, client unregistered
- **Write Failure**: WritePump exits, connection closed

---

### UC-WS-003: Leave Chat Room
**Actor:** User in Chat Room  
**Preconditions:** User connected to WebSocket chat  
**Flow:**
1. User types `/quit` command or presses Ctrl+C
2. CLI intercepts quit signal
3. CLI constructs leave message with current room_id
4. CLI sends leave message via WebSocket (best effort)
5. CLI sends WebSocket close frame with normal closure code
6. CLI closes WebSocket connection
7. Server's ReadPump detects connection close
8. Server removes client from room via Hub.LeaveRoom channel
9. Server removes client from Hub via Hub.Unregister channel
10. Server closes client's SendChannel
11. Server broadcasts leave notification to remaining room members
12. CLI returns to command prompt

**User Actions:**
```bash
# Method 1: Type /quit command
/quit

# Method 2: Press Ctrl+C (interrupt signal)
^C

# Both result in clean disconnect
```

**WebSocket Protocol:**
```json
// 1. Leave Message (Client → Server)
{
    "type": "leave",
    "room_id": 1
}

// 2. WebSocket Close Frame
// Code: 1000 (Normal Closure)
// Reason: ""

// 3. Leave Broadcast (Server → Remaining Room Members)
{
    "type": "system",
    "room_id": 1,
    "user_id": "system",
    "user_name": "System_Admin",
    "content": "[alice] has left the chat.",
    "timestamp": "2025-12-01T14:40:15Z"
}
```

**CLI Output:**
```bash
# User presses Ctrl+C or types /quit

Closing connection...
Warning: failed to send leave message: <error> (if any)
```

**Server Cleanup Sequence:**
```
1. ReadPump detects connection close
2. defer function executes:
   - Hub.Unregister <- client
   - conn.Close()
3. Hub processes Unregister:
   - Remove client from room
   - Broadcast leave notification
   - Delete client from Clients map
   - Close SendChannel
4. WritePump exits (SendChannel closed)
```

**Server Logs:**
```
INFO Client left room room_id=1 client_id=2c8a250e-b03a-4853-9643-89e6f2f17e7c
INFO Client removed from room room_id=1 client_id=2c8a250e-b03a-4853-9643-89e6f2f17e7c
INFO Broadcasting message room_id=1 message="[alice] has left the chat."
INFO Client unregistered client_id=2c8a250e-b03a-4853-9643-89e6f2f17e7c
INFO Websocket info: send channel closed, closing connection client_id=...
```

**Success Criteria:**
- ✅ Client disconnected cleanly
- ✅ Leave notification sent to room (best effort)
- ✅ Client removed from room's client map
- ✅ Client removed from Hub's client map
- ✅ SendChannel closed properly
- ✅ No goroutine leaks (ReadPump and WritePump exit)
- ✅ No resource leaks (WebSocket connection closed)

**Edge Cases:**
- **Network Failure**: ReadPump detects unexpected close, cleanup still happens
- **Server Restart**: All clients disconnected, must rejoin
- **Leave Message Fails**: Unregister still happens via defer in ReadPump
- **Already Left**: Warning logged, no error

---

### UC-WS-004: Multiple Rooms Isolation
**Actor:** Multiple Users in Different Rooms  
**Preconditions:** API server running, multiple users logged in  
**Flow:**
1. User A joins room for manga ID 1 (One Piece)
2. User B joins room for manga ID 2 (Naruto)
3. User A sends message "Love this arc!"
4. User B sends message "Believe it!"
5. Messages do NOT cross between rooms
6. Each user only sees messages from their own room

**Example:**
```bash
# Terminal 1 (Alice - One Piece room)
mangahub chat join --room 1
hello one piece fans!

# Terminal 2 (Bob - Naruto room)
mangahub chat join --room 2
hello naruto fans!

# Terminal 3 (Charlie - One Piece room)
mangahub chat join --room 1
hi alice!
```

**Message Isolation:**
```
# Alice sees:
[alice] hello one piece fans!
[charlie] hi alice!

# Bob sees:
[bob] hello naruto fans!

# Charlie sees:
🔔 You have joined the chat room for manga ID 1. Users online: 2
🔔 [charlie] has joined the chat.
[alice] hello one piece fans!
[charlie] hi alice!
```

**Server Architecture:**
- Hub maintains separate Room objects: `map[int64]*Room`
- Each Room has its own client map: `map[string]*Client`
- Broadcast only sends to clients in the same room
- Room ID = Manga ID (1:1 mapping)

**Success Criteria:**
- ✅ Messages isolated by room
- ✅ No message leakage between rooms
- ✅ Each room maintains independent user count
- ✅ Users can join multiple rooms sequentially (not simultaneously from one CLI)

---

### UC-WS-005: Room Switching
**Actor:** Authenticated User  
**Preconditions:** User currently in a chat room  
**Flow:**
1. User is in room A (manga ID 1)
2. User exits with `/quit`
3. User joins room B (manga ID 2) with new command
4. Server removes user from room A
5. Server adds user to room B
6. Both rooms receive appropriate notifications

**Example:**
```bash
# User starts in One Piece room
mangahub chat join --room 1
hello everyone!
/quit

# User switches to Naruto room
mangahub chat join --room 2
hi naruto fans!
```

**Server Behavior:**
- Each `chat join` command creates a NEW WebSocket connection
- Previous connection is closed before new one starts
- Rooms are independent - no shared state between connections

**Note:** Current implementation supports ONE room per CLI session.
To be in multiple rooms simultaneously, user must open multiple terminal windows.

**Success Criteria:**
- ✅ User successfully switches between rooms
- ✅ Clean disconnect from previous room
- ✅ Successful connection to new room
- ✅ Leave/join notifications sent to appropriate rooms

---

### UC-WS-006: Connection Heartbeat and Timeout
**Actor:** Connected WebSocket Client  
**Preconditions:** Active WebSocket connection  
**Flow:**
1. Client WritePump sends ping every 54 seconds
2. Server receives ping and responds with pong
3. Client ReadPump receives pong and updates read deadline
4. If pong not received within 60 seconds, ReadPump exits
5. If no message received within 60 seconds, connection times out

**Ping/Pong Configuration:**
```go
const (
    WriteWait      = 10 * time.Second    // Max time to write message
    PongWait       = 60 * time.Second    // Max time to wait for pong
    PingPeriod     = (PongWait * 9) / 10 // Send ping every 54s (90% of PongWait)
    MaxMessageSize = 512                 // Max message size in bytes
)
```

**Heartbeat Flow:**
```
┌─────────┐                              ┌─────────┐
│ Client  │                              │ Server  │
│WritePump│                              │ReadPump │
└────┬────┘                              └────┬────┘
     │                                        │
     │ Every 54 seconds:                     │
     │ Send PING frame                       │
     ├──────────────────────────────────────>│
     │                                        │
     │                                        │ Receive PING
     │                                        │ Auto-reply PONG
     │                                        │
     │<───────────────────────────────────────┤
     │ Receive PONG                           │
     │ Update read deadline (+60s)            │
     │                                        │
```

**Timeout Scenarios:**
```
Scenario 1: Normal operation
- Ping sent every 54s
- Pong received within 60s
- Connection stays alive indefinitely

Scenario 2: Network issue
- Ping sent at T=0
- No pong received
- Read deadline expires at T=60s
- ReadPump exits, connection closed

Scenario 3: Idle connection
- No messages sent/received
- Ping/pong keeps connection alive
- No timeout as long as pong responds
```

**Success Criteria:**
- ✅ Connection stays alive during idle periods
- ✅ Failed connections detected within 60 seconds
- ✅ Automatic cleanup on timeout
- ✅ No zombie connections

---

### UC-WS-002: Send and Receive Chat Messages
**Actor:** Authenticated User in Chat Room  
**Preconditions:** User connected to chat room via WebSocket  
**Flow:**
1. User types message in CLI
2. Client sends chat message via WebSocket
3. Server receives message from client
4. Server validates message format and room membership
5. Server broadcasts message to all room members
6. All connected clients receive message in real-time
7. Each client displays formatted message with username

**Message Input:**
```
# User types in terminal
hello everyone!
```

**WebSocket Protocol:**
```json
// Chat Message (Client → Server)
{
    "type": "chat",
    "content": "hello everyone!"
}

// Broadcast Message (Server → All Clients in Room)
{
    "type": "chat",
    "room_id": 1,
    "user_id": "df76ef64-641c-49ba-bc78-c6875aefe582",
    "user_name": "chatuser1",
    "content": "hello everyone!",
    "timestamp": "2025-12-01T14:10:25Z"
}
```

**Client Display:**
```
[chatuser1] hello everyone!
[user2] Hi there!
[chatuser1] anyone reading One Piece?
```

**Success Criteria:**
- Message sent successfully
- All room members receive message
- Username and timestamp included
- Messages displayed in order
- Real-time delivery (< 100ms latency)

**Special Commands:**
```bash
/quit   # Exit chat and close connection
```

---

### UC-WS-003: Leave Chat Room
**Actor:** User in Chat Room  
**Preconditions:** User connected to WebSocket chat  
**Flow:**
1. User types `/quit` or closes terminal
2. Client sends leave message to server
3. Client closes WebSocket connection
4. Server removes client from room
5. Server broadcasts leave notification to remaining members
6. Server unregisters client from Hub
7. CLI returns to command prompt

**WebSocket Protocol:**
```json
// Leave Message (Client → Server)
{
    "type": "leave",
    "room_id": 1
}

// Leave Broadcast (Server → Remaining Room Members)
{
    "type": "system",
    "room_id": 1,
    "user_id": "system",
    "user_name": "System_Admin",
    "content": "chatuser1 has left the room",
    "timestamp": "2025-12-01T14:15:30Z"
}
```

**Success Criteria:**
- Client disconnected cleanly
- Leave notification sent to room
- Client removed from Hub
- No resource leaks
- Connection properly closed

---

### WebSocket Server Architecture

**Components:**

1. **Hub** - Central connection and room manager
   - Runs in dedicated goroutine with event loop
   - Manages global client registry: `map[string]*Client`
   - Manages room registry: `map[int64]*Room`
   - Processes requests via buffered channels:
     - `Register`: New client registration
     - `Unregister`: Client disconnect cleanup
     - `JoinRoom`: Client joins manga room
     - `LeaveRoom`: Client leaves room
     - `Broadcast`: Message distribution
   - Thread-safe with sync.RWMutex

2. **Client** - Individual WebSocket connection handler
   - **Properties**:
     - `ID`: Unique client identifier (user_id)
     - `UserID`: From JWT claims
     - `UserName`: From JWT claims
     - `RoomID`: Current room (manga_id), -1 if not in room
     - `Conn`: gorilla/websocket connection
     - `SendChannel`: Buffered channel (256 messages)
     - `Hub`: Reference to central hub
   
   - **ReadPump Goroutine**:
     - Reads messages from WebSocket
     - Parses JSON and validates message type
     - Enriches messages with user info
     - Sends to Hub for processing
     - Handles ping/pong for keepalive
     - Cleanup on disconnect via defer
   
   - **WritePump Goroutine**:
     - Writes messages from SendChannel to WebSocket
     - Sends periodic pings (every 54s)
     - Handles write timeouts (10s)
     - Exits gracefully when SendChannel closes

3. **Room** - Manga-specific chat room
   - **Properties**:
     - `ID`: Manga ID (room_id)
     - `Title`: Manga title
     - `Clients`: Map of connected clients `map[string]*Client`
     - `mu`: sync.RWMutex for thread safety
   
   - **Methods**:
     - `AddUser(client)`: Adds client to room
     - `RemoveUser(client)`: Removes client from room
     - `Broadcast(message)`: Sends message to all room clients
     - `GetUserCount()`: Returns current user count
     - `GetClients()`: Returns snapshot of clients
   
   - **Lifecycle**:
     - Auto-created when first user joins
     - Persists while users are connected
     - No automatic cleanup (rooms remain in memory)

4. **Message** - Communication protocol
   ```go
   type Message struct {
       Type      MessageType `json:"type"`       // join, leave, chat, system, typing
       RoomID    int64       `json:"room_id"`    // Manga ID
       UserID    string      `json:"user_id"`    // Sender UUID
       UserName  string      `json:"user_name"`  // Sender username
       Content   string      `json:"content"`    // Message text
       Timestamp time.Time   `json:"timestamp"`  // UTC timestamp
   }
   ```
   
   - **Message Types**:
     - `join`: User joins room (triggers room creation/join)
     - `leave`: User leaves room (triggers cleanup)
     - `chat`: User message (broadcasted to room)
     - `system`: Server notification (join/leave events)
     - `typing`: Typing indicator (optional, not implemented in CLI)

**Message Flow Diagram:**

```text
┌─────────┐    WebSocket     ┌──────────┐    Register    ┌─────────┐
│   CLI   │─────────────────>│  Client  │───────────────>│   Hub   │
│ (User)  │                  │(ReadPump)│                │(Manager)│
└─────────┘                  └────┬─────┘                └────┬────┘
     │                            │                           │
     │ Type message               │                           │
     │──────────────────────────>│                           │
     │                            │                           │
     │                            │ Receive JSON              │
     │                            │ Parse Message             │
     │                            │ Enrich with user info     │
     │                            │                           │
     │                            │ Hub.Broadcast <- msg      │
     │                            │─────────────────────────>│
     │                            │                           │
     │                            │                      Get Room
     │                            │                      room.Broadcast(msg)
     │                            │                           │
     │                            │                      For each client:
     │                            │                        client.SendChannel <- msg
     │                            │                           │
     │                            │<──────────────────────────┤
     │                            │ WritePump receives msg    │
     │                       WritePump                        │
     │                       conn.WriteJSON(msg)              │
     │<───────────────────────────┤                           │
     │ Display: [user] message    │                           │
```

**Authentication Flow:**

```text
1. CLI: Add JWT to Authorization header during connection
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5...

2. Gin Router: HTTP Upgrade Request received
   GET /ws HTTP/1.1
   Upgrade: websocket

3. Gin Middleware: AuthMiddleware validates token
   - Extract token from header
   - Verify signature with secret key
   - Parse claims (user_id, username, exp)
   - Set userID and claims in gin.Context

4. WebSocket Handler: Extract user info from context
   userID, _ := c.Get("userID")
   claims, _ := c.Get("claims")
   userName := claims.(*service.Claims).Username

5. WebSocket Handler: Upgrade connection
   conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

6. WebSocket Handler: Create authenticated client
   client := NewClient(userID, userID, userName, NilRoomID, conn, hub)

7. Hub: Register authenticated client
   hub.Register <- client

8. Client: Start ReadPump and WritePump goroutines
   go client.ReadPump()
   go client.WritePump()
```

**Concurrency Model:**

```text
Main Goroutine
├── HTTP Server
│   └── For each WebSocket connection:
│       ├── AuthMiddleware (validate JWT)
│       ├── WSHandler (upgrade connection)
│       └── Spawn 2 goroutines:
│           ├── ReadPump  (read from WS, send to Hub)
│           └── WritePump (read from SendChannel, write to WS)
│
└── Hub (1 goroutine)
    ├── Listen on channels:
    │   ├── Register   (add client)
    │   ├── Unregister (remove client)
    │   ├── JoinRoom   (add client to room)
    │   ├── LeaveRoom  (remove client from room)
    │   └── Broadcast  (send message to room)
    │
    └── For each registered Client:
        └── SendChannel (buffered, 256 messages)
```

**Thread Safety:**
- Hub: Protected by `sync.RWMutex`, all operations via channels
- Room: Protected by `sync.RWMutex` for client map access
- Client: SendChannel is thread-safe (Go channels)
- No shared mutable state between goroutines

**Configuration:**

```go
// Connection Parameters
const (
    WriteWait      = 10 * time.Second    // Max time to write message
    PongWait       = 60 * time.Second    // Max time to wait for pong
    PingPeriod     = 54 * time.Second    // Send ping every 54s
    MaxMessageSize = 512                 // Max message size (bytes)
)

// WebSocket Endpoint
URL: ws://localhost:8084/ws
Port: 8084 (same as HTTP API server)
Protocol: ws:// (not wss:// in development)

// Upgrader Configuration
upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true  // Allow all origins (development only)
    },
}
```

**Error Scenarios:**

| Scenario | Detection | Handling | Recovery |
|----------|-----------|----------|----------|
| Invalid Message Format | ReadPump JSON unmarshal error | Log error, continue reading | User can retry |
| Send Channel Full | SendChannel <- blocks | Drop message, log warning | System continues |
| Connection Lost | ReadPump read error | Close connection, unregister | User must reconnect |
| Write Timeout | WritePump write deadline exceeded | Close connection | User must reconnect |
| Client Not in Room | Message validation | Ignore message, log warning | User should rejoin |
| Duplicate Registration | Hub.Register | Replace old client | New connection active |
| Pong Timeout | ReadPump deadline | Close connection, cleanup | User must reconnect |

**Performance Characteristics:**
- **Latency**: < 100ms for local network
- **Throughput**: 1000+ messages/second per room
- **Concurrent Users**: Limited by system resources (tested with 100+)
- **Memory**: ~50KB per connection (buffers + goroutines)
- **Goroutines**: 2 per connection + 1 for Hub

**Deployment Notes:**
- WebSocket and HTTP share same port (8084)
- Route registered: `GET /ws`
- Requires authentication middleware
- No TLS in development (use wss:// in production)
- CORS enabled for all origins (restrict in production)

---

## Data Flow Architecture

### HTTP API Progress Update Flow
```
┌─────────┐     POST /api/progress/:id      ┌─────────────┐
│   CLI   │────────────────────────────────>│  API Server │
└─────────┘     Authorization: Bearer JWT   │  (Gin)      │
                                             └──────┬──────┘
                                                    │
                                             ┌──────▼──────┐
                                             │  Middleware │
                                             │  - Auth     │
                                             │  - JWT      │
                                             └──────┬──────┘
                                                    │
                                             ┌──────▼──────────┐
                                             │  Handler        │
                                             │  - Validate     │
                                             │  - Extract JWT  │
                                             └──────┬──────────┘
                                                    │
                                             ┌──────▼──────────┐
                                             │  Service        │
                                             │  - Business     │
                                             │    Logic        │
                                             └──────┬──────────┘
                                                    │
                                             ┌──────▼──────────┐
                                             │  Repository     │
                                             │  - GORM ORM     │
                                             └──────┬──────────┘
                                                    │
                                             ┌──────▼──────────┐
                                             │  PostgreSQL     │
                                             │  (Persistent)   │
                                             └─────────────────┘
```

### TCP Sync Progress Update Flow
```
┌─────────┐     progress sync --manga-id 1    ┌─────────────┐
│   CLI   │────────────────────────────────────>│  Temp TCP   │
└─────────┘                                     │  Connection │
                                                └──────┬──────┘
                                                       │
                                                ┌──────▼──────────┐
                                                │  TCP Server     │
                                                │  - Auth         │
                                                │  - Validate     │
                                                └──────┬──────────┘
                                                       │
                                          ┌────────────┴────────────┐
                                          │                         │
                                   ┌──────▼──────────┐   ┌─────────▼────────┐
                                   │  Redis Cache    │   │  Write Queue     │
                                   │  (Immediate)    │   │  (Async Channel) │
                                   │  TTL: 90 days   │   │  Buffer: 10k     │
                                   └─────────────────┘   └─────────┬────────┘
                                                                    │
                                                         ┌──────────▼──────────┐
                                                         │  Batch Writer       │
                                                         │  (Every 5 min or    │
                                                         │   1000 records)     │
                                                         └──────────┬──────────┘
                                                                    │
                                                         ┌──────────▼──────────┐
                                                         │  PostgreSQL         │
                                                         │  (Persistent)       │
                                                         │  - FK Constraints   │
                                                         └─────────────────────┘
```

### gRPC Progress Update Flow
```
┌─────────┐     grpc progress update --manga-id 1    ┌─────────────┐
│   CLI   │─────────────────────────────────────────>│  gRPC       │
└─────────┘     User authenticated (JWT stored)      │  Client     │
                                                      └──────┬──────┘
                                                             │
                                                      ┌──────▼──────────┐
                                                      │  gRPC Server    │
                                                      │  (HTTP/2)       │
                                                      │  Port: 8083     │
                                                      └──────┬──────────┘
                                                             │
                                                      ┌──────▼──────────┐
                                                      │  MangaService   │
                                                      │  - GetManga     │
                                                      │  - SearchManga  │
                                                      │  - UpdateProg   │
                                                      └──────┬──────────┘
                                                             │
                                                      ┌──────▼──────────┐
                                                      │  Repository     │
                                                      │  - GORM ORM     │
                                                      │  - Validation   │
                                                      └──────┬──────────┘
                                                             │
                                                      ┌──────▼──────────┐
                                                      │  PostgreSQL     │
                                                      │  (Persistent)   │
                                                      │  - Upsert       │
                                                      │  - FK Check     │
                                                      └─────────────────┘
```

### Authentication Flow
```
┌─────────┐     POST /api/auth/login          ┌─────────────┐
│   CLI   │────────────────────────────────────>│  API Server │
└────┬────┘     {username, password}           └──────┬──────┘
     │                                                 │
     │                                          ┌──────▼──────────┐
     │                                          │  Auth Service   │
     │                                          │  - bcrypt       │
     │                                          │  - JWT Sign     │
     │                                          └──────┬──────────┘
     │                                                 │
     │         {access_token, refresh_token}   ┌──────▼──────────┐
     │◄────────────────────────────────────────│  PostgreSQL     │
     │                                          │  (users table)  │
     │                                          └─────────────────┘
     │
┌────▼─────────────┐
│  System Keyring  │
│  - Windows DPAPI │
│  - macOS Keychain│
│  - Linux Secret  │
│    Service       │
└──────────────────┘
```

---

## API Endpoints Reference

### Authentication Endpoints

| Method | Endpoint | Auth Required | Description |
|--------|----------|---------------|-------------|
| POST | `/api/auth/register` | No | Register new user |
| POST | `/api/auth/login` | No | Login user |
| POST | `/api/auth/refresh` | Yes (Refresh Token) | Refresh access token |
| POST | `/api/auth/logout` | Yes | Logout user (client-side) |

### Manga Library Endpoints

| Method | Endpoint | Auth Required | Description |
|--------|----------|---------------|-------------|
| GET | `/api/manga?search=<query>` | Yes | Search manga |
| GET | `/api/manga/<id>` | Yes | Get manga by ID |

### Progress Tracking Endpoints

| Method | Endpoint | Auth Required | Description |
|--------|----------|---------------|-------------|
| GET | `/api/progress` | Yes | Get all user progress |
| GET | `/api/progress/<manga_id>` | Yes | Get progress for manga |
| POST | `/api/progress/<manga_id>` | Yes | Update/create progress |

### TCP Sync Protocol Messages

| Message Type | Direction | Description |
|--------------|-----------|-------------|
| `auth` | Client → Server | Initial authentication |
| `auth_success` | Server → Client | Authentication successful |
| `auth_error` | Server → Client | Authentication failed |
| `progress_update` | Client → Server | Update reading progress |
| `heartbeat` | Client ↔ Server | Keep connection alive |
| `disconnect` | Client → Server | Close connection |

### gRPC Service Methods

| Service | Method | Request | Response | Description |
|---------|--------|---------|----------|-------------|
| MangaService | `GetManga` | `GetMangaRequest` | `GetMangaResponse` | Get manga by ID |
| MangaService | `SearchManga` | `SearchRequest` | `SearchResponse` | Search manga by query |
| MangaService | `UpdateProgress` | `UpdateProgressRequest` | `UpdateProgressResponse` | Update reading progress |

**gRPC Server Details:**
- **Default Port:** 8083
- **Protocol:** HTTP/2
- **Transport Security:** Insecure (development)
- **Connection Timeout:** 5 seconds
- **RPC Timeout:** 10 seconds per operation

---

## Database Schema

### users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### manga Table
```sql
CREATE TABLE manga (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE,
    total_chapters INTEGER,
    status VARCHAR(50) DEFAULT 'ongoing',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### user_progress Table
```sql
CREATE TABLE user_progress (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manga_id BIGINT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    current_chapter INTEGER DEFAULT 0,
    status TEXT CHECK (status IN ('reading', 'completed', 'plan_to_read', 'dropped')),
    rating INTEGER CHECK (rating >= 1 AND rating <= 10),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_progress_user_manga_unique UNIQUE (user_id, manga_id)
);

CREATE INDEX idx_user_progress_lookup ON user_progress(user_id, manga_id);
```

---

## DTO Alignment

### HTTP API ProgressResponse DTO
```go
type ProgressResponse struct {
    UserID    string `json:"user_id"`
    MangaID   int64  `json:"manga_id"`
    Chapter   int    `json:"chapter"`
    Status    string `json:"status"`
    UpdatedAt string `json:"updated_at"`
}
```

### TCP ProgressData DTO
```go
type ProgressData struct {
    ID             int64     `json:"id,omitempty"`
    UserID         string    `json:"user_id"`
    MangaID        int64     `json:"manga_id"`
    CurrentChapter int       `json:"current_chapter"`
    Status         string    `json:"status"`
    Rating         *int      `json:"rating,omitempty"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

**Mapping:**
- HTTP `chapter` = TCP `current_chapter` = DB `current_chapter`
- Both use same status enum: reading, completed, plan_to_read, dropped
- TCP includes optional `rating` field (1-10)
- All timestamps use RFC3339/ISO 8601 format

---

## Environment Variables

### HTTP API Server
```bash
PORT=8084
DATABASE_URL=postgres://user:pass@localhost:5432/mangahub?sslmode=disable
JWT_SECRET=your-secret-key-here
JWT_ACCESS_EXPIRY=1h
JWT_REFRESH_EXPIRY=168h  # 7 days
```

### TCP Sync Server
```bash
TCP_PORT=8081
REDIS_URL=redis://localhost:6379
DATABASE_URL=postgres://user:pass@localhost:5432/mangahub?sslmode=disable
JWT_SECRET=your-secret-key-here  # Must match API server
```

### CLI Environment
```bash
MANGAHUB_API_URL=http://localhost:8084
MANGAHUB_TCP_SERVER=localhost:8081
MANGAHUB_GRPC_ADDR=localhost:8083
```

---

## Error Handling

### HTTP API Error Responses
```json
{
  "error": "error message",
  "code": "ERROR_CODE"
}
```

**Common Error Codes:**
- `INVALID_CREDENTIALS` - Login failed
- `TOKEN_EXPIRED` - Access token expired
- `UNAUTHORIZED` - Missing or invalid token
- `FORBIDDEN` - User not allowed
- `NOT_FOUND` - Resource doesn't exist
- `VALIDATION_ERROR` - Invalid input
- `DATABASE_ERROR` - Database operation failed

### TCP Protocol Errors
```json
{
  "type": "error",
  "code": "ERROR_CODE",
  "message": "Human readable message"
}
```

**Common Error Codes:**
- `INVALID_USER_ID` - Missing or malformed user_id
- `FORBIDDEN` - Unauthorized progress update
- `INVALID_DATA` - Invalid manga_id or chapter
- `SAVE_FAILED` - Database/Redis operation failed

---

## Performance Characteristics

### HTTP API
- **Average Response Time:** < 50ms
- **Authentication:** < 10ms (JWT validation)
- **Database Queries:** < 20ms (indexed lookups)

### TCP Sync
- **Connection Setup:** < 100ms
- **Redis Write:** < 5ms (immediate)
- **PostgreSQL Batch:** Every 5 minutes or 1000 records
- **Batch Write Time:** < 500ms for 1000 records
- **Connection Keepalive:** 30-second heartbeat

### Storage
- **Redis TTL:** 90 days
- **PostgreSQL:** Unlimited retention
- **Batch Queue:** 10,000 record buffer

---

## Security Considerations

### Authentication
- Passwords hashed with bcrypt (cost factor 10)
- JWT tokens signed with HMAC-SHA256
- Access tokens expire after 1 hour
- Refresh tokens expire after 7 days
- Tokens stored securely in system keyring

### Authorization
- User can only access their own progress
- User ID extracted from JWT claims
- Foreign key constraints enforce data integrity

### TCP Connection
- JWT authentication required
- Session-based authorization
- User ID validated on every operation
- Connection timeout after 10 seconds idle

### Input Validation
- Status enum validated
- Chapter numbers validated (>= 0)
- Manga ID foreign key constraint
- User ID UUID format validated

---

## Testing Scenarios

### Successful Flows
1. **Complete User Journey:**
   - Register → Login → Search Manga → Update Progress → View History
   
2. **TCP Sync Flow:**
   - Login → Sync Connect → Sync Progress → Verify Redis → Wait for Batch

3. **Token Refresh:**
   - Login → Wait 1 hour → Make request → Auto-refresh → Request succeeds

### Error Scenarios
1. **Invalid Progress Update:**
   - Backward progress without --force → Error
   - Chapter exceeds total_chapters → Error
   - Invalid status value → Validation error

2. **Authentication Failures:**
   - Wrong password → Login failed
   - Expired token without refresh → 401 Unauthorized
   - Invalid JWT signature → 401 Unauthorized

3. **TCP Sync Failures:**
   - Foreign key violation → Batch write failed (logged, not returned to client)
   - Invalid user_id → Error response
   - Unauthorized update → Forbidden error

---

## CLI Command Reference

### Complete Command List

```bash
# Authentication
mangahub auth register --username <name> --email <email> --password <pass>
mangahub auth login --username <name> --password <pass>
mangahub auth logout
mangahub auth status

# Manga Library
mangahub library search "<query>"
mangahub library get --manga-id <id>

# Progress Tracking (HTTP)
mangahub progress update --manga-id <id> --chapter <num> [--status <status>] [--force]
mangahub progress history [--manga-id <id>]

# TCP Sync
mangahub sync connect [--server <host:port>]
mangahub sync disconnect
mangahub sync status
mangahub sync monitor

# Progress Tracking (TCP)
mangahub progress sync --manga-id <id> --chapter <num> [--status <status>]
mangahub progress sync-status

# gRPC Service Operations
mangahub grpc manga get --id <manga-id> [--grpc-addr <host:port>]
mangahub grpc manga search --query "<search-term>" [--limit <N>] [--offset <N>] [--grpc-addr <host:port>]
mangahub grpc progress update --manga-id <id> --chapter <number> [--status <status>] [--title "<title>"] [--grpc-addr <host:port>]
```

---

## Future Enhancements (Not Yet Implemented)

1. **WebSocket Support** - Real-time bi-directional sync
2. **Conflict Resolution** - Advanced merge strategies
3. **Offline Mode** - Queue operations when offline
4. **Multi-device Sync** - Automatic cross-device synchronization
5. **Progress Analytics** - Reading statistics and trends
6. **Rating System** - Rate manga and chapters
7. **Notification Service** - UDP-based notifications
8. **Advanced gRPC Features** - Streaming RPCs, authentication middleware

---

## Troubleshooting

### CLI Issues

**Problem:** `not authenticated, please login first`
- **Solution:** Run `mangahub auth login`

**Problem:** `failed to connect to TCP server`
- **Solution:** Check TCP server is running: `docker ps | grep mangahub-tcp`

**Problem:** `token has invalid claims: token is expired`
- **Solution:** CLI automatically refreshes. If it fails, run `mangahub auth login`

### Server Issues

**Problem:** `batch_insert_failed: foreign key constraint violation`
- **Solution:** This is expected. Manga ID doesn't exist in database. Use existing manga IDs.

**Problem:** `progress_save_failed`
- **Solution:** Check Redis connection and database connectivity

**Problem:** `authentication_failed`
- **Solution:** Verify JWT_SECRET matches between API and TCP servers

---

**Document Version:** 1.1  
**Last Updated:** November 26, 2025  
**Maintained By:** MangaHub Development Team  
**Status:** Current Implementation - gRPC Services Added
