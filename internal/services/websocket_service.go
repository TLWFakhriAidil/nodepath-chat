package services

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// WebSocketService handles real-time messaging for high-performance communication
type WebSocketService struct {
	// Connection management
	connections map[string]*ConnectionInfo
	connMutex   sync.RWMutex

	// Message broadcasting
	broadcast chan *BroadcastMessage

	// Connection limits for performance
	maxConnections int
	currentConns   int
	connCountMutex sync.RWMutex

	// Graceful shutdown support
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	// Connection cleanup
	cleanupTicker *time.Ticker
}

// ConnectionInfo holds connection details with metadata for leak prevention
type ConnectionInfo struct {
	Conn      *websocket.Conn
	LastPing  time.Time
	LastPong  time.Time
	CreatedAt time.Time
	cancel    context.CancelFunc
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
	ctx, cancel := context.WithCancel(context.Background())

	ws := &WebSocketService{
		connections:    make(map[string]*ConnectionInfo),
		broadcast:      make(chan *BroadcastMessage, 1000), // Buffered channel for performance
		maxConnections: maxConnections,
		ctx:            ctx,
		cancel:         cancel,
		done:           make(chan struct{}),
		cleanupTicker:  time.NewTicker(30 * time.Second), // Cleanup every 30 seconds
	}

	// Start the broadcast handler
	go ws.handleBroadcasts()

	// Start connection cleanup routine
	go ws.cleanupStaleConnections()

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

// registerConnection adds a new WebSocket connection with proper metadata tracking
func (ws *WebSocketService) registerConnection(deviceID string, conn *websocket.Conn) {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()

	// Close existing connection if any
	if existingConnInfo, exists := ws.connections[deviceID]; exists {
		existingConnInfo.Conn.Close()
		if existingConnInfo.cancel != nil {
			existingConnInfo.cancel()
		}
	}

	// Create connection context for graceful shutdown
	connCtx, connCancel := context.WithCancel(ws.ctx)

	// Create connection info with metadata
	connInfo := &ConnectionInfo{
		Conn:      conn,
		LastPing:  time.Now(),
		LastPong:  time.Now(),
		CreatedAt: time.Now(),
		cancel:    connCancel,
	}

	ws.connections[deviceID] = connInfo

	// Update connection count
	ws.connCountMutex.Lock()
	ws.currentConns++
	ws.connCountMutex.Unlock()

	// Set up ping/pong handlers for this connection
	conn.SetPongHandler(func(string) error {
		ws.connMutex.Lock()
		if connInfo, exists := ws.connections[deviceID]; exists {
			connInfo.LastPong = time.Now()
		}
		ws.connMutex.Unlock()
		return nil
	})

	// Start ping routine for this connection
	go ws.pingConnection(deviceID, connCtx)
}

// unregisterConnection removes a WebSocket connection with proper cleanup
func (ws *WebSocketService) unregisterConnection(deviceID string) {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()

	if connInfo, exists := ws.connections[deviceID]; exists {
		// Cancel the connection context
		if connInfo.cancel != nil {
			connInfo.cancel()
		}

		// Close the connection gracefully
		connInfo.Conn.Close()

		delete(ws.connections, deviceID)

		// Update connection count
		ws.connCountMutex.Lock()
		if ws.currentConns > 0 {
			ws.currentConns--
		}
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

// BroadcastMessage sends a message to a specific device or all devices
func (ws *WebSocketService) BroadcastMessageBytes(deviceID string, message []byte) {
	ws.connMutex.RLock()
	defer ws.connMutex.RUnlock()

	if deviceID == "" {
		// Broadcast to all devices
		for _, connInfo := range ws.connections {
			ws.sendToConnectionBytes(connInfo, message)
		}
	} else {
		// Send to specific device
		if connInfo, exists := ws.connections[deviceID]; exists {
			ws.sendToConnectionBytes(connInfo, message)
		}
	}
}

// handleBroadcasts processes broadcast messages in a separate goroutine
func (ws *WebSocketService) handleBroadcasts() {
	for msg := range ws.broadcast {
		ws.connMutex.RLock()

		if len(msg.Targets) > 0 {
			// Send to specific targets
			for _, deviceID := range msg.Targets {
				if connInfo, exists := ws.connections[deviceID]; exists {
					ws.sendToConnection(connInfo, msg, deviceID)
				}
			}
		} else {
			// Broadcast to all connections
			for deviceID, connInfo := range ws.connections {
				ws.sendToConnection(connInfo, msg, deviceID)
			}
		}

		ws.connMutex.RUnlock()
	}
}

// sendToConnection sends a message to a specific WebSocket connection
func (ws *WebSocketService) sendToConnection(connInfo *ConnectionInfo, msg *BroadcastMessage, deviceID string) {
	// Set write deadline for performance
	connInfo.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	err := connInfo.Conn.WriteJSON(msg)
	if err != nil {
		logrus.WithError(err).WithField("device_id", deviceID).Error("Failed to send WebSocket message")
		// Remove the problematic connection
		go ws.unregisterConnection(deviceID)
	}
}

// sendToConnectionBytes sends a message to a specific connection with proper error handling
func (ws *WebSocketService) sendToConnectionBytes(connInfo *ConnectionInfo, message []byte) {
	defer func() {
		if r := recover(); r != nil {
			// Connection might be closed, ignore the error
		}
	}()

	// Set write deadline to prevent hanging
	connInfo.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err := connInfo.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		// Connection is probably closed, we should unregister it
		// Find the device ID for this connection and unregister
		ws.connMutex.Lock()
		for deviceID, conn := range ws.connections {
			if conn == connInfo {
				delete(ws.connections, deviceID)
				// Update connection count
				ws.connCountMutex.Lock()
				if ws.currentConns > 0 {
					ws.currentConns--
				}
				ws.connCountMutex.Unlock()
				break
			}
		}
		ws.connMutex.Unlock()
		connInfo.Conn.Close()
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

// pingConnection sends periodic ping messages to maintain connection health
func (ws *WebSocketService) pingConnection(deviceID string, ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Ping every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ws.connMutex.RLock()
			connInfo, exists := ws.connections[deviceID]
			ws.connMutex.RUnlock()

			if !exists {
				return
			}

			// Send ping
			connInfo.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := connInfo.Conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				// Connection is dead, unregister it
				ws.unregisterConnection(deviceID)
				return
			}

			// Update last ping time
			ws.connMutex.Lock()
			if connInfo, exists := ws.connections[deviceID]; exists {
				connInfo.LastPing = time.Now()
			}
			ws.connMutex.Unlock()
		}
	}
}

// cleanupStaleConnections removes connections that haven't responded to pings
func (ws *WebSocketService) cleanupStaleConnections() {
	ticker := time.NewTicker(60 * time.Second) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			ws.connMutex.Lock()
			staleConnections := make([]string, 0)
			now := time.Now()

			for deviceID, connInfo := range ws.connections {
				// If no pong received for 2 minutes, consider connection stale
				if now.Sub(connInfo.LastPong) > 2*time.Minute {
					staleConnections = append(staleConnections, deviceID)
				}
			}
			ws.connMutex.Unlock()

			// Remove stale connections
			for _, deviceID := range staleConnections {
				ws.unregisterConnection(deviceID)
			}
		}
	}
}

// Shutdown gracefully shuts down the WebSocket service
func (ws *WebSocketService) Shutdown() {
	// Cancel the context to stop all goroutines
	ws.cancel()

	// Close all connections
	ws.connMutex.Lock()
	for deviceID, connInfo := range ws.connections {
		if connInfo.cancel != nil {
			connInfo.cancel()
		}
		connInfo.Conn.Close()
		delete(ws.connections, deviceID)
	}
	ws.connMutex.Unlock()

	// Reset connection count
	ws.connCountMutex.Lock()
	ws.currentConns = 0
	ws.connCountMutex.Unlock()
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
		"current_connections":  currentConns,
		"max_connections":      ws.maxConnections,
		"broadcast_queue_size": len(ws.broadcast),
		"broadcast_queue_cap":  cap(ws.broadcast),
	}
}
