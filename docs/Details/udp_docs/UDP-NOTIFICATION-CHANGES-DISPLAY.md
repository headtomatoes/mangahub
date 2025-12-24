# UDP Notification - Exact Field Changes Display

## Overview
Enhanced the UDP notification system to display **exactly what fields changed** in update notifications, including before/after values for fields like current chapter numbers.

## Changes Made

### 1. Enhanced Notification Structure (`internal/microservices/udp-server/notification.go`)

Added a new `FieldChange` struct to track individual field updates:

```go
type FieldChange struct {
    Field    string      `json:"field"`
    OldValue interface{} `json:"old_value,omitempty"`
    NewValue interface{} `json:"new_value"`
}
```

Updated the `Notification` struct to include changes:
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

### 2. Enhanced Chapter Notifications

#### New Method: `NewMangaUpdateNotificationWithDetails`
Creates notifications with detailed field change information:
```go
func NewMangaUpdateNotificationWithDetails(
    mangaID int64, 
    title string, 
    fieldChanges []FieldChange
) *Notification
```

#### Updated Chapter Notifications
Now includes current chapter in the changes array:
```go
Changes: []FieldChange{
    {
        Field:    "current_chapter",
        OldValue: oldChapter,  // Optional
        NewValue: newChapter,
    },
}
```

### 3. New Notifier Methods (`internal/ingestion/mangadex/notifier.go`)

Added `NotifyNewChapterWithPrevious` to send old and new chapter info:
```go
func (n *Notifier) NotifyNewChapterWithPrevious(
    mangaID int64, 
    title string, 
    oldChapter, 
    newChapter int
)
```

### 4. UDP Server Enhancement (`cmd/udp-server/main.go`)

Updated `/notify/new-chapter` endpoint to accept old chapter:
```go
var payload struct {
    MangaID    int64  `json:"manga_id"`
    Title      string `json:"title"`
    Chapter    int    `json:"chapter"`
    OldChapter *int   `json:"old_chapter,omitempty"`  // NEW
}
```

Added new server method `NotifyNewChapterWithPrevious` in `internal/microservices/udp-server/server.go`:
```go
func (s *Server) NotifyNewChapterWithPrevious(
    ctx context.Context, 
    mangaID int64, 
    title string, 
    oldChapter, 
    newChapter int
) error
```

### 5. Enhanced Client Display (`cmd/udp-client/main.go`)

Added `displayNotification` function with rich formatting:
```go
func displayNotification(notif map[string]interface{})
```

#### Display Format Example:
```
============================================================
📢 NEW_CHAPTER
📚 Title: One Piece
💬 New chapter 1095 available (was chapter 1094)
🕐 Time: 2025-12-02 14:30:45

🔄 Changes:
  • current_chapter: 1094 → 1095

📖 Chapter: 1095
============================================================
```

## Usage Examples

### Example 1: New Chapter Notification with Previous Chapter

**Server-side (MangaDex sync):**
```go
// When detecting new chapters
s.notifier.NotifyNewChapterWithPrevious(
    manga.ID, 
    manga.Title, 
    baseline,              // Old chapter (e.g., 1094)
    int(extracted.ChapterNumber)  // New chapter (e.g., 1095)
)
```

**Client Display:**
```
============================================================
📢 NEW_CHAPTER
📚 Title: Jujutsu Kaisen
💬 New chapter 245 available (was chapter 244)
🕐 Time: 2025-12-02 14:35:12

🔄 Changes:
  • current_chapter: 244 → 245

📖 Chapter: 245
============================================================
```

### Example 2: Manga Update Notification

**Server-side:**
```go
changes := []string{"description", "cover image", "status"}
server.NotifyMangaUpdate(ctx, mangaID, title, changes)
```

**Client Display:**
```
============================================================
📢 MANGA_UPDATE
📚 Title: Demon Slayer
💬 Updated: description, cover image, and status
🕐 Time: 2025-12-02 14:40:00

📝 Updated fields: description, cover image, status
============================================================
```

### Example 3: New Manga Notification

**Client Display:**
```
============================================================
📢 NEW_MANGA
📚 Title: New Popular Manga
💬 New manga added: New Popular Manga
🕐 Time: 2025-12-02 14:45:00
============================================================
```

## Notification JSON Structure

### New Chapter with Changes
```json
{
  "type": "NEW_CHAPTER",
  "manga_id": 123,
  "title": "One Piece",
  "message": "New chapter 1095 available (was chapter 1094)",
  "timestamp": "2025-12-02T14:30:45Z",
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

### Manga Update with Multiple Fields
```json
{
  "type": "MANGA_UPDATE",
  "manga_id": 456,
  "title": "Demon Slayer",
  "message": "Updated: description and cover image",
  "timestamp": "2025-12-02T14:40:00Z",
  "data": {
    "updated_fields": ["description", "cover_image"]
  }
}
```

## Benefits

1. **Clear Change Visibility**: Users can see exactly what changed (e.g., chapter 244 → 245)
2. **Rich Formatting**: Emoji icons and structured display make notifications easy to read
3. **Detailed Information**: Shows both old and new values for updated fields
4. **Extensible**: Can easily add more field changes for other update types
5. **Backward Compatible**: Existing notifications still work without the changes field

## Integration Points

### MangaDex Sync Integration
The MangaDex sync service (`internal/ingestion/mangadex/workflows.go`) automatically uses the enhanced notification when checking for new chapters:

```go
// In checkMangaChapters function
s.notifier.NotifyNewChapterWithPrevious(
    manga.ID, 
    manga.Title, 
    baseline,           // Previous highest chapter
    int(extracted.ChapterNumber)  // New chapter
)
```

### Manual API Updates
When updating manga through the API (`internal/microservices/http-api/service/mangaService.go`), the system tracks which fields changed:

```go
changes := []string{}
if m.Description != nil {
    changes = append(changes, "description")
}
if m.CoverURL != nil {
    changes = append(changes, "cover image")
}
// Send notification with specific changes
notifyMangaUpdate(id, existing.Title, changes)
```

## Testing

### Test New Chapter Notification
```bash
# Start UDP server
./bin/udp-server

# In another terminal, send test notification
curl -X POST http://localhost:8085/notify/new-chapter \
  -H "Content-Type: application/json" \
  -d '{
    "manga_id": 123,
    "title": "Test Manga",
    "chapter": 50,
    "old_chapter": 49
  }'
```

### Test Manga Update Notification
```bash
curl -X POST http://localhost:8085/notify/manga-update \
  -H "Content-Type: application/json" \
  -d '{
    "manga_id": 123,
    "title": "Test Manga",
    "changes": ["description", "cover_image", "status"]
  }'
```

## Future Enhancements

Potential additions to the change tracking system:

1. **Rating Changes**: Track when average rating changes
   ```json
   {
     "field": "average_rating",
     "old_value": 4.5,
     "new_value": 4.7
   }
   ```

2. **Status Changes**: Track manga completion status
   ```json
   {
     "field": "status",
     "old_value": "ongoing",
     "new_value": "completed"
   }
   ```

3. **Author Changes**: Track author information updates
   ```json
   {
     "field": "author",
     "old_value": "Author A",
     "new_value": "Author B"
   }
   ```

4. **Volume Updates**: Track volume releases
   ```json
   {
     "field": "total_volumes",
     "old_value": 10,
     "new_value": 11
   }
   ```
