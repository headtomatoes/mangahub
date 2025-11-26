package models

import "time"

type Comment struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	MangaID   int64     `json:"manga_id" gorm:"not null;index"`
	Content   string    `json:"content" gorm:"not null;type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Associations
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Manga Manga `json:"manga,omitempty" gorm:"foreignKey:MangaID;constraint:OnDelete:CASCADE;"`
}

func (Comment) TableName() string {
	return "comments"
}
