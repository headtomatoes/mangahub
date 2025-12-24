package websocket

import (
	"context"
	"mangahub/internal/microservices/http-api/models"

	//"time"

	"gorm.io/gorm"
)

type ChatMessageRepository interface {
	Create(ctx context.Context, message *models.ChatMessage) error
	GetByRoomID(ctx context.Context, roomID int64, limit int) ([]models.ChatMessage, error)
	GetByUserID(ctx context.Context, userID string, limit int) ([]models.ChatMessage, error)
	DeleteByID(ctx context.Context, messageID int64) error
}

type chatMessageRepository struct {
	db *gorm.DB
}

func NewChatMessageRepository(db *gorm.DB) ChatMessageRepository {
	return &chatMessageRepository{db: db}
}

func (r *chatMessageRepository) Create(ctx context.Context, message *models.ChatMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *chatMessageRepository) GetByRoomID(ctx context.Context, roomID int64, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *chatMessageRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *chatMessageRepository) DeleteByID(ctx context.Context, messageID int64) error {
	return r.db.WithContext(ctx).
		Delete(&models.ChatMessage{}, messageID).Error
}
