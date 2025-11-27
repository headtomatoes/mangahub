package websocket

import (
	"time"

	"github.com/gorilla/websocket"
	//"log/slog"
)

// Individual client connection handler
// for easy tracking purpose: id room = id manga

const ( // ping pong(2-way heartbeat) to keep connection alive
	WriteWait      = 10 * time.Second    // max time write a message to the peer
	PongWait       = 60 * time.Second    // max time to wait for pong from peer => no pong = no connection
	PingPeriod     = (PongWait * 9) / 10 // max time to send ping to peer => before pong wait expires || 90% of pong wait time => 10% to allow for any network delay or jitter
	MaxMessageSize = 512                 // maximum message size allowed from peer
)

type Client struct {
	ID          string          // unique client ID
	UserID      string          // user ID get from auth token(JWT.claims)
	UserName    string          // user name get from auth token(JWT.claims)
	RoomID      int64           // manga ID
	Conn        *websocket.Conn // WebSocket connection
	SendChannel chan []byte     // channel for outbound(chan <-) messages
	Hub         *Hub            // reference to the central Hub
}

// constructor new client
func NewClient(id, userID, userName string, roomID int64, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:          id,
		UserID:      userID,
		UserName:    userName,
		RoomID:      roomID,
		Conn:        conn,
		SendChannel: make(chan []byte, MaxMessageSize/2), // buffered channel to MaxMessageSize
		Hub:         hub,
	}
}

// ReadPump:
func (c *Client) ReadPump() {}

// WritePump:
func (c *Client) WritePump() {}

// SendMessage:
func (c *Client) SendMessage(message []byte) error {
	return nil
}

// Close:
func (c *Client) Close() error {
	return nil
}
