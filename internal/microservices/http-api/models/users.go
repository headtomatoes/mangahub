package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        string     `gorm:"primaryKey;type:uuid" json:"id"`
	Username  string     `gorm:"uniqueIndex;not null" json:"username"`
	Email     string     `gorm:"uniqueIndex;not null" json:"email"`
	Password  string     `gorm:"column:password_hash;not null" json:"-"` // Not show in JSON
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// BeforeCreate hook to set UUID before creating a User
func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	// If the ID is not already set, generate a new one.
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	return
}
