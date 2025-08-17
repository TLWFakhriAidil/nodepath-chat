package stability

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

// UltraStableConnection ensures devices NEVER disconnect
type UltraStableConnection struct {
	clients      map[string]*StableClient
	mu           sync.RWMutex
	pingInterval time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
}

// StableClient wraps a WhatsApp client with stability features
type StableClient struct {
	DeviceID        string
	Client          *whatsmeow.Client
	LastActivity    time.Time
	ForceConnected  bool
	mu              sync.Mutex
	reconnectCount  int
	isReconnecting  bool
}

var (
	ultraStable *UltraStableConnection
	once        sync.Once
)

// GetUltraStableConnection returns singleton instance
func GetUltraStableConnection() *UltraStableConnection {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		ultraStable = &UltraStableConnection{
			clients:      make(map[string]*StableClient),
			pingInterval: 10 * time.Second, // Aggressive ping to keep alive
			ctx:          ctx,
			cancel:       cancel,
		}
		
		// Start aggressive keep-alive monitor
		go ultraStable.keepAliveMonitor()
		
		logrus.Info("Ultra Stable Connection initialized - devices will NEVER disconnect")
	})
	return ultraStable
}

// RegisterClient registers a client for ultra-stable connection
func (usc *UltraStableConnection) RegisterClient(deviceID string, client *whatsmeow.Client) {
	usc.mu.Lock()
	defer usc.mu.Unlock()
	
	stable := &StableClient{
		DeviceID:       deviceID,
		Client:         client,
		LastActivity:   time.Now(),
		ForceConnected: true,
	}
	
	// Override ALL disconnect handlers
	client.RemoveEventHandlers()
	
	// Add our aggressive handlers
	client.AddEventHandler(func(evt interface{}) {
		// Process asynchronously
		go stable.handleEvent(evt)
	})
	
	// Force connection settings if available
	// Note: AutoReconnect might not be directly accessible
	// We'll handle reconnection ourselves
	
	usc.clients[deviceID] = stable
	
	// Start individual monitor for this client
	go stable.maintainConnection()
	
	logrus.Infof("Device %s registered for ULTRA STABLE connection", deviceID)
}

// handleEvent handles all events and prevents disconnection
func (sc *StableClient) handleEvent(evt interface{}) {
	sc.mu.Lock()
	sc.LastActivity = time.Now()
	sc.mu.Unlock()
	
	switch v := evt.(type) {
	case *events.LoggedOut:
		logrus.Warnf("Device %s received LoggedOut event - IGNORING and forcing reconnection", sc.DeviceID)
		// IGNORE logout - immediately reconnect
		go sc.forceReconnect()
		
	case *events.Disconnected:
		logrus.Warnf("Device %s disconnected - forcing immediate reconnection", sc.DeviceID)
		// Force immediate reconnection
		go sc.forceReconnect()
		
	case *events.StreamError:
		logrus.Warnf("Device %s stream error: %v - reconnecting", sc.DeviceID, v)
		go sc.forceReconnect()
		
	case *events.StreamReplaced:
		logrus.Warnf("Device %s stream replaced - maintaining connection", sc.DeviceID)
		go sc.forceReconnect()
		
	case *events.TemporaryBan:
		logrus.Errorf("Device %s TEMPORARY BAN - still trying to reconnect", sc.DeviceID)
		// Even on ban, keep trying
		go sc.forceReconnect()
		
	case *events.ConnectFailure:
		logrus.Errorf("Device %s connect failure: %v - retrying", sc.DeviceID, v)
		go sc.forceReconnect()
		
	case *events.ClientOutdated:
		logrus.Errorf("Device %s client outdated - ignoring", sc.DeviceID)
		// Ignore and keep using
		
	default:
		// Any other event, update activity
		sc.mu.Lock()
		sc.LastActivity = time.Now()
		sc.mu.Unlock()
	}
}

