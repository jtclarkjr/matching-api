package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"matching-api/internal/database"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (configure appropriately for production)
		return true
	},
}

// Connection represents a WebSocket connection
type Connection struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub maintains active WebSocket connections
type Hub struct {
	connections map[string][]*Connection
	broadcast   chan []byte
	register    chan *Connection
	unregister  chan *Connection
	mutex       sync.RWMutex
}

var hub = &Hub{
	connections: make(map[string][]*Connection),
	broadcast:   make(chan []byte),
	register:    make(chan *Connection),
	unregister:  make(chan *Connection),
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	ChatID  string      `json:"chat_id,omitempty"`
	Message interface{} `json:"message,omitempty"`
	UserID  string      `json:"user_id,omitempty"`
}

func init() {
	go hub.run()
}

// run handles hub operations
func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mutex.Lock()
			h.connections[conn.UserID] = append(h.connections[conn.UserID], conn)
			h.mutex.Unlock()
			log.Printf("User %s connected via WebSocket", conn.UserID)

		case conn := <-h.unregister:
			h.mutex.Lock()
			if connections, ok := h.connections[conn.UserID]; ok {
				for i, c := range connections {
					if c == conn {
						h.connections[conn.UserID] = append(connections[:i], connections[i+1:]...)
						close(c.Send)
						break
					}
				}
				if len(h.connections[conn.UserID]) == 0 {
					delete(h.connections, conn.UserID)
				}
			}
			h.mutex.Unlock()
			log.Printf("User %s disconnected from WebSocket", conn.UserID)

		case message := <-h.broadcast:
			var wsMsg WebSocketMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				continue
			}

			// Find users who should receive this message
			if wsMsg.ChatID != "" {
				h.broadcastToChat(wsMsg.ChatID, message)
			}
		}
	}
}

// broadcastToChat sends a message to all users in a specific chat
func (h *Hub) broadcastToChat(chatID string, message []byte) {
	// Get users in this chat from database
	if database.DB == nil {
		return
	}

	query := `
		SELECT m.user1_id, m.user2_id 
		FROM chats c
		JOIN matches m ON c.match_id = m.id
		WHERE c.id = $1 AND m.is_active = true
	`

	var user1ID, user2ID string
	err := database.DB.QueryRow(query, chatID).Scan(&user1ID, &user2ID)
	if err != nil {
		log.Printf("Error querying chat users: %v", err)
		return
	}

	// Send to both users
	h.sendToUser(user1ID, message)
	h.sendToUser(user2ID, message)
}

// sendToUser sends a message to a specific user
func (h *Hub) sendToUser(userID string, message []byte) {
	h.mutex.RLock()
	connections, ok := h.connections[userID]
	h.mutex.RUnlock()

	if !ok {
		return
	}

	for _, conn := range connections {
		select {
		case conn.Send <- message:
		default:
			close(conn.Send)
		}
	}
}

// HandleWebSocket handles WebSocket connections for real-time chat
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogError("Failed to upgrade to WebSocket", err)
		return
	}

	// Create connection
	connection := &Connection{
		UserID: user.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	// Register connection
	hub.register <- connection

	// Start goroutines for reading and writing
	go connection.writePump()
	go connection.readPump()
}

// readPump handles incoming WebSocket messages
func (c *Connection) readPump() {
	defer func() {
		hub.unregister <- c
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message (e.g., typing indicators, message sending)
		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error unmarshaling WebSocket message: %v", err)
			continue
		}

		// Process different message types
		switch wsMsg.Type {
		case "typing":
			// Broadcast typing indicator to other user in chat
			wsMsg.UserID = c.UserID
			if data, err := json.Marshal(wsMsg); err == nil {
				hub.broadcast <- data
			}
		case "ping":
			// Send pong response
			pong := WebSocketMessage{Type: "pong"}
			if data, err := json.Marshal(pong); err == nil {
				c.Send <- data
			}
		}
	}
}

// writePump handles outgoing WebSocket messages
func (c *Connection) writePump() {
	defer func() {
		if err := c.Conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			return
		}
	}
}

// BroadcastMessage sends a message to all users in a chat via WebSocket
func BroadcastMessage(chatID string, message *models.Message) {
	wsMsg := WebSocketMessage{
		Type:    "new_message",
		ChatID:  chatID,
		Message: message,
	}

	if data, err := json.Marshal(wsMsg); err == nil {
		hub.broadcast <- data
	}
}
