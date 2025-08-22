package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// SimpleHub manages Fiber WebSocket connections
type SimpleHub struct {
	connections map[*websocket.Conn]bool
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
	broadcast   chan []byte
	mu          sync.RWMutex
}

// NewSimpleHub creates a new simple WebSocket hub
func NewSimpleHub() *SimpleHub {
	return &SimpleHub{
		connections: make(map[*websocket.Conn]bool),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
		broadcast:   make(chan []byte, 256),
	}
}

// Run starts the hub
func (h *SimpleHub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn] = true
			h.mu.Unlock()
			log.Printf("WebSocket client registered, total: %d", len(h.connections))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Printf("WebSocket client unregistered, total: %d", len(h.connections))

		case message := <-h.broadcast:
			h.mu.RLock()
			for conn := range h.connections {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Error writing to WebSocket client: %v", err)
					// Remove failed connection
					delete(h.connections, conn)
					conn.Close()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterConnection registers a WebSocket connection
func (h *SimpleHub) RegisterConnection(conn *websocket.Conn) {
	h.register <- conn
}

// UnregisterConnection unregisters a WebSocket connection
func (h *SimpleHub) UnregisterConnection(conn *websocket.Conn) {
	h.unregister <- conn
}

// BroadcastToAll broadcasts a message to all connected clients
func (h *SimpleHub) BroadcastToAll(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case h.broadcast <- data:
		return nil
	default:
		log.Println("Warning: Broadcast channel full, dropping message")
		return nil
	}
}
