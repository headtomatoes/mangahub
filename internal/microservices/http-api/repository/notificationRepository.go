package repository

import (
    "context"
    "mangahub/internal/microservices/http-api/models"
    "gorm.io/gorm"
)

type NotificationRepository interface {
    Create(ctx context.Context, notification *models.Notification) error
    GetUnreadByUser(ctx context.Context, userID string) ([]models.Notification, error)
    MarkAsRead(ctx context.Context, notificationID int64) error
    MarkAllAsRead(ctx context.Context, userID string) error
}

type notificationRepository struct {
    db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
    return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
    return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepository) GetUnreadByUser(ctx context.Context, userID string) ([]models.Notification, error) {
    var notifications []models.Notification
    err := r.db.WithContext(ctx).
        Where("user_id = ? AND read = false", userID).
        Order("created_at DESC").
        Find(&notifications).Error
    return notifications, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID int64) error {
    return r.db.WithContext(ctx).
        Model(&models.Notification{}).
        Where("id = ?", notificationID).
        Update("read", true).Error
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
    return r.db.WithContext(ctx).
        Model(&models.Notification{}).
        Where("user_id = ?", userID).
        Update("read", true).Error
}