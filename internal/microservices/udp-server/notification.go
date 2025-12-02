package udp

import (
	"encoding/json"
	"fmt"
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
	Changes   []FieldChange    `json:"changes,omitempty"`
}

// FieldChange represents a specific field that was updated
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value"`
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
		Message:   fmt.Sprintf("New chapter %d available", chapter),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"chapter": chapter,
		},
		Changes: []FieldChange{
			{
				Field:    "current_chapter",
				NewValue: chapter,
			},
		},
	}
}

// NewMangaUpdateNotification creates a notification for manga update with changed fields
func NewMangaUpdateNotification(mangaID int64, title string, changes []string) *Notification {
	message := "Manga updated"
	if len(changes) > 0 {
		message = "Updated: " + joinChanges(changes)
	}

	return &Notification{
		Type:      NotificationMangaUpdate,
		MangaID:   mangaID,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"updated_fields": changes,
		},
	}
}

// NewMangaUpdateNotificationWithDetails creates a notification with detailed field changes
func NewMangaUpdateNotificationWithDetails(mangaID int64, title string, fieldChanges []FieldChange) *Notification {
	var changeNames []string
	for _, fc := range fieldChanges {
		changeNames = append(changeNames, fc.Field)
	}

	message := "Manga updated"
	if len(changeNames) > 0 {
		message = "Updated: " + joinChanges(changeNames)
	}

	return &Notification{
		Type:      NotificationMangaUpdate,
		MangaID:   mangaID,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"updated_fields": changeNames,
		},
		Changes: fieldChanges,
	}
}

// joinChanges joins change field names into a readable string
func joinChanges(changes []string) string {
	if len(changes) == 0 {
		return "unknown fields"
	}
	if len(changes) == 1 {
		return changes[0]
	}
	if len(changes) == 2 {
		return changes[0] + " and " + changes[1]
	}
	// More than 2: "field1, field2, and field3"
	last := changes[len(changes)-1]
	rest := changes[:len(changes)-1]
	result := ""
	for i, c := range rest {
		if i > 0 {
			result += ", "
		}
		result += c
	}
	return result + ", and " + last
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
