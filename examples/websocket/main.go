// WebSocket example demonstrating real-time communication
package main

import (
        "fmt"
        "log"
        "sync"
        "time"
        "github.com/grandpaej/smallapi"
)

// Message represents a WebSocket message
type Message struct {
        Type      string    `json:"type"`
        Content   string    `json:"content"`
        Username  string    `json:"username"`
        Timestamp time.Time `json:"timestamp"`
        Room      string    `json:"room,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
        ID       string
        Username string
        Room     string
        WS       *smallapi.WebSocket
        Send     chan Message
}

// Hub manages all WebSocket connections
type Hub struct {
        clients    map[string]*Client
        rooms      map[string]map[string]*Client
        register   chan *Client
        unregister chan *Client
        broadcast  chan Message
        mutex      sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
        return &Hub{
                clients:    make(map[string]*Client),
                rooms:      make(map[string]map[string]*Client),
                register:   make(chan *Client),
                unregister: make(chan *Client),
                broadcast:  make(chan Message),
        }
}

// Run starts the hub's main loop
func (h *Hub) Run() {
        for {
                select {
                case client := <-h.register:
                        h.registerClient(client)
                        
                case client := <-h.unregister:
                        h.unregisterClient(client)
                        
                case message := <-h.broadcast:
                        h.broadcastMessage(message)
                }
        }
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
        h.mutex.Lock()
        defer h.mutex.Unlock()
        
        h.clients[client.ID] = client
        
        // Add to room
        if h.rooms[client.Room] == nil {
                h.rooms[client.Room] = make(map[string]*Client)
        }
        h.rooms[client.Room][client.ID] = client
        
        log.Printf("Client %s (%s) joined room %s", client.ID, client.Username, client.Room)
        
        // Notify room about new user
        joinMessage := Message{
                Type:      "user_joined",
                Content:   fmt.Sprintf("%s joined the room", client.Username),
                Username:  "System",
                Timestamp: time.Now(),
                Room:      client.Room,
        }
        
        h.broadcastToRoom(joinMessage, client.Room, client.ID)
        
        // Send current room stats to the new user
        roomStats := h.getRoomStats(client.Room)
        statsMessage := Message{
                Type:      "room_stats",
                Content:   fmt.Sprintf("Room: %s, Users: %d", client.Room, roomStats["user_count"]),
                Username:  "System",
                Timestamp: time.Now(),
        }
        
        select {
        case client.Send <- statsMessage:
        default:
                close(client.Send)
                delete(h.clients, client.ID)
        }
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
        h.mutex.Lock()
        defer h.mutex.Unlock()
        
        if _, ok := h.clients[client.ID]; ok {
                delete(h.clients, client.ID)
                
                // Remove from room
                if room, exists := h.rooms[client.Room]; exists {
                        delete(room, client.ID)
                        
                        // Clean up empty rooms
                        if len(room) == 0 {
                                delete(h.rooms, client.Room)
                        }
                }
                
                close(client.Send)
                
                log.Printf("Client %s (%s) left room %s", client.ID, client.Username, client.Room)
                
                // Notify room about user leaving
                leaveMessage := Message{
                        Type:      "user_left",
                        Content:   fmt.Sprintf("%s left the room", client.Username),
                        Username:  "System",
                        Timestamp: time.Now(),
                        Room:      client.Room,
                }
                
                h.broadcastToRoom(leaveMessage, client.Room, "")
        }
}

// broadcastMessage sends a message to the appropriate recipients
func (h *Hub) broadcastMessage(message Message) {
        if message.Room != "" {
                h.broadcastToRoom(message, message.Room, "")
        } else {
                h.broadcastToAll(message)
        }
}

// broadcastToRoom sends a message to all clients in a specific room
func (h *Hub) broadcastToRoom(message Message, room string, excludeClientID string) {
        h.mutex.RLock()
        roomClients := h.rooms[room]
        h.mutex.RUnlock()
        
        for clientID, client := range roomClients {
                if clientID == excludeClientID {
                        continue
                }
                
                select {
                case client.Send <- message:
                default:
                        close(client.Send)
                        delete(h.clients, clientID)
                        delete(roomClients, clientID)
                }
        }
}

// broadcastToAll sends a message to all connected clients
func (h *Hub) broadcastToAll(message Message) {
        h.mutex.RLock()
        clients := make(map[string]*Client)
        for k, v := range h.clients {
                clients[k] = v
        }
        h.mutex.RUnlock()
        
        for clientID, client := range clients {
                select {
                case client.Send <- message:
                default:
                        close(client.Send)
                        delete(h.clients, clientID)
                }
        }
}

// getRoomStats returns statistics for a room
func (h *Hub) getRoomStats(room string) map[string]interface{} {
        h.mutex.RLock()
        defer h.mutex.RUnlock()
        
        roomClients := h.rooms[room]
        usernames := make([]string, 0, len(roomClients))
        
        for _, client := range roomClients {
                usernames = append(usernames, client.Username)
        }
        
        return map[string]interface{}{
                "user_count": len(roomClients),
                "users":      usernames,
        }
}

// handleClient handles a WebSocket client connection
func (h *Hub) handleClient(client *Client) {
        defer func() {
                h.unregister <- client
                client.WS.Close()
        }()
        
        // Start goroutine to send messages to client
        go func() {
                defer client.WS.Close()
                for message := range client.Send {
                        if err := client.WS.WriteJSON(message); err != nil {
                                log.Printf("Error writing message: %v", err)
                                return
                        }
                }
        }()
        
        // Read messages from client
        for {
                var message Message
                if err := client.WS.ReadJSON(&message); err != nil {
                        log.Printf("Error reading message: %v", err)
                        break
                }
                
                // Set metadata
                message.Username = client.Username
                message.Timestamp = time.Now()
                message.Room = client.Room
                
                // Handle different message types
                switch message.Type {
                case "chat":
                        h.broadcast <- message
                        
                case "ping":
                        pongMessage := Message{
                                Type:      "pong",
                                Content:   "Server is alive",
                                Username:  "System",
                                Timestamp: time.Now(),
                        }
                        select {
                        case client.Send <- pongMessage:
                        default:
                                return
                        }
                        
                case "room_stats":
                        stats := h.getRoomStats(client.Room)
                        statsMessage := Message{
                                Type:      "room_stats",
                                Content:   fmt.Sprintf("Room: %s, Users: %d", client.Room, stats["user_count"]),
                                Username:  "System",
                                Timestamp: time.Now(),
                        }
                        select {
                        case client.Send <- statsMessage:
                        default:
                                return
                        }
                        
                default:
                        // Echo unknown message types
                        h.broadcast <- message
                }
        }
}

func main() {
        app := smallapi.New()
        hub := NewHub()
        
        // Start the hub
        go hub.Run()
        
        // Middleware
        app.Use(smallapi.Logger())
        app.Use(smallapi.CORS())
        
        // Serve static files (HTML, CSS, JS for WebSocket client)
        app.Static("/", "./static")
        
        // WebSocket endpoint
        app.Get("/ws", func(c *smallapi.Context) {
                // Get connection parameters
                username := c.QueryDefault("username", "Anonymous")
                room := c.QueryDefault("room", "general")
                
                // Generate client ID
                clientID := fmt.Sprintf("%s_%d", username, time.Now().UnixNano())
                
                // Upgrade to WebSocket
                err := c.Upgrade(func(ws *smallapi.WebSocket) {
                        client := &Client{
                                ID:       clientID,
                                Username: username,
                                Room:     room,
                                WS:       ws,
                                Send:     make(chan Message, 256),
                        }
                        
                        // Register client
                        hub.register <- client
                        
                        // Handle client connection
                        hub.handleClient(client)
                })
                
                if err != nil {
                        log.Printf("WebSocket upgrade failed: %v", err)
                }
        })
        
        // REST API endpoints for chat management
        api := app.Group("/api")
        
        // Get room list
        api.Get("/rooms", func(c *smallapi.Context) {
                hub.mutex.RLock()
                rooms := make([]map[string]interface{}, 0, len(hub.rooms))
                for room, clients := range hub.rooms {
                        rooms = append(rooms, map[string]interface{}{
                                "name":       room,
                                "user_count": len(clients),
                        })
                }
                hub.mutex.RUnlock()
                
                c.JSON(map[string]interface{}{
                        "rooms": rooms,
                        "total": len(rooms),
                })
        })
        
        // Get room details
        api.Get("/rooms/:room", func(c *smallapi.Context) {
                room := c.Param("room")
                stats := hub.getRoomStats(room)
                
                c.JSON(map[string]interface{}{
                        "room":  room,
                        "stats": stats,
                })
        })
        
        // Broadcast message to room via REST API
        api.Post("/rooms/:room/message", func(c *smallapi.Context) {
                room := c.Param("room")
                
                var req struct {
                        Content  string `json:"content" validate:"required"`
                        Username string `json:"username"`
                }
                
                if err := c.JSON(&req); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": "Invalid JSON format",
                        })
                        return
                }
                
                if err := c.Validate(&req); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": err.Error(),
                        })
                        return
                }
                
                if req.Username == "" {
                        req.Username = "API"
                }
                
                message := Message{
                        Type:      "chat",
                        Content:   req.Content,
                        Username:  req.Username,
                        Timestamp: time.Now(),
                        Room:      room,
                }
                
                hub.broadcast <- message
                
                c.JSON(map[string]string{
                        "message": "Message sent successfully",
                })
        })
        
        // Server stats
        api.Get("/stats", func(c *smallapi.Context) {
                hub.mutex.RLock()
                totalClients := len(hub.clients)
                totalRooms := len(hub.rooms)
                hub.mutex.RUnlock()
                
                c.JSON(map[string]interface{}{
                        "total_clients": totalClients,
                        "total_rooms":   totalRooms,
                        "uptime":        time.Since(time.Now().Add(-1 * time.Hour)), // Simplified
                })
        })
        
        // Simple chat page (HTML)
        app.Get("/chat", func(c *smallapi.Context) {
                chatHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>SmallAPI WebSocket Chat</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
        #messages { border: 1px solid #ccc; height: 400px; overflow-y: scroll; padding: 10px; margin-bottom: 10px; }
        .message { margin-bottom: 10px; }
        .system { color: #666; font-style: italic; }
        .user { color: #333; }
        #input { width: 70%; padding: 10px; }
        #send { padding: 10px 20px; }
        #username, #room { margin: 5px; padding: 5px; }
    </style>
</head>
<body>
    <h1>SmallAPI WebSocket Chat</h1>
    <div>
        <input type="text" id="username" placeholder="Username" value="User1">
        <input type="text" id="room" placeholder="Room" value="general">
        <button onclick="connect()">Connect</button>
        <button onclick="disconnect()">Disconnect</button>
    </div>
    <div id="status">Disconnected</div>
    <div id="messages"></div>
    <div>
        <input type="text" id="input" placeholder="Type a message..." disabled>
        <button id="send" onclick="sendMessage()" disabled>Send</button>
    </div>

    <script>
        let ws = null;
        
        function connect() {
            const username = document.getElementById('username').value || 'Anonymous';
            const room = document.getElementById('room').value || 'general';
            
            ws = new WebSocket('ws://localhost:8080/ws?username=' + username + '&room=' + room);
            
            ws.onopen = function() {
                document.getElementById('status').textContent = 'Connected to room: ' + room;
                document.getElementById('input').disabled = false;
                document.getElementById('send').disabled = false;
            };
            
            ws.onmessage = function(event) {
                const message = JSON.parse(event.data);
                addMessage(message);
            };
            
            ws.onclose = function() {
                document.getElementById('status').textContent = 'Disconnected';
                document.getElementById('input').disabled = true;
                document.getElementById('send').disabled = true;
            };
        }
        
        function disconnect() {
            if (ws) {
                ws.close();
                ws = null;
            }
        }
        
        function sendMessage() {
            const input = document.getElementById('input');
            if (input.value && ws) {
                const message = {
                    type: 'chat',
                    content: input.value
                };
                ws.send(JSON.stringify(message));
                input.value = '';
            }
        }
        
        function addMessage(message) {
            const messages = document.getElementById('messages');
            const div = document.createElement('div');
            div.className = 'message ' + (message.username === 'System' ? 'system' : 'user');
            div.innerHTML = '<strong>' + message.username + ':</strong> ' + message.content + 
                           ' <small>(' + new Date(message.timestamp).toLocaleTimeString() + ')</small>';
            messages.appendChild(div);
            messages.scrollTop = messages.scrollHeight;
        }
        
        document.getElementById('input').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });
    </script>
</body>
</html>
        `
                c.HTML(chatHTML)
        })
        
        // Redirect root to chat
        app.Get("/", func(c *smallapi.Context) {
                c.Redirect("/chat")
        })
        
        app.Run(":8080")
}
