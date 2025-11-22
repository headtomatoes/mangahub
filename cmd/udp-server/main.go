package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"mangahub/database"
	"mangahub/internal/microservices/http-api/repository"
	udp "mangahub/internal/microservices/udp-server"
)
 
func main() {
	// Load environment variables
	port := os.Getenv("UDP_PORT")
	if port == "" {
		port = "8082"
	}

	// Open GORM DB (same as API server)
	db, err := database.OpenGorm()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// Initialize repositories
	libraryRepo := repository.NewLibraryRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create and start UDP server
	server, err := udp.NewServer(port, libraryRepo, notificationRepo, userRepo)
	if err != nil {
		log.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start an optional HTTP trigger that allows other services to ask the UDP server
	// to broadcast notifications. The HTTP trigger listens on UDP_HTTP_PORT (default 8085).
	httpPort := os.Getenv("UDP_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8085"
	}

	go func() {
		mux := http.NewServeMux()

		// new manga
		mux.HandleFunc("/notify/new-manga", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var payload struct {
				MangaID int64  `json:"manga_id"`
				Title   string `json:"title"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := server.NotifyNewManga(payload.MangaID, payload.Title); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusAccepted)
		})

		// new chapter
		mux.HandleFunc("/notify/new-chapter", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var payload struct {
				MangaID int64  `json:"manga_id"`
				Title   string `json:"title"`
				Chapter int    `json:"chapter"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.NotifyNewChapter(ctx, payload.MangaID, payload.Title, payload.Chapter); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusAccepted)
		})

		// manga update (generic update) -> notify only library users
		mux.HandleFunc("/notify/manga-update", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var payload struct {
				MangaID int64  `json:"manga_id"`
				Title   string `json:"title"`
				Message string `json:"message"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// build a MANGA_UPDATE notification and broadcast to library users (stores unread)
			notif := &udp.Notification{
				Type:      udp.NotificationMangaUpdate,
				MangaID:   payload.MangaID,
				Title:     payload.Title,
				Message:   payload.Message,
				Timestamp: time.Now(),
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.GetBroadcaster().BroadcastToLibraryUsers(ctx, payload.MangaID, notif); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusAccepted)
		})

		addr := ":" + httpPort
		log.Printf("HTTP trigger for UDP server listening on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("HTTP trigger server error: %v", err)
		}
	}()

	log.Printf("Starting UDP notification server on port %s", port)
	if err := server.Start(); err != nil {
		log.Fatalf("UDP server error: %v", err)
	}
}
