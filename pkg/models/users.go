package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `json:"id" db:"id"` // use UUID for unique user identification rather than incremental IDs
	Username      string     `json:"username" db:"username"`
	Passwork_hash string     `json:"-" db:"password_hash"`
	Email         string     `json:"email" db:"email"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	Last_login    *time.Time `json:"last_login,omitempty" db:"last_login"`
}
