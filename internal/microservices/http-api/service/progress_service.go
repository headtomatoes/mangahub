package service

import (
    "context"
    "errors"
    "mangahub/internal/microservices/http-api/models"
    "mangahub/internal/microservices/http-api/repository"
    "time"
)

type ProgressService interface {
    Update(ctx context.Context, userID string, mangaID int64, currentChapter, page int, status string) error
    Get(ctx context.Context, userID string, mangaID int64) (*models.UserProgress, error)
    GetByUser(ctx context.Context, userID string) ([]models.UserProgress, error)
}

type progressService struct {
    repo      repository.ProgressRepository
    mangaRepo *repository.MangaRepo
}

func NewProgressService(repo repository.ProgressRepository, mangaRepo *repository.MangaRepo) ProgressService {
    return &progressService{
        repo:      repo,
        mangaRepo: mangaRepo,
    }
}

func (s *progressService) Update(ctx context.Context, userID string, mangaID int64, currentChapter, page int, status string) error {
    // Validate manga exists
    manga, err := s.mangaRepo.GetByID(ctx, mangaID)
    if err != nil {
        return errors.New("manga not found")
    }
    
    // Validate chapter number
    if manga.TotalChapters != nil && currentChapter > *manga.TotalChapters {
        return errors.New("chapter number exceeds total chapters")
    }
    
    // Validate status
    validStatuses := map[string]bool{
        "reading":       true,
        "completed":     true,
        "plan_to_read":  true,
        "dropped":       true,
    }
    if status != "" && !validStatuses[status] {
        return errors.New("invalid status")
    }
    
    progress := &models.UserProgress{
        UserID:         userID,
        MangaID:        mangaID,
        CurrentChapter: currentChapter,
        Page:           page,
        Status:         status,
        UpdatedAt:      time.Now(),
    }
    
    return s.repo.Update(ctx, progress)
}

func (s *progressService) Get(ctx context.Context, userID string, mangaID int64) (*models.UserProgress, error) {
    return s.repo.Get(ctx, userID, mangaID)
}

func (s *progressService) GetByUser(ctx context.Context, userID string) ([]models.UserProgress, error) {
    return s.repo.GetByUser(ctx, userID)
}