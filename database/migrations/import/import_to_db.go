package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// Structures matching the JSON file
// ScrapedData is defined in mangadex_scraper.go to avoid duplicate type declarations.

func main() {
	log.Println("Starting database import...")

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://mangahub:mcfc1213@localhost:5432/mangahub?sslmode=disable"
		log.Println("Using default database URL (localhost)")
	} else {
		// Replace 'db' hostname with 'localhost' when running outside Docker
		dbURL = strings.ReplaceAll(dbURL, "@db:", "@localhost:")
		log.Println("Using database URL from environment (adjusted for localhost)")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v\nMake sure PostgreSQL is running: docker compose up -d db", err)
	}
	log.Println("✓ Successfully connected to database")

	// Read JSON file - look in parent directory
	jsonFile := "../scraped_data.json"
	if len(os.Args) > 1 {
		jsonFile = os.Args[1]
	}

	log.Printf("Reading data from %s...", jsonFile)
	data, err := readJSONFile(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	log.Printf("✓ Loaded %d mangas and %d genres from JSON", len(data.Mangas), len(data.Genres))

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Import genres
	log.Println("\n=== Importing Genres ===")
	genreIDMap, err := importGenres(tx, data.Genres)
	if err != nil {
		log.Fatalf("Failed to import genres: %v", err)
	}
	log.Printf("✓ Successfully imported %d genres", len(genreIDMap))

	// Import manga
	log.Println("\n=== Importing Manga ===")
	mangaCount, relationCount, err := importMangas(tx, data.Mangas, genreIDMap)
	if err != nil {
		log.Fatalf("Failed to import manga: %v", err)
	}
	log.Printf("✓ Successfully imported %d manga entries", mangaCount)
	log.Printf("✓ Created %d manga-genre relationships", relationCount)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("\n=== Import Summary ===")
	log.Printf("✓ Genres: %d", len(genreIDMap))
	log.Printf("✓ Manga: %d", mangaCount)
	log.Printf("✓ Relationships: %d", relationCount)
	log.Println("✓ Database import completed successfully!")
}

func readJSONFile(filename string) (*ScrapedData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var data ScrapedData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &data, nil
}

func importGenres(tx *sql.Tx, genres []Genre) (map[string]int, error) {
	genreIDMap := make(map[string]int)

	stmt, err := tx.Prepare(`
		INSERT INTO genres (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, genre := range genres {
		var genreID int
		err := stmt.QueryRow(genre.Name).Scan(&genreID)
		if err != nil {
			log.Printf("⚠ Warning: Failed to insert genre %s: %v", genre.Name, err)
			continue
		}
		genreIDMap[genre.Name] = genreID
		log.Printf("  [%d/%d] ✓ %s (ID: %d)", i+1, len(genres), genre.Name, genreID)
	}

	return genreIDMap, nil
}

func importMangas(tx *sql.Tx, mangas []Manga, genreIDMap map[string]int) (int, int, error) {
	// Prepare manga insert statement
	mangaStmt, err := tx.Prepare(`
		INSERT INTO manga (slug, title, author, status, total_chapters, description, cover_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			author = EXCLUDED.author,
			status = EXCLUDED.status,
			total_chapters = EXCLUDED.total_chapters,
			description = EXCLUDED.description,
			cover_url = EXCLUDED.cover_url
		RETURNING id
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to prepare manga statement: %w", err)
	}
	defer mangaStmt.Close()

	// Prepare manga_genres insert statement
	genreStmt, err := tx.Prepare(`
		INSERT INTO manga_genres (manga_id, genre_id)
		VALUES ($1, $2)
		ON CONFLICT (manga_id, genre_id) DO NOTHING
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to prepare manga_genres statement: %w", err)
	}
	defer genreStmt.Close()

	mangaCount := 0
	relationCount := 0

	for i, manga := range mangas {
		// Generate unique slug if needed
		slug := manga.Slug
		if slug == "" {
			slug = generateSlug(manga.Title)
		}

		// Insert manga
		var mangaID int64
		err := mangaStmt.QueryRow(
			slug,
			manga.Title,
			manga.Author,
			manga.Status,
			manga.TotalChapters,
			manga.Description,
			manga.CoverURL,
		).Scan(&mangaID)

		if err != nil {
			log.Printf("⚠ [%d/%d] Warning: Failed to insert manga %s: %v", i+1, len(mangas), manga.Title, err)
			continue
		}

		// Insert manga_genres relationships
		genreCount := 0
		for _, genreName := range manga.Genres {
			genreID, exists := genreIDMap[genreName]
			if !exists {
				// Try to find genre ID from database
				var dbGenreID int
				err := tx.QueryRow("SELECT id FROM genres WHERE name = $1", genreName).Scan(&dbGenreID)
				if err != nil {
					continue
				}
				genreID = dbGenreID
				genreIDMap[genreName] = genreID
			}

			_, err := genreStmt.Exec(mangaID, genreID)
			if err == nil {
				genreCount++
				relationCount++
			}
		}

		mangaCount++
		if mangaCount%10 == 0 || i == len(mangas)-1 {
			log.Printf("  [%d/%d] ✓ %s (%d genres)", i+1, len(mangas), manga.Title, genreCount)
		}
	}

	return mangaCount, relationCount, nil
}

func generateSlug(title string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	// Remove multiple consecutive hyphens
	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}
