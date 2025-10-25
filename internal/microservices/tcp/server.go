package tcp

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// server entry point would go here

// server struct and methods
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
}

// constructor for Server
func NewServer(addr string) *TCPServer {
	return &TCPServer{
		Addr:     addr,
		Manager:  NewConnectionManager(),
		quitChan: make(chan struct{}),
	}
}

// method to start the server
func (s *TCPServer) Start() error {
	// listen for incoming connections
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("failed to start TCP server, error: %v", err)
	}
	defer listener.Close()
	fmt.Printf("TCP server started on %s\n", s.Addr)

	// accept connections in a loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("failed to accept connection, error: %v\n", err)
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
	s.Manager.AddConnection(client)                // register connection with manager
	client.Listen()                                // start listening for messages from client
	s.Manager.RemoveConnection(client)             // unregister connection on exit
}

// stop the server
func (s *TCPServer) Stop() {
	close(s.quitChan)                                                         // signal all goroutines to shutdown
	s.Manager.BroadcastSystemMessage("Server is shutting down in 5 seconds.") // notify clients
	time.Sleep(5 * time.Second)                                               // wait for a moment to allow clients to process the shutdown message
	s.Manager.CloseAllConnections()                                           // close all active connections
	s.wg.Wait()                                                               // wait for all goroutines to finish
}
