# AniList Sync Service Guide

## Overview

The AniList Sync Service integrates manga data from AniList.co into MangaHub using GraphQL API queries. It provides:

- **Initial Bulk Sync**: Import 150 popular manga entries
- **New Manga Detection**: Poll for newly added manga every 24 hours
- **Chapter Updates**: Check for chapter count changes every 48 hours
- **Real-time Notifications**: UDP notifications for all updates

## Architecture

```
┌─────────────────────────────────────────┐
│         AniList GraphQL API             │
│  https://graphql.anilist.co             │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      AniList Sync Service               │
│  • GraphQL Client                       │
│  • Rate Limiter (1 req/s)             │
│  • Worker Pool (10 workers)             │
│  • Metadata Extractor                   │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│        PostgreSQL Database              │
│  • manga (anilist_id column)            │
│  • sync_state tracking                  │
└─────────────────────────────────────────┘
```

## Features

### 1. Initial Sync
- Fetches 50 most popular manga
- Extracts complete metadata (title, author, genres, etc.)
- Stores in database with AniList ID tracking
- Sends UDP notifications for each new manga

### 2. New Manga Polling
- Runs every 24 hours
- Detects manga updated since last poll
- Uses `updatedAt` timestamp for tracking
- Incremental updates only

### 3. Chapter Update Checking
- Runs every 48 hours
- Checks tracked manga for chapter count changes
- Updates metadata if changed
- Sends notifications for chapter updates

## Configuration

Environment variables:

```bash
ANILIST_SYNC_INITIAL_COUNT=50  # Number of manga for initial sync
ANILIST_SYNC_WORKERS=10         # Concurrent workers
ANILIST_RATE_CONCURRENCY=5      # Max concurrent API calls
DATABASE_URL=postgres://...      # Database connection
UDP_SERVER_URL=http://localhost:8085  # Notification server
```

## Usage

### Run with Docker
```bash
docker-compose up anilist-sync
```

### Run Standalone
```bash
go run cmd/anilist-sync/main.go
```

### Development with Hot Reload
```bash
air -c .air.anilist-sync.toml
```

## API Rate Limits

AniList allows ~90 requests per minute:
- Rate limiter: 1.5 requests/second
- Burst capacity: 5 requests
- Exponential backoff on errors (1s → 32s max)

## Database Schema

The service uses these columns in the `manga` table:

```sql
anilist_id                  INTEGER UNIQUE
anilist_last_synced_at      TIMESTAMPTZ
anilist_last_chapter_check  TIMESTAMPTZ
```

## Notification Events

1. **New Manga**: `type: "new_manga", source: "anilist"`
2. **Chapter Update**: `type: "chapter_update", old_chapters, new_chapters`
3. **Manga Update**: `type: "manga_update"` (metadata changes)

## Comparison with MangaDex Sync

| Feature | MangaDex | AniList |
|---------|----------|---------|
| API Type | REST | GraphQL |
| Rate Limit | 5/sec | 1.5/sec |
| Chapter Data | Full chapters | Count only |
| Metadata Quality | High | Very High |
| Update Frequency | Real-time | Delayed |
| Cover Images | Multiple | Large + Medium |

## Troubleshooting

### High Error Rate
- Check rate limit configuration
- Verify network connectivity
- Review AniList API status

### Missing Manga
- Check `anilist_id` column exists
- Verify initial sync completed
- Review sync_state table

### Duplicate Entries
- Ensure unique constraint on `anilist_id`
- Check slug generation logic

## Monitoring

Check sync status:
```sql
SELECT * FROM sync_state WHERE sync_type LIKE 'anilist%';
```

View recently synced manga:
```sql
SELECT id, title, anilist_id, anilist_last_synced_at 
FROM manga 
WHERE anilist_id IS NOT NULL 
ORDER BY anilist_last_synced_at DESC 
LIMIT 10;
```