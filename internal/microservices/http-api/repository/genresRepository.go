package repository

import (
	"context"
	"fmt"

	"mangahub/internal/microservices/http-api/models"

	"gorm.io/gorm"
)

type GenreRepo struct {
	db *gorm.DB
}

func NewGenreRepo(db *gorm.DB) *GenreRepo {
	return &GenreRepo{db: db}
}

func (r *GenreRepo) GetAll(ctx context.Context) ([]models.Genre, error) {
	var list []models.Genre
	if err := r.db.WithContext(ctx).Order("name asc").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("get genres: %w", err)
	}
	return list, nil
}

func (r *GenreRepo) Create(ctx context.Context, g *models.Genre) error {
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		return fmt.Errorf("create genre: %w", err)
	}
	return nil
}

// GetMangasByGenre returns mangas associated with the given genre id.
// Preloads Genres on each manga.
func (r *GenreRepo) GetMangasByGenre(ctx context.Context, genreID int64) ([]models.Manga, error) {
	var list []models.Manga
	if err := r.db.WithContext(ctx).
		Model(&models.Manga{}).
		Joins("JOIN manga_genres mg ON mg.manga_id = manga.id").
		Where("mg.genre_id = ?", genreID).
		Preload("Genres").
		Order("created_at desc").
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("get mangas by genre: %w", err)
	}
	return list, nil
}
