package models

import "time"

type Manga struct {
	ID            int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Slug          *string    `json:"slug,omitempty" gorm:"uniqueIndex;size:200"`
	Title         string     `json:"title" gorm:"not null"`
	Author        *string    `json:"author,omitempty"`
	Status        *string    `json:"status,omitempty"`
	TotalChapters *int       `json:"total_chapters,omitempty"`
	Description   *string    `json:"description,omitempty"`
	AverageRating *float64   `json:"average_rating,omitempty" gorm:"type:decimal(3,2)"`
	CoverURL      *string    `json:"cover_url,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty" gorm:"autoCreateTime"`

	// association
	Genres []Genre `json:"genres,omitempty" gorm:"many2many:manga_genres;constraint:OnDelete:CASCADE;"`
}

func (Manga) TableName() string {
	return "manga"
}
