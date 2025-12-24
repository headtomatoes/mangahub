# UDP Notification System - Change Tracking Flow

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         NOTIFICATION SOURCES                         │
└─────────────────────────────────────────────────────────────────────┘
                                   │
        ┌──────────────────────────┼──────────────────────────┐
        │                          │                          │
        ▼                          ▼                          ▼
┌───────────────┐          ┌──────────────┐         ┌──────────────┐
│  MangaDex     │          │   HTTP API   │         │   Manual     │
│  Sync Service │          │   Service    │         │   Trigger    │
└───────────────┘          └──────────────┘         └──────────────┘
        │                          │                          │
        │ Chapter Detection        │ Field Updates            │ Test/Debug
        │                          │                          │
        ▼                          ▼                          ▼
┌─────────────────────────────────────────────────────────────────────┐
│                   UDP SERVER HTTP ENDPOINTS                          │
│                                                                      │
│  POST /notify/new-chapter     - New chapter notifications          │
│  POST /notify/manga-update    - Manga field updates                │
│  POST /notify/new-manga       - New manga additions                │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                   NOTIFICATION PROCESSING                            │
│                                                                      │
│  1. Parse incoming request                                          │
│  2. Create Notification with FieldChange array                      │
│  3. Determine target users (all or library users)                   │
│  4. Broadcast via UDP                                               │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     CONNECTED UDP CLIENTS                            │
│                                                                      │
│  Client 1 (subscribed to manga 1,2,3)                               │
│  Client 2 (subscribed to manga 4,5,6)                               │
│  Client 3 (subscribed to all)                                       │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      ENHANCED DISPLAY                                │
│                                                                      │
│  ════════════════════════════════════════                           │
│  📢 NEW_CHAPTER                                                      │
│  📚 Title: One Piece                                                 │
│  💬 New chapter 1095 available (was chapter 1094)                   │
│  🕐 Time: 2025-12-02 14:30:45                                       │
│                                                                      │
│  🔄 Changes:                                                         │
│    • current_chapter: 1094 → 1095                                   │
│                                                                      │
│  📖 Chapter: 1095                                                    │
│  ════════════════════════════════════════════                       │
└─────────────────────────────────────────────────────────────────────┘
```

## Data Flow Example: New Chapter Detection

### Step 1: MangaDex Sync Detects New Chapter
```go
// In: internal/ingestion/mangadex/workflows.go
func (s *SyncService) checkMangaChapters(...) {
    baseline := 1094  // Old chapter from DB
    newChapter := 1095  // Detected from MangaDex
    
    // Send notification with old and new chapter
    s.notifier.NotifyNewChapterWithPrevious(
        manga.ID,      // 123
        manga.Title,   // "One Piece"
        baseline,      // 1094
        newChapter     // 1095
    )
}
```

### Step 2: Notifier Sends HTTP Request
```go
// In: internal/ingestion/mangadex/notifier.go
payload := map[string]interface{}{
    "manga_id":    123,
    "title":       "One Piece",
    "chapter":     1095,
    "old_chapter": 1094,  // NEW!
}

// POST http://localhost:8085/notify/new-chapter
```

### Step 3: UDP Server Creates Notification
```go
// In: cmd/udp-server/main.go
server.NotifyNewChapterWithPrevious(
    ctx,
    payload.MangaID,    // 123
    payload.Title,      // "One Piece"
    *payload.OldChapter, // 1094
    payload.Chapter     // 1095
)
```

### Step 4: Notification Object Created
```go
// In: internal/microservices/udp-server/server.go
notification := NewChapterNotification(mangaID, title, newChapter)
notification.Changes = []FieldChange{
    {
        Field:    "current_chapter",
        OldValue: 1094,
        NewValue: 1095,
    },
}
notification.Message = "New chapter 1095 available (was chapter 1094)"
```

### Step 5: JSON Serialization
```json
{
  "type": "NEW_CHAPTER",
  "manga_id": 123,
  "title": "One Piece",
  "message": "New chapter 1095 available (was chapter 1094)",
  "timestamp": "2025-12-02T14:30:45.123Z",
  "data": {
    "chapter": 1095
  },
  "changes": [
    {
      "field": "current_chapter",
      "old_value": 1094,
      "new_value": 1095
    }
  ]
}
```

### Step 6: Broadcast to Library Users
```go
// In: internal/microservices/udp-server/broadcaster.go
// Find all users who have manga 123 in their library
users := getLibraryUsers(123)

