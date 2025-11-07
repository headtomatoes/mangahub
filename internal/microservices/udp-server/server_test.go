package udp

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"mangahub/internal/microservices/http-api/models"
)

func TestServer_Integration(t *testing.T) {
	// Create mock repos
	mockLibRepo := &mockLibraryRepo{
		userIDs: []string{"test-user-1", "test-user-2"},
	}
	mockNotifRepo := &mockNotificationRepo{
		notifications: make([]*models.Notification, 0),
	}

	// Create server on random port
	// mock user repo for server
	mockUsers1 := &mockUserRepo{ids: []string{"test-user-1", "test-user-2"}}

	server, err := NewServer("0", mockLibRepo, mockNotifRepo, mockUsers1)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	go func() {
		server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual port the server is listening on
	serverAddr := server.conn.LocalAddr().(*net.UDPAddr)

	t.Run("Subscribe and Receive Confirmation", func(t *testing.T) {
		// Create client
		clientConn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer clientConn.Close()

		// Send subscribe request
		subscribeReq := SubscribeRequest{
			Type:   "SUBSCRIBE",
			UserID: "test-user-1",
		}
		data, _ := json.Marshal(subscribeReq)
		_, err = clientConn.Write(data)
		if err != nil {
			t.Fatalf("Failed to send subscribe: %v", err)
		}

		// Read confirmation
		buffer := make([]byte, 4096)
		clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := clientConn.Read(buffer)
		if err != nil {
			t.Fatalf("Failed to read confirmation: %v", err)
		}

		var notification Notification
		if err := json.Unmarshal(buffer[:n], &notification); err != nil {
			t.Fatalf("Failed to parse confirmation: %v", err)
		}

		if notification.Type != NotificationSubscribe {
			t.Errorf("Expected SUBSCRIBE confirmation, got %s", notification.Type)
		}

		// Verify subscriber count
		if server.SubscriberCount() != 1 {
			t.Errorf("Expected 1 subscriber, got %d", server.SubscriberCount())
		}
	})

	t.Run("Receive Broadcast Notification", func(t *testing.T) {
		// Create client
		clientConn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer clientConn.Close()

		// Subscribe first
		subscribeReq := SubscribeRequest{
			Type:   "SUBSCRIBE",
			UserID: "test-user-2",
		}
		data, _ := json.Marshal(subscribeReq)
		clientConn.Write(data)

		// Read confirmation and discard
		buffer := make([]byte, 4096)
		clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		clientConn.Read(buffer)

		// Broadcast a notification
		go func() {
			time.Sleep(100 * time.Millisecond)
			server.NotifyNewManga(123, "Test Manga")
		}()

		// Read broadcast
		clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := clientConn.Read(buffer)
		if err != nil {
			t.Fatalf("Failed to read broadcast: %v", err)
		}

		var notification Notification
		if err := json.Unmarshal(buffer[:n], &notification); err != nil {
			t.Fatalf("Failed to parse broadcast: %v", err)
		}

		if notification.Type != NotificationNewManga {
			t.Errorf("Expected NEW_MANGA notification, got %s", notification.Type)
		}
		if notification.MangaID != 123 {
			t.Errorf("Expected MangaID 123, got %d", notification.MangaID)
		}
		if notification.Title != "Test Manga" {
			t.Errorf("Expected title 'Test Manga', got '%s'", notification.Title)
		}
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		// Create client
		clientConn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer clientConn.Close()

		// Subscribe first
		subscribeReq := SubscribeRequest{
			Type:   "SUBSCRIBE",
			UserID: "test-user-3",
		}
		data, _ := json.Marshal(subscribeReq)
		clientConn.Write(data)

		// Read confirmation
		buffer := make([]byte, 4096)
		clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		clientConn.Read(buffer)

		initialCount := server.SubscriberCount()

		// Unsubscribe
		unsubscribeReq := SubscribeRequest{
			Type:   "UNSUBSCRIBE",
			UserID: "test-user-3",
		}
		data, _ = json.Marshal(unsubscribeReq)
		clientConn.Write(data)

		// Read confirmation
		clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := clientConn.Read(buffer)
		if err != nil {
			t.Fatalf("Failed to read unsubscribe confirmation: %v", err)
		}

		var notification Notification
		if err := json.Unmarshal(buffer[:n], &notification); err != nil {
			t.Fatalf("Failed to parse unsubscribe confirmation: %v", err)
		}

		if notification.Type != NotificationUnsubscribe {
			t.Errorf("Expected UNSUBSCRIBE confirmation, got %s", notification.Type)
		}

		// Wait a bit for processing
		time.Sleep(100 * time.Millisecond)

		// Verify subscriber count decreased
		if server.SubscriberCount() >= initialCount {
			t.Errorf("Expected subscriber count to decrease from %d, got %d",
				initialCount, server.SubscriberCount())
		}
	})

	t.Run("Ping Pong", func(t *testing.T) {
		// Create client
		clientConn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer clientConn.Close()

		// Subscribe first
		subscribeReq := SubscribeRequest{
			Type:   "SUBSCRIBE",
			UserID: "test-user-ping",
		}
		data, _ := json.Marshal(subscribeReq)
		clientConn.Write(data)

		// Read confirmation
		buffer := make([]byte, 4096)
		clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		clientConn.Read(buffer)

		// Send ping
		pingReq := SubscribeRequest{
			Type:   "PING",
			UserID: "test-user-ping",
		}
		data, _ = json.Marshal(pingReq)
		clientConn.Write(data)

		// Read pong
		clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := clientConn.Read(buffer)
		if err != nil {
			t.Fatalf("Failed to read pong: %v", err)
		}

		response := string(buffer[:n])
		if response != `{"type":"PONG"}` {
			t.Errorf("Expected PONG response, got: %s", response)
		}
	})

	// Cleanup
	server.Shutdown()
}

func TestServer_NotifyNewChapter_StoresForOfflineUsers(t *testing.T) {
	// Create mock repos
	mockLibRepo := &mockLibraryRepo{
		userIDs: []string{"online-user", "offline-user1", "offline-user2"},
	}
	mockNotifRepo := &mockNotificationRepo{
		notifications: make([]*models.Notification, 0),
	}

	// Create server
	// mock user repo for server
	mockUsers2 := &mockUserRepo{ids: []string{"online-user", "offline-user1", "offline-user2"}}

	server, err := NewServer("0", mockLibRepo, mockNotifRepo, mockUsers2)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	go func() {
		server.Start()
	}()
	time.Sleep(100 * time.Millisecond)

	serverAddr := server.conn.LocalAddr().(*net.UDPAddr)

	// Subscribe only one user (online)
	clientConn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer clientConn.Close()

	subscribeReq := SubscribeRequest{
		Type:   "SUBSCRIBE",
		UserID: "online-user",
	}
	data, _ := json.Marshal(subscribeReq)
	clientConn.Write(data)

	// Read confirmation
	buffer := make([]byte, 4096)
	clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
	clientConn.Read(buffer)

	// Send new chapter notification
	ctx := context.Background()
	err = server.NotifyNewChapter(ctx, 123, "Test Manga", 10)
	if err != nil {
		t.Fatalf("NotifyNewChapter failed: %v", err)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify notifications were stored for ALL users (online + offline)
	if len(mockNotifRepo.notifications) != 3 {
		t.Errorf("Expected 3 notifications stored (1 online + 2 offline), got %d",
			len(mockNotifRepo.notifications))
	}

	// Verify notification details
	for _, notif := range mockNotifRepo.notifications {
		if notif.MangaID != 123 {
			t.Errorf("Expected MangaID 123, got %d", notif.MangaID)
		}
		if notif.Type != string(NotificationNewChapter) {
			t.Errorf("Expected type NEW_CHAPTER, got %s", notif.Type)
		}
		if notif.Read {
			t.Error("New notification should not be marked as read")
		}
	}

	// Cleanup
	server.Shutdown()
}
