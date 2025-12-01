package websocket

import (
	"log/slog"
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
	RoomID      int64           // room that client is connected to (manga ID) => can develop to multiple rooms later || like multi devices or 1 connection per room
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

// ReadPump: continuously reads messages from the WebSocket connection
// handle read-side(deadlines, ping pong, close-conn) of the WebSocket connection
// run in its own goroutine (per connection)
func (c *Client) ReadPump() {
	// on exit unregister client and close connection
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	// set read limit for incoming messages
	c.Conn.SetReadLimit(MaxMessageSize)
	// set initial read deadline (pong wait)
	c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	// set pong handler to update read deadline on pong receipt
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	// main read loop
	for {
		// read message from WebSocket connection
		_, messageData, err := c.Conn.ReadMessage()
		if err != nil {
			// if unexpected close error(log and break loop)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Websocket error: unexpected close error", "error", err)
			}
			break
		}
		// process incoming message

		// parse message
		message, err := MessageFromJSON(messageData)
		if err != nil {
			slog.Error("Websocket error: invalid message format", "error", err)
			continue
		}

		// set client info to message
		message.UserID = c.UserID
		message.UserName = c.UserName
		message.RoomID = c.RoomID
		message.Timestamp = time.Now().UTC()

		// with diff type of message handle accordingly
		switch message.Type {
		case TypeJoin:
			// Update client's room ID from the message
			c.RoomID = message.RoomID
			c.Hub.JoinRoom <- &RoomActions{ // send join room action to hub
				RoomID: message.RoomID,
				Client: c,
			}
			slog.Info("Client joined room", "room_id", message.RoomID, "client_id", c.ID)
		case TypeLeave:
			// ensure client leaves the correct room
			message.RoomID = c.RoomID
			c.Hub.LeaveRoom <- &RoomActions{ // send leave room action to hub
				RoomID: message.RoomID,
				Client: c,
			}
			slog.Info("Client left room", "room_id", message.RoomID, "client_id", c.ID)
		case TypeChat:
			// get room ID from client.RoomID to ensure correct room
			message.RoomID = c.RoomID
			// broadcast chat message to room via hub
			c.Hub.Broadcast <- message
		case TypeTyping:
			// get room ID from client.RoomID to ensure correct room
			message.RoomID = c.RoomID
			// broadcast typing indicator to room via hub
			c.Hub.Broadcast <- message
		default:
			slog.Warn("Websocket warning: unknown message type", "type", message.Type)
		}
	}
}

// WritePump: continuously listens on an internal channel for outbound messages
// handle write-side(periodic ping, write-deadlines) of the WebSocket connection
// run in its own goroutine
func (c *Client) WritePump() {
	// set ticker for periodic ping messages
	ticker := time.NewTicker(PingPeriod)
	// setup defer to stop ticker and close connection on exit
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	// main write loop
	for {
		// wait for message to send or ping ticker
		select {
		// if message is sent and ok
		case message, ok := <-c.SendChannel:
			// set write deadline
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				// not ok = channel closed
				// send close message
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				slog.Info("Websocket info: send channel closed, closing connection", "client_id", c.ID)
				return
			}
			// write message to WebSocket connection
			// get next writer for text message
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				slog.Error("Websocket error: failed to get next writer", "error", err)
				return
			}

			// write message data to writer
			w.Write(message)
			//slog.Debug("Message written to WebSocket writer", "client_id", c.ID)
			// close writer to flush message
			if err := w.Close(); err != nil {
				slog.Error("Websocket error: failed to close writer", "error", err)
				return
			}
		case <-ticker.C:
			// send periodic ping
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Websocket error: failed to send ping message", "error", err)
				return
			}
		}
	}
}

// SendMessage: sends a message to the client's send channel
func (c *Client) SendMessage(message *Message) error {
	data, err := message.ToJSON()
	if err != nil {
		slog.Error("Failed to send message: unable to convert message to JSON", "error", err)
		return err
	}
	select {
	case c.SendChannel <- data:
		// message sent successfully
		slog.Debug("Message sent to client", "client_id", c.ID, "message_type", message.Type)
		return nil
	default:
		// send channel is full, unable to send message
		slog.Warn("Send channel full, unable to send message", "client_id", c.ID)
		return nil
	}
}

// Close: closes the client's WebSocket connection
func (c *Client) Close() error {
	return c.Conn.Close()
}
