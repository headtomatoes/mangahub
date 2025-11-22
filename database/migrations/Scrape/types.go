package main

// Shared data structures for MangaDex scraping and database import

// ScrapedData represents the complete scraped data structure
type ScrapedData struct {
	Mangas []Manga `json:"mangas"`
	Genres []Genre `json:"genres"`
}

// Manga represents manga information
type Manga struct {
	ID            string   `json:"id"`
	Slug          string   `json:"slug"`
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	Status        string   `json:"status"`
	TotalChapters int      `json:"total_chapters"`
	Description   string   `json:"description"`
	CoverURL      string   `json:"cover_url"`
	Genres        []string `json:"genres"`
}

// Genre represents genre/tag information
type Genre struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
