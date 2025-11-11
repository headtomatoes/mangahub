# MangaDex Scraping Scripts

This directory contains scripts for scraping manga data from MangaDex.org and importing it into the database.

## Overview

The scraping system consists of three main components:

1. **mangadex_scraper.go** - Scrapes manga and genre data from MangaDex API
2. **import_to_db.go** - Imports the scraped JSON data into PostgreSQL database
3. **scrape_and_import.sh** - Orchestrates the entire process

## Prerequisites

- Go 1.21 or higher
- PostgreSQL database running
- MangaDex API key (optional but recommended)
- Required Go packages:
  ```bash
  go get github.com/lib/pq
  ```

## Setup

1. **Configure Environment Variables**

   Make sure your `.env` file in the project root contains:
   ```env
   DATABASE_URL=postgres://mangahub:mcfc1213@localhost:5432/mangahub?sslmode=disable
   MANGADX_API_KEY=your-api-key-here
   ```

2. **Database Migration**

   Ensure your database tables are created:
   ```bash
   # From project root
   make migrate-up
   # or run migrations manually
   ```

## Usage

### Option 1: Run Complete Process (Recommended)

Run the shell script to scrape and import in one go:

```bash
cd scripts
chmod +x scrape_and_import.sh
./scrape_and_import.sh
```

### Option 2: Run Steps Individually

**Step 1: Scrape Data from MangaDex**

```bash
cd scripts
go run mangadex_scraper.go
```

This will:
- Fetch genres/tags from MangaDex API
- Fetch up to 100 manga entries (configurable in code)
- Save data to `scraped_data.json`

**Step 2: Import Data to Database**

```bash
cd scripts
go run import_to_db.go scraped_data.json
```

This will:
- Read data from `scraped_data.json`
- Insert genres into `genres` table
- Insert manga into `manga` table
- Create relationships in `manga_genres` table

## Output

### JSON File Structure

The `scraped_data.json` file has the following structure:

```json
{
  "mangas": [
    {
      "id": "mangadex-id",
      "slug": "manga-slug",
      "title": "Manga Title",
      "author": "Author Name",
      "status": "ongoing",
      "total_chapters": 100,
      "description": "Description text...",
      "cover_url": "https://...",
      "genres": ["Action", "Adventure", "Fantasy"]
    }
  ],
  "genres": [
    {
      "id": "genre-id",
      "name": "Action"
    }
  ]
}
```

### Database Tables

The scripts populate three tables:

1. **genres** - Genre names and IDs
2. **manga** - Manga information (title, author, status, etc.)
3. **manga_genres** - Many-to-many relationship between manga and genres

## Configuration

### Adjusting Scrape Limits

In `mangadex_scraper.go`, modify the limit in main():

```go
mangas, err := scrapeManga(apiKey, 100) // Change 100 to desired number
```

### Rate Limiting

The scraper includes a 1-second delay between API requests to respect MangaDex rate limits:

```go
time.Sleep(1 * time.Second)
```

### API Parameters

The scraper fetches manga with these filters (in `fetchMangaBatch`):
- Content ratings: safe, suggestive
- Only manga with available chapters
- Ordered by rating (descending)

Modify the `params` in `fetchMangaBatch()` to adjust filters.

## Error Handling

- Both scripts include comprehensive error handling and logging
- Database operations use transactions (automatic rollback on error)
- Conflicts are handled with UPSERT logic (ON CONFLICT clauses)
- Warnings are logged for non-critical errors

## Troubleshooting

### Database Connection Issues

If you get connection errors:
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Test connection manually
psql "postgres://mangahub:mcfc1213@localhost:5432/mangahub"
```

### API Rate Limiting

If you hit rate limits:
- Increase the delay in `scrapeManga()` function
- Reduce the number of manga being scraped
- Ensure you're using an API key

### Duplicate Data

The import script uses `ON CONFLICT` clauses to handle duplicates:
- Existing genres are updated
- Existing manga (by slug) are updated
- Duplicate manga_genres relationships are ignored

## Development

### Adding New Fields

To add new fields from MangaDex API:

1. Update the `MangaAttribute` struct in `mangadex_scraper.go`
2. Update the `Manga` struct in both scripts
3. Update the `convertToManga()` function
4. Update the database schema if needed
5. Update the INSERT statement in `import_to_db.go`

### Testing

Test individual components:

```bash
# Test scraper only
go run mangadex_scraper.go

# Test import with existing JSON
go run import_to_db.go test_data.json
```

## API Reference

### MangaDex API Endpoints Used

- **Tags/Genres**: `GET https://api.mangadex.org/manga/tag`
- **Manga List**: `GET https://api.mangadex.org/manga`

### Documentation

- [MangaDex API Documentation](https://api.mangadex.org/docs/)
- [MangaDex API GitHub](https://github.com/mangadex-pub/api-docs)

## Notes

- The scraper respects MangaDex's rate limits
- Cover URLs are placeholders and may need additional API calls
- Author information requires parsing relationships (currently set to "Unknown")
- The script focuses on safe/suggestive content ratings only
- Data is scraped in batches of 20 (MangaDex API limit)

## License

This is part of the MangaHub project for educational purposes.
