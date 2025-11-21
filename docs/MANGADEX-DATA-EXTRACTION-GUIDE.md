# MangaDex Complete Data Extraction Guide

## Overview

This guide explains how to extract **all manga metadata** from the MangaDex API during initial sync and continuous polling, ensuring consistency with your existing `scraped_data.json` format.

---

## 🎯 Required Metadata Fields

When syncing manga from MangaDex, you must extract and store these fields:

| Field | Required | Source | Example |
|-------|----------|--------|---------|
| **mangadex_id** | ✅ Yes | `data.id` | `"a77742b1-befd-49a4-bff5-1ad4e6b0ef7b"` |
| **title** | ✅ Yes | `attributes.title.en` | `"Chainsaw Man"` |
| **author** | ✅ Yes | `relationships[type="author"]` | `"Fujimoto Tatsuki"` |
| **status** | ✅ Yes | `attributes.status` | `"ongoing"`, `"completed"`, `"hiatus"` |
| **total_chapters** | ✅ Yes | `attributes.lastChapter` | `180` |
| **description** | ✅ Yes | `attributes.description.en` | `"Broke young man..."` |
| **cover_url** | ✅ Yes | `relationships[type="cover_art"]` | `"https://uploads.mangadex.org/covers/..."` |
| **genres** | ✅ Yes | `attributes.tags[group="genre"]` | `["Action", "Comedy"]` |
| **slug** | ⚙️ Generated | From title | `"chainsaw-man"` |

---

## 📡 MangaDex API Request Pattern

### ✅ CORRECT: Include Relationships

Always use the `includes[]` parameter to fetch related data in a single request:

```bash
GET https://api.mangadex.org/manga?limit=100&includes[]=author&includes[]=cover_art&includes[]=artist
```

**Why?** This prevents making separate API calls for each manga to get author/cover data.

### ❌ WRONG: Without Includes

```bash
GET https://api.mangadex.org/manga?limit=100
# ❌ This only returns manga IDs without author/cover info
# You'd need 100 additional API calls to get complete data!
```

---

## 🔄 Complete Data Extraction Flow

### Step 1: API Request with All Includes

```go
// Construct API URL with all necessary includes
params := url.Values{}
params.Add("limit", "100")
params.Add("offset", fmt.Sprintf("%d", offset))
params.Add("order[createdAt]", "desc")
params.Add("includes[]", "author")
params.Add("includes[]", "cover_art")
params.Add("includes[]", "artist")

apiURL := fmt.Sprintf("%s/manga?%s", MangaDexAPIBase, params.Encode())
```

### Step 2: Parse API Response

```go
type MangaDexResponse struct {
    Result   string        `json:"result"`
    Data     []MangaData   `json:"data"`
    Limit    int           `json:"limit"`
    Offset   int           `json:"offset"`
    Total    int           `json:"total"`
}

type MangaData struct {
    ID            string         `json:"id"`
    Type          string         `json:"type"`
    Attributes    MangaAttribute `json:"attributes"`
    Relationships []Relationship `json:"relationships"`
}

type MangaAttribute struct {
    Title         map[string]string   `json:"title"`
    Description   map[string]string   `json:"description"`
    Status        string              `json:"status"`
    LastChapter   string              `json:"lastChapter"`
    Tags          []Tag               `json:"tags"`
    CreatedAt     string              `json:"createdAt"`
    UpdatedAt     string              `json:"updatedAt"`
}

type Relationship struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`
    Attributes map[string]interface{} `json:"attributes"`
}

type Tag struct {
    ID         string        `json:"id"`
    Type       string        `json:"type"`
    Attributes TagAttributes `json:"attributes"`
}

type TagAttributes struct {
    Name  map[string]string `json:"name"`
    Group string            `json:"group"` // "genre", "theme", "format"
}
```

### Step 3: Extract All Fields

```go
func ExtractMangaMetadata(apiManga MangaData) (*MangaMetadata, error) {
    metadata := &MangaMetadata{
        MangaDexID: apiManga.ID,
    }
    
    // 1. Title (prefer English)
    if enTitle, ok := apiManga.Attributes.Title["en"]; ok {
        metadata.Title = enTitle
    } else {
        // Fallback to first available language
        for _, title := range apiManga.Attributes.Title {
            metadata.Title = title
            break
        }
    }
    
    // 2. Description (prefer English)
    if enDesc, ok := apiManga.Attributes.Description["en"]; ok {
        metadata.Description = enDesc
    } else {
        for _, desc := range apiManga.Attributes.Description {
            metadata.Description = desc
            break
        }
    }
    
    // 3. Status (direct mapping)
    metadata.Status = apiManga.Attributes.Status
    
    // 4. Total Chapters (parse from lastChapter string)
    if apiManga.Attributes.LastChapter != "" {
        if chapters, err := strconv.Atoi(apiManga.Attributes.LastChapter); err == nil {
            metadata.TotalChapters = chapters
        }
    }
    
    // 5. Extract Author from relationships
    for _, rel := range apiManga.Relationships {
        if rel.Type == "author" {
            if name, ok := rel.Attributes["name"].(string); ok {
                metadata.Author = name
                break
            }
        }
    }
    
    // 6. Extract Cover URL from relationships
    for _, rel := range apiManga.Relationships {
        if rel.Type == "cover_art" {
            if fileName, ok := rel.Attributes["fileName"].(string); ok {
                metadata.CoverURL = fmt.Sprintf(
                    "https://uploads.mangadex.org/covers/%s/%s",
                    apiManga.ID,
                    fileName,
                )
                break
            }
        }
    }
    
    // 7. Extract Genres (filter tags by group="genre")
    for _, tag := range apiManga.Attributes.Tags {
        if tag.Attributes.Group == "genre" {
            if enName, ok := tag.Attributes.Name["en"]; ok {
                metadata.Genres = append(metadata.Genres, enName)
            }
        }
    }
    
    // 8. Generate Slug
    metadata.Slug = GenerateSlug(metadata.Title)
    
    return metadata, nil
}

