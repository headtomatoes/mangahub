package client

// udp_client.go = handles UDP client functionality for the mangahubCLI application.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// UDPClient represents a UDP client for real-time notifications
type UDPClient struct {
	serverAddr string
	conn       *net.UDPConn
	userID     string
	connected  bool
	mu         sync.RWMutex
	stopChan   chan struct{}
	stats      UDPStats
}

// UDPStats holds UDP connection statistics
type UDPStats struct {
	ConnectedAt           time.Time
	NotificationsReceived int
	LastNotification      time.Time
	LastPing              time.Time
	Uptime                time.Duration
}

// UDPNotification represents a notification message from the server
type UDPNotification struct {
	Type      string                 `json:"type"`
	MangaID   int64                  `json:"manga_id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// subscribeRequest for UDP server
type subscribeRequest struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
}

// NewUDPClient creates a new UDP client
func NewUDPClient(serverAddr string) *UDPClient {
	return &UDPClient{
		serverAddr: serverAddr,
		stopChan:   make(chan struct{}),
		stats: UDPStats{
			ConnectedAt: time.Now(),
		},
	}
}

// Connect establishes connection to UDP server and subscribes
func (c *UDPClient) Connect(userID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	// Dial UDP server
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to UDP server: %w", err)
	}

	c.conn = conn
	c.userID = userID
	c.connected = true
	c.stats.ConnectedAt = time.Now()

	// Send SUBSCRIBE message
	sub := subscribeRequest{
		Type:   "SUBSCRIBE",
		UserID: userID,
	}

	subBytes, err := json.Marshal(sub)
	if err != nil {
		c.conn.Close()
		c.connected = false
		return fmt.Errorf("failed to marshal subscribe request: %w", err)
	}

	if _, err := c.conn.Write(subBytes); err != nil {
		c.conn.Close()
		c.connected = false
		return fmt.Errorf("failed to send subscribe request: %w", err)
	}

	log.Printf("✓ Subscribed to notifications (User ID: %s)", userID)

	return nil
}

// StartListening starts listening for incoming notifications
func (c *UDPClient) StartListening() error {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return fmt.Errorf("not connected to UDP server")
	}
	c.mu.RUnlock()

	// Start ping routine to keep connection alive
	go c.pingRoutine()

	// Start listening routine
	go c.listenRoutine()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("\n📡 Listening for notifications... (Press Ctrl+C to stop)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	<-sigChan

	fmt.Println("\n\n🛑 Stopping notification listener...")
	return c.Disconnect()
}

// listenRoutine handles incoming UDP messages
func (c *UDPClient) listenRoutine() {
	buffer := make([]byte, 8192)

	for {
		select {
		case <-c.stopChan:
			return
		default:
			// Set read deadline to avoid blocking forever
			c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			n, err := c.conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout is expected, continue loop
				}
				log.Printf("Error reading UDP message: %v", err)
				continue
			}

			if n == 0 {
				continue
			}

			// Process the notification
			c.handleNotification(buffer[:n])
		}
	}
}

// handleNotification processes incoming notification messages
func (c *UDPClient) handleNotification(data []byte) {
	var notification UDPNotification
	if err := json.Unmarshal(data, &notification); err != nil {
		log.Printf("Failed to parse notification: %v", err)
		return
	}

	// Update stats
	c.mu.Lock()
	c.stats.NotificationsReceived++
	c.stats.LastNotification = time.Now()
	c.mu.Unlock()

	// Format and display notification
	c.displayNotification(&notification)
}

// displayNotification formats and prints a notification
func (c *UDPClient) displayNotification(n *UDPNotification) {
	fmt.Println("\n┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓")

	// Display based on notification type
	switch n.Type {
	case "SUBSCRIBE":
		fmt.Printf("  ✅ %s\n", n.Message)

	case "UNSUBSCRIBE":
		fmt.Printf("  👋 %s\n", n.Message)

	case "NEW_MANGA":
		fmt.Printf("  🆕 NEW MANGA ADDED!\n")
		fmt.Printf("  📖 Title: %s\n", n.Title)
		fmt.Printf("  🆔 Manga ID: %d\n", n.MangaID)
		fmt.Printf("  💬 %s\n", n.Message)

	case "NEW_CHAPTER":
		fmt.Printf("  📚 NEW CHAPTER AVAILABLE!\n")
		fmt.Printf("  📖 Manga: %s\n", n.Title)
		fmt.Printf("  🆔 Manga ID: %d\n", n.MangaID)
		if n.Data != nil {
			if chapter, ok := n.Data["chapter"]; ok {
				fmt.Printf("  📄 Chapter: %.0f\n", chapter)
			}
		}
		fmt.Printf("  💬 %s\n", n.Message)

	case "MANGA_UPDATE":
		fmt.Printf("  🔄 MANGA UPDATED!\n")
		fmt.Printf("  📖 Title: %s\n", n.Title)
		fmt.Printf("  💬 %s\n", n.Message)

	case "PONG":
		// Don't display pong messages
		return

	default:
		fmt.Printf("  📨 NOTIFICATION\n")
		fmt.Printf("  Type: %s\n", n.Type)
		if n.Title != "" {
			fmt.Printf("  Title: %s\n", n.Title)
		}
		fmt.Printf("  Message: %s\n", n.Message)
	}

	fmt.Printf("  ⏰ Time: %s\n", n.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Println("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛")
}

// pingRoutine sends periodic PING messages to keep connection alive
func (c *UDPClient) pingRoutine() {
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

			pingMsg := subscribeRequest{
				Type:   "PING",
				UserID: c.userID,
			}

			data, err := json.Marshal(pingMsg)
			if err != nil {
				c.mu.Unlock()
				continue
			}

			_, err = c.conn.Write(data)
			if err == nil {
				c.stats.LastPing = time.Now()
			}
			c.mu.Unlock()
		}
	}
}

// Disconnect closes the UDP connection
func (c *UDPClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// Send UNSUBSCRIBE message
	unsub := subscribeRequest{
		Type:   "UNSUBSCRIBE",
		UserID: c.userID,
	}

	unsubBytes, _ := json.Marshal(unsub)
	_, _ = c.conn.Write(unsubBytes)

	// Stop routines
	close(c.stopChan)

	// Close connection
	if c.conn != nil {
		c.conn.Close()
	}

	c.connected = false
	log.Println("✓ Disconnected from UDP server")

	return nil
}

// GetStats returns the current connection statistics
func (c *UDPClient) GetStats() UDPStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	if c.connected {
		stats.Uptime = time.Since(c.stats.ConnectedAt)
	}
	return stats
}

// IsConnected returns whether the client is connected
func (c *UDPClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// PrintStats displays connection statistics
func (c *UDPClient) PrintStats() {
	stats := c.GetStats()

	fmt.Println("\n📊 UDP Connection Statistics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Connected at:         %s\n", stats.ConnectedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Uptime:               %s\n", formatDuration(stats.Uptime))
	fmt.Printf("  Notifications received: %d\n", stats.NotificationsReceived)

	if !stats.LastNotification.IsZero() {
		fmt.Printf("  Last notification:    %s\n", stats.LastNotification.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("  Last notification:    N/A\n")
	}

	if !stats.LastPing.IsZero() {
		fmt.Printf("  Last ping:            %s\n", stats.LastPing.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// PrettyPrintJSON formats JSON for display
func PrettyPrintJSON(data []byte) string {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		return string(data)
	}
	return pretty.String()
}
