package tcp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// we know that the messages are gonna be protobuf in the future => memory efficiency
// thus the expected message size is small and fixed
// apply the size limiting to the message

const MaxMessageSize = 1024 * 1024          // 1MB max message size
const MaxDeadlineDuration = 5 * time.Minute // 5min max read timeout duration

type ClientConnection struct {
	ID            string // unique identifier = key in map
	conn          net.Conn
	Writer        *bufio.Writer
	Manager       *ConnectionManager // reference to the connection manager for use the broadcast method
	Limiter       *rate.Limiter      // rate limiter for rate of sending messages
	UserID        string             // authenticated user ID (from JWT)
	Username      string             // authenticated username (from JWT)
	Authenticated bool               // whether the connection is authenticated
}

// constructor for Connection
func NewClientConnection(conn net.Conn, manager *ConnectionManager) *ClientConnection {
	return &ClientConnection{
		ID:      uuid.NewString(),
		conn:    conn,
		Writer:  bufio.NewWriter(conn),
		Manager: manager,
		Limiter: rate.NewLimiter(rate.Limit(10), 20), // 10 msgs/sec with burst of 20
		// the limiter auto depletes tokens when Allow is called and refills over time
	}
}

// method to listen for incoming data
func (c *ClientConnection) Listen() {
	defer c.conn.Close()              // close the connection
	reader := bufio.NewReader(c.conn) // buffered reader for efficient reading

	c.Manager.logger.Info("client_started_listening",
		"client_id", c.ID,
		"remote_addr", c.conn.RemoteAddr().String(),
	)
	// Set initial deadline for read operations
	c.conn.SetReadDeadline(time.Now().Add(MaxDeadlineDuration))

	for {
		// Read until newline delimiter (messages are newline-terminated)
		line, err := reader.ReadBytes('\n')
		if err != nil { //if error occurred during read, check the type
			if errors.Is(err, io.EOF) { // check for client disconnection or EOF signal
				c.Manager.logger.Info("client_disconnected",
					"client_id", c.ID,
				)
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() { // check for timeout error
				c.Manager.logger.Warn("client_read_timeout",
					"client_id", c.ID,
				)
				break
			}
			// Check for connection closed errors which are expected during shutdown
			// On Windows: "wsarecv: An established connection was aborted by the software in your host machine."
			//             "wsarecv: An existing connection was forcibly closed by the remote host."
			// On Linux: "use of closed network connection"
			if errors.Is(err, net.ErrClosed) ||
				strings.Contains(err.Error(), "closed network connection") ||
				strings.Contains(err.Error(), "connection was aborted") ||
				strings.Contains(err.Error(), "forcibly closed") {
				break // Exit silently for expected shutdown errors
			}
			c.Manager.logger.Error("client_read_error",
				"client_id", c.ID,
				"error", err,
			)
			continue
		}

		// reset deadline on successful read
		c.conn.SetReadDeadline(time.Now().Add(MaxDeadlineDuration))

		// Check message size (protect against oversized messages)
		if len(line) > MaxMessageSize {
			c.Manager.logger.Warn(
				"message_too_large",
				"client_id", c.ID,
				"size", len(line),
				"max_size", MaxMessageSize,
			)
			continue
		}

		// check rate limit
		if !c.Limiter.Allow() { // returns true if a token is available then consumes it
			c.Manager.logger.Warn(
				"rate_limit_exceeded",
				"client_id", c.ID,
			)
			// send json error message back to client
			c.Send([]byte(`{"type":"error","message":"Rate limit exceeded"}`))
			continue
		}

		// process the incoming message
		var msg Message                                    // custom struct to hold the incoming message
		if err := json.Unmarshal(line, &msg); err != nil { // parse JSON message into struct
			c.Manager.logger.Warn(
				"invalid_json_received",
				"client_id", c.ID,
				"error", err.Error(),
			)
			continue
		}

		// handle different message types
		switch msg.Type {
		case "progress_update":
			c.HandleProgressMessage(msg.Data)
		default:
			// Broadcast any valid JSON message (for flexibility and testing)
			c.Manager.logger.Info("broadcasting_message",
				"message_type", msg.Type,
				"client_id", c.ID,
			)
			payload, err := json.Marshal(map[string]any{
				"type":      msg.Type,
				"data":      msg.Data,
				"timestamp": time.Now().Unix(),
				"client_id": c.ID,
			})
			if err != nil {
				c.Manager.logger.Error("failed_to_marshal_broadcast_message",
					"client_id", c.ID,
					"error", err.Error(),
				)
				continue
			}
			c.Manager.Broadcast(payload)
		}
	}
}

// method to handle incoming messages in this case is the
// progress messages
func (c *ClientConnection) HandleProgressMessage(data map[string]any) {
	// Extract data
	userID, _ := data["user_id"].(string)
	mangaID, _ := data["manga_id"].(float64)
	chapter, _ := data["chapter"].(float64) // JSON numbers are float64

	// Save to repository (works with both Redis-only and Hybrid)
	if c.Manager.progressRepo != nil {
		progressData := &ProgressData{
			UserID:     userID,
			MangaID:    int64(mangaID),
			Chapter:    int64(chapter),
			LastReadAt: time.Now(),
			Status:     "reading",
		}

		if err := c.Manager.progressRepo.SaveProgress(progressData); err != nil {
			c.Manager.logger.Error("progress_save_failed",
				"client_id", c.ID,
				"user_id", userID,
				"error", err.Error(),
			)
			return
		}

		c.Manager.logger.Info("progress_saved",
			"client_id", c.ID,
			"user_id", userID,
			"manga_id", int64(mangaID),
			"chapter", int64(chapter),
		)
	}

	// Broadcast to other clients
	payload, _ := json.Marshal(map[string]any{
		"type":      "progress_broadcast",
		"data":      data,
		"timestamp": time.Now().Unix(),
	})
	c.Manager.Broadcast(payload)
}

// method to send data over the connection
func (c *ClientConnection) Send(data []byte) error {
	//=> data + "\n" then flush to the io.Writer buffer
	if _, err := c.Writer.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	if _, err := c.Writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}
	if err := c.Writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	return nil
}

// method to close the connection
func (c *ClientConnection) Close() {
	c.conn.Close()
}
