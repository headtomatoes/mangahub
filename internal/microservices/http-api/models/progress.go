package models

import "time"

// UserProgress represents the progress of a user reading a manga.
type UserProgress struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         string    `gorm:"type:uuid;not null;index:idx_user_manga" json:"user_id"`
	MangaID        int64     `gorm:"not null;index:idx_user_manga" json:"manga_id"`
	CurrentChapter int       `gorm:"default:0" json:"current_chapter"`
	Status         string    `gorm:"type:text" json:"status"`
	UpdatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}
