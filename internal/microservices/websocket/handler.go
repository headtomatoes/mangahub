package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// HTTP upgrade handler to WebSocket connections

var upgrader = websocket.Upgrader{}

// WSHandler: handle upgrade request from HTTP connection to WebSocket
func WSHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {}
}
