package dto

import "time"

// AddToLibraryRequest: payload to add manga to user's library
type AddToLibraryRequest struct {
    MangaID int64 `json:"manga_id" binding:"required"`
}

// LibraryResponse: response for a library item
type LibraryResponse struct {
    ID        int64         `json:"id"`
    MangaID   int64         `json:"manga_id"`
    Manga     MangaResponse `json:"manga,omitempty"`
    AddedAt   time.Time     `json:"added_at"`
}

// LibraryListResponse: list of library items
type LibraryListResponse struct {
    Items []LibraryResponse `json:"items"`
    Total int               `json:"total"`
}