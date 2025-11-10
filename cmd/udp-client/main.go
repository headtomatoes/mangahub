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

	fmt.Print("email: ")
	email, _ := reader.ReadString('\n')
	email = trimNewline(email)

	fmt.Print("password: ")
	password, _ := reader.ReadString('\n')
	password = trimNewline(password)

	// Send login request to API server
	loginReq := loginRequest{Username: username, Email: email, Password: password}
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

			var pretty bytes.Buffer
			if err := json.Indent(&pretty, buf[:n], "", "  "); err != nil {
				fmt.Printf("received: %s\n", string(buf[:n]))
			} else {
				fmt.Printf("notification:\n%s\n", pretty.String())
			}
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