type MangaMetadata struct {
    MangaDexID    string
    Title         string
    Author        string
    Status        string
    TotalChapters int
    Description   string
    CoverURL      string
    Genres        []string
    Slug          string
}
```

### Step 4: Store in Database

```go
func StoreManga(ctx context.Context, db *gorm.DB, metadata *MangaMetadata) (int64, error) {
    // 1. Insert/Update manga
    manga := &models.Manga{
        MangaDexID:    &metadata.MangaDexID,
        Slug:          &metadata.Slug,
        Title:         metadata.Title,
        Author:        &metadata.Author,
        Status:        &metadata.Status,
        TotalChapters: &metadata.TotalChapters,
        Description:   &metadata.Description,
        CoverURL:      &metadata.CoverURL,
        LastSyncedAt:  &now,
    }
    
    // Upsert: insert or update if mangadex_id exists
    result := db.WithContext(ctx).
        Clauses(clause.OnConflict{
            Columns: []clause.Column{{Name: "mangadex_id"}},
            DoUpdates: clause.AssignmentColumns([]string{
                "title", "author", "status", "total_chapters",
                "description", "cover_url", "last_synced_at", "updated_at",
            }),
        }).
        Create(manga)
    
    if result.Error != nil {
        return 0, result.Error
    }
    
    // 2. Insert genres and create relationships
    for _, genreName := range metadata.Genres {
        // Upsert genre
        genre := &models.Genre{Name: genreName}
        db.WithContext(ctx).
            Clauses(clause.OnConflict{
                Columns:   []clause.Column{{Name: "name"}},
                DoNothing: true,
            }).
            Create(genre)
        
        // Get genre ID
        var genreID int64
        db.WithContext(ctx).
            Model(&models.Genre{}).
            Select("id").
            Where("name = ?", genreName).
            Scan(&genreID)
        
        // Link manga to genre
        db.WithContext(ctx).Exec(`
            INSERT INTO manga_genres (manga_id, genre_id)
            VALUES (?, ?)
            ON CONFLICT (manga_id, genre_id) DO NOTHING
        `, manga.ID, genreID)
    }
    
    return manga.ID, nil
}
```

---

## 🔍 Verification Checklist

After implementing the sync service, verify that all metadata is extracted:

### ✅ Data Completeness Check

```sql
-- Check for missing author names
SELECT COUNT(*) FROM manga WHERE author IS NULL OR author = '';
-- Should be 0 or very low

-- Check for missing descriptions
SELECT COUNT(*) FROM manga WHERE description IS NULL OR description = '';
-- Some manga may not have descriptions on MangaDex

-- Check for missing cover URLs
SELECT COUNT(*) FROM manga WHERE cover_url IS NULL OR cover_url = '';
-- Should be 0

-- Check for missing genres
SELECT m.id, m.title, COUNT(mg.genre_id) as genre_count
FROM manga m
LEFT JOIN manga_genres mg ON m.id = mg.manga_id
WHERE m.mangadex_id IS NOT NULL
GROUP BY m.id, m.title
HAVING COUNT(mg.genre_id) = 0;
-- Should return very few results

-- Check for missing total_chapters
SELECT COUNT(*) FROM manga WHERE total_chapters IS NULL OR total_chapters = 0;
-- Some ongoing manga may have 0 chapters published
```

### 📊 Sample Data Comparison

Compare your synced data with existing scraped data:

```sql
-- Sample manga record
SELECT 
    id,
    mangadex_id,
    slug,
    title,
    author,
    status,
    total_chapters,
    LEFT(description, 100) as description_preview,
    cover_url
