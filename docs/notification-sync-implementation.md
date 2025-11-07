# Notification Sync Implementation for Offline Users

## Overview
This document describes how the MangaHub notification system handles offline users and syncs missed notifications when they reconnect.

## Requirements
✅ **Since UDP does not guarantee delivery for offline users:**
- The system saves all unread notifications to the database for both online and offline users
- When a user reconnects, the server detects it and automatically pushes missed updates via UDP

## Implementation Details

### 1. Notification Storage (Already Implemented)
**File:** `internal/microservices/udp-server/broadcaster.go`

```go
// BroadcastToLibraryUsers sends notification AND stores it for offline users
func (b *Broadcaster) BroadcastToLibraryUsers(ctx context.Context, mangaID int64, notification *Notification) error {
    // Get all users who have this manga in their library
    userIDs, err := b.libraryRepo.GetUserIDsByMangaID(ctx, mangaID)
    
    // Store notification in database for ALL users (online and offline)
    for _, userID := range userIDs {
        dbNotification := &models.Notification{
            UserID:  userID,
            Type:    string(notification.Type),
            MangaID: mangaID,
            Title:   notification.Title,
            Message: notification.Message,
            Read:    false, // Initially unread
        }
        b.notificationRepo.Create(ctx, dbNotification)
    }
    
    // Send to currently online subscribers via UDP
    subscribers := b.subManager.GetByUserIDs(userIDs)
    // ... send UDP notifications ...
}
```

**Key Points:**
- Notifications are saved to the database for **all users** (both online and offline)
- Each notification has a `read` flag (default: `false`)
- Online users receive immediate UDP notification
- Offline users will get notified when they reconnect

### 2. Sync on Reconnection (NEW Implementation)
**File:** `internal/microservices/udp-server/server.go`

#### When User Subscribes (Reconnects):
```go
case "SUBSCRIBE":
    s.subManager.Add(req.UserID, addr)
    log.Printf("User %s subscribed from %s", req.UserID, addr.String())
    
    // Send confirmation
    confirmation := &Notification{
        Type:      NotificationSubscribe,
        Message:   "Successfully subscribed to notifications",
        Timestamp: time.Now(),
    }
    s.conn.WriteToUDP(confirmation.ToJSON(), addr)
    
    // SYNC: Push missed notifications to reconnecting user
    go s.syncMissedNotifications(req.UserID, addr)
```

#### Sync Method:
```go
func (s *Server) syncMissedNotifications(userID string, addr *net.UDPAddr) {
    // Get all unread notifications from database
    unreadNotifs, err := s.notificationRepo.GetUnreadByUser(ctx, userID)
    
    // Send each notification via UDP
    for _, dbNotif := range unreadNotifs {
        notification := &Notification{
            Type:      NotificationType(dbNotif.Type),
            MangaID:   dbNotif.MangaID,
            Title:     dbNotif.Title,
            Message:   dbNotif.Message,
            Timestamp: dbNotif.CreatedAt,
        }
        s.conn.WriteToUDP(notification.ToJSON(), addr)
        time.Sleep(50 * time.Millisecond) // Avoid overwhelming client
    }
}
```

**Key Points:**
- Runs asynchronously (in a goroutine) to not block the subscribe confirmation
- Retrieves all unread notifications for the user
- Sends them one by one via UDP
- Includes a small delay (50ms) between sends to avoid overwhelming the client

### 3. Mark as Read (Via HTTP API)
**File:** `internal/microservices/http-api/handler/notification_handler.go`

Users can mark notifications as read via HTTP endpoints:
- `GET /api/notifications/unread` - Fetch all unread notifications
- `PUT /api/notifications/:id/read` - Mark specific notification as read
- `PUT /api/notifications/read-all` - Mark all as read

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ NEW CHAPTER PUBLISHED                                        │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
    ┌────────────────────────────┐
    │ BroadcastToLibraryUsers()  │
    └────────────┬───────────────┘
                 │
        ┌────────┴────────┐
        ▼                 ▼
┌───────────────┐  ┌─────────────────┐
│ Save to DB    │  │ Send UDP to     │
│ for ALL users │  │ ONLINE users    │
│ (read=false)  │  │ only            │
└───────────────┘  └─────────────────┘
        │
        │
        ▼
┌─────────────────────────────────────┐
│ OFFLINE USER                         │
│ - Notification stored in DB          │
│ - Will receive when reconnecting     │
└──────────────┬──────────────────────┘
               │
               │ (User reconnects)
               ▼
    ┌──────────────────────┐
    │ SUBSCRIBE request    │
    └──────────┬───────────┘
               │
               ▼
    ┌──────────────────────────────┐
    │ syncMissedNotifications()    │
    │ - Fetch unread from DB       │
    │ - Send all via UDP           │
    └──────────────────────────────┘
               │
               ▼
    ┌──────────────────────────────┐
    │ User receives all missed     │
    │ notifications via UDP        │
    └──────────────────────────────┘
               │
               ▼
    ┌──────────────────────────────┐
    │ User marks as read via       │
    │ HTTP API (optional)          │
    └──────────────────────────────┘
```

## Testing Scenarios

### Scenario 1: User is Online
1. New chapter published
2. Notification saved to DB (read=false)
3. UDP notification sent immediately
4. User receives notification in real-time

### Scenario 2: User is Offline
1. New chapter published
2. Notification saved to DB (read=false)
3. UDP send fails (user offline)
4. **User reconnects later**
5. **Server automatically syncs all missed notifications**
6. User receives all missed notifications

### Scenario 3: User Misses Multiple Notifications
1. Multiple chapters published while user offline
2. All notifications saved to DB
3. User reconnects
4. Server sends all missed notifications (with 50ms delay between each)
5. User receives all updates in order

## Advantages of This Approach

1. **Reliability:** No notification is lost - all are persisted in DB
2. **Automatic Sync:** Users automatically receive missed updates on reconnection
3. **UDP Efficiency:** Real-time notifications for online users
4. **Fallback:** HTTP API allows users to manually fetch missed notifications if needed
5. **Scalable:** Async processing doesn't block the subscription flow

## Future Enhancements (Optional)

1. **Auto-mark as read after sync:** Automatically mark notifications as read after successful UDP delivery
2. **Batch sync:** Send notifications in batches instead of one by one
3. **Sync confirmation:** Client sends ACK after receiving synced notifications
4. **Priority queue:** Send most important notifications first
5. **Max sync limit:** Only sync last N notifications to avoid overload
