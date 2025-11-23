package dto

import (
	"time"

	"mangahub/internal/microservices/http-api/models"
)

// CreateRatingDTO for creating or updating a rating
type CreateRatingDTO struct {
	Rating int `json:"rating" binding:"required,min=1,max=10"`
}

// RatingResponse for returning rating information (for list view - without IDs)
type RatingResponse struct {
	Username  string    `json:"username"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromModelToRatingResponse converts a Rating model to RatingResponse DTO
func FromModelToRatingResponse(rating *models.Rating) *RatingResponse {
	return &RatingResponse{
		Username:  rating.User.Username,
		Rating:    rating.Rating,
		CreatedAt: rating.CreatedAt,
		UpdatedAt: rating.UpdatedAt,
	}
}

// UserRatingResponse for returning user's own rating
type UserRatingResponse struct {
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PaginatedRatingResponse for returning paginated ratings
type PaginatedRatingResponse struct {
	Data       []RatingResponse `json:"data"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	Total      int              `json:"total"`
	TotalPages int              `json:"total_pages"`
}

// NewPaginatedRatingResponse creates a paginated rating response
func NewPaginatedRatingResponse(data []RatingResponse, total, page, pageSize int) *PaginatedRatingResponse {
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &PaginatedRatingResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
