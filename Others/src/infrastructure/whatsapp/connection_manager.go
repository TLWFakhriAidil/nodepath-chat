package whatsapp

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// ConnectionManager handles robust connection management for WhatsApp clients
type ConnectionManager struct {
	clients    map[string]*ManagedClient
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// ManagedClient wraps a WhatsApp client with connection management
type ManagedClient struct {
	Client              *whatsmeow.Client
	DeviceID            string
	LastConnected       time.Time
	ReconnectAttempts   int
	ConsecutiveFailures int
	mu                  sync.RWMutex
}

var (
	connManager     *ConnectionManager
	connManagerOnce sync.Once
)

// GetConnectionManager returns singleton connection manager
func GetConnectionManager() *ConnectionManager {
	connManagerOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		connManager = &ConnectionManager{
			clients: make(map[string]*ManagedClient),
			ctx:     ctx,
			cancel:  cancel,
		}
		go connManager.monitorConnections()
	})
	return connManager
}

// AddClient adds a client to managed connections
func (cm *ConnectionManager) AddClient(deviceID string, client *whatsmeow.Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.clients[deviceID] = &ManagedClient{
		Client:        client,
		DeviceID:      deviceID,
		LastConnected: time.Now(),
	}
	
	// Set client properties for better stability
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true
	
	// Add disconnect handler
	client.AddEventHandler(func(evt interface{}) {
		// Process asynchronously to prevent blocking
		go func() {
			switch evt.(type) {
			case *events.Disconnected:
				logrus.Warnf("Device %s disconnected, will attempt reconnection", deviceID)
				cm.handleDisconnect(deviceID)
			}
		}()
	})
}

// handleDisconnect manages reconnection attempts
func (cm *ConnectionManager) handleDisconnect(deviceID string) {
	cm.mu.RLock()
	mc, exists := cm.clients[deviceID]
	cm.mu.RUnlock()
	
	if !exists {
		return
	}
	
	// Don't reconnect if manually disconnected
	if !mc.Client.EnableAutoReconnect {
		return
	}
	
	go func() {
		// Wait a bit before reconnecting to avoid rapid reconnects
		time.Sleep(5 * time.Second)
		
		mc.mu.Lock()
		mc.ReconnectAttempts++
		mc.mu.Unlock()
		
		// Try to reconnect with exponential backoff
		backoff := time.Second * 5
		maxBackoff := time.Minute * 5
		
		for i := 0; i < 10; i++ {
			if !mc.Client.IsConnected() {
				logrus.Infof("Reconnection attempt %d for device %s", i+1, deviceID)
				
				err := mc.Client.Connect()
				if err == nil {
					// Wait for connection to stabilize
					time.Sleep(2 * time.Second)
					
					if mc.Client.IsConnected() && mc.Client.IsLoggedIn() {
						mc.mu.Lock()
						mc.LastConnected = time.Now()
						mc.ConsecutiveFailures = 0
						mc.mu.Unlock()
						
						logrus.Infof("Device %s reconnected successfully", deviceID)
						
						// Send presence to confirm connection
						mc.Client.SendPresence(types.PresenceAvailable)
						return
					}
				}
				
				logrus.Errorf("Reconnection attempt %d failed for device %s: %v", i+1, deviceID, err)
			}
			
			// Exponential backoff
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
		
		mc.mu.Lock()
		mc.ConsecutiveFailures++
		mc.mu.Unlock()
		
		logrus.Errorf("Failed to reconnect device %s after 10 attempts", deviceID)
	}()
}

// monitorConnections periodically checks connection health
func (cm *ConnectionManager) monitorConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.checkConnections()
		}
	}
}

// checkConnections verifies all connections are healthy
func (cm *ConnectionManager) checkConnections() {
	cm.mu.RLock()
	clients := make(map[string]*ManagedClient)
	for k, v := range cm.clients {
		clients[k] = v
	}
	cm.mu.RUnlock()
	
	for deviceID, mc := range clients {
		if mc.Client == nil {
			continue
		}
		
		if !mc.Client.IsConnected() {
			mc.mu.RLock()
			lastConnected := mc.LastConnected
			failures := mc.ConsecutiveFailures
			mc.mu.RUnlock()
			
			// If disconnected for more than 5 minutes and many failures, might be banned
			if time.Since(lastConnected) > 5*time.Minute && failures > 5 {
				logrus.Warnf("Device %s has been disconnected for %v with %d failures, might be banned",
					deviceID, time.Since(lastConnected), failures)
			} else {
				// Try to reconnect
				cm.handleDisconnect(deviceID)
			}
		} else {
			// Device is connected - no need to send presence
			// This reduces pattern detection by WhatsApp
		}
	}
}

// GetClient returns a managed client
func (cm *ConnectionManager) GetClient(deviceID string) (*whatsmeow.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	mc, exists := cm.clients[deviceID]
	if !exists {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}
	
	return mc.Client, nil
}

// RemoveClient removes a client from management
func (cm *ConnectionManager) RemoveClient(deviceID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if mc, exists := cm.clients[deviceID]; exists {
		mc.Client.EnableAutoReconnect = false
		delete(cm.clients, deviceID)
	}
}
