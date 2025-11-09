package service

import (
	"context"
	"errors"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/repository"
)

var (
	ErrProgressNotFound       = errors.New("progress not found")
	ErrFailedToUpdateProgress = errors.New("failed to update progress")
	ErrFailedToGetAllProgress = errors.New("failed to get all progress")
	ErrFailedToGetProgress    = errors.New("failed to get progress")
	ErrFailedToDeleteProgress = errors.New("failed to delete progress")
)

type progressService struct {
	progressRepo repository.ProgressRepository
}

type ProgressService interface {
	GetAllProgress(ctx context.Context, userID string) (*[]models.UserProgress, error)
	GetProgressByMangaID(ctx context.Context, userID string, mangaID int64) (*models.UserProgress, error)
	UpdateProgress(ctx context.Context, progress *models.UserProgress) error
	DeleteProgress(ctx context.Context, userID string, mangaID int64) error
}

func NewProgressService(progressRepo repository.ProgressRepository) ProgressService {
	return &progressService{progressRepo: progressRepo}
}

func (s *progressService) GetAllProgress(ctx context.Context, userID string) (*[]models.UserProgress, error) {
	progressList, err := s.progressRepo.GetAllProgress(ctx, userID)
	if err != nil {
		return nil, ErrFailedToGetAllProgress
	}
	return progressList, nil
}

func (s *progressService) GetProgressByMangaID(ctx context.Context, userID string, mangaID int64) (*models.UserProgress, error) {
	progress, err := s.progressRepo.GetProgressByMangaID(ctx, userID, mangaID)
	if err != nil {
		return nil, ErrFailedToGetProgress
	}
	return progress, nil
}
func (s *progressService) UpdateProgress(ctx context.Context, progress *models.UserProgress) error {
	if err := s.progressRepo.UpdateProgress(ctx, progress); err != nil {
		return ErrFailedToUpdateProgress
	}
	return nil
}
func (s *progressService) DeleteProgress(ctx context.Context, userID string, mangaID int64) error {
	if err := s.progressRepo.DeleteProgress(ctx, userID, mangaID); err != nil {
		return ErrFailedToDeleteProgress
	}
	return nil
}
