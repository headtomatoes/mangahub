package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// loginRequest matches the server DTO for login
type loginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// authResponse matches the server auth response
type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	ExpiresIn    int64  `json:"expires_in"`
}

// subscribeRequest for UDP server
type subscribeRequest struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("username: ")
	username, _ := reader.ReadString('\n')
	username = trimNewline(username)

	fmt.Print("password: ")
	password, _ := reader.ReadString('\n')
	password = trimNewline(password)

	// Send login request to API server (email can be empty for login)
	loginReq := loginRequest{Username: username, Password: password}
	body, _ := json.Marshal(loginReq)

	resp, err := http.Post("http://localhost:8084/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("failed to call /auth/login: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("login failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var auth authResponse
	if err := json.Unmarshal(respBody, &auth); err != nil {
		log.Fatalf("failed to parse auth response: %v", err)
	}

	log.Printf("login success: user_id=%s username=%s", auth.UserID, auth.Username)

	// Connect to UDP server and subscribe
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8082")
	if err != nil {
		log.Fatalf("failed to resolve udp addr: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("failed to dial udp: %v", err)
	}
	defer conn.Close()

	sub := subscribeRequest{Type: "SUBSCRIBE", UserID: auth.UserID}
	subBytes, _ := json.Marshal(sub)

	if _, err := conn.Write(subBytes); err != nil {
		log.Fatalf("failed to send subscribe: %v", err)
	}
	log.Printf("sent SUBSCRIBE for user %s to %s", auth.UserID, udpAddr.String())

	// Read incoming notifications
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				log.Printf("udp read error: %v", err)
				return
			}
			if n == 0 {
				continue
			}

			// Parse and display notification with enhanced formatting
			var notification map[string]interface{}
			if err := json.Unmarshal(buf[:n], &notification); err != nil {
				fmt.Printf("received: %s\n", string(buf[:n]))
				continue
			}

			// Display notification with clear formatting
			displayNotification(notification)
		}
	}()

	// optional: send PING every 30s to keep server activity updated
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			_, _ = conn.Write([]byte(`{"type":"PING","user_id":"` + auth.UserID + `"}`))
		}
	}()

	<-stop
	// Send unsubscribe before exit
	unsub := subscribeRequest{Type: "UNSUBSCRIBE", UserID: auth.UserID}
	ub, _ := json.Marshal(unsub)
	_, _ = conn.Write(ub)
	log.Println("exiting")
}

func trimNewline(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	if len(s) > 0 && s[len(s)-1] == '\r' {
		s = s[:len(s)-1]
	}
	return s
}

// displayNotification formats and displays notification with enhanced details
func displayNotification(notif map[string]interface{}) {
	fmt.Println("\n" + strings.Repeat("=", 60))

	notifType, _ := notif["type"].(string)
	title, _ := notif["title"].(string)
	message, _ := notif["message"].(string)
	timestamp, _ := notif["timestamp"].(string)

	// Parse timestamp for better display
	var timeStr string
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		timeStr = t.Format("2006-01-02 15:04:05")
	} else {
		timeStr = timestamp
	}

	fmt.Printf("📢 %s\n", notifType)
	fmt.Printf("📚 Title: %s\n", title)
	fmt.Printf("💬 %s\n", message)
	fmt.Printf("🕐 Time: %s\n", timeStr)

	// Display changes if available
	if changes, ok := notif["changes"].([]interface{}); ok && len(changes) > 0 {
		fmt.Println("\n🔄 Changes:")
		for _, change := range changes {
			if changeMap, ok := change.(map[string]interface{}); ok {
				field, _ := changeMap["field"].(string)
				oldValue := changeMap["old_value"]
				newValue := changeMap["new_value"]

				if oldValue != nil {
					fmt.Printf("  • %s: %v → %v\n", field, oldValue, newValue)
				} else {
					fmt.Printf("  • %s: %v\n", field, newValue)
				}
			}
		}
	}

	// Display additional data if available
	if data, ok := notif["data"].(map[string]interface{}); ok && len(data) > 0 {
		if chapter, ok := data["chapter"].(float64); ok {
			fmt.Printf("\n📖 Chapter: %.0f\n", chapter)
		}
		if updatedFields, ok := data["updated_fields"].([]interface{}); ok && len(updatedFields) > 0 {
			fmt.Print("\n📝 Updated fields: ")
			fields := make([]string, 0, len(updatedFields))
			for _, f := range updatedFields {
				if fieldStr, ok := f.(string); ok {
					fields = append(fields, fieldStr)
				}
			}
			fmt.Println(strings.Join(fields, ", "))
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}