// Send to each subscribed user
for _, user := range users {
    conn.WriteToUDP(jsonData, user.addr)
}
```

### Step 7: Client Receives and Displays
```go
// In: cmd/udp-client/main.go
notification := parseJSON(data)
displayNotification(notification)

// Output:
// ════════════════════════════════════════
// 📢 NEW_CHAPTER
// 📚 Title: One Piece
// 💬 New chapter 1095 available (was chapter 1094)
// 🕐 Time: 2025-12-02 14:30:45
//
// 🔄 Changes:
//   • current_chapter: 1094 → 1095
//
// 📖 Chapter: 1095
// ════════════════════════════════════════
```

## Data Flow Example: Manga Field Updates

### Step 1: API Receives Update Request
```go
// In: internal/microservices/http-api/service/mangaService.go
func (s *mangaService) Update(ctx context.Context, id int64, m *models.Manga) {
    changes := []string{}
    
    if m.Description != nil {
        existing.Description = m.Description
        changes = append(changes, "description")
    }
    if m.CoverURL != nil {
        existing.CoverURL = m.CoverURL
        changes = append(changes, "cover image")
    }
    if m.Status != nil {
        existing.Status = m.Status
        changes = append(changes, "status")
    }
    
    // Notify with specific changes
    notifyMangaUpdate(id, existing.Title, changes)
}
```

### Step 2: HTTP Request to UDP Server
```json
POST http://localhost:8085/notify/manga-update
{
  "manga_id": 456,
  "title": "Demon Slayer",
  "changes": ["description", "cover image", "status"]
}
```

### Step 3: Notification Created
```go
notification := NewMangaUpdateNotification(
    456,
    "Demon Slayer",
    []string{"description", "cover image", "status"}
)
// Message: "Updated: description, cover image, and status"
```

### Step 4: Client Display
```
════════════════════════════════════════
📢 MANGA_UPDATE
📚 Title: Demon Slayer
💬 Updated: description, cover image, and status
🕐 Time: 2025-12-02 14:35:20

📝 Updated fields: description, cover image, status
════════════════════════════════════════
```

## Key Components

### FieldChange Struct
```go
type FieldChange struct {
    Field    string      `json:"field"`        // e.g., "current_chapter"
    OldValue interface{} `json:"old_value"`    // e.g., 1094 (optional)
    NewValue interface{} `json:"new_value"`    // e.g., 1095 (required)
}
```

### Notification Struct (Enhanced)
```go
type Notification struct {
    Type      NotificationType `json:"type"`
    MangaID   int64            `json:"manga_id"`
    Title     string           `json:"title"`
    Message   string           `json:"message"`
    Timestamp time.Time        `json:"timestamp"`
    Data      interface{}      `json:"data,omitempty"`
    Changes   []FieldChange    `json:"changes,omitempty"`  // NEW
}
```

## Notification Types Supported

| Type | Description | Changes Field | Example |
|------|-------------|---------------|---------|
| `NEW_CHAPTER` | New chapter released | current_chapter: old → new | 1094 → 1095 |
| `MANGA_UPDATE` | Manga metadata updated | Field names only | [description, cover_image] |
| `NEW_MANGA` | New manga added | N/A | - |

## Benefits of This Design

✅ **Precise Information**: Shows exactly what changed  
✅ **Historical Context**: Old values provide context  
✅ **Multiple Changes**: Can track multiple field updates  
✅ **Extensible**: Easy to add new change types  
✅ **User-Friendly**: Clear visual formatting  
✅ **Backward Compatible**: Works with existing clients  
✅ **Performance**: Minimal overhead  

## Future Enhancements

1. **Detailed Field Values**: Track actual content changes
   ```json
   {
     "field": "description",
     "old_value": "Short description...",
     "new_value": "Updated longer description..."
   }
   ```

2. **Aggregate Changes**: Group related updates
   ```json
   {
     "field": "metadata",
     "changes": ["author", "artist", "serialization"]
   }
   ```

3. **Change Timestamps**: When each field was updated
   ```json
   {
     "field": "status",
     "old_value": "ongoing",
     "new_value": "completed",
     "changed_at": "2025-12-02T14:30:00Z"
   }
   ```
