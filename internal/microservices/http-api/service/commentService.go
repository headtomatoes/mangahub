package service

import (
	"context"
	"errors"

	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/repository"

	"gorm.io/gorm"
)

type CommentService interface {
	CreateComment(userID string, mangaID int64, content string) (*dto.CommentResponse, error)
	UpdateComment(commentID int64, userID string, content string) (*dto.CommentResponse, error)
	DeleteComment(commentID int64, userID string) error
	GetCommentByID(commentID int64) (*dto.CommentResponse, error)
	GetMangaComments(mangaID int64, page, pageSize int) (*dto.PaginatedCommentResponse, error)
	GetUserComments(userID string, page, pageSize int) (*dto.PaginatedCommentResponse, error)
}

type commentService struct {
	commentRepo repository.CommentRepository
	mangaRepo   *repository.MangaRepo
}

func NewCommentService(commentRepo repository.CommentRepository, mangaRepo *repository.MangaRepo) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		mangaRepo:   mangaRepo,
	}
}

// CreateComment creates a new comment for a manga
func (s *commentService) CreateComment(userID string, mangaID int64, content string) (*dto.CommentResponse, error) {
	ctx := context.Background()

	// Check if manga exists
	_, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("manga not found")
		}
		return nil, err
	}

	// Create new comment
	comment := &models.Comment{
		UserID:  userID,
		MangaID: mangaID,
		Content: content,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	// Reload with user data
	comment, err = s.commentRepo.GetByID(comment.ID)
	if err != nil {
		return nil, err
	}

	return dto.FromModelToCommentResponse(comment), nil
}

// UpdateComment updates an existing comment
func (s *commentService) UpdateComment(commentID int64, userID string, content string) (*dto.CommentResponse, error) {
	// Get existing comment
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		return nil, err
	}

	// Check ownership
	if comment.UserID != userID {
		return nil, errors.New("you don't have permission to update this comment")
	}

	// Update content
	comment.Content = content
	if err := s.commentRepo.Update(comment); err != nil {
		return nil, err
	}

	// Reload with user data
	comment, err = s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, err
	}

	return dto.FromModelToCommentResponse(comment), nil
}

// DeleteComment deletes a comment
func (s *commentService) DeleteComment(commentID int64, userID string) error {
	return s.commentRepo.Delete(commentID, userID)
}

// GetCommentByID retrieves a comment by ID
func (s *commentService) GetCommentByID(commentID int64) (*dto.CommentResponse, error) {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		return nil, err
	}

	return dto.FromModelToCommentResponse(comment), nil
}

// GetMangaComments retrieves all comments for a manga with pagination
func (s *commentService) GetMangaComments(mangaID int64, page, pageSize int) (*dto.PaginatedCommentResponse, error) {
	ctx := context.Background()

	// Check if manga exists
	_, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("manga not found")
		}
		return nil, err
	}

	comments, total, err := s.commentRepo.GetByManga(mangaID, page, pageSize)
	if err != nil {
		return nil, err
	}

	commentResponses := make([]dto.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		commentResponses = append(commentResponses, *dto.FromModelToCommentResponse(&comment))
	}

	return dto.NewPaginatedCommentResponse(commentResponses, int(total), page, pageSize), nil
}

// GetUserComments retrieves all comments by a user with pagination
func (s *commentService) GetUserComments(userID string, page, pageSize int) (*dto.PaginatedCommentResponse, error) {
	comments, total, err := s.commentRepo.GetByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	commentResponses := make([]dto.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		commentResponses = append(commentResponses, *dto.FromModelToCommentResponse(&comment))
	}

	return dto.NewPaginatedCommentResponse(commentResponses, int(total), page, pageSize), nil
}
