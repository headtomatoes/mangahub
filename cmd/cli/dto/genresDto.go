package dto

import "mangahub/internal/microservices/http-api/models"

// CreateGenreDTO for POST /api/genres
type CreateGenreDTO struct {
    Name string `json:"name" binding:"required"`
}

type GenreResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

type GenreIDsRequest struct {
    GenreIDs []int64 `json:"genre_ids" binding:"required"`
}

func GenreFromModel(g models.Genre) GenreResponse {
    return GenreResponse{
        ID:   g.ID,
        Name: g.Name,
    }
}