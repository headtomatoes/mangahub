
# Notification Sync Implementation for Offline Users (Detailed Code Workflow)

## What I'm adding
This file now expands the existing explanation into a developer-focused, code-level workflow for the UDP notification mechanism implemented in this repository. It references the main files in `internal/microservices/udp-server/` and the related HTTP notification handlers so you can trace the end-to-end flow.

## Goals
- Ensure no notification is lost when users are offline (persistence + sync)
- Use UDP for low-latency pushes to online users
- Provide a reconnection sync path for missed notifications

## Files / Components Involved
- `internal/microservices/udp-server/broadcaster.go` — prepares notifications for users and persists them
- `internal/microservices/udp-server/server.go` — UDP server loop; handles SUBSCRIBE, UNSUBSCRIBE and sync on reconnect
- `internal/microservices/udp-server/subscriber.go` — subscription manager (maps userID -> UDP address, manages online set)
- `internal/microservices/udp-server/notification.go` — serialization / message shapes used over UDP (e.g. `Notification.ToJSON()`)
- `internal/microservices/http-api/handler/notification_handler.go` — HTTP endpoints to fetch/mark notifications read
- Database repos: notificationRepo, libraryRepo used by broadcaster and server

## Data shapes and contracts
- Notification (UDP payload): { Type, MangaID, Title, Message, Timestamp }
- DB Notification model: { ID, UserID, Type, MangaID, Title, Message, Read, CreatedAt }
- Subscription request: client sends a small UDP message (e.g. { Action: "SUBSCRIBE", UserID: "..." }) to register its current UDP address

## End-to-end Workflow (step-by-step)

1) New content event (publisher side)
   - A new chapter / content event triggers the broadcaster flow. In code: `Broadcaster.BroadcastToLibraryUsers(ctx, mangaID, notification)`.

2) Resolve recipients
   - `BroadcastToLibraryUsers` calls `libraryRepo.GetUserIDsByMangaID(ctx, mangaID)` to get all user IDs who have the manga in their library.

3) Persist notifications for all recipients
   - For each userID, the broadcaster creates a `models.Notification` with `Read = false` and calls `notificationRepo.Create(ctx, dbNotification)`.
   - This ensures durability: even if UDP delivery fails, the notification is stored.

4) Send realtime UDP to online subscribers
   - The broadcaster asks the subscription manager for active addresses: `subscribers := b.subManager.GetByUserIDs(userIDs)` (returns only online users)
   - For each online subscriber it serializes the `Notification` (via `Notification.ToJSON()`) and calls `s.conn.WriteToUDP(payload, addr)`.
   - The code treats UDP as best-effort: errors are logged, but persistence guarantees later sync.

5) Client-side: subscription
   - Client opens a UDP socket and sends a SUBSCRIBE message containing its `UserID`.
   - Server receives the SUBSCRIBE and registers the mapping in `subManager` (userID -> net.UDPAddr).

6) Subscribe acknowledgement and non-blocking sync
   - Server sends a confirmation UDP message back to the client right away: a small `Notification` of type `NotificationSubscribe`.
   - To avoid blocking the confirmation, the server spawns a goroutine: `go s.syncMissedNotifications(req.UserID, addr)` to push missed notifications asynchronously.

7) Sync missed notifications (reconnection path)
   - `syncMissedNotifications(userID, addr)`:
     - Uses `notificationRepo.GetUnreadByUser(ctx, userID)` to fetch all unread notifications for that user.
     - Iterates them, converts each DB model to UDP `Notification` payload, and sends via `s.conn.WriteToUDP(payload, addr)`.
     - Sleeps a small amount between sends (e.g. 50ms) to avoid overwhelming clients or network bursts.
     - The method does not assume UDP delivery; it's a push-only sync. If you need strong delivery, add ACKs or switch to TCP for sync or use reliable retransmit.

8) Mark read / cleanup
   - The user may mark notifications as read via HTTP endpoints provided by `notification_handler.go`:
     - `GET /api/notifications/unread` — to list unread
     - `PUT /api/notifications/:id/read` — to mark one as read
     - `PUT /api/notifications/read-all` — to mark all as read
   - Optionally, the server could auto-mark notifications as read after successful sync (not implemented by default).

## Subscription manager (behavioral notes)
- The subManager keeps an in-memory mapping of online users to their UDP addresses. Typical operations:
  - Add(userID, addr)
  - Remove(userID)
  - GetByUserIDs([]userID) -> []addr
- Since this is in-memory, server restarts clear online state; persisted DB guarantees re-sync on next SUBSCRIBE.

## Error handling, guarantees and trade-offs
- UDP is low-latency but unreliable. The system guarantees eventual delivery by persisting notifications in DB and re-sending on reconnect.
- Current behavior is "at-least-stored" but not strictly "at-least-once" over UDP; clients may receive duplicates (server should let clients dedupe by notification ID or timestamp).
- syncMissedNotifications pushes unread items but does not wait for client ACKs — adding ACKs would allow marking delivered notifications as read automatically.
- For large backlogs, consider batching (e.g. send an array of notifications) or limiting N recent notifications to avoid flooding clients.

## Edge cases and recommended handling
- Client reconnects from a new NAT/port: SUBSCRIBE registers the new `net.UDPAddr`, and sync will go to the new address.
- Multiple devices per user: subManager may only store one address per userID; to support multiple devices store a list of addresses per user.
- Large backlog for a user: either stream in chunks and request client ACKs per-chunk or limit to last-N notifications.
- Clock skew: use server-created timestamps (`CreatedAt`) for ordering rather than client timestamps.

## Tests and verification
- Unit tests exist near `internal/microservices/udp-server/` (files like `server_test.go`, `subscriber_test.go`, `broadcaster_test.go`, `notification_test.go`). Run them with:

```bash
# from repository root
go test ./internal/microservices/udp-server -v
```

- Manual smoke test:
  1. Start the UDP server (run the binary under `cmd/udp-server` or `go run` the server main).
  2. Use the provided `udp-client` (in `cmd/udp-client`) or a small test client to SUBSCRIBE.
  3. Trigger a broadcast (via code or an API that produces a new chapter event).
  4. Observe immediate UDP delivery for online client and DB persistence for all users.

## Small improvements to consider (non-breaking)
- Add an ACK/confirm message from client to server for sync deliveries; server can then mark those notifications as delivered/read.
- Implement batching for sync (send an array of notifications as a single UDP payload) to reduce per-packet overhead.
- Support multiple addresses per user (multiple devices) in `subManager`.
- Add metrics for: notifications persisted, realtime UDP sent, sync count per reconnect, and average sync time.

## Summary (requirements coverage)
- Persist notifications for all users: Done (broadcaster stores unread notifications in DB).
- Deliver realtime notifications via UDP to online users: Done (broadcaster sends to addresses from `subManager`).
- Sync missed notifications on reconnect: Done (server `syncMissedNotifications` is triggered after SUBSCRIBE).

If you want, I can also:
- Add client-side example (Go or JS) showing SUBSCRIBE and ACK flow.
- Implement auto-mark-as-read-after-ACK in the server and update HTTP handlers accordingly.

