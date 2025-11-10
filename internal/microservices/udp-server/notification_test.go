package udp

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewMangaNotification(t *testing.T) {
	notification := NewMangaNotification(123, "Test Manga")

	if notification.Type != NotificationNewManga {
		t.Errorf("Expected type %s, got %s", NotificationNewManga, notification.Type)
	}
	if notification.MangaID != 123 {
		t.Errorf("Expected MangaID 123, got %d", notification.MangaID)
	}
	if notification.Title != "Test Manga" {
		t.Errorf("Expected title 'Test Manga', got '%s'", notification.Title)
	}
	if notification.Message != "New manga added: Test Manga" {
		t.Errorf("Unexpected message: %s", notification.Message)
	}
}

func TestNewChapterNotification(t *testing.T) {
	notification := NewChapterNotification(456, "Test Manga", 10)

	if notification.Type != NotificationNewChapter {
		t.Errorf("Expected type %s, got %s", NotificationNewChapter, notification.Type)
	}
	if notification.MangaID != 456 {
		t.Errorf("Expected MangaID 456, got %d", notification.MangaID)
	}
	if notification.Title != "Test Manga" {
		t.Errorf("Expected title 'Test Manga', got '%s'", notification.Title)
	}

	// Check data
	data, ok := notification.Data.(map[string]interface{})
	if !ok {
		t.Error("Expected Data to be a map")
	}
	if chapter, ok := data["chapter"]; !ok || chapter != 10 {
		t.Errorf("Expected chapter 10 in data, got %v", data)
	}
}

func TestNotification_ToJSON(t *testing.T) {
	notification := &Notification{
		Type:      NotificationNewManga,
		MangaID:   789,
		Title:     "Test Manga",
		Message:   "Test Message",
		Timestamp: time.Now(),
	}

	data, err := notification.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	// Parse back
	var parsed Notification
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed.Type != notification.Type {
		t.Errorf("Type mismatch: expected %s, got %s", notification.Type, parsed.Type)
	}
	if parsed.MangaID != notification.MangaID {
		t.Errorf("MangaID mismatch: expected %d, got %d", notification.MangaID, parsed.MangaID)
	}
	if parsed.Title != notification.Title {
		t.Errorf("Title mismatch: expected %s, got %s", notification.Title, parsed.Title)
	}
}

func TestParseSubscribeRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantReq *SubscribeRequest
	}{
		{
			name:    "Valid SUBSCRIBE",
			input:   `{"type":"SUBSCRIBE","user_id":"user123"}`,
			wantErr: false,
			wantReq: &SubscribeRequest{Type: "SUBSCRIBE", UserID: "user123"},
		},
		{
			name:    "Valid UNSUBSCRIBE",
			input:   `{"type":"UNSUBSCRIBE","user_id":"user456"}`,
			wantErr: false,
			wantReq: &SubscribeRequest{Type: "UNSUBSCRIBE", UserID: "user456"},
		},
		{
			name:    "Invalid JSON",
			input:   `{invalid json}`,
			wantErr: true,
			wantReq: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseSubscribeRequest([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSubscribeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if req.Type != tt.wantReq.Type {
					t.Errorf("Type = %v, want %v", req.Type, tt.wantReq.Type)
				}
				if req.UserID != tt.wantReq.UserID {
					t.Errorf("UserID = %v, want %v", req.UserID, tt.wantReq.UserID)
				}
			}
		})
	}
}
