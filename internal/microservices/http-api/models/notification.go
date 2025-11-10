package models

import "time"

type Notification struct {
    ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
    Type      string    `gorm:"not null" json:"type"` // NEW_MANGA, NEW_CHAPTER, MANGA_UPDATE
    MangaID   int64     `json:"manga_id"`
    Title     string    `json:"title"`
    Message   string    `json:"message"`
    Read      bool      `gorm:"default:false" json:"read"`
    CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
    
    // Associations
    User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Manga *Manga `gorm:"foreignKey:MangaID" json:"manga,omitempty"`
}

func (Notification) TableName() string {
    return "notifications"
}
