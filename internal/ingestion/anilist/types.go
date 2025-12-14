package anilist

import (
    "fmt"
    "html"
    "regexp"
    "strings"
    "time"
)

// ============================================
// API RESPONSE STRUCTURES
// ============================================

// MangaPageResponse represents a paginated response
type MangaPageResponse struct {
    Page PageData `json:"Page"`
}

// PageData contains pagination info and media list
type PageData struct {
    PageInfo PageInfo    `json:"pageInfo"`
    Media    []MediaData `json:"media"`
}

// PageInfo contains pagination metadata
type PageInfo struct {
    Total       int  `json:"total"`
    CurrentPage int  `json:"currentPage"`
    LastPage    int  `json:"lastPage"`
    HasNextPage bool `json:"hasNextPage"`
    PerPage     int  `json:"perPage"`
}

// MediaResponse wraps a single media item
type MediaResponse struct {
    Media MediaData `json:"Media"`
}

// MediaData represents a manga entry from AniList
type MediaData struct {
    ID           int          `json:"id"`
    IDMal        *int         `json:"idMal"`
    Title        TitleData    `json:"title"`
    Description  *string      `json:"description"`
    Status       string       `json:"status"` // FINISHED, RELEASING, NOT_YET_RELEASED, CANCELLED
    Chapters     *int         `json:"chapters"`
    Volumes      *int         `json:"volumes"`
    CoverImage   CoverImage   `json:"coverImage"`
    BannerImage  *string      `json:"bannerImage"`
    Genres       []string     `json:"genres"`
    Tags         []Tag        `json:"tags"`
    AverageScore *int         `json:"averageScore"` // 0-100
    Staff        StaffData    `json:"staff"`
    StartDate    FuzzyDate    `json:"startDate"`
    EndDate      *FuzzyDate   `json:"endDate"`
    UpdatedAt    int64        `json:"updatedAt"` // Unix timestamp
}

// TitleData contains title variants
type TitleData struct {
    English *string `json:"english"`
    Romaji  *string `json:"romaji"`
    Native  *string `json:"native"`
}

// CoverImage contains cover URLs
type CoverImage struct {
    Large  *string `json:"large"`
    Medium *string `json:"medium"`
}

// Tag represents a genre/theme tag
type Tag struct {
    Name string `json:"name"`
    Rank *int   `json:"rank"` // Relevance score
}

// StaffData contains staff information
type StaffData struct {
    Edges []StaffEdge `json:"edges"`
}

// StaffEdge represents a staff member
type StaffEdge struct {
    Role string    `json:"role"`
    Node StaffNode `json:"node"`
}

// StaffNode contains staff name
type StaffNode struct {
    Name StaffName `json:"name"`
}

// StaffName contains full name
type StaffName struct {
    Full string `json:"full"`
}

// FuzzyDate represents a date with optional components
type FuzzyDate struct {
    Year  *int `json:"year"`
    Month *int `json:"month"`
    Day   *int `json:"day"`
}

// ============================================
// EXTRACTED METADATA STRUCTURES
// ============================================

// ExtractedManga represents manga metadata ready for database
type ExtractedManga struct {
    AniListID     int
    Title         string
    Slug          string
    Author        string
    Status        string
    TotalChapters int
    Description   string
    CoverURL      string
    AverageRating float64
    Genres        []string
    UpdatedAt     time.Time
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// ExtractMangaMetadata extracts complete metadata from AniList API response
func ExtractMangaMetadata(apiManga MediaData) (*ExtractedManga, error) {
    extracted := &ExtractedManga{
        AniListID: apiManga.ID,
    }

    // 1. Title (prefer English, fallback to Romaji, then Native)
    if apiManga.Title.English != nil && *apiManga.Title.English != "" {
        extracted.Title = *apiManga.Title.English
    } else if apiManga.Title.Romaji != nil && *apiManga.Title.Romaji != "" {
        extracted.Title = *apiManga.Title.Romaji
    } else if apiManga.Title.Native != nil {
        extracted.Title = *apiManga.Title.Native
    }

    if extracted.Title == "" {
        return nil, fmt.Errorf("manga %d has no title", apiManga.ID)
    }

    // 2. Generate slug
    extracted.Slug = GenerateSlug(extracted.Title)

    // 3. Description (clean HTML tags)
    if apiManga.Description != nil {
        extracted.Description = CleanDescription(*apiManga.Description)
    }

    // 4. Status (map AniList status to our format)
    extracted.Status = mapAniListStatus(apiManga.Status)

    // 5. Total chapters
    if apiManga.Chapters != nil {
        extracted.TotalChapters = *apiManga.Chapters
    }

    // 6. Extract author from staff (role="Story")
    for _, edge := range apiManga.Staff.Edges {
        if edge.Role == "Story" || edge.Role == "Story & Art" {
            extracted.Author = edge.Node.Name.Full
            break
        }
    }

    // 7. Cover URL
    if apiManga.CoverImage.Large != nil {
        extracted.CoverURL = *apiManga.CoverImage.Large
    } else if apiManga.CoverImage.Medium != nil {
        extracted.CoverURL = *apiManga.CoverImage.Medium
    }

    // 8. Average rating (convert from 0-100 to 0-10)
    if apiManga.AverageScore != nil {
        extracted.AverageRating = float64(*apiManga.AverageScore) / 10.0
    }

    // 9. Genres
    extracted.Genres = apiManga.Genres

    // 10. Updated timestamp
    if apiManga.UpdatedAt > 0 {
        extracted.UpdatedAt = time.Unix(apiManga.UpdatedAt, 0)
    }

    return extracted, nil
}

// mapAniListStatus converts AniList status to our format
func mapAniListStatus(status string) string {
    switch status {
    case "FINISHED":
        return "completed"
    case "RELEASING":
        return "ongoing"
    case "NOT_YET_RELEASED":
        return "upcoming"
    case "CANCELLED":
        return "cancelled"
    case "HIATUS":
        return "hiatus"
    default:
        return "ongoing"
    }
}

// CleanDescription removes HTML tags and decodes entities
func CleanDescription(desc string) string {
    // Remove HTML tags
    re := regexp.MustCompile(`<[^>]*>`)
    cleaned := re.ReplaceAllString(desc, "")

    // Decode HTML entities
    cleaned = html.UnescapeString(cleaned)

    // Trim whitespace
    cleaned = strings.TrimSpace(cleaned)

    return cleaned
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

// ToTime converts FuzzyDate to time.Time
func (fd *FuzzyDate) ToTime() *time.Time {
    if fd == nil || fd.Year == nil {
        return nil
    }

    year := *fd.Year
    month := 1
    day := 1

    if fd.Month != nil {
        month = *fd.Month
    }
    if fd.Day != nil {
        day = *fd.Day
    }

    t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
    return &t
}