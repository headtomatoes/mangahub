package repository

import (
	//"context"
	"mangahub/internal/microservices/http-api/models"

	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations.
// added context to resilient for cancellation and timeouts also better tracing (but for now we just make it simple)
type UserRepository interface {
	Create(user *models.User) error
	FindByUsername(username string) (*models.User, error)
	FindByID(id string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
}

// userRepository is the GORM implementation of UserRepository.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository in a GORM implementation
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	// .WithContext(ctx) to pass context to gorm
	// check for the error if the user is not found
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		// then we return nil and the error
		// prevent returning a zero-value user struct => which make GORM think user is found => erroneous behavior
		// we do it for all query methods
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
