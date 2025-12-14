package anilist

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
    udpServerURL string
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
func (n *Notifier) NotifyNewManga(mangaID int64, titleEN *string, titleRomaji *string) {
    go func() {
        title := "Unknown Title"
        if titleEN != nil && *titleEN != "" {
            title = *titleEN
        } else if titleRomaji != nil && *titleRomaji != "" {
            title = *titleRomaji
        }

        payload := map[string]interface{}{
            "type":     "new_manga",
            "manga_id": mangaID,
            "title":    title,
            "source":   "anilist",
        }

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := n.sendNotification(ctx, "/notify/new-manga", payload); err != nil {
            log.Printf("[AniListNotifier] Failed to send new manga notification: %v", err)
        }
    }()
}

// NotifyChapterUpdate sends notification for chapter count update
func (n *Notifier) NotifyChapterUpdate(mangaID int64, title string, oldChapters, newChapters int) {
    go func() {
        payload := map[string]interface{}{
            "type":         "chapter_update",
            "manga_id":     mangaID,
            "title":        title,
            "old_chapters": oldChapters,
            "new_chapters": newChapters,
            "source":       "anilist",
        }

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := n.sendNotification(ctx, "/notify/chapter-update", payload); err != nil {
            log.Printf("[AniListNotifier] Failed to send chapter update notification: %v", err)
        }
    }()
}

// NotifyMangaUpdate sends notification for manga metadata update
func (n *Notifier) NotifyMangaUpdate(mangaID int64, title string) {
    go func() {
        payload := map[string]interface{}{
            "type":     "manga_update",
            "manga_id": mangaID,
            "title":    title,
            "source":   "anilist",
        }

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := n.sendNotification(ctx, "/notify/manga-update", payload); err != nil {
            log.Printf("[AniListNotifier] Failed to send manga update notification: %v", err)
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