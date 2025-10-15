package models

import (
	"database/sql"
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
	Last_login    *time.Time `json:"last_login,omitempty" db:"last_login"`
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// queries and methods for user repository would go here

// Example: CreateUser, GetUserByID, GetUserByUsername, UpdateUser, DeleteUser, etc.
// These methods would use prepared statements to interact with the database securely
// and would handle errors appropriately.

func (r *UserRepository) CreateUser(email, username, passwordHash string) (*User, error) {
	user := &User{
		ID:            uuid.New(),
		Email:         email,
		Username:      username,
		Passwork_hash: passwordHash,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `INSERT INTO users (id, email, username, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(query, user.ID, user.Email, user.Username, user.Passwork_hash, user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (*User, error) {
	query := `SELECT id, email, username, password_hash, created_at, updated_at, last_login FROM users WHERE id = $1`

	var user User
	var lastLogin sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Passwork_hash,
		&user.CreatedAt,
		&lastLogin,
	)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.Last_login = &lastLogin.Time
	} else {
		user.Last_login = nil
	}

	return &user, nil
}
