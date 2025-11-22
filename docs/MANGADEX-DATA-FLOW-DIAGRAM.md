# MangaDex to Database Data Flow

## Visual Data Mapping

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          MangaDex API Response                               │
├─────────────────────────────────────────────────────────────────────────────┤
│ {                                                                            │
│   "data": {                                                                  │
│     "id": "a77742b1-befd-49a4-bff5-1ad4e6b0ef7b",  ──────────┐              │
│     "type": "manga",                                          │              │
│     "attributes": {                                           │              │
│       "title": {                                              │              │
│         "en": "Chainsaw Man" ────────────────────────────┐   │              │
│       },                                                  │   │              │
│       "description": {                                    │   │              │
│         "en": "Broke young man..." ──────────────────┐   │   │              │
│       },                                              │   │   │              │
│       "status": "ongoing", ───────────────────────┐  │   │   │              │
│       "lastChapter": "180", ──────────────────┐   │  │   │   │              │
│       "tags": [                               │   │  │   │   │              │
│         {                                     │   │  │   │   │              │
│           "attributes": {                     │   │  │   │   │              │
│             "name": {"en": "Action"},  ───┐  │   │  │   │   │              │
│             "group": "genre"               │  │   │  │   │   │              │
│           }                                │  │   │  │   │   │              │
│         },                                 │  │   │  │   │   │              │
│         {                                  │  │   │  │   │   │              │
│           "attributes": {                  │  │   │  │   │   │              │
│             "name": {"en": "Comedy"}, ──┐ │  │   │  │   │   │              │
│             "group": "genre"            │ │  │   │  │   │   │              │
│           }                             │ │  │   │  │   │   │              │
│         }                               │ │  │   │  │   │   │              │
│       ]                                 │ │  │   │  │   │   │              │
│     },                                  │ │  │   │  │   │   │              │
│     "relationships": [                  │ │  │   │  │   │   │              │
│       {                                 │ │  │   │  │   │   │              │
│         "type": "author",               │ │  │   │  │   │   │              │
│         "attributes": {                 │ │  │   │  │   │   │              │
│           "name": "Fujimoto Tatsuki" ─┐ │ │  │   │  │   │   │              │
│         }                             │ │ │  │   │  │   │   │              │
│       },                              │ │ │  │   │  │   │   │              │
│       {                               │ │ │  │   │  │   │   │              │
│         "type": "cover_art",          │ │ │  │   │  │   │   │              │
│         "attributes": {               │ │ │  │   │  │   │   │              │
│           "fileName": "abc.jpg" ────┐ │ │ │  │   │  │   │   │              │
│         }                           │ │ │ │  │   │  │   │   │              │
│       }                             │ │ │ │  │   │  │   │   │              │
│     ]                               │ │ │ │  │   │  │   │   │              │
│   }                                 │ │ │ │  │   │  │   │   │              │
│ }                                   │ │ │ │  │   │  │   │   │              │
└─────────────────────────────────────┼─┼─┼─┼──┼───┼──┼───┼───┼──────────────┘
                                      │ │ │ │  │   │  │   │   │
                        ┌─────────────┘ │ │ │  │   │  │   │   │
                        │ ┌─────────────┘ │ │  │   │  │   │   │
                        │ │ ┌─────────────┘ │  │   │  │   │   │
                        │ │ │ ┌─────────────┘  │   │  │   │   │
                        │ │ │ │ ┌──────────────┘   │  │   │   │
                        │ │ │ │ │ ┌────────────────┘  │   │   │
                        │ │ │ │ │ │ ┌──────────────────┘   │   │
                        │ │ │ │ │ │ │                      │   │
                        ▼ ▼ ▼ ▼ ▼ ▼ ▼                      ▼   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Data Extraction Layer                              │
