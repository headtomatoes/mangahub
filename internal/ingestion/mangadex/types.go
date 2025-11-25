package mangadex

import (
	"fmt"
	"strings"
	"time"
)

// ============================================
// API RESPONSE STRUCTURES
// ============================================

// MangaListResponse represents the response from GET /manga
type MangaListResponse struct {
	Result   string      `json:"result"`
	Response string      `json:"response"`
	Data     []MangaData `json:"data"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Total    int         `json:"total"`
}

// MangaData represents a single manga entry from the API
type MangaData struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Attributes    MangaAttributes `json:"attributes"`
	Relationships []Relationship  `json:"relationships"`
}

// MangaAttributes contains manga metadata
type MangaAttributes struct {
	Title         map[string]string   `json:"title"`
	AltTitles     []map[string]string `json:"altTitles"`
	Description   map[string]string   `json:"description"`
	Status        string              `json:"status"` // "ongoing", "completed", "hiatus", "cancelled"
	Year          int                 `json:"year"`
	ContentRating string              `json:"contentRating"`
	LastVolume    string              `json:"lastVolume"`
	LastChapter   string              `json:"lastChapter"`
	Tags          []Tag               `json:"tags"`
	CreatedAt     string              `json:"createdAt"`
	UpdatedAt     string              `json:"updatedAt"`
}

// Tag represents a genre or theme tag
type Tag struct {
	ID         string        `json:"id"`
	Type       string        `json:"type"`
	Attributes TagAttributes `json:"attributes"`
}

// TagAttributes contains tag metadata
type TagAttributes struct {
	Name  map[string]string `json:"name"`
	Group string            `json:"group"` // "genre", "theme", "format"
}

// Relationship represents related entities (author, artist, cover_art)
type Relationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "author", "artist", "cover_art"
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// ChapterListResponse represents the response from GET /manga/{id}/feed
type ChapterListResponse struct {
	Result   string        `json:"result"`
	Response string        `json:"response"`
	Data     []ChapterData `json:"data"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
	Total    int           `json:"total"`
}

// ChapterData represents a single chapter entry
type ChapterData struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Attributes ChapterAttributes `json:"attributes"`
}

// ChapterAttributes contains chapter metadata
type ChapterAttributes struct {
	Volume             string `json:"volume"`
	Chapter            string `json:"chapter"`
	Title              string `json:"title"`
	TranslatedLanguage string `json:"translatedLanguage"`
	ExternalURL        string `json:"externalUrl"`
	PublishAt          string `json:"publishAt"`
	ReadableAt         string `json:"readableAt"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	Pages              int    `json:"pages"`
	Version            int    `json:"version"`
}

// ============================================
// EXTRACTED METADATA STRUCTURES
// ============================================

// ExtractedManga represents manga metadata extracted and ready for database
type ExtractedManga struct {
	MangaDexID    string
	Title         string
	Slug          string
	Author        string
	Status        string
	TotalChapters int
	Description   string
	CoverURL      string
	Genres        []string
	CreatedAt     time.Time
}

// ExtractedChapter represents chapter metadata ready for database
type ExtractedChapter struct {
	MangaDexChapterID string
	ChapterNumber     float64
	Title             string
	Volume            string
	Pages             int
	PublishedAt       time.Time
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// ExtractMangaMetadata extracts complete metadata from MangaDex API response
func ExtractMangaMetadata(apiManga MangaData) (*ExtractedManga, error) {
	extracted := &ExtractedManga{
		MangaDexID: apiManga.ID,
	}

	// 1. Title (prefer English, fallback to first available)
	if enTitle, ok := apiManga.Attributes.Title["en"]; ok {
		extracted.Title = enTitle
	} else {
		// Get first available title
		for _, t := range apiManga.Attributes.Title {
			extracted.Title = t
			break
		}
	}

	if extracted.Title == "" {
		return nil, fmt.Errorf("manga %s has no title", apiManga.ID)
	}

	// 2. Generate slug
	extracted.Slug = GenerateSlug(extracted.Title)

	// 3. Description (prefer English)
	if enDesc, ok := apiManga.Attributes.Description["en"]; ok {
		extracted.Description = enDesc
	} else {
		for _, d := range apiManga.Attributes.Description {
			extracted.Description = d
			break
		}
	}

	// 4. Status
	extracted.Status = apiManga.Attributes.Status
	if extracted.Status == "" {
		extracted.Status = "ongoing"
	}

	// 5. Total chapters (parse lastChapter field)
	if apiManga.Attributes.LastChapter != "" {
		// Try to parse as integer
		var chapters int
		fmt.Sscanf(apiManga.Attributes.LastChapter, "%d", &chapters)
		extracted.TotalChapters = chapters
	}

	// 6. Extract author from relationships
	for _, rel := range apiManga.Relationships {
		if rel.Type == "author" {
			if name, ok := rel.Attributes["name"].(string); ok {
				extracted.Author = name
				break
			}
		}
	}

	// 7. Extract cover URL
	for _, rel := range apiManga.Relationships {
		if rel.Type == "cover_art" {
			if fileName, ok := rel.Attributes["fileName"].(string); ok {
				// Construct cover URL: https://uploads.mangadex.org/covers/{manga-id}/{fileName}
				extracted.CoverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", apiManga.ID, fileName)
				break
			}
		}
	}

	// 8. Extract genres (filter tags where group="genre")
	for _, tag := range apiManga.Attributes.Tags {
		if tag.Attributes.Group == "genre" {
			if enName, ok := tag.Attributes.Name["en"]; ok {
				extracted.Genres = append(extracted.Genres, enName)
			}
		}
	}

	// 9. Parse created_at timestamp
	if apiManga.Attributes.CreatedAt != "" {
		createdAt, err := time.Parse(time.RFC3339, apiManga.Attributes.CreatedAt)
		if err == nil {
			extracted.CreatedAt = createdAt
		}
	}

	return extracted, nil
}

// ExtractChapterMetadata extracts chapter metadata from API response
func ExtractChapterMetadata(apiChapter ChapterData) (*ExtractedChapter, error) {
	extracted := &ExtractedChapter{
		MangaDexChapterID: apiChapter.ID,
		Title:             apiChapter.Attributes.Title,
		Volume:            apiChapter.Attributes.Volume,
		Pages:             apiChapter.Attributes.Pages,
	}

	// Parse chapter number
	if apiChapter.Attributes.Chapter != "" {
		var chapterNum float64
		fmt.Sscanf(apiChapter.Attributes.Chapter, "%f", &chapterNum)
		extracted.ChapterNumber = chapterNum
	}

	// Parse published_at timestamp
	if apiChapter.Attributes.PublishAt != "" {
		publishedAt, err := time.Parse(time.RFC3339, apiChapter.Attributes.PublishAt)
		if err == nil {
			extracted.PublishedAt = publishedAt
		}
	}

	return extracted, nil
}

// GenerateSlug creates a URL-friendly slug from title
func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters, keep only alphanumeric and hyphens
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

	return strings.Trim(slug, "-")
}

// GetPreferredTitle extracts title preferring English
func GetPreferredTitle(titleMap map[string]string) string {
	if enTitle, ok := titleMap["en"]; ok {
		return enTitle
	}
	// Return first available
	for _, title := range titleMap {
		return title
	}
	return ""
}

// GetPreferredDescription extracts description preferring English
func GetPreferredDescription(descMap map[string]string) string {
	if enDesc, ok := descMap["en"]; ok {
		return enDesc
	}
	// Return first available
	for _, desc := range descMap {
		return desc
	}
	return ""
}
