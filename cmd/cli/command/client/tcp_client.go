package client

// tcp_client.go = handles TCP client functionality for the mangahubCLI application.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// TCPClient represents a TCP client for real-time sync
type TCPClient struct {
	serverAddr string
	conn       net.Conn
	connected  bool
	sessionID  string
	stats      ConnectionStats
	mu         sync.RWMutex
	stopChan   chan struct{}
	monitoring bool
}

// ConnectionStats holds connection statistics
type ConnectionStats struct {
	Uptime           time.Duration
	LastHeartbeat    time.Time
	MessagesSent     int
	MessagesReceived int
	LastSync         time.Time
	Conflicts        int
	ConnectedAt      time.Time
}

// SyncMessage represents a sync update message
type SyncMessage struct {
	Direction       string    // "incoming", "outgoing", "conflict"
	Device          string    // Device name
	MangaTitle      string    // Manga title
	Chapter         int       // Chapter number
	ConflictChapter int       // Conflicting chapter (for conflict type)
	Timestamp       time.Time // Message timestamp
}

// TCPMessage represents the wire protocol message
type TCPMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewTCPClient creates a new TCP client
func NewTCPClient(serverAddr string) *TCPClient {
	return &TCPClient{
		serverAddr: serverAddr,
		stopChan:   make(chan struct{}),
		stats: ConnectionStats{
			ConnectedAt: time.Now(),
		},
	}
}

// Connect establishes connection to TCP server
func (c *TCPClient) Connect(username, token string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Establish TCP connection
	conn, err := net.DialTimeout("tcp", c.serverAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	c.conn = conn
	c.connected = true
	c.stats.ConnectedAt = time.Now()

	// Send authentication message
	authMsg := TCPMessage{
		Type: "auth",
		Data: map[string]any{
			"username": username,
			"token":    token,
		},
		Timestamp: time.Now(),
	}

	if err := c.sendMessage(&authMsg); err != nil {
		c.conn.Close()
		c.connected = false
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Wait for auth response
	response, err := c.readMessage()
	if err != nil {
		c.conn.Close()
		c.connected = false
		return fmt.Errorf("authentication response failed: %w", err)
	}

	if response.Type == "auth_success" {
		if sessionID, ok := response.Data["session_id"].(string); ok {
			c.sessionID = sessionID
		}

		// Start heartbeat
		go c.heartbeatLoop()

		return nil
	}

	c.conn.Close()
	c.connected = false
	return fmt.Errorf("authentication rejected")
}

// Disconnect closes the connection
func (c *TCPClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// Stop monitoring if active
	if c.monitoring {
		close(c.stopChan)
		c.monitoring = false
	}

	// Send disconnect message
	disconnectMsg := TCPMessage{
		Type:      "disconnect",
		Data:      map[string]interface{}{},
		Timestamp: time.Now(),
	}
	_ = c.sendMessage(&disconnectMsg)

	// Close connection
	if c.conn != nil {
		c.conn.Close()
	}

	c.connected = false
	return nil
}

// IsConnected returns connection status
func (c *TCPClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetSessionID returns the session ID
func (c *TCPClient) GetSessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}

// GetStats returns connection statistics
func (c *TCPClient) GetStats() ConnectionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	if c.connected {
		stats.Uptime = time.Since(c.stats.ConnectedAt)
	}
	return stats
}

// SendProgressUpdate sends a progress update to the server
func (c *TCPClient) SendProgressUpdate(mangaID int64, chapter int, status, userID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected")
	}

	msg := TCPMessage{
		Type: "progress_update",
		Data: map[string]any{
			"user_id":  userID,
			"manga_id": mangaID,
			"chapter":  chapter,
			"status":   status,
		},
		Timestamp: time.Now(),
	}

	err := c.sendMessage(&msg)
	if err == nil {
		c.stats.MessagesSent++
		c.stats.LastSync = time.Now()
	}
	return err
}

// StartMonitoring starts monitoring for real-time updates
func (c *TCPClient) StartMonitoring(msgChan chan<- SyncMessage) {
	c.mu.Lock()
	c.monitoring = true
	c.mu.Unlock()

	for {
		select {
		case <-c.stopChan:
			return
		default:
			// Read message with timeout
			c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			msg, err := c.readMessage()
			if err != nil {
				// Timeout or error, continue
				continue
			}

			c.mu.Lock()
			c.stats.MessagesReceived++
			c.mu.Unlock()

			// Parse and send to channel
			syncMsg := c.parseMessage(msg)
			if syncMsg != nil {
				select {
				case msgChan <- *syncMsg:
				default:
					// Channel full, skip
				}
			}
		}
	}
}

// StopMonitoring stops the monitoring loop
func (c *TCPClient) StopMonitoring() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.monitoring {
		close(c.stopChan)
		c.monitoring = false
		c.stopChan = make(chan struct{}) // Reset for next monitoring session
	}
}

// heartbeatLoop sends periodic heartbeats
func (c *TCPClient) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.mu.Lock()
			if !c.connected {
				c.mu.Unlock()
				return
			}

			heartbeat := TCPMessage{
				Type:      "heartbeat",
				Data:      map[string]interface{}{},
				Timestamp: time.Now(),
			}

			err := c.sendMessage(&heartbeat)
			if err == nil {
				c.stats.LastHeartbeat = time.Now()
			}
			c.mu.Unlock()
		}
	}
}

// sendMessage sends a message over the TCP connection
func (c *TCPClient) sendMessage(msg *TCPMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Add newline delimiter
	data = append(data, '\n')

	_, err = c.conn.Write(data)
	return err
}

// readMessage reads a message from the TCP connection
func (c *TCPClient) readMessage() (*TCPMessage, error) {
	reader := bufio.NewReader(c.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	var msg TCPMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// parseMessage converts TCP message to SyncMessage
func (c *TCPClient) parseMessage(msg *TCPMessage) *SyncMessage {
	switch msg.Type {
	case "progress_update":
		return &SyncMessage{
			Direction:  "incoming",
			Device:     getStringField(msg.Data, "device", "unknown"),
			MangaTitle: getStringField(msg.Data, "manga_title", "Unknown"),
			Chapter:    getIntField(msg.Data, "chapter", 0),
			Timestamp:  msg.Timestamp,
		}
	case "progress_broadcast":
		return &SyncMessage{
			Direction:  "outgoing",
			MangaTitle: getStringField(msg.Data, "manga_title", "Unknown"),
			Chapter:    getIntField(msg.Data, "chapter", 0),
			Timestamp:  msg.Timestamp,
		}
	case "conflict":
		return &SyncMessage{
			Direction:       "conflict",
			MangaTitle:      getStringField(msg.Data, "manga_title", "Unknown"),
			Chapter:         getIntField(msg.Data, "local_chapter", 0),
			ConflictChapter: getIntField(msg.Data, "remote_chapter", 0),
			Timestamp:       msg.Timestamp,
		}
	}
	return nil
}

// Helper functions
func getStringField(data map[string]interface{}, key, defaultVal string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultVal
}

func getIntField(data map[string]interface{}, key string, defaultVal int) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	if val, ok := data[key].(int); ok {
		return val
	}
	return defaultVal
}
