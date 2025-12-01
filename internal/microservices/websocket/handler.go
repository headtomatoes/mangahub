package websocket

import (
	"mangahub/internal/microservices/http-api/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// HTTP upgrade handler to WebSocket connections

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// allow all origins for development purpose; can restrict later
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSHandler: handle upgrade request from HTTP connection to WebSocket
func WSHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get user info from JWT middleware
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found"})
			return
		}

		// Get username from claims
		claims, claimsExists := c.Get("claims")
		userName := "Unknown"
		if claimsExists {
			if claimsData, ok := claims.(*service.Claims); ok {
				userName = claimsData.Username
			}
		}

		// upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
			return
		}

		// create new client
		client := NewClient(
			userID.(string), // unique client ID (we use userID for now)
			userID.(string), // user ID from JWT
			userName,        // user name from JWT
			NilRoomID,       // initially not in any room
			conn,            // WebSocket connection
			hub,             // reference to the central Hub
		)

		// register client to hub
		hub.Register <- client

		// start goroutines for read and write pumps
		go client.ReadPump()
		go client.WritePump()

	}
}
