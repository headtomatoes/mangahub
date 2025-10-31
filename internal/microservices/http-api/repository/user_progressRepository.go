package repository

import (
    "context"
    "fmt"
    "mangahub/internal/microservices/http-api/models"

    "gorm.io/gorm"
)

type ProgressRepository interface {
    Update(ctx context.Context, progress *models.UserProgress) error
    Get(ctx context.Context, userID string, mangaID int64) (*models.UserProgress, error)
    GetByUser(ctx context.Context, userID string) ([]models.UserProgress, error)
}

type progressRepository struct {
    db *gorm.DB
}

func NewProgressRepository(db *gorm.DB) ProgressRepository {
    return &progressRepository{db: db}
}

func (r *progressRepository) Update(ctx context.Context, progress *models.UserProgress) error {
    // Use Save to insert or update (upsert behavior)
    if err := r.db.WithContext(ctx).Save(progress).Error; err != nil {
        return fmt.Errorf("update progress: %w", err)
    }
    return nil
}

func (r *progressRepository) Get(ctx context.Context, userID string, mangaID int64) (*models.UserProgress, error) {
    var progress models.UserProgress
    
    if err := r.db.WithContext(ctx).
        Where("user_id = ? AND manga_id = ?", userID, mangaID).
        First(&progress).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil // No progress yet
        }
        return nil, fmt.Errorf("get progress: %w", err)
    }
    
    return &progress, nil
}

func (r *progressRepository) GetByUser(ctx context.Context, userID string) ([]models.UserProgress, error) {
    var progress []models.UserProgress
    
    if err := r.db.WithContext(ctx).
        Preload("Manga").
        Where("user_id = ?", userID).
        Order("updated_at DESC").
        Find(&progress).Error; err != nil {
        return nil, fmt.Errorf("get user progress: %w", err)
    }
    
    return progress, nil
}