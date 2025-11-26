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

func GenreFromModel(g models.Genre) GenreResponse {
	return GenreResponse{
		ID:   g.ID,
		Name: g.Name,
	}
}
