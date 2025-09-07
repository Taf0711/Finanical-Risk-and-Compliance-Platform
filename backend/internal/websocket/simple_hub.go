package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// SimpleHub manages WebSocket connections with a simpler interface
type SimpleHub struct {
	connections map[*websocket.Conn]bool
	broadcast   chan []byte
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
	mu          sync.RWMutex
}

// NewSimpleHub creates a new simple WebSocket hub
func NewSimpleHub() *SimpleHub {
	return &SimpleHub{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan []byte, 256),
		register:    make(chan *websocket.Conn, 256),
		unregister:  make(chan *websocket.Conn, 256),
	}
}

// Run starts the hub's main loop
func (h *SimpleHub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn] = true
			h.mu.Unlock()
			log.Printf("SimpleHub: WebSocket connection registered (%d total)", len(h.connections))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				h.mu.Unlock()
				log.Printf("SimpleHub: WebSocket connection unregistered (%d remaining)", len(h.connections))
			} else {
				h.mu.Unlock()
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			for conn := range h.connections {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("SimpleHub: Error writing to connection: %v", err)
					// Remove the connection if write fails
					delete(h.connections, conn)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterConnection adds a WebSocket connection to the hub
func (h *SimpleHub) RegisterConnection(conn *websocket.Conn) {
	h.register <- conn
}

// UnregisterConnection removes a WebSocket connection from the hub
func (h *SimpleHub) UnregisterConnection(conn *websocket.Conn) {
	h.unregister <- conn
}

// BroadcastToAll sends a message to all connected WebSocket clients
func (h *SimpleHub) BroadcastToAll(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case h.broadcast <- data:
	default:
		log.Println("SimpleHub: Broadcast channel full, dropping message")
	}

	return nil
}

// GetConnectionCount returns the number of active connections
func (h *SimpleHub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}