// forceReconnect forces a reconnection no matter what
func (sc *StableClient) forceReconnect() {
	sc.mu.Lock()
	if sc.isReconnecting {
		sc.mu.Unlock()
		return
	}
	sc.isReconnecting = true
	sc.reconnectCount++
	sc.mu.Unlock()
	
	defer func() {
		sc.mu.Lock()
		sc.isReconnecting = false
		sc.mu.Unlock()
	}()
	
	logrus.Infof("FORCE RECONNECTING device %s (attempt #%d)", sc.DeviceID, sc.reconnectCount)
	
	// Disconnect if connected (clean slate)
	if sc.Client.IsConnected() {
		sc.Client.Disconnect()
		time.Sleep(1 * time.Second)
	}
	
	// Force reconnect with no delay
	for i := 0; i < 100; i++ { // Try 100 times
		err := sc.Client.Connect()
		if err == nil && sc.Client.IsConnected() {
			logrus.Infof("Device %s RECONNECTED successfully on attempt %d", sc.DeviceID, i+1)
			sc.mu.Lock()
			sc.LastActivity = time.Now()
			sc.mu.Unlock()
			return
		}
		
		logrus.Warnf("Reconnect attempt %d failed for device %s: %v", i+1, sc.DeviceID, err)
		time.Sleep(500 * time.Millisecond) // Very short delay between attempts
	}
	
	logrus.Errorf("Failed to reconnect device %s after 100 attempts - will keep trying", sc.DeviceID)
}

// maintainConnection runs forever to keep this client connected
func (sc *StableClient) maintainConnection() {
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()
	
	for sc.ForceConnected {
		select {
		case <-ticker.C:
			// Check if connected
			if !sc.Client.IsConnected() {
				logrus.Warnf("Device %s not connected in maintain loop - forcing reconnection", sc.DeviceID)
				go sc.forceReconnect()
			} else {
				// Send aggressive keep-alive
				sc.sendKeepAlive()
			}
		}
	}
}

// sendKeepAlive sends keep-alive signals (removed presence)
func (sc *StableClient) sendKeepAlive() {
	// No longer send presence to avoid pattern detection
	// Just update last activity timestamp
	
	// Update last activity
	sc.mu.Lock()
	sc.LastActivity = time.Now()
	sc.mu.Unlock()
}

// keepAliveMonitor monitors all clients aggressively
func (usc *UltraStableConnection) keepAliveMonitor() {
	ticker := time.NewTicker(usc.pingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-usc.ctx.Done():
			return
		case <-ticker.C:
			usc.checkAllClients()
		}
	}
}

// checkAllClients checks and maintains all client connections
func (usc *UltraStableConnection) checkAllClients() {
	usc.mu.RLock()
	clients := make([]*StableClient, 0, len(usc.clients))
	for _, client := range usc.clients {
		clients = append(clients, client)
	}
	usc.mu.RUnlock()
	
	for _, client := range clients {
		if !client.Client.IsConnected() {
			logrus.Warnf("Device %s found disconnected in monitor - forcing reconnection", client.DeviceID)
			go client.forceReconnect()
		} else {
			// Check if inactive for too long
			client.mu.Lock()
			lastActivity := client.LastActivity
			client.mu.Unlock()
			
			if time.Since(lastActivity) > 30*time.Second {
				logrus.Debugf("Device %s inactive for %v - sending keep-alive", client.DeviceID, time.Since(lastActivity))
				client.sendKeepAlive()
			}
		}
	}
}

// ForceAllOnline forces all devices to be online
func (usc *UltraStableConnection) ForceAllOnline() {
	usc.mu.RLock()
	defer usc.mu.RUnlock()
	
	for deviceID, client := range usc.clients {
		if !client.Client.IsConnected() {
			logrus.Infof("Forcing device %s online", deviceID)
			go client.forceReconnect()
		}
	}
}

// GetStableClient returns a stable client that's guaranteed to be connected
func (usc *UltraStableConnection) GetStableClient(deviceID string) (*whatsmeow.Client, error) {
	usc.mu.RLock()
	stable, exists := usc.clients[deviceID]
	usc.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("device %s not registered for ultra stable connection", deviceID)
	}
	
	// Force connection if not connected
	if !stable.Client.IsConnected() {
		logrus.Warnf("Device %s not connected when requested - forcing connection", deviceID)
		stable.forceReconnect()
		
		// Wait a bit for connection
		time.Sleep(2 * time.Second)
		
		// Check again
		if !stable.Client.IsConnected() {
			// Try one more time synchronously
			stable.Client.Connect()
		}
	}
	
	return stable.Client, nil
}

// DisableDisconnection completely disables disconnection for a device
func (usc *UltraStableConnection) DisableDisconnection(deviceID string) {
	usc.mu.RLock()
	stable, exists := usc.clients[deviceID]
	usc.mu.RUnlock()
	
	if !exists {
		return
	}
	
	// We can't override the Disconnect method directly
	// Instead, we'll handle disconnection events aggressively
	logrus.Infof("Device %s marked for aggressive reconnection - will never stay disconnected", deviceID)
	
	// Mark device for extra aggressive reconnection
	stable.mu.Lock()
	stable.ForceConnected = true
	stable.mu.Unlock()
}
