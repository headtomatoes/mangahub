package udp

import (
	"net"
	"testing"
	"time"
)

func TestSubscriberManager_AddAndRemove(t *testing.T) {
	sm := NewSubscriberManager(5 * time.Minute)

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")

	// Test Add
	sm.Add("user1", addr)

	if sm.Count() != 1 {
		t.Errorf("Expected 1 subscriber, got %d", sm.Count())
	}

	// Test GetByUserID
	sub, exists := sm.GetByUserID("user1")
	if !exists {
		t.Error("Expected subscriber to exist")
	}
	if sub.UserID != "user1" {
		t.Errorf("Expected UserID 'user1', got '%s'", sub.UserID)
	}

	// Test Remove
	sm.Remove("user1")
	if sm.Count() != 0 {
		t.Errorf("Expected 0 subscribers after removal, got %d", sm.Count())
	}
}

func TestSubscriberManager_GetAll(t *testing.T) {
	sm := NewSubscriberManager(5 * time.Minute)

	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12346")

	sm.Add("user1", addr1)
	sm.Add("user2", addr2)

	subs := sm.GetAll()
	if len(subs) != 2 {
		t.Errorf("Expected 2 subscribers, got %d", len(subs))
	}
}

func TestSubscriberManager_GetByUserIDs(t *testing.T) {
	sm := NewSubscriberManager(5 * time.Minute)

	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12346")
	addr3, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12347")

	sm.Add("user1", addr1)
	sm.Add("user2", addr2)
	sm.Add("user3", addr3)

	// Get specific users
	userIDs := []string{"user1", "user3"}
	subs := sm.GetByUserIDs(userIDs)

	if len(subs) != 2 {
		t.Errorf("Expected 2 subscribers, got %d", len(subs))
	}

	// Verify correct users
	found := make(map[string]bool)
	for _, sub := range subs {
		found[sub.UserID] = true
	}

	if !found["user1"] || !found["user3"] {
		t.Error("Did not get expected users")
	}
	if found["user2"] {
		t.Error("Got unexpected user2")
	}
}

func TestSubscriberManager_CleanupInactive(t *testing.T) {
	sm := NewSubscriberManager(100 * time.Millisecond)

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	sm.Add("user1", addr)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	sm.CleanupInactive()

	if sm.Count() != 0 {
		t.Errorf("Expected 0 subscribers after cleanup, got %d", sm.Count())
	}
}

func TestSubscriberManager_UpdateActivity(t *testing.T) {
	sm := NewSubscriberManager(100 * time.Millisecond)

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	sm.Add("user1", addr)

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Update activity
	sm.UpdateActivity("user1")

	// Wait for original timeout
	time.Sleep(60 * time.Millisecond)

	// Should still be active because we updated
	sm.CleanupInactive()

	if sm.Count() != 1 {
		t.Errorf("Expected 1 subscriber after activity update, got %d", sm.Count())
	}
}

func TestSubscriberManager_Concurrent(t *testing.T) {
	sm := NewSubscriberManager(5 * time.Minute)

	// Test concurrent adds
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
			sm.Add(string(rune('a'+id)), addr)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if sm.Count() != 10 {
		t.Errorf("Expected 10 subscribers, got %d", sm.Count())
	}
}
