package tcp

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

// server entry point would go here

// server struct and methods
// when using pointers we need to define explicitly in the constructor, avoid nil pointer dereference
type TCPServer struct {
	Addr string
	// server address
	Manager *ConnectionManager
	// custom manager for handling connections
	// use * for shared access and efficient updates to the same manager instance
	// across multiple goroutines that handle connections
	AuthService *TCPAuthService
	// authentication service for validating JWT tokens
	quitChan chan struct{}
	// shutdown signal channel
	// when closed, all goroutines listening for this channel will initiate shutdown
	wg sync.WaitGroup
	// wait group for collection of goroutines to finish
	logger *slog.Logger
	// structured logger for logging server events
	batchWriterCtx    context.Context
	batchWriterCancel context.CancelFunc
	// context for batch writer lifecycle
}

// NewServer creates a TCP server with Redis-only storage (backward compatible)
func NewServer(addrTCP, addrRedis string) *TCPServer {
	logger := slog.Default()                             // Use default logger for now, can be customized later
	progressRepo, err := NewProgressRedisRepo(addrRedis) // create new progress repository
	if err != nil {
		logger.Error("failed_to_create_progress_repository", "error", err.Error())
		return nil
	}
	manager := NewConnectionManager(progressRepo) // create new connection manager
	manager.logger = logger                       // then we can set the logger of the manager struct to use the same logger

	return &TCPServer{
		Addr:     addrTCP,
		Manager:  manager,
		quitChan: make(chan struct{}),
		logger:   logger,
	}
}

// NewServerWithHybridStorage creates a TCP server with Redis + PostgreSQL hybrid storage
func NewServerWithHybridStorage(addrTCP, addrRedis string, db *sql.DB, jwtSecret string) *TCPServer {
	logger := slog.Default()

	// Create Redis repository
	redisRepo, err := NewProgressRedisRepo(addrRedis)
	if err != nil {
		logger.Error("failed_to_create_redis_repository", "error", err.Error())
		return nil
	}

	// Create PostgreSQL repository
	postgresRepo := NewProgressPostgresRepo(db)

	// Create hybrid repository
	hybridRepo := NewHybridProgressRepository(redisRepo, postgresRepo)

	// Create connection manager with hybrid repo
	manager := NewConnectionManager(hybridRepo)
	manager.logger = logger

	// Create authentication service
	authService := NewTCPAuthService(jwtSecret)

	// Create context for batch writer
	ctx, cancel := context.WithCancel(context.Background())

	// Start batch writer in background
	go hybridRepo.StartBatchWriter(ctx)

	logger.Info("server_created_with_hybrid_storage",
		"redis_addr", addrRedis,
		"postgres", "enabled",
		"auth", "enabled",
	)

	return &TCPServer{
		Addr:              addrTCP,
		Manager:           manager,
		AuthService:       authService,
		quitChan:          make(chan struct{}),
		logger:            logger,
		batchWriterCtx:    ctx,
		batchWriterCancel: cancel,
	}
}

// NewServerWithMockRedis creates a server without Redis for testing
func NewServerWithMockRedis(addrTCP string) *TCPServer {
	logger := slog.Default()
	progressRepo := &ProgressRedisRepo{
		client: nil, // nil client for testing - won't be used
		ctx:    nil,
	}
	manager := NewConnectionManager(progressRepo)
	manager.logger = logger

	return &TCPServer{
		Addr:     addrTCP,
		Manager:  manager,
		quitChan: make(chan struct{}),
		logger:   logger,
	}
}

// method to start the server
func (s *TCPServer) Start() error {
	// listen for incoming connections
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		s.logger.Error(
			"failed_to_start_server",
			"error", err.Error(),
		)
		return fmt.Errorf("failed to start TCP server, error: %v", err)
	}
	defer listener.Close()
	s.logger.Info("server_started",
		"addr", s.Addr,
	)
	// accept connections in a loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Error(
				"failed_to_accept_connection",
				"error", err.Error(),
			)
			continue
		}
		// add +1 to wait group for the new connection handler goroutine
		s.wg.Add(1)
		// use an anonymous function to handle the connection
		// for 1. encapsulating the connection handling logic
		// 2. passing the conn variable correctly to avoid closure issues
		// 3. accessing the server's wait group to signal when done
		go func(conn net.Conn) {
			defer s.wg.Done()
			s.handleConnection(conn)
		}(conn)
	}
}

