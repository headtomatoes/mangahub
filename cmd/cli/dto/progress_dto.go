package dto

// DTOs for progress-related operations in HTTP API

type GetProgressByMangaIDRequest struct {
	UserID  string `json:"user_id" binding:"required,uuid"`
	MangaID int64  `json:"manga_id" binding:"required,gt=0"`
}

type UpdateProgressRequest struct {
	MangaID    int64  `json:"manga_id" binding:"required,gt=0"`
	MangaTitle string `json:"manga_title,omitempty"`
	Chapter    int    `json:"chapter" binding:"required,min=0"`
	Status     string `json:"status" binding:"required,oneof=reading completed plan_to_read dropped"`
}

type DeleteProgressRequest struct {
	UserID  string `json:"user_id" binding:"required,uuid"`
	MangaID int64  `json:"manga_id" binding:"required,uuid"`
}

type ProgressResponse struct {
	UserID      string `json:"user_id"`
	MangaID     int64  `json:"manga_id"`
	MangaTitle  string `json:"manga_title,omitempty"`
	Chapter     int    `json:"chapter"`
	Status      string `json:"status"`
	LastChapter int    `json:"last_chapter"`
	UpdatedAt   string `json:"updated_at"`
}

type ProgressHistoryResponse struct {
	History []ProgressResponse `json:"history"`
	Total   int                `json:"total"`
}