├─────────────────────────────────────────────────────────────────────────────┤
│ ExtractMangaMetadata():                                                      │
│   1. coverURL = "https://uploads.mangadex.org/covers/" + id + "/" + fileName│
│   2. slug = generateSlug(title)  // "chainsaw-man"                          │
│   3. totalChapters = parseInt(lastChapter)  // 180                          │
│   4. Filter tags where group == "genre"                                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          PostgreSQL Database                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                          manga table                                 │   │
│  ├──────────────┬───────────────────────────────────────────────────────┤   │
│  │ id           │ 123 (auto-generated)                                  │   │
│  │ mangadex_id  │ "a77742b1-befd-49a4-bff5-1ad4e6b0ef7b" ◄─────────────┤   │
│  │ slug         │ "chainsaw-man" ◄──────────────────────────────────────┤   │
│  │ title        │ "Chainsaw Man" ◄──────────────────────────────────────┤   │
│  │ author       │ "Fujimoto Tatsuki" ◄──────────────────────────────────┤   │
│  │ status       │ "ongoing" ◄───────────────────────────────────────────┤   │
│  │ total_chap..│ 180 ◄─────────────────────────────────────────────────┤   │
│  │ description  │ "Broke young man..." ◄────────────────────────────────┤   │
│  │ cover_url    │ "https://uploads.mangadex.org/covers/a77.../abc.jpg"◄┤   │
│  │ created_at   │ 2025-11-21 10:00:00                                   │   │
│  │ last_synced..│ 2025-11-21 10:00:00                                   │   │
│  └──────────────┴───────────────────────────────────────────────────────┘   │
│                                    │                                         │
│                                    │ Foreign Key (manga_id)                  │
│                                    ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      manga_genres table                              │   │
│  ├───────────┬──────────┬──────────────────────────────────────────────┤   │
│  │ manga_id  │ genre_id │                                              │   │
│  ├───────────┼──────────┤                                              │   │
│  │ 123       │ 5        │ ◄──── Links to Action                        │   │
│  │ 123       │ 12       │ ◄──── Links to Comedy                        │   │
│  └───────────┴──────────┘                                              │   │
│                  │                                                           │
│                  │ Foreign Key (genre_id)                                   │
│                  ▼                                                           │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         genres table                                 │   │
│  ├─────┬──────────────────────────────────────────────────────────────┤   │
│  │ id  │ name                                                          │   │
│  ├─────┼──────────────────────────────────────────────────────────────┤   │
│  │ 5   │ "Action" ◄────────────────────────────────────────────────────┤   │
│  │ 12  │ "Comedy" ◄────────────────────────────────────────────────────┤   │
│  └─────┴──────────────────────────────────────────────────────────────┘   │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Field-by-Field Extraction

### 1. Title Extraction
```
API: attributes.title = {"en": "Chainsaw Man", "ja": "チェンソーマン"}
                            ↓
Logic: Prefer "en", fallback to any available
                            ↓
DB: title = "Chainsaw Man"
```

### 2. Author Extraction
```
API: relationships = [
       {type: "author", attributes: {name: "Fujimoto Tatsuki"}},
       {type: "artist", attributes: {name: "Fujimoto Tatsuki"}}
     ]
                            ↓
Logic: Find where type == "author", extract attributes.name
                            ↓
DB: author = "Fujimoto Tatsuki"
```

### 3. Cover URL Construction
```
API: {
       id: "a77742b1-befd-49a4-bff5-1ad4e6b0ef7b",
       relationships: [{
         type: "cover_art",
         attributes: {fileName: "abc123.jpg"}
       }]
     }
                            ↓
Logic: Construct URL = "https://uploads.mangadex.org/covers/" + id + "/" + fileName
                            ↓
DB: cover_url = "https://uploads.mangadex.org/covers/a77742b1-befd-49a4-bff5-1ad4e6b0ef7b/abc123.jpg"
```

### 4. Genre Extraction
```
API: tags = [
       {attributes: {name: {en: "Action"}, group: "genre"}},
       {attributes: {name: {en: "Shounen"}, group: "demographic"}},  ← SKIP (not genre)
       {attributes: {name: {en: "Comedy"}, group: "genre"}},
       {attributes: {name: {en: "Gore"}, group: "theme"}}  ← SKIP (not genre)
     ]
                            ↓
Logic: Filter where group == "genre", extract name.en
                            ↓
Genres: ["Action", "Comedy"]
                            ↓
DB: genres table: INSERT ["Action", "Comedy"] (if not exist)
    manga_genres table: INSERT (manga_id=123, genre_id=5), (manga_id=123, genre_id=12)
```

