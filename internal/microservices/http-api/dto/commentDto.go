package dto

import (
	"time"

	"mangahub/internal/microservices/http-api/models"
)

// CreateCommentDTO for creating a comment
type CreateCommentDTO struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

// UpdateCommentDTO for updating a comment
type UpdateCommentDTO struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

// CommentResponse for returning comment information (for list view - without IDs)
type CommentResponse struct {
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromModelToCommentResponse converts a Comment model to CommentResponse DTO
func FromModelToCommentResponse(comment *models.Comment) *CommentResponse {
	return &CommentResponse{
		Username:  comment.User.Username,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}

// PaginatedCommentResponse for returning paginated comments
type PaginatedCommentResponse struct {
	Data       []CommentResponse `json:"data"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	Total      int               `json:"total"`
	TotalPages int               `json:"total_pages"`
}

// NewPaginatedCommentResponse creates a paginated comment response
func NewPaginatedCommentResponse(data []CommentResponse, total, page, pageSize int) *PaginatedCommentResponse {
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &PaginatedCommentResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
