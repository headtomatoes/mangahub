package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// MangaDex API Base URL
const (
	MangaDexAPIBase = "https://api.mangadex.org"
	MangaEndpoint   = "/manga"
	GenreEndpoint   = "/manga/tag"
)

// MangaDex API response structures
type MangaDexResponse struct {
	Result   string      `json:"result"`
	Response string      `json:"response"`
	Data     []MangaData `json:"data"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Total    int         `json:"total"`
}

type MangaData struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    MangaAttribute `json:"attributes"`
	Relationships []Relationship `json:"relationships"`
}

type Relationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes map[string]interface{} `json:"attributes"`
}

type MangaAttribute struct {
	Title                  map[string]string   `json:"title"`
	AltTitles              []map[string]string `json:"altTitles"`
	Description            map[string]string   `json:"description"`
	Status                 string              `json:"status"`
	Year                   int                 `json:"year"`
	ContentRating          string              `json:"contentRating"`
	Tags                   []Tag               `json:"tags"`
	OriginalLanguage       string              `json:"originalLanguage"`
	LastVolume             string              `json:"lastVolume"`
	LastChapter            string              `json:"lastChapter"`
	PublicationDemographic string              `json:"publicationDemographic"`
	CreatedAt              string              `json:"createdAt"`
	UpdatedAt              string              `json:"updatedAt"`
}

type Tag struct {
	ID         string        `json:"id"`
	Type       string        `json:"type"`
	Attributes TagAttributes `json:"attributes"`
}

type TagAttributes struct {
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	Group       string            `json:"group"`
	Version     int               `json:"version"`
}

type GenreResponse struct {
	Result   string `json:"result"`
	Response string `json:"response"`
	Data     []Tag  `json:"data"`
}

func main() {
	log.Println("Starting MangaDex scraper...")

	// Get API key from environment or use default
	apiKey := os.Getenv("MANGADX_API_KEY")

	// Step 1: Scrape genres/tags
	log.Println("Fetching genres from MangaDex...")
	genres, err := scrapeGenres(apiKey)
	if err != nil {
		log.Fatalf("Failed to scrape genres: %v", err)
	}
	log.Printf("Successfully scraped %d genres", len(genres))

	// Step 2: Scrape manga
	log.Println("Fetching manga from MangaDex...")
	mangas, err := scrapeManga(apiKey, 500) // Scrape 500 manga entries
	if err != nil {
		log.Fatalf("Failed to scrape manga: %v", err)
	}
	log.Printf("Successfully scraped %d manga entries", len(mangas))

	// Step 3: Save to JSON files
	scrapedData := ScrapedData{
		Mangas: mangas,
		Genres: genres,
	}

	// Save to parent directory so import script can find it
	outputPath := "../scraped_data.json"
	if err := saveToJSON(scrapedData, outputPath); err != nil {
		log.Fatalf("Failed to save data: %v", err)
	}

	log.Println("Scraping completed successfully!")
	log.Printf("Data saved to %s", outputPath)
	log.Printf("Summary: %d manga entries, %d genres", len(mangas), len(genres))
}

func scrapeGenres(apiKey string) ([]Genre, error) {
	endpoint := MangaDexAPIBase + GenreEndpoint

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch genres: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var genreResp GenreResponse
	if err := json.NewDecoder(resp.Body).Decode(&genreResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	genres := make([]Genre, 0)
	for _, tag := range genreResp.Data {
		// Only include genre tags
		if tag.Attributes.Group == "genre" {
			name := tag.Attributes.Name["en"]
			if name == "" {
				// Fallback to first available language
				for _, n := range tag.Attributes.Name {
					name = n
					break
				}
			}
			if name != "" {
				genres = append(genres, Genre{
					ID:   tag.ID,
					Name: name,
				})
				log.Printf("  - Found genre: %s", name)
			}
		}
	}

	return genres, nil
}

func scrapeManga(apiKey string, limit int) ([]Manga, error) {
	mangas := make([]Manga, 0)
	perPage := 20 // MangaDex limit per request

	// Create rate limiter: 5 requests per second (MangaDex API limit)
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 5)
	ctx := context.Background()

	// Use channels for concurrent scraping
	type batchResult struct {
		mangas []Manga
		offset int
		err    error
	}

	resultChan := make(chan batchResult, 10)
	var wg sync.WaitGroup

	// Calculate number of batches needed
	numBatches := (limit + perPage - 1) / perPage

	// Launch goroutines for parallel fetching
	for i := 0; i < numBatches; i++ {
		currentOffset := i * perPage
		if currentOffset >= limit {
			break
		}

		wg.Add(1)
		go func(offset int) {
			defer wg.Done()

			// Wait for rate limiter
			if err := limiter.Wait(ctx); err != nil {
				resultChan <- batchResult{err: err}
				return
			}

			log.Printf("Fetching manga batch (offset: %d/%d)...", offset, limit)

			batch, err := fetchMangaBatch(apiKey, perPage, offset)
			if err != nil {
				resultChan <- batchResult{err: err, offset: offset}
				return
			}

			resultChan <- batchResult{mangas: batch, offset: offset}
			log.Printf("  - Fetched %d manga from offset %d", len(batch), offset)
		}(currentOffset)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results (they may arrive out of order)
	batchMap := make(map[int][]Manga)
	for result := range resultChan {
		if result.err != nil {
			log.Printf("⚠ Warning: Failed to fetch batch at offset %d: %v", result.offset, result.err)
			continue
		}
		batchMap[result.offset] = result.mangas
	}

	// Sort and combine results in order
	for i := 0; i < numBatches; i++ {
		currentOffset := i * perPage
		if batch, ok := batchMap[currentOffset]; ok {
			mangas = append(mangas, batch...)
		}
	}

	return mangas, nil
}

func fetchMangaBatch(apiKey string, limit, offset int) ([]Manga, error) {
	endpoint := MangaDexAPIBase + MangaEndpoint

	// Build query parameters
	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	params.Add("includes[]", "cover_art")
	params.Add("includes[]", "author")
	params.Add("contentRating[]", "safe")
	params.Add("contentRating[]", "suggestive")
	params.Add("order[rating]", "desc")
	params.Add("hasAvailableChapters", "true")

	fullURL := endpoint + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manga: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var mangaResp MangaDexResponse
	if err := json.NewDecoder(resp.Body).Decode(&mangaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	mangas := make([]Manga, 0)
	for _, data := range mangaResp.Data {
		manga := convertToManga(data)
		mangas = append(mangas, manga)
	}

	return mangas, nil
}

func convertToManga(data MangaData) Manga {
	attr := data.Attributes

	// Get English title or first available
	title := attr.Title["en"]
	if title == "" {
		for _, t := range attr.Title {
			title = t
			break
		}
	}

	// Get English description or first available
	description := attr.Description["en"]
	if description == "" {
		for _, d := range attr.Description {
			description = d
			break
		}
	}

	// Extract author from relationships
	author := "Unknown"
	for _, rel := range data.Relationships {
		if rel.Type == "author" {
			if nameAttr, ok := rel.Attributes["name"].(string); ok {
				author = nameAttr
				break
			}
		}
	}

	// Extract genre names
	genres := make([]string, 0)
	for _, tag := range attr.Tags {
		if tag.Attributes.Group == "genre" {
			genreName := tag.Attributes.Name["en"]
			if genreName == "" {
				for _, n := range tag.Attributes.Name {
					genreName = n
					break
				}
			}
			if genreName != "" {
				genres = append(genres, genreName)
			}
		}
	}

	// Map status
	status := "ongoing"
	switch attr.Status {
	case "completed":
		status = "completed"
	case "hiatus":
		status = "hiatus"
	case "ongoing":
		status = "ongoing"
	default:
		status = "ongoing"
	}

	// Parse total chapters
	totalChapters := 0
	if attr.LastChapter != "" {
		fmt.Sscanf(attr.LastChapter, "%d", &totalChapters)
	}

	// Generate slug from title
	slug := generateSlug(title)

	// Cover URL
	coverURL := fmt.Sprintf("https://mangadex.org/covers/%s.jpg", data.ID)

	return Manga{
		ID:            data.ID,
		Slug:          slug,
		Title:         title,
		Author:        author, // Now parsed from relationships
		Status:        status,
		TotalChapters: totalChapters,
		Description:   description,
		CoverURL:      coverURL,
		Genres:        genres,
	}
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

func saveToJSON(data ScrapedData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	log.Printf("Data saved to %s", filename)
	return nil
}
