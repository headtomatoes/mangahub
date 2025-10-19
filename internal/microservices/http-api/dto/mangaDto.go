package dto

import (
    "mangahub/internal/microservices/http-api/models"
    "time"
)

// CreateMangaDTO used for POST /api/manga
type CreateMangaDTO struct {
    Slug          *string `json:"slug,omitempty"` // optional client slug
    Title         string  `json:"title" binding:"required"`
    Author        *string `json:"author,omitempty"`
    Status        *string `json:"status,omitempty"`
    TotalChapters *int    `json:"total_chapters,omitempty"`
    Description   *string `json:"description,omitempty"`
    CoverURL      *string `json:"cover_url,omitempty"`
}

// UpdateMangaDTO used for PUT /api/manga/:id (partial updates allowed)
type UpdateMangaDTO struct {
    Title         *string `json:"title,omitempty"`
    Author        *string `json:"author,omitempty"`
    Status        *string `json:"status,omitempty"`
    TotalChapters *int    `json:"total_chapters,omitempty"`
    Description   *string `json:"description,omitempty"`
    CoverURL      *string `json:"cover_url,omitempty"`
    Slug          *string `json:"slug,omitempty"`
}

// MangaResponse DTO for responses
type MangaResponse struct {
    ID            int64      `json:"id"`
    Slug          *string    `json:"slug,omitempty"`
    Title         string     `json:"title"`
    Author        *string    `json:"author,omitempty"`
    Status        *string    `json:"status,omitempty"`
    TotalChapters *int       `json:"total_chapters,omitempty"`
    Description   *string    `json:"description,omitempty"`
    CoverURL      *string    `json:"cover_url,omitempty"`
    CreatedAt     *time.Time `json:"created_at,omitempty"`
}

// Converters
func (d CreateMangaDTO) ToModel() models.Manga {
    return models.Manga{
        Slug:          d.Slug,
        Title:         d.Title,
        Author:        d.Author,
        Status:        d.Status,
        TotalChapters: d.TotalChapters,
        Description:   d.Description,
        CoverURL:      d.CoverURL,
    }
}

func (d UpdateMangaDTO) ApplyTo(m *models.Manga) {
    if d.Title != nil {
        m.Title = *d.Title
    }
    if d.Author != nil {
        m.Author = d.Author
    }
    if d.Status != nil {
        m.Status = d.Status
    }
    if d.TotalChapters != nil {
        m.TotalChapters = d.TotalChapters
    }
    if d.Description != nil {
        m.Description = d.Description
    }
    if d.CoverURL != nil {
        m.CoverURL = d.CoverURL
    }
    if d.Slug != nil {
        m.Slug = d.Slug
    }
}

func FromModelToResponse(m models.Manga) MangaResponse {
    return MangaResponse{
        ID:            m.ID,
        Slug:          m.Slug,
        Title:         m.Title,
        Author:        m.Author,
        Status:        m.Status,
        TotalChapters: m.TotalChapters,
        Description:   m.Description,
        CoverURL:      m.CoverURL,
        CreatedAt:     m.CreatedAt,
    }
}