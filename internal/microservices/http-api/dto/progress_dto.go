package dto

// DTOs for progress-related operations in HTTP API

type GetProgressByMangaIDRequest struct {
	UserID  string `json:"user_id" binding:"required,uuid"`
	MangaID int64  `json:"manga_id" binding:"required,gt=0"`
}

type UpdateProgressRequest struct {
	UserID  string `json:"user_id" binding:"required,uuid"`
	MangaID int64  `json:"manga_id" binding:"required,gt=0"`
	Chapter int    `json:"chapter" binding:"required,min=0"`
	Status  string `json:"status" binding:"required,oneof=reading completed on-hold dropped plan-to-read"`
}

type DeleteProgressRequest struct {
	UserID  string `json:"user_id" binding:"required,uuid"`
	MangaID int64  `json:"manga_id" binding:"required,uuid"`
}

type ProgressResponse struct {
	UserID    string `json:"user_id"`
	MangaID   int64  `json:"manga_id"`
	Chapter   int    `json:"chapter"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}
