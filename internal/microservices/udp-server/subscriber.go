package udp

import (
	"net"
	"sync"
	"time"
)

// Subscriber represents a connected client
type Subscriber struct {
	UserID   string
	Addr     *net.UDPAddr
	LastSeen time.Time
	Active   bool
}

// SubscriberManager manages all subscribers
type SubscriberManager struct {
	mu          sync.RWMutex
	subscribers map[string]*Subscriber // userID -> Subscriber
	timeout     time.Duration
}

// NewSubscriberManager creates a new subscriber manager
func NewSubscriberManager(timeout time.Duration) *SubscriberManager {
	return &SubscriberManager{
		subscribers: make(map[string]*Subscriber),
		timeout:     timeout,
	}
}

// Add adds or updates a subscriber
func (sm *SubscriberManager) Add(userID string, addr *net.UDPAddr) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.subscribers[userID] = &Subscriber{
		UserID:   userID,
		Addr:     addr,
		LastSeen: time.Now(),
		Active:   true,
	}
}

// Remove removes a subscriber
func (sm *SubscriberManager) Remove(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.subscribers, userID)
}

// UpdateActivity updates the last seen time for a subscriber
func (sm *SubscriberManager) UpdateActivity(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sub, exists := sm.subscribers[userID]; exists {
		sub.LastSeen = time.Now()
		sub.Active = true
	}
}

// GetAll returns all active subscribers
func (sm *SubscriberManager) GetAll() []*Subscriber {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subs := make([]*Subscriber, 0, len(sm.subscribers))
	for _, sub := range sm.subscribers {
		if sub.Active {
			subs = append(subs, sub)
		}
	}
	return subs
}

// GetByUserID returns a subscriber by user ID
func (sm *SubscriberManager) GetByUserID(userID string) (*Subscriber, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sub, exists := sm.subscribers[userID]
	return sub, exists
}

// GetByUserIDs returns subscribers for specific user IDs
func (sm *SubscriberManager) GetByUserIDs(userIDs []string) []*Subscriber {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subs := make([]*Subscriber, 0, len(userIDs))
	for _, userID := range userIDs {
		if sub, exists := sm.subscribers[userID]; exists && sub.Active {
			subs = append(subs, sub)
		}
	}
	return subs
}

// CleanupInactive removes inactive subscribers
func (sm *SubscriberManager) CleanupInactive() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for userID, sub := range sm.subscribers {
		if now.Sub(sub.LastSeen) > sm.timeout {
			delete(sm.subscribers, userID)
		}
	}
}

// Count returns the number of active subscribers
func (sm *SubscriberManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.subscribers)
}

// StartCleanupRoutine starts a goroutine to periodically cleanup inactive subscribers
func (sm *SubscriberManager) StartCleanupRoutine(interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.CleanupInactive()
		case <-done:
			return
		}
	}
}
