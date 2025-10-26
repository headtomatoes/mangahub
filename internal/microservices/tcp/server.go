package tcp

import (
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
	quitChan chan struct{}
	// shutdown signal channel
	// when closed, all goroutines listening for this channel will initiate shutdown
	wg sync.WaitGroup
	// wait group for collection of goroutines to finish
	logger *slog.Logger
	// structured logger for logging server events
}

// constructor for Server
func NewServer(addr string) *TCPServer {
	logger := slog.Default()          // Use default logger for now, can be customized later
	manager := NewConnectionManager() // create new connection manager
	manager.logger = logger           // then we can set the logger of the manager struct to use the same logger

	return &TCPServer{
		Addr:     addr,
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
	client := NewClientConnection(conn, s.Manager) // create new client connection that wrap around manager//
	// auth process prototype maybe use the http api auth 1st the proceed to tcp
	// if !s.authenticateClient(client) {
	// 	conn.Close()
	// 	return
	// }
	s.Manager.AddConnection(client)    // register connection with manager
	client.Listen()                    // start listening for messages from client
	s.Manager.RemoveConnection(client) // unregister connection when Listen exits
}

// stop the server
func (s *TCPServer) Stop() {
	close(s.quitChan)                                                         // signal all goroutines to shutdown
	s.Manager.BroadcastSystemMessage("Server is shutting down in 5 seconds.") // notify clients
	time.Sleep(5 * time.Second)                                               // wait for a moment to allow clients to process the shutdown message
	s.Manager.CloseAllConnections()                                           // close all active connections
	s.wg.Wait()                                                               // wait for all goroutines to finish
}

// authenticate client prototype
// func (s *TCPServer) authenticateClient(c *ClientConnection) bool {
// 	// placeholder for authentication logic
// 	// e.g., read initial auth message, validate credentials, etc.
// 	return true // assume always successful for prototype
// }
