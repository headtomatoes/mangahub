package repository

import (
	"context"
	"fmt"
	"strings"

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

// SearchByTitle performs case-insensitive partial match on title, author and slug.
// Splits query into tokens and requires each token to appear in at least one of the fields.
// Example: "one piece oda" -> WHERE (title ILIKE '%one%' OR author ILIKE '%one%' OR slug ILIKE '%one%')
//
//	AND (title ILIKE '%piece%' OR author ILIKE '%piece%' OR slug ILIKE '%piece%') ...
func (r *MangaRepo) SearchByTitle(ctx context.Context, title string) ([]models.Manga, error) {
	var list []models.Manga
	tokens := strings.Fields(title)
	db := r.db.WithContext(ctx)

	// if empty tokens, return empty list
	if len(tokens) == 0 {
		return list, nil
	}

	clauses := make([]string, 0, len(tokens))
	args := make([]interface{}, 0, len(tokens)*3)
	for _, t := range tokens {
		p := "%" + t + "%"
		// use COALESCE to avoid NULL author/slug causing ILIKE failure
		clauses = append(clauses, "(title ILIKE ? OR COALESCE(author,'') ILIKE ? OR COALESCE(slug,'') ILIKE ?)")
		args = append(args, p, p, p)
	}

	where := strings.Join(clauses, " AND ")
	if err := db.Where(where, args...).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("search manga by title/author: %w", err)
	}
	return list, nil
}
