package udp

import (
	"context"
	"fmt"
	"log"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/repository"
	"net"
	"sync"
)

type Broadcaster struct {
	conn             *net.UDPConn
	subManager       *SubscriberManager
	libraryRepo      repository.LibraryRepository
	notificationRepo repository.NotificationRepository
	mu               sync.RWMutex
}

func NewBroadcaster(
	conn *net.UDPConn,
	subManager *SubscriberManager,
	libraryRepo repository.LibraryRepository,
	notificationRepo repository.NotificationRepository,
) *Broadcaster {
	return &Broadcaster{
		conn:             conn,
		subManager:       subManager,
		libraryRepo:      libraryRepo,
		notificationRepo: notificationRepo,
	}
}

// BroadcastToLibraryUsers sends notification AND stores it for offline users
func (b *Broadcaster) BroadcastToLibraryUsers(ctx context.Context, mangaID int64, notification *Notification) error {
	data, err := notification.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Get all users who have this manga in their library
	userIDs, err := b.libraryRepo.GetUserIDsByMangaID(ctx, mangaID)
	if err != nil {
		return fmt.Errorf("failed to get library users: %w", err)
	}

	if len(userIDs) == 0 {
		log.Printf("No users found for manga ID %d", mangaID)
		return nil
	}

	// Store notification in database for ALL users (online and offline)
	for _, userID := range userIDs {
		dbNotification := &models.Notification{
			UserID:  userID,
			Type:    string(notification.Type),
			MangaID: mangaID,
			Title:   notification.Title,
			Message: notification.Message,
			Read:    false,
		}
		if err := b.notificationRepo.Create(ctx, dbNotification); err != nil {
			log.Printf("Failed to store notification for user %s: %v", userID, err)
		}
	}

	// Send to currently online subscribers via UDP
	subscribers := b.subManager.GetByUserIDs(userIDs)
	var wg sync.WaitGroup

	for _, sub := range subscribers {
		wg.Add(1)
		go func(s *Subscriber) {
			defer wg.Done()
			if err := b.sendToSubscriber(s, data); err != nil {
				log.Printf("Failed to send to %s: %v", s.UserID, err)
			}
		}(sub)
	}

	wg.Wait()
	log.Printf("Notification sent to %d online users and stored for %d total users (manga ID %d)",
		len(subscribers), len(userIDs), mangaID)

	return nil
}

// BroadcastToAll sends notification to all active subscribers
func (b *Broadcaster) BroadcastToAll(notification *Notification) error {
	data, err := notification.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	subscribers := b.subManager.GetAll()
	var wg sync.WaitGroup

	for _, sub := range subscribers {
		wg.Add(1)
		go func(s *Subscriber) {
			defer wg.Done()
			if err := b.sendToSubscriber(s, data); err != nil {
				log.Printf("Failed to send to %s: %v", s.UserID, err)
			}
		}(sub)
	}

	wg.Wait()
	log.Printf("Broadcast sent to %d subscribers", len(subscribers))
	return nil
}

// sendToSubscriber sends data to a specific subscriber
func (b *Broadcaster) sendToSubscriber(sub *Subscriber, data []byte) error {
	_, err := b.conn.WriteToUDP(data, sub.Addr)
	if err != nil {
		sub.Active = false
		return err
	}
	return nil
}
