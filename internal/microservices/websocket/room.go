package websocket

import (
	"log/slog"
	"sync"
)

// Room = 1 manga chat room for multiple clients
type Room struct {
	ID      int64              // manga ID
	Title   string             // manga title
	Clients map[string]*Client // map[clientID] -> *Client
	mu      sync.RWMutex       // mutex for concurrent access
}

// NewRoom creates a new chat Room
func NewRoom(id int64, title string) *Room {
	return &Room{
		ID:      id,
		Title:   title,
		Clients: make(map[string]*Client),
	}
}

// AddUser: adds new client to the room
func (r *Room) AddUser(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// check if client already exists
	if r.Clients[c.ID] == nil {
		slog.Info("Client added to room", "room_id", r.ID, "client_id", c.ID)
		r.Clients[c.ID] = c
	} else {
		slog.Warn("Client already in room", "room_id", r.ID, "client_id", c.ID)
	}
}

// RemoveUser: removes client from the room
func (r *Room) RemoveUser(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// check if client exists
	if r.Clients[c.ID] != nil {
		slog.Info("Client removed from room", "room_id", r.ID, "client_id", c.ID)
		delete(r.Clients, c.ID)
	} else {
		slog.Warn("Client not found in room", "room_id", r.ID, "client_id", c.ID)
	}
}

// Broadcast: broadcasts message to all clients in the room
func (r *Room) Broadcast(message []byte) {}

// GetUserCount: returns the number of clients in the room
func (r *Room) GetUserCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Clients)
}

// GetClients: returns copy of clients list in the room
func (r *Room) GetClients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// create a tmp slice to hold clients
	clients := make([]*Client, 0, len(r.Clients))
	for _, client := range r.Clients {
		clients = append(clients, client)
	}
	return clients
}
