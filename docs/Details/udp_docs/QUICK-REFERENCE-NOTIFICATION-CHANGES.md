# UDP Notification Changes - Quick Reference

## What Was Added

### ✅ Exact Field Change Tracking
- Shows **what changed** (e.g., current_chapter, description, cover_image)
- Shows **old value → new value** (e.g., chapter 1094 → 1095)
- Works for all notification types

### ✅ Enhanced Client Display
- **Emoji icons** for visual clarity
- **Structured format** with clear sections
- **Before/after values** for changes
- **Timestamp formatting** in readable format

### ✅ Chapter Update Details
- Shows previous chapter number
- Shows new chapter number  
- Clear arrow notation: `1094 → 1095`

## Client Display Format

```
============================================================
📢 [NOTIFICATION_TYPE]
📚 Title: [Manga Title]
💬 [Message describing the update]
🕐 Time: [Formatted timestamp]

🔄 Changes:
  • [field_name]: [old_value] → [new_value]
  • [field_name]: [new_value]

📖 Chapter: [chapter_number]
📝 Updated fields: [list of fields]
============================================================
```

## Real Examples

### New Chapter (with previous chapter)
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

### Manga Update (multiple fields)
```
============================================================
📢 MANGA_UPDATE
📚 Title: Jujutsu Kaisen
💬 Updated: description, cover image, and status
🕐 Time: 2025-12-02 14:35:20

📝 Updated fields: description, cover image, status
============================================================
```

### New Manga
```
============================================================
📢 NEW_MANGA
📚 Title: New Amazing Series
💬 New manga added: New Amazing Series
🕐 Time: 2025-12-02 14:40:00
============================================================
```

## Files Modified

1. ✅ `internal/microservices/udp-server/notification.go`
   - Added `FieldChange` struct
   - Enhanced `Notification` struct
   - Added `NewMangaUpdateNotificationWithDetails()`

2. ✅ `internal/microservices/udp-server/server.go`
   - Added `NotifyNewChapterWithPrevious()` method

3. ✅ `cmd/udp-server/main.go`
   - Enhanced `/notify/new-chapter` endpoint
   - Accepts `old_chapter` parameter

4. ✅ `internal/ingestion/mangadex/notifier.go`
   - Added `NotifyNewChapterWithPrevious()` method

5. ✅ `internal/ingestion/mangadex/workflows.go`
   - Uses new notification method with old chapter

6. ✅ `cmd/udp-client/main.go`
   - Added `displayNotification()` function
   - Enhanced display formatting

## Testing Commands

### Start Services
```bash
# Terminal 1: Start UDP server
./bin/udp-server

# Terminal 2: Start UDP client (login first)
./bin/udp-client
```

### Send Test Notifications
```bash
# New chapter with old chapter info
curl -X POST http://localhost:8085/notify/new-chapter \
  -H "Content-Type: application/json" \
  -d '{"manga_id": 1, "title": "Test", "chapter": 50, "old_chapter": 49}'

# Manga update with multiple fields
curl -X POST http://localhost:8085/notify/manga-update \
  -H "Content-Type: application/json" \
  -d '{"manga_id": 1, "title": "Test", "changes": ["description", "cover_image"]}'
```

## Key Benefits

✅ **See exactly what changed** - No more guessing  
✅ **Before/after values** - Clear comparison  
✅ **Current chapter displayed** - Always visible  
✅ **Multiple field updates** - All changes shown  
✅ **Beautiful formatting** - Easy to read  
✅ **Backward compatible** - Old notifications still work  

## JSON Structure

```json
{
  "type": "NEW_CHAPTER",
  "manga_id": 123,
  "title": "Manga Title",
  "message": "New chapter 50 available (was chapter 49)",
  "timestamp": "2025-12-02T14:30:45Z",
  "data": {
    "chapter": 50
  },
  "changes": [
    {
      "field": "current_chapter",
      "old_value": 49,
      "new_value": 50
    }
  ]
}
```
