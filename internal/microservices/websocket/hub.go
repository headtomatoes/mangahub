package websocket

import (
	"sync"
)

// Central hub managing all connections and rooms
// Each WebSocket connection runs in its own goroutine
// but they all communicate through channels to avoid race conditions.

// Hub maintains the set of active clients and rooms, sending messages to the clients.
type Hub struct {
	Clients    map[string]*Client // Registered clients
	Rooms      map[int64]*Room    // Active rooms mapped by manga ID
	Broadcast  chan *Message      // Inbound messages( <- channel) from the clients
	Register   chan *Client       // Register requests from the clients = join room request
	Unregister chan *Client       // Unregister requests from clients = leave room request
	JoinRoom   chan *RoomActions  // Join room action = happened
	LeaveRoom  chan *RoomActions  // Leave room action = happened
	mu         sync.RWMutex       // mutex for concurrent access
}

// RoomActions defines actions leave/join on rooms of specific clients
type RoomActions struct {
	RoomID int64
	Client *Client
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Rooms:      make(map[int64]*Room),
		Broadcast:  make(chan *Message, MaxMessageSize/2), // buffered channel to hold messages
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		JoinRoom:   make(chan *RoomActions),
		LeaveRoom:  make(chan *RoomActions),
	}
}

// Run: starts the Hub's main loop to process incoming channels
func (h *Hub) Run() {}

// RegisterClient: registers a new client
func (h *Hub) RegisterClient(c *Client) {}

// UnregisterClient: unregisters a client
func (h *Hub) UnregisterClient(c *Client) {}

// HandleJoinRoom: handles client joining a room
func (h *Hub) HandleJoinRoom(action *RoomActions) {}

// HandleLeaveRoom: handles client leaving a room
func (h *Hub) HandleLeaveRoom(action *RoomActions) {}

// BroadcastMessage: broadcasts a message to the appropriate room
func (h *Hub) BroadcastMessage(message *Message) {}

// GetRoom: retrieves or creates a room by ID
func (h *Hub) GetRoom(roomID int64) *Room
