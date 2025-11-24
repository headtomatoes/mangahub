package models

import "time"

type Rating struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	MangaID   int64     `json:"manga_id" gorm:"not null;index"`
	Rating    int       `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 10"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Associations
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Manga Manga `json:"manga,omitempty" gorm:"foreignKey:MangaID;constraint:OnDelete:CASCADE;"`
}

func (Rating) TableName() string {
	return "ratings"
}