FROM manga
WHERE mangadex_id IS NOT NULL
LIMIT 5;
```

Expected output should match the structure of your `scraped_data.json`:
```
id  | mangadex_id                           | slug         | title        | author            | status   | total_chapters
----|---------------------------------------|--------------|--------------|-------------------|----------|---------------
123 | a77742b1-befd-49a4-bff5-1ad4e6b0ef7b | chainsaw-man | Chainsaw Man | Fujimoto Tatsuki  | ongoing  | 180
```

---

## 🚨 Common Mistakes to Avoid

### ❌ Mistake 1: Not Using `includes[]`
```go
// BAD: Makes N+1 API calls
for _, manga := range mangas {
    authorResp := api.Get("/author/" + manga.AuthorID)  // Extra API call!
    coverResp := api.Get("/cover/" + manga.CoverID)     // Extra API call!
}

// GOOD: Includes relationships in single request
mangas := api.Get("/manga?includes[]=author&includes[]=cover_art")
```

### ❌ Mistake 2: Ignoring Language Fallbacks
```go
// BAD: Only gets English titles
title := apiManga.Attributes.Title["en"]  // Returns empty string if no English title!

// GOOD: Fallback to any available language
title := apiManga.Attributes.Title["en"]
if title == "" {
    for _, t := range apiManga.Attributes.Title {
        title = t
        break
    }
}
```

### ❌ Mistake 3: Not Filtering Genre Tags
```go
// BAD: Includes all tags (themes, formats, etc.)
for _, tag := range apiManga.Attributes.Tags {
    genres = append(genres, tag.Attributes.Name["en"])
}

// GOOD: Only includes actual genres
for _, tag := range apiManga.Attributes.Tags {
    if tag.Attributes.Group == "genre" {
        genres = append(genres, tag.Attributes.Name["en"])
    }
}
```

### ❌ Mistake 4: Incorrect Cover URL Construction
```go
// BAD: Cover URL is not directly in API response
coverURL := apiManga.Attributes.CoverURL  // This field doesn't exist!

// GOOD: Construct from manga ID + cover fileName
for _, rel := range apiManga.Relationships {
    if rel.Type == "cover_art" {
        fileName := rel.Attributes["fileName"].(string)
        coverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", 
            apiManga.ID, fileName)
    }
}
```

---

## 📝 Example: Complete Initial Sync Function

```go
func InitialSync(ctx context.Context, limit int) error {
    log.Printf("Starting initial sync of %d manga...", limit)
    
    offset := 0
    batchSize := 100
    totalSynced := 0
    
    for totalSynced < limit {
        // 1. Fetch manga batch with all includes
        params := url.Values{}
        params.Add("limit", fmt.Sprintf("%d", batchSize))
        params.Add("offset", fmt.Sprintf("%d", offset))
        params.Add("order[createdAt]", "desc")
        params.Add("includes[]", "author")
        params.Add("includes[]", "cover_art")
        
        apiURL := fmt.Sprintf("%s/manga?%s", MangaDexAPIBase, params.Encode())
        
        // Rate limit: wait before request
        if err := rateLimiter.Wait(ctx); err != nil {
            return err
        }
        
        resp, err := client.Get(apiURL)
        if err != nil {
            return fmt.Errorf("API request failed: %w", err)
        }
        
        var apiResp MangaDexResponse
        if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
            return err
        }
        resp.Body.Close()
        
        if len(apiResp.Data) == 0 {
            break  // No more manga
        }
        
        // 2. Process each manga
        for _, apiManga := range apiResp.Data {
            // Extract complete metadata
            metadata, err := ExtractMangaMetadata(apiManga)
            if err != nil {
                log.Printf("Failed to extract metadata for %s: %v", apiManga.ID, err)
                continue
            }
            
            // Store in database with genres
            mangaID, err := StoreManga(ctx, db, metadata)
            if err != nil {
                log.Printf("Failed to store manga %s: %v", metadata.Title, err)
                continue
            }
            
            log.Printf("✓ Synced: %s (ID: %d, Author: %s, Genres: %v, Chapters: %d)",
                metadata.Title, mangaID, metadata.Author, metadata.Genres, metadata.TotalChapters)
            
            totalSynced++
            if totalSynced >= limit {
                break
            }
        }
        
        offset += batchSize
    }
    
    log.Printf("Initial sync completed: %d manga imported", totalSynced)
    return nil
}
```

---

## 🎓 Summary

**Key Takeaways:**

1. ✅ **Always use `includes[]`** to get author and cover data in one API call
2. ✅ **Extract all 8 required fields**: mangadex_id, title, author, status, total_chapters, description, cover_url, genres
3. ✅ **Filter tags by `group="genre"`** to get only actual genres
4. ✅ **Prefer English language** but fallback to any available language
5. ✅ **Construct cover URLs** using the pattern: `https://uploads.mangadex.org/covers/{manga_id}/{fileName}`
6. ✅ **Parse `lastChapter`** as an integer for total_chapters
7. ✅ **Generate slugs** from titles for URL-friendly identifiers
8. ✅ **Store genres** in separate table with many-to-many relationship

This ensures your MangaDex sync produces **identical data structure** to your existing `scraped_data.json` format! 🎉
