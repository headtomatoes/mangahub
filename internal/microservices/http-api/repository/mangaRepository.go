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

// type MangaRepo interface {
//     GetByID(ctx context.Context, id int64) (*models.Manga, error)
//     Search(ctx context.Context, query string, limit, offset int) ([]*models.Manga, int64, error) // returns items, totalCount, error
// }

func NewMangaRepo(db *gorm.DB) *MangaRepo {
	return &MangaRepo{db: db}
}

func (r *MangaRepo) GetAll(ctx context.Context, page, pageSize int) ([]models.Manga, int64, error) {
	var list []models.Manga
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&models.Manga{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Fetch paginated results
	if err := r.db.WithContext(ctx).
		Preload("Genres").
		Order("created_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *MangaRepo) GetByID(ctx context.Context, id int64) (*models.Manga, error) {
	var m models.Manga
	if err := r.db.WithContext(ctx).Preload("Genres").First(&m, id).Error; err != nil {
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

func (r *MangaRepo) GetGenresByManga(ctx context.Context, mangaID int64) ([]models.Genre, error) {
	var m models.Manga
	if err := r.db.WithContext(ctx).Preload("Genres").First(&m, mangaID).Error; err != nil {
		return nil, fmt.Errorf("get manga: %w", err)
	}
	return m.Genres, nil
}

func (r *MangaRepo) AddGenresToManga(ctx context.Context, mangaID int64, genreIDs []int64) error {
	if len(genreIDs) == 0 {
		return nil
	}
	tx := r.db.WithContext(ctx).Begin()
	var m models.Manga
	if err := tx.First(&m, mangaID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("manga not found: %w", err)
	}
	// build genre placeholders (only IDs needed)
	genres := make([]models.Genre, 0, len(genreIDs))
	for _, id := range genreIDs {
		genres = append(genres, models.Genre{ID: id})
	}
	if err := tx.Model(&m).Association("Genres").Append(&genres); err != nil {
		tx.Rollback()
		return fmt.Errorf("append genres: %w", err)
	}
	return tx.Commit().Error
}

func (r *MangaRepo) RemoveGenresFromManga(ctx context.Context, mangaID int64, genreIDs []int64) error {
	if len(genreIDs) == 0 {
		return nil
	}
	tx := r.db.WithContext(ctx).Begin()
	var m models.Manga
	if err := tx.First(&m, mangaID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("manga not found: %w", err)
	}
	genres := make([]models.Genre, 0, len(genreIDs))
	for _, id := range genreIDs {
		genres = append(genres, models.Genre{ID: id})
	}
	if err := tx.Model(&m).Association("Genres").Delete(&genres); err != nil {
		tx.Rollback()
		return fmt.Errorf("remove genres: %w", err)
	}
	return tx.Commit().Error
}
