package repository

import (
    "context"
    "fmt"
    "mangahub/internal/microservices/http-api/models"

    "gorm.io/gorm"
)

type LibraryRepository interface {
    Add(ctx context.Context, userID string, mangaID int64) error
    Remove(ctx context.Context, userID string, mangaID int64) error
    List(ctx context.Context, userID string) ([]models.UserLibrary, error)
    Exists(ctx context.Context, userID string, mangaID int64) (bool, error)
}

type libraryRepository struct {
    db *gorm.DB
}

func NewLibraryRepository(db *gorm.DB) LibraryRepository {
    return &libraryRepository{db: db}
}

func (r *libraryRepository) Add(ctx context.Context, userID string, mangaID int64) error {
    library := &models.UserLibrary{
        UserID:  userID,
        MangaID: mangaID,
    }
    
    if err := r.db.WithContext(ctx).Create(library).Error; err != nil {
        return fmt.Errorf("add to library: %w", err)
    }
    return nil
}

func (r *libraryRepository) Remove(ctx context.Context, userID string, mangaID int64) error {
    result := r.db.WithContext(ctx).
        Where("user_id = ? AND manga_id = ?", userID, mangaID).
        Delete(&models.UserLibrary{})
    
    if result.Error != nil {
        return fmt.Errorf("remove from library: %w", result.Error)
    }
    
    if result.RowsAffected == 0 {
        return fmt.Errorf("manga not found in library")
    }
    
    return nil
}

func (r *libraryRepository) List(ctx context.Context, userID string) ([]models.UserLibrary, error) {
    var library []models.UserLibrary
    
    if err := r.db.WithContext(ctx).
        Preload("Manga").
        Where("user_id = ?", userID).
        Order("added_at DESC").
        Find(&library).Error; err != nil {
        return nil, fmt.Errorf("list library: %w", err)
    }
    
    return library, nil
}

func (r *libraryRepository) Exists(ctx context.Context, userID string, mangaID int64) (bool, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&models.UserLibrary{}).
        Where("user_id = ? AND manga_id = ?", userID, mangaID).
        Count(&count).Error; err != nil {
        return false, err
    }
    return count > 0, nil
}