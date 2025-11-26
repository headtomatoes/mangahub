package websocket

// Central hub managing all connections and rooms
// Each WebSocket connection runs in its own goroutine
// but they all communicate through channels to avoid race conditions.
