# UDP Client CLI Guide

## Overview

The UDP client CLI allows users to receive real-time push notifications from the MangaHub server. Users can subscribe to notifications about new manga, new chapters, and manga updates.

## Features

- **Real-time Notifications**: Receive instant push notifications via UDP
- **Auto-reconnect**: Periodic ping messages keep the connection alive
- **Notification Types**:
  - New manga additions
  - New chapter releases (for manga in your library)
  - Manga updates
- **Connection Statistics**: Track notifications received and connection uptime

## Prerequisites

- You must be logged in (`mangahubCLI auth login`)
- UDP server must be running (default: `127.0.0.1:8082`)

## Commands

### 1. Listen for Notifications

Start listening for real-time notifications:

```bash
mangahubCLI udp listen
```

**Options:**
- `--server`: Specify UDP server address (default: `127.0.0.1:8082`)

**Example:**
```bash
mangahubCLI udp listen --server 192.168.1.100:8082
```

**What it does:**
1. Connects to the UDP notification server
2. Subscribes using your authenticated user ID
3. Displays incoming notifications in real-time
4. Sends periodic ping messages (every 30 seconds)
5. Press `Ctrl+C` to stop and disconnect

**Output Example:**
```
🔌 Connecting to UDP notification server...
   Server: 127.0.0.1:8082
   User: john_doe (ID: user123)

✓ Subscribed to notifications (User ID: user123)

📡 Listening for notifications... (Press Ctrl+C to stop)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  📚 NEW CHAPTER AVAILABLE!
  📖 Manga: One Piece
  🆔 Manga ID: 42
  📄 Chapter: 1095
  💬 New chapter available
  ⏰ Time: 2025-12-01 14:30:45
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### 2. Test Connection

Test the UDP server connection:

```bash
mangahubCLI udp test
```

**What it does:**
- Connects to the UDP server
- Subscribes with your user ID
- Verifies the connection is successful
- Disconnects immediately

**Use case:** Quick check to ensure the UDP server is reachable before starting a listening session.

### 3. View Statistics

Display information about UDP connection statistics:

```bash
mangahubCLI udp stats
```

**Note:** Stats are only available during an active listening session.

## Notification Types

### 1. Subscription Confirmation
Received when you successfully subscribe:
```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  ✅ Successfully subscribed to notifications
  ⏰ Time: 2025-12-01 14:25:30
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### 2. New Manga
Notifies all users when new manga is added:
```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  🆕 NEW MANGA ADDED!
  📖 Title: Attack on Titan
  🆔 Manga ID: 99
  💬 New manga added: Attack on Titan
  ⏰ Time: 2025-12-01 15:00:00
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### 3. New Chapter
Notifies users who have the manga in their library:
```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  📚 NEW CHAPTER AVAILABLE!
  📖 Manga: Naruto
  🆔 Manga ID: 15
  📄 Chapter: 700
  💬 New chapter available
  ⏰ Time: 2025-12-01 16:30:00
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

## Configuration

### Environment Variables

- `MANGAHUB_UDP_ADDR`: Set default UDP server address
  ```bash
  export MANGAHUB_UDP_ADDR="192.168.1.100:8082"
  mangahubCLI udp listen
  ```

### Connection Settings

The UDP client automatically:
- Sends ping messages every **30 seconds** to keep the connection alive
- Handles server timeouts gracefully
- Reconnects and syncs missed notifications (handled by server)

## Architecture

### Client-Server Communication

1. **Subscribe**: Client sends SUBSCRIBE message with user ID
2. **Listen**: Client listens for incoming UDP packets
3. **Ping**: Client sends periodic PING messages
4. **Notifications**: Server broadcasts relevant notifications
5. **Unsubscribe**: Client sends UNSUBSCRIBE before disconnecting

### Message Format

All messages are JSON-encoded:

**Subscribe Request:**
```json
{
  "type": "SUBSCRIBE",
  "user_id": "user123"
}
```

**Notification:**
```json
{
  "type": "NEW_CHAPTER",
  "manga_id": 42,
  "title": "One Piece",
  "message": "New chapter available",
  "timestamp": "2025-12-01T14:30:45Z",
  "data": {
    "chapter": 1095
  }
}
```

## Troubleshooting

### Connection Failed
**Error:** `failed to connect: connection refused`

**Solution:**
- Ensure the UDP server is running
- Check the server address and port
- Verify firewall settings

### Not Logged In
**Error:** `not logged in, please run 'mangahubCLI auth login' first`

**Solution:**
```bash
mangahubCLI auth login
```

### No Notifications Received
**Possible causes:**
1. **No manga in library**: Add manga to your library to receive chapter notifications
2. **No new updates**: Server only sends notifications when events occur
3. **Connection timeout**: Check if connection is still active (ping messages should keep it alive)

## Best Practices

1. **Keep the session active**: The listener runs indefinitely until you press `Ctrl+C`
2. **Use in a separate terminal**: Run the listener in a dedicated terminal window
3. **Check library**: Ensure you have manga in your library to receive chapter notifications
4. **Test first**: Use `udp test` before starting a long listening session

## Examples

### Basic Usage
```bash
# Login first
mangahubCLI auth login

# Start listening
mangahubCLI udp listen
```

### Custom Server
```bash
# Connect to custom server
mangahubCLI udp listen --server 192.168.1.50:9000
```

### Quick Test
```bash
# Test connection without listening
mangahubCLI udp test
```

## Related Commands

- `mangahubCLI auth login` - Authenticate before using UDP client
- `mangahubCLI library add` - Add manga to library to receive chapter notifications
- `mangahubCLI library list` - View manga in your library

## Technical Details

### Connection Statistics

During a listening session, the client tracks:
- Connection start time
- Total notifications received
- Last notification timestamp
- Last ping timestamp
- Session uptime

Statistics are displayed when you disconnect (Ctrl+C).

### Network Protocol

- **Protocol**: UDP (User Datagram Protocol)
- **Port**: 8082 (default)
- **Format**: JSON
- **Max packet size**: 8192 bytes

### Implementation

The UDP client is implemented in:
- Client logic: `cmd/cli/command/client/udp_client.go`
- CLI commands: `cmd/cli/command/udp.go`
- Server: `internal/microservices/udp-server/`