### 5. Status Mapping
```
API: attributes.status = "ongoing" | "completed" | "hiatus" | "cancelled"
                            ↓
Logic: Direct mapping (MangaDex uses same values)
                            ↓
DB: status = "ongoing" | "completed" | "hiatus"
```

### 6. Total Chapters Parsing
```
API: attributes.lastChapter = "180" (string)
                            ↓
Logic: Parse as integer, handle empty/null
                            ↓
DB: total_chapters = 180 (integer)
```

### 7. Slug Generation
```
API: attributes.title.en = "Chainsaw Man"
                            ↓
Logic: 
  - Lowercase: "chainsaw man"
  - Replace spaces: "chainsaw-man"
  - Remove special chars
  - Remove duplicate hyphens
                            ↓
DB: slug = "chainsaw-man"
```

## Complete Data Consistency Check

Your synced data should match this structure from `scraped_data.json`:

```json
{
  "mangas": [
    {
      "id": "a77742b1-befd-49a4-bff5-1ad4e6b0ef7b",     // ← mangadex_id
      "slug": "chainsaw-man",                            // ← generated
      "title": "Chainsaw Man",                           // ← attributes.title.en
      "author": "Fujimoto Tatsuki",                      // ← relationships[author].name
      "status": "ongoing",                               // ← attributes.status
      "total_chapters": 180,                             // ← parse(attributes.lastChapter)
      "description": "Broke young man...",               // ← attributes.description.en
      "cover_url": "https://uploads.mangadex.org/...",   // ← constructed URL
      "genres": ["Action", "Comedy"]                     // ← filter tags by group="genre"
    }
  ]
}
```

## SQL Verification Queries

### Check Author Population
```sql
SELECT 
    COUNT(*) as total_manga,
    COUNT(author) as with_author,
    COUNT(*) - COUNT(author) as missing_author,
    ROUND(100.0 * COUNT(author) / COUNT(*), 2) as author_percentage
FROM manga
WHERE mangadex_id IS NOT NULL;
```

Expected: `author_percentage` > 95%

### Check Genre Population
```sql
SELECT 
    m.title,
    m.author,
    COUNT(mg.genre_id) as genre_count,
    STRING_AGG(g.name, ', ') as genres
FROM manga m
LEFT JOIN manga_genres mg ON m.id = mg.manga_id
LEFT JOIN genres g ON mg.genre_id = g.id
WHERE m.mangadex_id IS NOT NULL
GROUP BY m.id, m.title, m.author
ORDER BY genre_count DESC
LIMIT 10;
```

Expected: Most manga should have 2-5 genres

### Check Total Chapters
```sql
SELECT 
    status,
    COUNT(*) as count,
    ROUND(AVG(total_chapters), 2) as avg_chapters,
    MIN(total_chapters) as min_chapters,
    MAX(total_chapters) as max_chapters
FROM manga
WHERE mangadex_id IS NOT NULL
GROUP BY status;
```

Expected:
- `completed` manga: avg 20-100 chapters
- `ongoing` manga: varies widely
- Some may have 0 (newly published)

### Check Cover URLs
```sql
SELECT 
    COUNT(*) as total,
    COUNT(cover_url) as with_cover,
    COUNT(*) - COUNT(cover_url) as missing_cover
FROM manga
WHERE mangadex_id IS NOT NULL;
```

Expected: All manga should have cover URLs (100%)

## Summary

✅ **8 Required Fields Extracted:**
1. mangadex_id ← `data.id`
2. title ← `attributes.title.en`
3. author ← `relationships[type="author"].attributes.name`
4. status ← `attributes.status`
5. total_chapters ← `parseInt(attributes.lastChapter)`
6. description ← `attributes.description.en`
7. cover_url ← `construct from relationships[type="cover_art"]`
8. genres ← `filter attributes.tags where group="genre"`

✅ **Generated Fields:**
- slug ← `generateSlug(title)`

✅ **Stored Across 3 Tables:**
- `manga` table: Core metadata
- `genres` table: Genre names (reusable)
- `manga_genres` table: Many-to-many relationships

This ensures **100% data consistency** with your existing scraped data format! 🎯
