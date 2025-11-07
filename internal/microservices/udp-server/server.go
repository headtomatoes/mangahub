package udp

import (
	"context"
	"fmt"
	"log"
	"mangahub/internal/microservices/http-api/repository"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server represents the UDP notification server
type Server struct {
	conn             *net.UDPConn
	subManager       *SubscriberManager
	broadcaster      *Broadcaster
	notificationRepo repository.NotificationRepository
	done             chan struct{}
}

// NewServer creates a new UDP server
func NewServer(port string, libraryRepo repository.LibraryRepository, notificationRepo repository.NotificationRepository) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP: %w", err)
	}

	subManager := NewSubscriberManager(5 * time.Minute)
	broadcaster := NewBroadcaster(conn, subManager, libraryRepo, notificationRepo)

	return &Server{
		conn:             conn,
		subManager:       subManager,
		broadcaster:      broadcaster,
		notificationRepo: notificationRepo,
		done:             make(chan struct{}),
	}, nil
}

// Start starts the UDP server
func (s *Server) Start() error {
	log.Printf("UDP server listening on %s", s.conn.LocalAddr().String())

	// Start cleanup routine
	go s.subManager.StartCleanupRoutine(1*time.Minute, s.done)

	// Start listening for incoming messages
	go s.handleIncomingMessages()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down UDP server...")
	return s.Shutdown()
}

// handleIncomingMessages handles incoming UDP messages
func (s *Server) handleIncomingMessages() {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-s.done:
			return
		default:
			n, addr, err := s.conn.ReadFromUDP(buffer)
			if err != nil {
				log.Printf("Error reading UDP message: %v", err)
				continue
			}

			go s.processMessage(buffer[:n], addr)
		}
	}
}

// processMessage processes incoming messages from clients
func (s *Server) processMessage(data []byte, addr *net.UDPAddr) {
	req, err := ParseSubscribeRequest(data)
	if err != nil {
		log.Printf("Failed to parse message from %s: %v", addr.String(), err)
		return
	}

	switch req.Type {
	case "SUBSCRIBE":
		s.subManager.Add(req.UserID, addr)
		log.Printf("User %s subscribed from %s", req.UserID, addr.String())

		// Send confirmation
		confirmation := &Notification{
			Type:      NotificationSubscribe,
			Message:   "Successfully subscribed to notifications",
			Timestamp: time.Now(),
		}
		if data, err := confirmation.ToJSON(); err == nil {
			s.conn.WriteToUDP(data, addr)
		}

		// SYNC: Push missed notifications to reconnecting user
		go s.syncMissedNotifications(req.UserID, addr)

	case "UNSUBSCRIBE":
		s.subManager.Remove(req.UserID)
		log.Printf("User %s unsubscribed", req.UserID)

		// Send confirmation
		confirmation := &Notification{
			Type:      NotificationUnsubscribe,
			Message:   "Successfully unsubscribed",
			Timestamp: time.Now(),
		}
		if data, err := confirmation.ToJSON(); err == nil {
			s.conn.WriteToUDP(data, addr)
		}

	case "PING":
		s.subManager.UpdateActivity(req.UserID)
		// Send pong
		s.conn.WriteToUDP([]byte(`{"type":"PONG"}`), addr)

	default:
		log.Printf("Unknown message type from %s: %s", addr.String(), req.Type)
	}
}

// syncMissedNotifications retrieves and sends all unread notifications to a reconnecting user
func (s *Server) syncMissedNotifications(userID string, addr *net.UDPAddr) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get unread notifications from database
	unreadNotifs, err := s.notificationRepo.GetUnreadByUser(ctx, userID)
	if err != nil {
		log.Printf("Failed to fetch unread notifications for user %s: %v", userID, err)
		return
	}

	if len(unreadNotifs) == 0 {
		log.Printf("No missed notifications for user %s", userID)
		return
	}

	log.Printf("Syncing %d missed notifications to user %s", len(unreadNotifs), userID)

	// Send each unread notification via UDP
	for _, dbNotif := range unreadNotifs {
		notification := &Notification{
			Type:      NotificationType(dbNotif.Type),
			MangaID:   dbNotif.MangaID,
			Title:     dbNotif.Title,
			Message:   dbNotif.Message,
			Timestamp: dbNotif.CreatedAt,
		}

		data, err := notification.ToJSON()
		if err != nil {
			log.Printf("Failed to marshal notification %d: %v", dbNotif.ID, err)
			continue
		}

		if _, err := s.conn.WriteToUDP(data, addr); err != nil {
			log.Printf("Failed to send notification %d to %s: %v", dbNotif.ID, addr.String(), err)
		} else {
			log.Printf("Synced notification %d to user %s", dbNotif.ID, userID)
		}

		// Small delay to avoid overwhelming the client
		time.Sleep(50 * time.Millisecond)
	}

	log.Printf("Sync completed for user %s", userID)
}

// NotifyNewManga broadcasts notification for new manga to all users
func (s *Server) NotifyNewManga(mangaID int64, title string) error {
	notification := NewMangaNotification(mangaID, title)
	return s.broadcaster.BroadcastToAll(notification)
}

// NotifyNewChapter broadcasts notification for new chapter to library users
func (s *Server) NotifyNewChapter(ctx context.Context, mangaID int64, title string, chapter int) error {
	notification := NewChapterNotification(mangaID, title, chapter)
	return s.broadcaster.BroadcastToLibraryUsers(ctx, mangaID, notification)
}

// GetBroadcaster returns the broadcaster instance
func (s *Server) GetBroadcaster() *Broadcaster {
	return s.broadcaster
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	close(s.done)
	return s.conn.Close()
}

// SubscriberCount returns the number of active subscribers
func (s *Server) SubscriberCount() int {
	return s.subManager.Count()
}
