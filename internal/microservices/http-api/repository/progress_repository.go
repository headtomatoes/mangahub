package repository

import (
	"context"
	"mangahub/internal/microservices/http-api/models"
	"time"

	"gorm.io/gorm"
)

type progressRepository struct {
	db *gorm.DB
}

type ProgressRepository interface {
	GetAllProgress(ctx context.Context, userID string) (*[]models.UserProgress, error)
	GetProgressByMangaID(ctx context.Context, userID string, mangaID int64) (*models.UserProgress, error)
	UpdateProgress(ctx context.Context, progress *models.UserProgress) error
	DeleteProgress(ctx context.Context, userID string, mangaID int64) error
}

func NewProgressRepository(db *gorm.DB) ProgressRepository {
	return &progressRepository{db: db}
}

func (r *progressRepository) GetAllProgress(ctx context.Context, userID string) (*[]models.UserProgress, error) {
	var list []models.UserProgress
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&list).Error; err != nil {
		return nil, err
	}
	return &list, nil
}
func (r *progressRepository) GetProgressByMangaID(ctx context.Context, userID string, mangaID int64) (*models.UserProgress, error) {
	var progress models.UserProgress
	if err := r.db.WithContext(ctx).Where("user_id = ? AND manga_id = ?", userID, mangaID).First(&progress).Error; err != nil {
		return nil, err
	}
	return &progress, nil
}

// Upsert progress record
func (r *progressRepository) UpdateProgress(ctx context.Context, progress *models.UserProgress) error {
	// Try to find existing progress
	var existing models.UserProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND manga_id = ?", progress.UserID, progress.MangaID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new record
		progress.UpdatedAt = time.Now()
		return r.db.WithContext(ctx).Create(progress).Error
	} else if err != nil {
		return err
	}

	// Update existing record
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"current_chapter": progress.CurrentChapter,
		"status":          progress.Status,
		"updated_at":      time.Now(),
	}).Error
}
func (r *progressRepository) DeleteProgress(ctx context.Context, userID string, mangaID int64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND manga_id = ?", userID, mangaID).Delete(&models.UserProgress{}).Error; err != nil {
		return err
	}
	return nil
}
