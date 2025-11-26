package repository

import (
	"errors"

	"mangahub/internal/microservices/http-api/models"

	"gorm.io/gorm"
)

type RatingRepository interface {
	Create(rating *models.Rating) error
	Update(rating *models.Rating) error
	Delete(userID string, mangaID int64) error
	GetByUserAndManga(userID string, mangaID int64) (*models.Rating, error)
	GetByManga(mangaID int64, page, pageSize int) ([]models.Rating, int64, error)
	CalculateAverageRating(mangaID int64) (float64, error)
	CountRatings(mangaID int64) (int64, error)
}

type ratingRepository struct {
	db *gorm.DB
}

func NewRatingRepository(db *gorm.DB) RatingRepository {
	return &ratingRepository{db: db}
}

// Create a new rating
func (r *ratingRepository) Create(rating *models.Rating) error {
	return r.db.Create(rating).Error
}

// Update an existing rating
func (r *ratingRepository) Update(rating *models.Rating) error {
	return r.db.Save(rating).Error
}

// Delete a rating by user and manga
func (r *ratingRepository) Delete(userID string, mangaID int64) error {
	result := r.db.Where("user_id = ? AND manga_id = ?", userID, mangaID).Delete(&models.Rating{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("rating not found")
	}
	return nil
}

// GetByUserAndManga retrieves a user's rating for a specific manga
func (r *ratingRepository) GetByUserAndManga(userID string, mangaID int64) (*models.Rating, error) {
	var rating models.Rating
	err := r.db.Where("user_id = ? AND manga_id = ?", userID, mangaID).
		Preload("User").
		First(&rating).Error
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

// GetByManga retrieves all ratings for a specific manga with pagination
func (r *ratingRepository) GetByManga(mangaID int64, page, pageSize int) ([]models.Rating, int64, error) {
	var ratings []models.Rating
	var total int64

	// Count total ratings
	if err := r.db.Model(&models.Rating{}).Where("manga_id = ?", mangaID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated ratings
	offset := (page - 1) * pageSize
	err := r.db.Where("manga_id = ?", mangaID).
		Preload("User").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&ratings).Error

	if err != nil {
		return nil, 0, err
	}

	return ratings, total, nil
}

// CalculateAverageRating calculates the average rating for a manga
func (r *ratingRepository) CalculateAverageRating(mangaID int64) (float64, error) {
	var avg struct {
		Average float64
	}

	err := r.db.Model(&models.Rating{}).
		Select("COALESCE(AVG(rating), 0) as average").
		Where("manga_id = ?", mangaID).
		Scan(&avg).Error

	if err != nil {
		return 0, err
	}

	return avg.Average, nil
}

// CountRatings counts the total number of ratings for a manga
func (r *ratingRepository) CountRatings(mangaID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.Rating{}).Where("manga_id = ?", mangaID).Count(&count).Error
	return count, err
}
