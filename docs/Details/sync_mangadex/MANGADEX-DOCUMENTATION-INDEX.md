# MangaDex Integration Documentation Index

## 📚 Documentation Overview

This directory contains comprehensive documentation for integrating MangaDex real-time data synchronization with your MangaHub system.

---

## 📄 Documents

### 1. **MANGADEX-REALTIME-PIPELINE-PROPOSAL.md** (Main Proposal)
**38.9 KB | Complete System Architecture**

The comprehensive proposal document covering:
- ✅ Complete system architecture with diagrams
- ✅ Database schema extensions (chapters, sync_state tables)
- ✅ Three synchronization workflows (initial, new manga, chapter updates)
- ✅ MangaDex API integration strategy with rate limiting
- ✅ UDP notification integration
- ✅ Service implementation design
- ✅ Docker deployment configuration
- ✅ Monitoring & observability
- ✅ Why this architecture is reliable
- ✅ 5-week implementation roadmap
- ✅ Security & testing considerations

**Start here** for the complete proposal overview.

---

### 2. **MANGADEX-DATA-EXTRACTION-GUIDE.md** (Implementation Guide)
**14.8 KB | Complete Metadata Extraction**

Detailed guide for extracting **all manga metadata** from MangaDex API:
- ✅ All 8 required fields (mangadex_id, title, author, status, total_chapters, description, cover_url, genres)
- ✅ Correct API request patterns with `includes[]` parameters
- ✅ Complete code examples for data extraction
- ✅ Genre filtering and slug generation
- ✅ Database storage with GORM
- ✅ Verification queries
- ✅ Common mistakes to avoid
- ✅ Complete initial sync function

**Use this** when implementing the sync service code.

---

### 3. **MANGADEX-DATA-FLOW-DIAGRAM.md** (Visual Reference)
**19.5 KB | Visual Data Mapping**

Visual ASCII diagrams showing:
- ✅ Complete data flow from MangaDex API → Database
- ✅ Field-by-field extraction examples
- ✅ Table relationship diagrams
- ✅ Step-by-step transformation logic
- ✅ SQL verification queries
- ✅ Data consistency checks

**Reference this** to understand the exact data transformations.

---

## 🎯 Quick Start Guide

### For Academic Review/Presentation:
1. Read **MANGADEX-REALTIME-PIPELINE-PROPOSAL.md** sections:
   - Section 1: System Architecture Overview (diagram)
   - Section 4: Synchronization Workflows
   - Section 9: Why This Architecture is Reliable

### For Implementation:
1. Read **MANGADEX-DATA-EXTRACTION-GUIDE.md** for code patterns
2. Reference **MANGADEX-DATA-FLOW-DIAGRAM.md** for field mapping
3. Follow **MANGADEX-REALTIME-PIPELINE-PROPOSAL.md** Section 10: Implementation Roadmap

### For Data Verification:
1. Use SQL queries from **MANGADEX-DATA-FLOW-DIAGRAM.md** bottom section
2. Compare output with your `scraped_data.json` structure
3. Run completeness checks from **MANGADEX-DATA-EXTRACTION-GUIDE.md**

---

## 🔑 Key Takeaways

### Complete Metadata Extraction
When syncing from MangaDex, you **must extract**:

1. ✅ **mangadex_id** (UUID from API)
2. ✅ **title** (English or fallback)
3. ✅ **author** (from relationships array)
4. ✅ **status** (ongoing/completed/hiatus)
5. ✅ **total_chapters** (parse from lastChapter)
6. ✅ **description** (English or fallback)
7. ✅ **cover_url** (construct from cover relationship)
8. ✅ **genres** (filter tags where group="genre")

### Critical API Usage
```bash
# ✅ CORRECT: Use includes[] to get all data in one request
GET /manga?limit=100&includes[]=author&includes[]=cover_art

# ❌ WRONG: Without includes (requires N+1 additional API calls)
GET /manga?limit=100
```

### Data Consistency
Your synced data will match the exact structure of `scraped_data.json`:

