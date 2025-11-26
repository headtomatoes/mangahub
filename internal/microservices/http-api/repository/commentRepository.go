package repository

import (
	"errors"

	"mangahub/internal/microservices/http-api/models"

	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *models.Comment) error
	Update(comment *models.Comment) error
	Delete(commentID int64, userID string) error
	GetByID(commentID int64) (*models.Comment, error)
	GetByManga(mangaID int64, page, pageSize int) ([]models.Comment, int64, error)
	GetByUser(userID string, page, pageSize int) ([]models.Comment, int64, error)
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// Create a new comment
func (r *commentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

// Update an existing comment
func (r *commentRepository) Update(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

// Delete a comment (only if user owns it)
func (r *commentRepository) Delete(commentID int64, userID string) error {
	result := r.db.Where("id = ? AND user_id = ?", commentID, userID).Delete(&models.Comment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("comment not found or you don't have permission to delete it")
	}
	return nil
}

// GetByID retrieves a comment by its ID
func (r *commentRepository) GetByID(commentID int64) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.Where("id = ?", commentID).
		Preload("User").
		First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetByManga retrieves all comments for a specific manga with pagination
func (r *commentRepository) GetByManga(mangaID int64, page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	// Count total comments
	if err := r.db.Model(&models.Comment{}).Where("manga_id = ?", mangaID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated comments
	offset := (page - 1) * pageSize
	err := r.db.Where("manga_id = ?", mangaID).
		Preload("User").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&comments).Error

	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// GetByUser retrieves all comments by a specific user with pagination
func (r *commentRepository) GetByUser(userID string, page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	// Count total comments
	if err := r.db.Model(&models.Comment{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated comments
	offset := (page - 1) * pageSize
	err := r.db.Where("user_id = ?", userID).
		Preload("User").
		Preload("Manga").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&comments).Error

	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}
