package udp

import (
	"context"
	"net"
	"testing"
	"time"

	"mangahub/internal/microservices/http-api/models"
)

// Mock repositories for testing
type mockLibraryRepo struct {
	userIDs []string
	err     error
}

func (m *mockLibraryRepo) GetUserIDsByMangaID(ctx context.Context, mangaID int64) ([]string, error) {
	return m.userIDs, m.err
}

func (m *mockLibraryRepo) Add(ctx context.Context, userID string, mangaID int64) error {
	return nil
}

func (m *mockLibraryRepo) Remove(ctx context.Context, userID string, mangaID int64) error {
	return nil
}

func (m *mockLibraryRepo) List(ctx context.Context, userID string) ([]models.UserLibrary, error) {
	return nil, nil
}

func (m *mockLibraryRepo) Exists(ctx context.Context, userID string, mangaID int64) (bool, error) {
	return false, nil
}

type mockNotificationRepo struct {
	notifications []*models.Notification
	err           error
}

func (m *mockNotificationRepo) Create(ctx context.Context, notification *models.Notification) error {
	m.notifications = append(m.notifications, notification)
	return m.err
}

func (m *mockNotificationRepo) GetUnreadByUser(ctx context.Context, userID string) ([]models.Notification, error) {
	return nil, nil
}

func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, notificationID int64) error {
	return nil
}

func (m *mockNotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	return nil
}

// mockUserRepo implements the user repository interface used by broadcaster tests
type mockUserRepo struct {
	ids []string
	err error
}

func (m *mockUserRepo) GetAllIDs(ctx context.Context) ([]string, error) {
	return m.ids, m.err
}

// Implement other UserRepository methods as no-ops for tests
func (m *mockUserRepo) Create(user *models.User) error {
	return nil
}

func (m *mockUserRepo) FindByUsername(username string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindByID(id string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindByEmail(email string) (*models.User, error) {
	return nil, nil
}

func TestBroadcaster_BroadcastToAll(t *testing.T) {
	// Create a UDP connection for testing
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	// Create subscriber manager and add subscribers
	subManager := NewSubscriberManager(5 * time.Minute)
	clientAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	subManager.Add("user1", clientAddr)

	// Create broadcaster
	mockLibRepo := &mockLibraryRepo{}
	mockNotifRepo := &mockNotificationRepo{}
	mockUserRepo := &mockUserRepo{ids: []string{"user1"}}
	broadcaster := NewBroadcaster(conn, subManager, mockLibRepo, mockNotifRepo, mockUserRepo)

	// Test broadcast
	notification := NewMangaNotification(123, "Test Manga")
	err = broadcaster.BroadcastToAll(notification)

	if err != nil {
		t.Errorf("BroadcastToAll failed: %v", err)
	}
}

func TestBroadcaster_BroadcastToLibraryUsers(t *testing.T) {
	// Create a UDP connection for testing
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	// Create subscriber manager and add subscribers
	subManager := NewSubscriberManager(5 * time.Minute)
	clientAddr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	clientAddr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12346")
	subManager.Add("user1", clientAddr1)
	subManager.Add("user2", clientAddr2)

	// Create mock repos
	mockLibRepo := &mockLibraryRepo{
		userIDs: []string{"user1", "user2", "user3"}, // user3 is offline
	}
	mockNotifRepo := &mockNotificationRepo{
		notifications: make([]*models.Notification, 0),
	}

	// Create broadcaster
	mockUserRepo := &mockUserRepo{ids: []string{"user1", "user2", "user3"}}
	broadcaster := NewBroadcaster(conn, subManager, mockLibRepo, mockNotifRepo, mockUserRepo)

	// Test broadcast
	ctx := context.Background()
	notification := NewChapterNotification(123, "Test Manga", 5)
	err = broadcaster.BroadcastToLibraryUsers(ctx, 123, notification)

	if err != nil {
		t.Errorf("BroadcastToLibraryUsers failed: %v", err)
	}

	// Verify notifications were stored for all users
	if len(mockNotifRepo.notifications) != 3 {
		t.Errorf("Expected 3 notifications stored, got %d", len(mockNotifRepo.notifications))
	}

	// Verify user IDs
	userIDsStored := make(map[string]bool)
	for _, notif := range mockNotifRepo.notifications {
		userIDsStored[notif.UserID] = true
	}

	if !userIDsStored["user1"] || !userIDsStored["user2"] || !userIDsStored["user3"] {
		t.Error("Not all users received notification in database")
	}
}

func TestBroadcaster_BroadcastToLibraryUsers_NoUsers(t *testing.T) {
	// Create a UDP connection for testing
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	subManager := NewSubscriberManager(5 * time.Minute)

	// Create mock repos with no users
	mockLibRepo := &mockLibraryRepo{
		userIDs: []string{}, // No users
	}
	mockNotifRepo := &mockNotificationRepo{
		notifications: make([]*models.Notification, 0),
	}

	mockUserRepo := &mockUserRepo{ids: []string{}}
	broadcaster := NewBroadcaster(conn, subManager, mockLibRepo, mockNotifRepo, mockUserRepo)

	// Test broadcast
	ctx := context.Background()
	notification := NewChapterNotification(999, "Empty Manga", 1)
	err = broadcaster.BroadcastToLibraryUsers(ctx, 999, notification)

	// Should not error even with no users
	if err != nil {
		t.Errorf("BroadcastToLibraryUsers failed with no users: %v", err)
	}

	// No notifications should be stored
	if len(mockNotifRepo.notifications) != 0 {
		t.Errorf("Expected 0 notifications stored, got %d", len(mockNotifRepo.notifications))
	}
}
