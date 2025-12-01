package client

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

// ws_client.go = handles WebSocket client functionality for the mangahubCLI application.
func JoinChatRoom(roomID, token string) error {
	// Parse roomID to int64
	roomIDInt, err := strconv.ParseInt(roomID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid room ID: %w", err)
	}

	// Build WebSocket URL
	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:8084", // Adjust host/port as needed
		Path:   "/ws",
	}

	// Connect with auth header
	header := http.Header{}
	header.Add("Authorization", "Bearer "+token)

	fmt.Printf("\n🔌 Connecting to chat room %s...\n", roomID)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	fmt.Printf("✅ Connected! Type your messages (or /quit to exit)\n\n")

	// Send join message
	joinMsg := map[string]any{
		"type":    "join",
		"room_id": roomIDInt,
	}
	if err := conn.WriteJSON(joinMsg); err != nil {
		return err
	}

	// Channel for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Goroutine to receive messages
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var msg map[string]any
			if err := conn.ReadJSON(&msg); err != nil {
				log.Println("Read error:", err)
				return
			}

			// Pretty print message
			PrintMessage(msg)
		}
	}()

	// Goroutine to send messages
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if text == "/quit" {
				interrupt <- os.Interrupt
				return
			}

			chatMsg := map[string]any{
				"type":    "chat",
				"content": text,
			}
			if err := conn.WriteJSON(chatMsg); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}()

	// Wait for interrupt
	<-interrupt
	log.Println("Closing connection...")

	// Send leave message
	leaveMsg := map[string]any{
		"type":    "leave",
		"room_id": roomIDInt,
	}
	conn.WriteJSON(leaveMsg)

	return nil
}

func PrintMessage(msg map[string]any) {
	msgType := msg["type"].(string)

	switch msgType {
	case "system":
		color.Yellow("🔔 %s", msg["content"])

	case "chat":
		username := msg["username"].(string)
		content := msg["content"].(string)
		color.Cyan("[%s] %s", username, content)

	case "typing":
		username := msg["username"].(string)
		color.HiBlack("%s is typing...", username)
	}
}
