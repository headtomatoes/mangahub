package models

import "time"

type ChatMessage struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomID    int64     `gorm:"not null;index:idx_chat_messages_room_id" json:"room_id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	UserName  string    `gorm:"not null" json:"user_name"`
	Message   string    `gorm:"not null" json:"message"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	// Associations
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Manga *Manga `gorm:"foreignKey:RoomID" json:"manga,omitempty"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
