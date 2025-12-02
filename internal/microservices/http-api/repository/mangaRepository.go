package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/models"

	"gorm.io/gorm"
)

type MangaRepo struct {
	db *gorm.DB
}

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

	// Fetch paginated results (without genres for better performance)
	if err := r.db.WithContext(ctx).
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

	if len(tokens) == 0 {
		return list, nil
	}

	clauses := make([]string, 0, len(tokens))
	args := make([]interface{}, 0, len(tokens)*3)
	for _, t := range tokens {
		p := "%" + t + "%"
		clauses = append(clauses, "(title ILIKE ? OR COALESCE(author,'') ILIKE ? OR COALESCE(slug,'') ILIKE ?)")
		args = append(args, p, p, p)
	}

	where := strings.Join(clauses, " AND ")
	if err := db.Where(where, args...).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("search manga by title/author: %w", err)
	}
	return list, nil
}

// AdvancedSearch performs full-text search with multiple filters
func (r *MangaRepo) AdvancedSearch(ctx context.Context, filters dto.SearchFilters) ([]models.Manga, int64, error) {
	var list []models.Manga
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Manga{})

	// Full-text search on title, author, description, slug
	if filters.Query != "" {
		tokens := strings.Fields(filters.Query)
		if len(tokens) > 0 {
			clauses := make([]string, 0, len(tokens))
			args := make([]interface{}, 0, len(tokens)*4)
			for _, t := range tokens {
				p := "%" + t + "%"
				clauses = append(clauses, "(title ILIKE ? OR COALESCE(author,'') ILIKE ? OR COALESCE(description,'') ILIKE ? OR COALESCE(slug,'') ILIKE ?)")
				args = append(args, p, p, p, p)
			}
			where := strings.Join(clauses, " AND ")
			db = db.Where(where, args...)
		}
	}

	// Filter by status (ongoing, completed, hiatus)
	if filters.Status != "" {
		db = db.Where("LOWER(status) = LOWER(?)", filters.Status)
	}

	// Filter by minimum average rating
	if filters.MinRating != nil {
		db = db.Where("average_rating >= ?", *filters.MinRating)
	}

	// Filter by genres (many-to-many relationship)
	if len(filters.Genres) > 0 {
		genreConditions := make([]string, 0, len(filters.Genres))
		genreArgs := make([]interface{}, 0, len(filters.Genres))

		for _, g := range filters.Genres {
			// Check if it's a numeric ID or name
			if id, err := strconv.ParseInt(g, 10, 64); err == nil {
				genreConditions = append(genreConditions, "genres.id = ?")
				genreArgs = append(genreArgs, id)
			} else {
				genreConditions = append(genreConditions, "LOWER(genres.name) = LOWER(?)")
				genreArgs = append(genreArgs, g)
			}
		}

		if len(genreConditions) > 0 {
			db = db.Joins("JOIN manga_genres ON manga_genres.manga_id = manga.id").
				Joins("JOIN genres ON genres.id = manga_genres.genre_id").
				Where(strings.Join(genreConditions, " OR "), genreArgs...).
				Group("manga.id").
				Having("COUNT(DISTINCT genres.id) >= ?", len(filters.Genres))
		}
	}

	// Count total matching records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count manga: %w", err)
	}

	// Apply sorting
	switch filters.SortBy {
	case "popularity", "rating":
		db = db.Order("average_rating DESC NULLS LAST")
	case "recent":
		db = db.Order("created_at DESC")
	case "title":
		db = db.Order("title ASC")
	default:
		db = db.Order("created_at DESC")
	}

	// Pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Fetch results without preloading genres for better performance
	if err := db.Limit(pageSize).
		Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("search manga: %w", err)
	}

	return list, total, nil
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
