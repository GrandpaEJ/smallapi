package smallapi

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// WebSocket represents a WebSocket connection
type WebSocket struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

// WebSocketHandler defines the WebSocket handler function signature
type WebSocketHandler func(*WebSocket)

// Upgrade upgrades an HTTP connection to WebSocket
func (c *Context) Upgrade(handler WebSocketHandler) error {
	// Check if it's a WebSocket upgrade request
	if c.Request.Header.Get("Connection") != "Upgrade" ||
		c.Request.Header.Get("Upgrade") != "websocket" {
		c.Status(400).JSON(map[string]string{
			"error": "Not a WebSocket upgrade request",
		})
		return fmt.Errorf("not a websocket upgrade request")
	}
	
	// Get the WebSocket key
	key := c.Request.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		c.Status(400).JSON(map[string]string{
			"error": "Missing Sec-WebSocket-Key header",
		})
		return fmt.Errorf("missing websocket key")
	}
	
	// Calculate the accept key
	acceptKey := calculateAcceptKey(key)
	
	// Hijack the connection
	hijacker, ok := c.Response.(http.Hijacker)
	if !ok {
		c.Status(500).JSON(map[string]string{
			"error": "WebSocket upgrade not supported",
		})
		return fmt.Errorf("websocket upgrade not supported")
	}
	
	conn, bufrw, err := hijacker.Hijack()
	if err != nil {
		c.Status(500).JSON(map[string]string{
			"error": "Failed to hijack connection",
		})
		return err
	}
	
	// Send the WebSocket handshake response
	response := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n"+
			"\r\n",
		acceptKey,
	)
	
	if _, err := bufrw.Write([]byte(response)); err != nil {
		conn.Close()
		return err
	}
	
	if err := bufrw.Flush(); err != nil {
		conn.Close()
		return err
	}
	
	// Create WebSocket wrapper
	ws := &WebSocket{
		conn:   conn,
		reader: bufrw.Reader,
		writer: bufrw.Writer,
	}
	
	// Handle the WebSocket connection
	go func() {
		defer conn.Close()
		handler(ws)
	}()
	
	return nil
}

// calculateAcceptKey calculates the WebSocket accept key
func calculateAcceptKey(key string) string {
	const websocketMagicString = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + websocketMagicString))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ReadMessage reads a message from the WebSocket connection
func (ws *WebSocket) ReadMessage() ([]byte, error) {
	// Simplified WebSocket frame reading
	// In production, you'd implement the full WebSocket protocol
	
	// Read frame header (simplified)
	header := make([]byte, 2)
	if _, err := ws.conn.Read(header); err != nil {
		return nil, err
	}
	
	// Extract payload length (simplified - only handles small payloads)
	payloadLen := int(header[1] & 0x7F)
	
	// Read mask key if present
	var maskKey []byte
	if header[1]&0x80 != 0 {
		maskKey = make([]byte, 4)
		if _, err := ws.conn.Read(maskKey); err != nil {
			return nil, err
		}
	}
	
	// Read payload
	payload := make([]byte, payloadLen)
	if _, err := ws.conn.Read(payload); err != nil {
		return nil, err
	}
	
	// Unmask payload if masked
	if maskKey != nil {
		for i := 0; i < len(payload); i++ {
			payload[i] ^= maskKey[i%4]
		}
	}
	
	return payload, nil
}

// WriteMessage writes a message to the WebSocket connection
func (ws *WebSocket) WriteMessage(data []byte) error {
	// Simplified WebSocket frame writing
	// In production, you'd implement the full WebSocket protocol
	
	// Create frame header
	var frame []byte
	
	// Fin bit + text opcode
	frame = append(frame, 0x81)
	
	// Payload length
	if len(data) < 126 {
		frame = append(frame, byte(len(data)))
	} else {
		// For larger payloads, you'd use extended length fields
		frame = append(frame, 126)
		frame = append(frame, byte(len(data)>>8))
		frame = append(frame, byte(len(data)&0xFF))
	}
	
	// Add payload
	frame = append(frame, data...)
	
	_, err := ws.conn.Write(frame)
	return err
}

// WriteJSON writes a JSON message to the WebSocket connection
func (ws *WebSocket) WriteJSON(v interface{}) error {
	data, err := jsonMarshal(v)
	if err != nil {
		return err
	}
	return ws.WriteMessage(data)
}

// ReadJSON reads a JSON message from the WebSocket connection
func (ws *WebSocket) ReadJSON(v interface{}) error {
	data, err := ws.ReadMessage()
	if err != nil {
		return err
	}
	return jsonUnmarshal(data, v)
}

// Close closes the WebSocket connection
func (ws *WebSocket) Close() error {
	return ws.conn.Close()
}

// WriteText writes a text message to the WebSocket connection
func (ws *WebSocket) WriteText(text string) error {
	return ws.WriteMessage([]byte(text))
}

// RemoteAddr returns the remote network address
func (ws *WebSocket) RemoteAddr() net.Addr {
	return ws.conn.RemoteAddr()
}

// LocalAddr returns the local network address
func (ws *WebSocket) LocalAddr() net.Addr {
	return ws.conn.LocalAddr()
}

// Simple JSON marshal/unmarshal functions (would use encoding/json in real implementation)
func jsonMarshal(v interface{}) ([]byte, error) {
	// Simplified JSON marshaling
	return []byte(fmt.Sprintf("%v", v)), nil
}

func jsonUnmarshal(data []byte, v interface{}) error {
	// Simplified JSON unmarshaling
	str := string(data)
	str = strings.TrimSpace(str)
	
	// This is a very basic implementation
	// In production, you'd use encoding/json
	switch ptr := v.(type) {
	case *string:
		*ptr = str
	case *map[string]interface{}:
		// Very basic JSON object parsing
		*ptr = map[string]interface{}{"data": str}
	}
	
	return nil
}
