package tcp

import (
	"fmt"
	"log/slog"
	"sync"
)

// ProgressRepository interface for abstraction (supports both Redis-only and Hybrid)
type ProgressRepository interface {
	SaveProgress(data *ProgressData) error
	GetProgress(userID string, mangaID int64) (*ProgressData, error)
	GetUserProgress(userID string) ([]*ProgressData, error)
	DeleteProgress(userID string, mangaID int64) error
}

type ConnectionManager struct {
	clients map[string]*ClientConnection
	// map store all active client connections
	// key: client ID, value: ClientConnection pointer
	// use pointer for efficient access and modification
	mu           sync.RWMutex       // read-write mutex for concurrent access
	logger       *slog.Logger       // pointer to structured logger for logging events
	progressRepo ProgressRepository // pointer to progress repository (can be Redis or Hybrid)
}

// constructor for ConnectionManager
func NewConnectionManager(progressRepo ProgressRepository) *ConnectionManager {
	return &ConnectionManager{ // return a pointer to a new ConnectionManager to share across goroutines
		clients:      make(map[string]*ClientConnection), // initialize empty map
		logger:       slog.Default(),                     // Initialize with default logger which can be customized later
		progressRepo: progressRepo,                       // Set the progress repository
	}
}

// method to add a new connection
func (m *ConnectionManager) AddConnection(client *ClientConnection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client.ID] = client // add the new client connection to the map by its ID
	m.logger.Info("client_added",
		"client_id", client.ID,
	)
}

// method to remove a connection
func (m *ConnectionManager) RemoveConnection(client *ClientConnection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, client.ID) // remove the client connection from the map by its ID
	m.logger.Info("client_removed",
		"client_id", client.ID,
	)
}

// method to close all connections
func (m *ConnectionManager) CloseAllConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, client := range m.clients { // iterate over all connectedclients
		func() { // use closure to ensure each client is closed properly
			defer func() { // recover from panic during close
				if r := recover(); r != nil {
					m.logger.Error("panic_closing_client",
						"client_id", id,
						"panic", r,
					)
				}
			}()
			client.Close() // close the client connection
			m.logger.Info("client_connection_closed",
				"client_id", id,
			)
		}()
	}
	m.clients = make(map[string]*ClientConnection)
	// reset the map,for clearing all references
	// allowing garbage collection
}

func (m *ConnectionManager) BroadcastSystemMessage(text string) {
	msg := []byte(fmt.Sprintf(`{"type":"system","message":"%s"}`, text))
	// construct system message payload in JSON format in byte slice for network transmission
	m.Broadcast(msg, "")
}

// method to broadcast a message to all clients except sender
// senderID is the ID of the client that sent the message (to exclude from broadcast)
func (m *ConnectionManager) BroadcastUserMessage(text, senderID string) {
	msg := []byte(fmt.Sprintf(`{"type":"user","message":"%s"}`, text))
	m.Broadcast(msg, senderID)
}

// hold a read lock while i/o operations are performed
// => starvation risk for writers if slow clients exist
// fix by using read lock only to copy the map of clients
// then release lock before sending messages
func (m *ConnectionManager) Broadcast(msg []byte, senderID string) {
	m.mu.RLock() // use read lock because we are only reading from the map, by that
	clients := make([]*ClientConnection, 0, len(m.clients))
	for _, c := range m.clients {
		clients = append(clients, c)
	}
	m.mu.RUnlock()
	// release lock before performing i/o operations
	var wg sync.WaitGroup // wait group to wait for all send operations to complete
	for _, c := range clients {
		wg.Add(1)                           // increment wait group counter
		go func(client *ClientConnection) { // launch goroutine for each send operation
			defer wg.Done()
			if err := client.Send(msg); err != nil {
				m.logger.Warn("failed_to_send_broadcast",
					"client_id", client.ID,
					"error", err.Error(),
				)
			}
		}(c) // pass client as argument to avoid closure issues
	}
	wg.Wait() // wait for all send operations to complete
	// Send to each client without holding lock
}
