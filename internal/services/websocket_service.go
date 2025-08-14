package services

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// WebSocketService handles real-time messaging for high-performance communication
type WebSocketService struct {
	// Connection management
	connections map[string]*websocket.Conn
	connMutex   sync.RWMutex
	
	// Message broadcasting
	broadcast chan *BroadcastMessage
	
	// Connection limits for performance
	maxConnections int
	currentConns   int
	connCountMutex sync.RWMutex
}

// BroadcastMessage represents a message to be broadcast
type BroadcastMessage struct {
	DeviceID string      `json:"device_id"`
	Type     string      `json:"type"`
	Data     interface{} `json:"data"`
	Targets  []string    `json:"targets,omitempty"` // Specific device IDs to target
}

// WebSocketMessage represents incoming WebSocket messages
type WebSocketMessage struct {
	Type      string      `json:"type"`
	DeviceID  string      `json:"device_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewWebSocketService creates a new WebSocket service optimized for high concurrency
func NewWebSocketService(maxConnections int) *WebSocketService {
	ws := &WebSocketService{
		connections:    make(map[string]*websocket.Conn),
		broadcast:      make(chan *BroadcastMessage, 1000), // Buffered channel for performance
		maxConnections: maxConnections,
	}
	
	// Start the broadcast handler
	go ws.handleBroadcasts()
	
	return ws
}

// HandleWebSocket handles WebSocket connections with performance optimizations
func (ws *WebSocketService) HandleWebSocket(c *fiber.Ctx) error {
	// Check connection limit
	ws.connCountMutex.RLock()
	currentConns := ws.currentConns
	ws.connCountMutex.RUnlock()
	
	if currentConns >= ws.maxConnections {
		logrus.Warn("WebSocket connection limit reached")
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "Too many connections",
		})
	}
	
	return websocket.New(func(conn *websocket.Conn) {
		deviceID := c.Query("device_id")
		if deviceID == "" {
			logrus.Error("Device ID is required for WebSocket connection")
			conn.Close()
			return
		}
		
		// Register connection
		ws.registerConnection(deviceID, conn)
		defer ws.unregisterConnection(deviceID)
		
		logrus.WithField("device_id", deviceID).Info("WebSocket connection established")
		
		// Handle incoming messages
		for {
			var msg WebSocketMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				logrus.WithError(err).Debug("WebSocket read error")
				break
			}
			
			msg.DeviceID = deviceID
			msg.Timestamp = time.Now()
			
			// Process the message
			ws.handleIncomingMessage(&msg)
		}
	})(c)
}

// registerConnection adds a new WebSocket connection
func (ws *WebSocketService) registerConnection(deviceID string, conn *websocket.Conn) {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	
	// Close existing connection if any
	if existingConn, exists := ws.connections[deviceID]; exists {
		existingConn.Close()
	}
	
	ws.connections[deviceID] = conn
	
	// Update connection count
	ws.connCountMutex.Lock()
	ws.currentConns++
	ws.connCountMutex.Unlock()
}

// unregisterConnection removes a WebSocket connection
func (ws *WebSocketService) unregisterConnection(deviceID string) {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	
	if _, exists := ws.connections[deviceID]; exists {
		delete(ws.connections, deviceID)
		
		// Update connection count
		ws.connCountMutex.Lock()
		ws.currentConns--
		ws.connCountMutex.Unlock()
		
		logrus.WithField("device_id", deviceID).Info("WebSocket connection closed")
	}
}

// BroadcastMessage sends a message to specific devices or all connected devices
func (ws *WebSocketService) BroadcastMessage(msg *BroadcastMessage) {
	select {
	case ws.broadcast <- msg:
		// Message queued successfully
	default:
		// Channel is full, log warning
		logrus.Warn("Broadcast channel is full, dropping message")
	}
}

// handleBroadcasts processes broadcast messages in a separate goroutine
func (ws *WebSocketService) handleBroadcasts() {
	for msg := range ws.broadcast {
		ws.connMutex.RLock()
		
		if len(msg.Targets) > 0 {
			// Send to specific targets
			for _, deviceID := range msg.Targets {
				if conn, exists := ws.connections[deviceID]; exists {
					ws.sendToConnection(conn, msg, deviceID)
				}
			}
		} else {
			// Broadcast to all connections
			for deviceID, conn := range ws.connections {
				ws.sendToConnection(conn, msg, deviceID)
			}
		}
		
		ws.connMutex.RUnlock()
	}
}

// sendToConnection sends a message to a specific WebSocket connection
func (ws *WebSocketService) sendToConnection(conn *websocket.Conn, msg *BroadcastMessage, deviceID string) {
	// Set write deadline for performance
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	
	err := conn.WriteJSON(msg)
	if err != nil {
		logrus.WithError(err).WithField("device_id", deviceID).Error("Failed to send WebSocket message")
		// Remove the problematic connection
		go ws.unregisterConnection(deviceID)
	}
}

// handleIncomingMessage processes incoming WebSocket messages
func (ws *WebSocketService) handleIncomingMessage(msg *WebSocketMessage) {
	logrus.WithFields(logrus.Fields{
		"device_id": msg.DeviceID,
		"type":      msg.Type,
	}).Debug("Received WebSocket message")
	
	// Handle different message types
	switch msg.Type {
	case "ping":
		// Respond with pong for keepalive
		ws.BroadcastMessage(&BroadcastMessage{
			DeviceID: msg.DeviceID,
			Type:     "pong",
			Data:     map[string]interface{}{"timestamp": time.Now()},
			Targets:  []string{msg.DeviceID},
		})
		
	case "status_update":
		// Handle status updates
		logrus.WithField("device_id", msg.DeviceID).Info("Device status updated")
		
	case "typing":
		// Handle typing indicators
		// Could broadcast to other relevant connections
		
	default:
		logrus.WithField("type", msg.Type).Warn("Unknown WebSocket message type")
	}
}

// GetConnectionCount returns the current number of WebSocket connections
func (ws *WebSocketService) GetConnectionCount() int {
	ws.connCountMutex.RLock()
	defer ws.connCountMutex.RUnlock()
	return ws.currentConns
}

// IsDeviceConnected checks if a specific device is connected via WebSocket
func (ws *WebSocketService) IsDeviceConnected(deviceID string) bool {
	ws.connMutex.RLock()
	defer ws.connMutex.RUnlock()
	_, exists := ws.connections[deviceID]
	return exists
}

// SendToDevice sends a message to a specific device
func (ws *WebSocketService) SendToDevice(deviceID string, msgType string, data interface{}) {
	ws.BroadcastMessage(&BroadcastMessage{
		DeviceID: deviceID,
		Type:     msgType,
		Data:     data,
		Targets:  []string{deviceID},
	})
}

// GetStats returns WebSocket service statistics
func (ws *WebSocketService) GetStats() map[string]interface{} {
	ws.connCountMutex.RLock()
	currentConns := ws.currentConns
	ws.connCountMutex.RUnlock()
	
	return map[string]interface{}{
		"current_connections": currentConns,
		"max_connections":     ws.maxConnections,
		"broadcast_queue_size": len(ws.broadcast),
		"broadcast_queue_cap":  cap(ws.broadcast),
	}
}