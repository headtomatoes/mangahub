package websocket

import (
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
func (r *Room) AddUser(c *Client) {}

// RemoveUser: removes client from the room
func (r *Room) RemoveUser(c *Client) {}

// Broadcast: broadcasts message to all clients in the room
func (r *Room) Broadcast(message []byte) {}

// GetUserCount: returns the number of clients in the room
func (r *Room) GetUserCount() int {
	return 0
}

// GetClients: returns copy of clients list in the room
func (r *Room) GetClients() []*Client {
	return nil
}