```json
{
  "id": "mangadex-uuid",
  "slug": "chainsaw-man",
  "title": "Chainsaw Man",
  "author": "Fujimoto Tatsuki",
  "status": "ongoing",
  "total_chapters": 180,
  "description": "Broke young man...",
  "cover_url": "https://uploads.mangadex.org/covers/...",
  "genres": ["Action", "Comedy"]
}
```

---

## 🏗️ System Architecture Summary

```
MangaDex API (Rate: 5 req/sec)
    ↓ includes[]=author&includes[]=cover_art
Sync Service (Polling every 5-15 min)
    ↓ Extract all 8 metadata fields
PostgreSQL Database
    ├── manga table (extended with mangadex_id, last_synced_at)
    ├── chapters table (new)
    ├── sync_state table (new)
    ├── genres table (existing)
    └── manga_genres table (existing)
    ↓ Trigger on new manga/chapters
UDP Notification Server (Port 8085)
    ↓ Broadcast via UDP
Connected Users (Real-time push)
Offline Users (Stored in notifications table)
```

---

## 📊 Database Schema Changes

### Extended `manga` Table
```sql
ALTER TABLE manga ADD COLUMN mangadex_id UUID UNIQUE;
ALTER TABLE manga ADD COLUMN last_synced_at TIMESTAMPTZ;
ALTER TABLE manga ADD COLUMN last_chapter_check TIMESTAMPTZ;
ALTER TABLE manga ADD COLUMN updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
```

### New `chapters` Table
```sql
CREATE TABLE chapters (
    id BIGSERIAL PRIMARY KEY,
    manga_id BIGINT REFERENCES manga(id),
    mangadex_chapter_id UUID UNIQUE NOT NULL,
    chapter_number DECIMAL(10, 2) NOT NULL,
    title TEXT,
    volume TEXT,
    pages INTEGER,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

### New `sync_state` Table
```sql
CREATE TABLE sync_state (
    id SERIAL PRIMARY KEY,
    sync_type TEXT NOT NULL UNIQUE,
    last_run_at TIMESTAMPTZ,
    last_success_at TIMESTAMPTZ,
    last_cursor TEXT,
    status TEXT DEFAULT 'idle',
    metadata JSONB
);
```

---

## 🔄 Synchronization Workflows

### 1. Initial Sync (One-Time)
- Import 100-200 manga with **all metadata**
- Store manga + genres + initial chapters
- Run once on first deployment

### 2. New Manga Detection (Every 5 min)
- Poll `/manga?createdAtSince={cursor}`
- Detect newly published manga
- Broadcast to all users via UDP

### 3. Chapter Updates (Every 15 min)
- Check 50 manga per cycle
- Poll `/manga/{id}/feed` for new chapters
- Notify users with manga in library

---

## 🔌 UDP Notification Integration

Your existing UDP server (port 8085) already supports:

```bash
# New manga (broadcast to all)
POST /notify/new-manga
{"manga_id": 123, "title": "Chainsaw Man"}

# New chapter (notify library users)
POST /notify/new-chapter
{"manga_id": 123, "title": "Chainsaw Man", "chapter": 42}

# Generic update (notify library users)
POST /notify/manga-update
{"manga_id": 123, "title": "...", "message": "..."}
```

---

## 🚀 Implementation Timeline

| Week | Tasks | Deliverables |
|------|-------|--------------|
| **1** | Database migrations, API client, rate limiter | Schema ready, client with retry logic |
| **2** | Initial sync, new manga polling, chapter detection | Core sync working |
| **3** | Error handling, circuit breaker, health checks | Production-ready reliability |
| **4** | Metrics, monitoring, prioritization | Observability complete |
| **5** | Integration tests, load tests, documentation | Ready for production |

---

## 📈 Success Metrics

After implementation, verify:

### Data Completeness
```sql
-- All manga should have author names
SELECT COUNT(*) FROM manga WHERE author IS NULL;  -- Should be ~0

-- All manga should have cover URLs
SELECT COUNT(*) FROM manga WHERE cover_url IS NULL;  -- Should be 0

