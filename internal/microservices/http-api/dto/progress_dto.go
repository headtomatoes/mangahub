package dto

import "time"

// UpdateProgressRequest: payload to update reading progress
type UpdateProgressRequest struct {
    CurrentChapter int    `json:"current_chapter" binding:"min=0"`
    Page           int    `json:"page" binding:"min=0"`
    Status         string `json:"status" binding:"omitempty,oneof=reading completed plan_to_read dropped"`
}

// ProgressResponse: response for reading progress
type ProgressResponse struct {
    MangaID        int64     `json:"manga_id"`
    CurrentChapter int       `json:"current_chapter"`
    Page           int       `json:"page"`
    Status         string    `json:"status"`
    UpdatedAt      time.Time `json:"updated_at"`
}