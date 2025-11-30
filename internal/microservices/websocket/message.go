package websocket

import (
	"encoding/json"
	"log/slog"
	"time"
)

// Message protocol definitions

// Message types and structures
type MessageType string

const ( //trigger when +
	TypeJoin   MessageType = "join"   // user joins a room
	TypeLeave  MessageType = "leave"  // user leaves a room
	TypeChat   MessageType = "chat"   // user chat a message
	TypeSystem MessageType = "system" // system message
	TypeTyping MessageType = "typing" // user is typing indicator
)

// Message structure for WebSocket communication
type Message struct {
	Type      MessageType `json:"type"`      // type of message (join, leave, chat, system, typing)
	RoomID    int64       `json:"room_id"`   // manga ID 1:1 room
	UserID    string      `json:"user_id"`   // sender user ID
	UserName  string      `json:"user_name"` // sender user name
	Content   string      `json:"content"`   // message content
	Timestamp time.Time   `json:"timestamp"` // time in UTC format
}

// constructor new message
func NewMessage(msgType MessageType, roomID int64, userID, userName, content string) *Message {
	return &Message{
		Type:      msgType,
		RoomID:    roomID,
		UserID:    userID,
		UserName:  userName,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}
}

// specify the message for system
func NewSystemMessage(roomID int64, content string) *Message {
	return &Message{
		Type:      TypeSystem,
		RoomID:    roomID,
		UserID:    "system",
		UserName:  "System_Admin",
		Content:   content,
		Timestamp: time.Now().UTC(),
	}
}

// ToJSON: marshal Message struct to JSON
func (m *Message) ToJSON() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		slog.Error("Failed to marshal message to JSON", "error", err)
		return nil, err
	}
	return data, nil
}

// MessageFromJSON: unmarshal JSON data to Message struct
func MessageFromJSON(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		slog.Error("Failed to unmarshal message from JSON", "error", err)
		return nil, err
	}
	return &msg, nil
}
