package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

type ClientConnection struct {
	ID      string // unique identifier = key in map
	conn    net.Conn
	Writer  *bufio.Writer
	Manager *ConnectionManager // reference to the connection manager for use the broadcast method
}

// constructor for Connection
func NewClientConnection(conn net.Conn, manager *ConnectionManager) *ClientConnection {
	return &ClientConnection{
		ID:      uuid.NewString(),
		conn:    conn,
		Writer:  bufio.NewWriter(conn),
		Manager: manager,
	}
}

// method to listen for incoming data
func (c *ClientConnection) Listen() {
	defer c.conn.Close()
	reader := bufio.NewReader(c.conn) // buffered reader for efficient reading
	for {
		line, err := reader.ReadBytes('\n') // read until newline character as message delimiter
		if err != nil {
			fmt.Println("Error reading from client:", err)
			break
		}

		var msg Message                                    // custom struct to hold the incoming message
		if err := json.Unmarshal(line, &msg); err != nil { // parse JSON message into struct
			fmt.Println("Invalid JSON:", err)
			continue
		}

		// handle different message types
		switch msg.Type {
		case "progress_update":
			c.HandleProgressMessage(msg.Data)
		default:
			fmt.Printf("Unknown message type from client %s: %s\n", c.ID, msg.Type)
		}
	}
}

// method to handle incoming messages in this case is the
// progress messages
func (c *ClientConnection) HandleProgressMessage(data map[string]interface{}) {
	payload, _ := json.Marshal(map[string]interface{}{ // construct broadcast message payload
		"type":      "progress_broadcast",
		"data":      data,
		"timestamp": time.Now().Unix(),
	})
	c.Manager.Broadcast(payload)
}

// method to send data over the connection
func (c *ClientConnection) Send(data []byte) {
	c.Writer.Write(data)         // write data to the buffer
	c.Writer.Write([]byte("\n")) // append newline as message delimiter as the end of message
	c.Writer.Flush()             // flush = forces the buffered data to be sent immediately
}

// method to close the connection
func (c *ClientConnection) Close() {
	c.conn.Close()
}