-- Most manga should have genres
SELECT AVG(genre_count) FROM (
    SELECT COUNT(genre_id) as genre_count
    FROM manga_genres
    GROUP BY manga_id
) t;  -- Should be 2-5
```

### Sync Performance
- Initial sync: 100-200 manga in < 5 minutes
- New manga detection: < 30 seconds per cycle
- Chapter updates: 50 manga checked in < 2 minutes
- API rate limit: Never exceeded (< 5 req/sec)

### Notification Delivery
- Online users: Real-time UDP push (< 100ms)
- Offline users: Stored in DB, synced on reconnect
- No duplicate notifications within 1-hour window

---

## 🛡️ Reliability Features

| Feature | Implementation | Benefit |
|---------|---------------|---------|
| **Rate Limiting** | Token bucket (5/sec) | Prevents API bans |
| **Retry Logic** | Exponential backoff | Handles transient failures |
| **Circuit Breaker** | Open after 10 failures | Protects from cascading failures |
| **Cursor-Based Sync** | Store last timestamp | No data loss |
| **Notification Persistence** | Store before UDP send | Offline user support |
| **Idempotency** | Upsert patterns | Safe to re-run |

---

## 🔍 Common Implementation Mistakes

### ❌ DON'T:
- Skip `includes[]` parameter (causes N+1 API calls)
- Only extract English titles/descriptions (some manga don't have English)
- Include all tags as genres (filter by `group="genre"`)
- Use non-existent `coverURL` field from API
- Parse `lastChapter` as string (convert to integer)

### ✅ DO:
- Always use `includes[]=author&includes[]=cover_art`
- Fallback to any available language if English missing
- Filter tags where `attributes.group == "genre"`
- Construct cover URL: `https://uploads.mangadex.org/covers/{manga_id}/{fileName}`
- Parse and store total_chapters as integer

---

## 📞 Support & Questions

If you need clarification on any section:
1. Check the detailed document listed above
2. Review the code examples in DATA-EXTRACTION-GUIDE
3. Reference the visual diagrams in DATA-FLOW-DIAGRAM
4. Consult the MangaDex API documentation: https://api.mangadex.org/docs/

---

## ✅ Pre-Implementation Checklist

Before starting implementation:

- [ ] Reviewed complete proposal (PIPELINE-PROPOSAL.md)
- [ ] Understood all 8 required metadata fields
- [ ] Database migrations planned (3 new/extended tables)
- [ ] MangaDex API key obtained
- [ ] Rate limiter pattern understood (5 req/sec)
- [ ] UDP notification endpoints tested
- [ ] Docker environment configured
- [ ] Monitoring strategy defined

---

## 🎓 Academic Highlights

For your project presentation/documentation, emphasize:

1. **Real-Time Architecture**: Polling-based sync with cursor pagination
2. **Fault Tolerance**: Multiple layers (retry, circuit breaker, persistence)
3. **Scalability**: Batch processing, rate limiting, horizontal scaling ready
4. **Data Integrity**: ACID transactions, unique constraints, foreign keys
5. **User Experience**: Real-time notifications + offline sync support
6. **Observability**: Metrics, logs, health checks
7. **Security**: API key management, input validation, rate limiting

---

## 📝 Document Change Log

| Date | Document | Changes |
|------|----------|---------|
| 2025-11-21 | PIPELINE-PROPOSAL | Initial creation - complete architecture |
| 2025-11-21 | DATA-EXTRACTION-GUIDE | Added complete metadata extraction guide |
| 2025-11-21 | DATA-FLOW-DIAGRAM | Added visual data flow diagrams |
| 2025-11-21 | All docs | Updated to emphasize complete metadata extraction |

---

## 🎯 Next Steps

1. **Review**: Read MANGADEX-REALTIME-PIPELINE-PROPOSAL.md
2. **Plan**: Decide on implementation timeline (5 weeks recommended)
3. **Setup**: Run database migrations (Section 2 of proposal)
4. **Implement**: Follow DATA-EXTRACTION-GUIDE.md for code
5. **Test**: Use verification queries from DATA-FLOW-DIAGRAM.md
6. **Deploy**: Configure Docker (Section 7 of proposal)
7. **Monitor**: Set up metrics (Section 8 of proposal)

---

**Total Documentation:** ~73 KB across 3 comprehensive documents

**Coverage:** Architecture, Implementation, Data Flow, Testing, Deployment, Monitoring

**Ready for:** Academic review, implementation, and production deployment 🚀
