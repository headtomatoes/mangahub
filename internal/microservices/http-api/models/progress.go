package models

import "time"

type UserProgress struct {
    UserID         string    `gorm:"type:uuid;primaryKey" json:"user_id"`
    MangaID        int64     `gorm:"primaryKey" json:"manga_id"`
    CurrentChapter int       `gorm:"default:0" json:"current_chapter"`
    Status         string    `gorm:"type:text;check:status IN ('reading','completed','plan_to_read','dropped')" json:"status"`
    Page           int       `gorm:"default:0" json:"page"`
    UpdatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
    
    // Associations
    User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Manga *Manga `gorm:"foreignKey:MangaID" json:"manga,omitempty"`
}

func (UserProgress) TableName() string {
    return "user_progress"
}