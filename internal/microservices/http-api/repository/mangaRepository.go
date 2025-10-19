package repository

import (
	"context"
	"fmt"

	"mangahub/internal/microservices/http-api/models"

	"gorm.io/gorm"
)

type MangaRepo struct {
	db *gorm.DB
}

func NewMangaRepo(db *gorm.DB) *MangaRepo {
	return &MangaRepo{db: db}
}

func (r *MangaRepo) GetAll(ctx context.Context) ([]models.Manga, error) {
	var list []models.Manga
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *MangaRepo) GetByID(ctx context.Context, id int64) (*models.Manga, error) {
	var m models.Manga
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MangaRepo) Create(ctx context.Context, m *models.Manga) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("create manga: %w", err)
	}
	// GORM will populate m.ID and m.CreatedAt
	return nil
}

func (r *MangaRepo) Update(ctx context.Context, id int64, m *models.Manga) error {
	// ensure ID set for Save
	m.ID = id
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("update manga: %w", err)
	}
	return nil
}

func (r *MangaRepo) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&models.Manga{}, id).Error; err != nil {
		return fmt.Errorf("delete manga: %w", err)
	}
	return nil
}
