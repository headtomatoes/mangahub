package websocket

import (
	"fmt"
	"log/slog"
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
func (h *Hub) Run() {
	// infinite loop to listen on channels
	for {
		// use select case statement to listen on multiple channels then execute corresponding action
		select {
		case client := <-h.Register:
			h.RegisterClient(client)
		case client := <-h.Unregister:
			h.UnregisterClient(client)
		case action := <-h.JoinRoom:
			h.HandleJoinRoom(action)
		case action := <-h.LeaveRoom:
			h.HandleLeaveRoom(action)
		case message := <-h.Broadcast:
			h.BroadcastMessage(message)
			// shut down case can be added later once needed
		}
	}
}

// RegisterClient: registers a new client
func (h *Hub) RegisterClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// check if client already exists
	// if not, add to clients map
	if h.Clients[c.ID] == nil {
		h.Clients[c.ID] = c
		slog.Info("Client registered", "client_id", c.ID)
	} else {
		slog.Warn("Client already registered", "client_id", c.ID)
	}
}

// UnregisterClient: unregisters a client
func (h *Hub) UnregisterClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.Clients[c.ID]; exists {
		// remove client from any room they are in
		// check if client is in a room or not
		if c.RoomID != NilRoomID {
			// check if room exists before removing user from it
			if room, roomExists := h.Rooms[c.RoomID]; roomExists {
				room.RemoveUser(c)

				// notify others in the room that user has left
				sysMsg := NewSystemMessage(
					c.RoomID,
					fmt.Sprintf("[%s] has left the chat.", c.UserName))
				room.Broadcast(sysMsg)
			}
			// remove client from hub's client map
			delete(h.Clients, c.ID)
			// close client's send channel to free resources
			close(c.SendChannel)
			slog.Info("Client unregistered", "client_id", c.ID)
		} else {
			slog.Warn("Client not in any room during unregistration", "client_id", c.ID)
		}
	} else {
		slog.Warn("Client not found during unregistration", "client_id", c.ID)
	}
}

// HandleJoinRoom: handles client joining a room
func (h *Hub) HandleJoinRoom(action *RoomActions) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// check if room exists, if not create it
	room, exists := h.Rooms[action.RoomID]
	if !exists {
		// Auto-create room for the manga
		room = NewRoom(action.RoomID, fmt.Sprintf("Manga %d Chat", action.RoomID))
		h.Rooms[action.RoomID] = room
		slog.Info("Room auto-created", "room_id", action.RoomID)
	}
	
	// because client 1:1 room => remove client from previous room if any
	if action.Client.RoomID != NilRoomID && action.Client.RoomID != action.RoomID {
		// check if previous room exists
		if prevRoom, prevExists := h.Rooms[action.Client.RoomID]; prevExists {
			prevRoom.RemoveUser(action.Client)
		}
	}

	// add client to the room
	room.AddUser(action.Client)
	// notify the client that they have joined the room
	joinMsg := NewSystemMessage(
		action.RoomID,
		fmt.Sprintf("You have joined the chat room for manga ID %d. Users online: %d", action.RoomID, room.GetUserCount()))
	action.Client.SendMessage(joinMsg)
	// notify others in the room that user has joined
	sysMsg := NewSystemMessage(
		action.RoomID,
		fmt.Sprintf("[%s] has joined the chat.", action.Client.UserName))
	room.Broadcast(sysMsg)
}

// HandleLeaveRoom: handles client leaving a room
func (h *Hub) HandleLeaveRoom(action *RoomActions) {
	h.mu.Lock()
	// get the room
	room, exists := h.Rooms[action.RoomID]
	h.mu.Unlock()
	if !exists {
		slog.Warn("Room not found for leaving", "room_id", action.RoomID)
		sysMsg := NewSystemMessage(
			action.RoomID,
			"Room does not exist.")
		action.Client.SendMessage(sysMsg)
		return
	}
	// remove client from the room
	room.RemoveUser(action.Client)
	// notify the client that they have left the room
	leaveMsg := NewSystemMessage(
		action.RoomID,
		fmt.Sprintf("You have left the chat room for manga ID %d.", action.RoomID))
	action.Client.SendMessage(leaveMsg)
	// notify others in the room that user has left
	sysMsg := NewSystemMessage(
		action.RoomID,
		fmt.Sprintf("[%s] has left the chat.", action.Client.UserName))
	room.Broadcast(sysMsg)
}

// BroadcastMessage: broadcasts a message to the appropriate room
func (h *Hub) BroadcastMessage(message *Message) {
	h.mu.RLock()
	// check if room that message is sent to exists
	room, exists := h.Rooms[message.RoomID]
	h.mu.RUnlock()
	if !exists {
		slog.Warn("Room not found for broadcasting message", "room_id", message.RoomID)
	}
	// log the broadcast action
	slog.Info("Broadcasting message",
		slog.Int64("room_id", message.RoomID),
		slog.String("user_id", message.UserID),
		slog.String("user_name", message.UserName),
		slog.String("message", message.Content))

	// broadcast to room
	room.Broadcast(message)
}

// GetRoom: retrieves a room by ID
func (h *Hub) GetRoom(roomID int64) (*Room, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// check if room exists
	room, exists := h.Rooms[roomID]
	if exists {
		return room, nil
	}
	// if not return error and nil
	return nil, fmt.Errorf("room with ID %d not found", roomID)
}
