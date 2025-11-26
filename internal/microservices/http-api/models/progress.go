package models

import "time"

// UserProgress represents the progress of a user reading a manga.
type UserProgress struct {
	UserID         string    `gorm:"type:uuid;not null;primaryKey;index:idx_user_manga" json:"user_id"`
	MangaID        int64     `gorm:"not null;primaryKey;index:idx_user_manga" json:"manga_id"`
	CurrentChapter int       `gorm:"default:0" json:"current_chapter"`
	Status         string    `gorm:"type:text" json:"status"`
	Page           int       `gorm:"default:0" json:"page"`
	UpdatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName overrides the table name used by UserProgress to `user_progress`
func (UserProgress) TableName() string {
	return "user_progress"
}
