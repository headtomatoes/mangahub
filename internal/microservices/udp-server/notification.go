package udp

import (
	"encoding/json"
	"time"
)

// NotificationType defines the type of notification
type NotificationType string

const (
	NotificationNewManga    NotificationType = "NEW_MANGA"
	NotificationNewChapter  NotificationType = "NEW_CHAPTER"
	NotificationMangaUpdate NotificationType = "MANGA_UPDATE"
	NotificationSubscribe   NotificationType = "SUBSCRIBE"
	NotificationUnsubscribe NotificationType = "UNSUBSCRIBE"
)

// Notification represents a notification message
type Notification struct {
	Type      NotificationType `json:"type"`
	MangaID   int64            `json:"manga_id"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Timestamp time.Time        `json:"timestamp"`
	Data      interface{}      `json:"data,omitempty"`
}

// NewMangaNotification creates a notification for new manga
func NewMangaNotification(mangaID int64, title string) *Notification {
	return &Notification{
		Type:      NotificationNewManga,
		MangaID:   mangaID,
		Title:     title,
		Message:   "New manga added: " + title,
		Timestamp: time.Now(),
	}
}
   
// NewChapterNotification creates a notification for new chapter
func NewChapterNotification(mangaID int64, title string, chapter int) *Notification {
	return &Notification{
		Type:      NotificationNewChapter,
		MangaID:   mangaID,
		Title:     title,
		Message:   "New chapter available",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"chapter": chapter,
		},
	}
}

// ToJSON converts notification to JSON bytes
func (n *Notification) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

// SubscribeRequest represents a subscription request from client
type SubscribeRequest struct {
	Type   string `json:"type"` // "SUBSCRIBE" or "UNSUBSCRIBE"
	UserID string `json:"user_id"`
}

// ParseSubscribeRequest parses incoming subscription request
func ParseSubscribeRequest(data []byte) (*SubscribeRequest, error) {
	var req SubscribeRequest
	err := json.Unmarshal(data, &req)
	return &req, err
}
