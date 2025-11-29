package mangadex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Notifier sends notifications to the UDP notification server
type Notifier struct {
	udpServerURL string // http://localhost:8085 or http://udp-server:8085
	httpClient   *http.Client
}

// NewNotifier creates a new notifier instance
func NewNotifier(udpServerURL string) *Notifier {
	return &Notifier{
		udpServerURL: udpServerURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// NotifyNewManga sends notification for a newly discovered manga (async, non-blocking)
func (n *Notifier) NotifyNewManga(mangaID int64, title string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		payload := map[string]interface{}{
			"manga_id": mangaID,
			"title":    title,
		}

		if err := n.sendNotification(ctx, "/notify/new-manga", payload); err != nil {
			log.Printf("[Notifier] Failed to send new manga notification for '%s': %v", title, err)
		} else {
			log.Printf("[Notifier] ✅ Sent new manga notification: %s (ID: %d)", title, mangaID)
		}
	}()
}

// NotifyNewChapter sends notification for a new chapter (async, non-blocking)
func (n *Notifier) NotifyNewChapter(mangaID int64, title string, chapter int) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		payload := map[string]interface{}{
			"manga_id": mangaID,
			"title":    title,
			"chapter":  chapter,
		}

		if err := n.sendNotification(ctx, "/notify/new-chapter", payload); err != nil {
			log.Printf("[Notifier] Failed to send chapter notification for '%s' ch.%d: %v", title, chapter, err)
		} else {
			log.Printf("[Notifier] ✅ Sent new chapter notification: %s - Chapter %d (ID: %d)", title, chapter, mangaID)
		}
	}()
}

// NotifyMangaUpdate sends notification for manga metadata update (async, non-blocking)
func (n *Notifier) NotifyMangaUpdate(mangaID int64, title string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		payload := map[string]interface{}{
			"manga_id": mangaID,
			"title":    title,
		}

		if err := n.sendNotification(ctx, "/notify/manga-update", payload); err != nil {
			log.Printf("[Notifier] Failed to send manga update notification for '%s': %v", title, err)
		} else {
			log.Printf("[Notifier] ✅ Sent manga update notification: %s (ID: %d)", title, mangaID)
		}
	}()
}

// sendNotification sends HTTP POST request to UDP server
func (n *Notifier) sendNotification(ctx context.Context, endpoint string, payload map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := n.udpServerURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// NotifyBatch sends multiple notifications in batch (for efficiency)
func (n *Notifier) NotifyBatch(notifications []map[string]interface{}, endpoint string) {
	for _, payload := range notifications {
		// Extract fields for async call
		mangaID, _ := payload["manga_id"].(int64)
		title, _ := payload["title"].(string)

		if endpoint == "/notify/new-chapter" {
			chapter, _ := payload["chapter"].(int)
			n.NotifyNewChapter(mangaID, title, chapter)
		} else if endpoint == "/notify/new-manga" {
			n.NotifyNewManga(mangaID, title)
		}
	}
}
