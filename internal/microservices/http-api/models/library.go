package models

import "time"

type UserLibrary struct {
    ID      int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID  string    `gorm:"type:uuid;not null;index" json:"user_id"`
    MangaID int64     `gorm:"not null;index" json:"manga_id"`
    AddedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"added_at"`
    
    // Associations
    User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Manga *Manga `gorm:"foreignKey:MangaID" json:"manga,omitempty"`
}

func (UserLibrary) TableName() string {
    return "user_library"
}