// handle connections/lifecycle of single client connection
func (s *TCPServer) handleConnection(conn net.Conn) {
	client := NewClientConnection(conn, s.Manager) // create new client connection that wrap around manager

	// Authenticate client if AuthService is available
	if s.AuthService != nil {
		if !s.authenticateClient(client) {
			s.logger.Warn("authentication_failed", "client_id", client.ID, "remote_addr", conn.RemoteAddr().String())
			conn.Close()
			return
		}
		s.logger.Info("client_authenticated", "client_id", client.ID, "user_id", client.UserID)
	}

	s.Manager.AddConnection(client)    // register connection with manager
	client.Listen()                    // start listening for messages from client
	s.Manager.RemoveConnection(client) // unregister connection when Listen exits
}

// authenticateClient handles the authentication handshake
func (s *TCPServer) authenticateClient(client *ClientConnection) bool {
	// Set authentication deadline (10 seconds)
	client.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read initial auth message
	reader := bufio.NewReader(client.conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		s.logger.Error("auth_read_error", "error", err.Error())
		return false
	}

	// Parse auth message
	var authMsg Message
	if err := json.Unmarshal(line, &authMsg); err != nil {
		s.logger.Error("auth_parse_error", "error", err.Error())
		return false
	}

	// Check message type
	if authMsg.Type != "auth" {
		s.logger.Warn("expected_auth_message", "got", authMsg.Type)
		return false
	}

	// Extract token
	token, ok := authMsg.Data["token"].(string)
	if !ok {
		s.logger.Warn("missing_token_in_auth_message")
		return false
	}

	// Validate token
	userID, username, err := s.AuthService.ValidateToken(token)
	if err != nil {
		s.logger.Warn("token_validation_failed", "error", err.Error())
		// Send auth failure response
		client.Send([]byte(`{"type":"auth_response","success":false,"error":"invalid token"}`))
		return false
	}

	// Set user info on client
	client.UserID = userID
	client.Username = username
	client.Authenticated = true

	// Send auth success response
	client.Send([]byte(fmt.Sprintf(`{"type":"auth_response","success":true,"user_id":"%s","username":"%s"}`, userID, username)))

	// Reset deadline for normal operations
	client.conn.SetReadDeadline(time.Now().Add(MaxDeadlineDuration))

	return true
}

// stop the server
func (s *TCPServer) Stop() {
	close(s.quitChan)                                                         // signal all goroutines to shutdown
	s.Manager.BroadcastSystemMessage("Server is shutting down in 5 seconds.") // notify clients
	time.Sleep(5 * time.Second)                                               // wait for a moment to allow clients to process the shutdown message
	s.Manager.CloseAllConnections()                                           // close all active connections
	s.wg.Wait()                                                               // wait for all goroutines to finish

	// Stop batch writer if it's running (hybrid mode)
	if s.batchWriterCancel != nil {
		s.logger.Info("stopping_batch_writer")
		s.batchWriterCancel()
	}

	// Close progress repository
	// Try to cast to HybridProgressRepository first
	if hybridRepo, ok := s.Manager.progressRepo.(*HybridProgressRepository); ok {
		s.logger.Info("closing_hybrid_repository")
		if err := hybridRepo.Close(); err != nil {
			s.logger.Error("failed_to_close_hybrid_repo", "error", err.Error())
		}
	} else if redisRepo, ok := s.Manager.progressRepo.(*ProgressRedisRepo); ok {
		// Fallback to Redis-only repository
		s.logger.Info("closing_redis_repository")
		if err := redisRepo.Close(); err != nil {
			s.logger.Error("failed_to_close_redis", "error", err.Error())
		}
	}

	s.logger.Info("server_stopped")
}

// authenticate client prototype
// func (s *TCPServer) authenticateClient(c *ClientConnection) bool {
// 	// placeholder for authentication logic
// 	// e.g., read initial auth message, validate credentials, etc.
// 	return true // assume always successful for prototype
// }
