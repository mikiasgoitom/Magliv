package http

import (
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

// We need to install the gorilla/websocket package.
// Run: go get github.com/gorilla/websocket

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    // In a production environment, you should check the origin.
    // For this demo, we can allow any origin.
    CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
    clients    map[*websocket.Conn]bool
    broadcast  chan []byte
    register   chan *websocket.Conn
    unregister chan *websocket.Conn
    mu         sync.Mutex
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
    return &Hub{
        broadcast:  make(chan []byte),
        register:   make(chan *websocket.Conn),
        unregister: make(chan *websocket.Conn),
        clients:    make(map[*websocket.Conn]bool),
    }
}

// Run starts the hub's event loop. It should be run in a separate goroutine.
func (h *Hub) Run() {
    for {
        select {
        case conn := <-h.register:
            h.mu.Lock()
            h.clients[conn] = true
            h.mu.Unlock()
            log.Println("[WS] Client registered")
        case conn := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[conn]; ok {
                delete(h.clients, conn)
                conn.Close()
                log.Println("[WS] Client unregistered")
            }
            h.mu.Unlock()
        case message := <-h.broadcast:
            h.mu.Lock()
            for conn := range h.clients {
                if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
                    log.Printf("[WS] Error writing to client: %v", err)
                    // The client might have disconnected, so we can unregister them.
                    go func(c *websocket.Conn) { h.unregister <- c }(conn)
                }
            }
            h.mu.Unlock()
        }
    }
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
    h.broadcast <- message
}

// ServeWs handles websocket requests from the peer.
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("[WS] Failed to upgrade connection: %v", err)
        return
    }

    // Register the new client.
    h.register <- conn

    // This is a blocking loop to keep the connection alive. When the client
    // disconnects, this loop will exit and the connection will be closed.
    // We don't need to read any messages from the client for this demo.
    for {
        if _, _, err := conn.ReadMessage(); err != nil {
            h.unregister <- conn
            break
        }
    }
}