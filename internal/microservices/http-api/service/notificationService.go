package service

import (
    "context"
    "errors"
    "mangahub/internal/microservices/http-api/models"
    "mangahub/internal/microservices/http-api/repository"
)

type NotificationService interface {
    GetUnread(ctx context.Context, userID string) ([]models.Notification, error)
    MarkAsRead(ctx context.Context, userID string, notificationID int64) error
    MarkAllAsRead(ctx context.Context, userID string) error
}

type notificationService struct {
    repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
    return &notificationService{repo: repo}
}

func (s *notificationService) GetUnread(ctx context.Context, userID string) ([]models.Notification, error) {
    return s.repo.GetUnreadByUser(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID string, notificationID int64) error {
    // Verify notification belongs to user
    notifications, err := s.repo.GetUnreadByUser(ctx, userID)
    if err != nil {
        return err
    }
    
    found := false
    for _, n := range notifications {
        if n.ID == notificationID {
            found = true
            break
        }
    }
    
    if !found {
        return errors.New("notification not found or already read")
    }
    
    return s.repo.MarkAsRead(ctx, notificationID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
    return s.repo.MarkAllAsRead(ctx, userID)
}