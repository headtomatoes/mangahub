package repository

import (
	"mangahub/internal/microservices/http-api/models"

	"gorm.io/gorm"
)

// RefreshTokenRepository handles database operations for refresh tokens
type RefreshTokenRepository interface {
	Create(refreshToken *models.RefreshToken) error
	FindByToken(tokenString string) (*models.RefreshToken, error)
	Revoke(tokenID string) error
	Delete(tokenID string) error
}

// refreshTokenRepository is the GORM implementation of RefreshTokenRepository
type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new instance of RefreshTokenRepository
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db} // note: what does it do here?
}

// Create: creates a new refresh token for a user with a specified TTL
func (r *refreshTokenRepository) Create(refreshToken *models.RefreshToken) error {
	return r.db.Create(refreshToken).Error
}

// FindByToken: look up the refresh token by its token string
func (r *refreshTokenRepository) FindByToken(tokenString string) (*models.RefreshToken, error) {
	// 1st var to hold the result
	// 2nd check for the 1st then return error if any
	var refreshToken models.RefreshToken
	if err := r.db.Where("token = ?", tokenString).First(&refreshToken).Error; err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

// Revoke: marks a refresh token as revoked
func (r *refreshTokenRepository) Revoke(tokenID string) error {
	return r.db.Model(&models.RefreshToken{}).Where("id = ?", tokenID).Update("revoked", true).Error
}

// Delete: removes a refresh token from the database based on its revoked status(true)
// can be use with time-based cleanup of revoked tokens or triggered cleanup
func (r *refreshTokenRepository) Delete(tokenID string) error {
	return r.db.Where("id = ?", tokenID).Delete(&models.RefreshToken{}).Error
}
