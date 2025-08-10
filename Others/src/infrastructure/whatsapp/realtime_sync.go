package whatsapp

import (
	"sync"
	"time"
	
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
)

// RealtimeSyncManager handles automatic real-time sync for all devices
type RealtimeSyncManager struct {
	mu              sync.RWMutex
	syncInterval    time.Duration
	isRunning       bool
	lastSyncTime    map[string]time.Time
	messageHandlers map[string]chan *events.Message
}

var (
	syncManager     *RealtimeSyncManager
	syncManagerOnce sync.Once
)

// GetRealtimeSyncManager returns the singleton sync manager
func GetRealtimeSyncManager() *RealtimeSyncManager {
	syncManagerOnce.Do(func() {
		syncManager = &RealtimeSyncManager{
			syncInterval:    30 * time.Second, // Check every 30 seconds
			lastSyncTime:    make(map[string]time.Time),
			messageHandlers: make(map[string]chan *events.Message),
		}
	})
	return syncManager
}

// StartRealtimeSync starts the automatic sync for all devices
func (rsm *RealtimeSyncManager) StartRealtimeSync() {
	rsm.mu.Lock()
	defer rsm.mu.Unlock()
	
	if rsm.isRunning {
		logrus.Info("Realtime sync already running")
		return
	}
	
	rsm.isRunning = true
	go rsm.syncLoop()
	
	logrus.Info("🚀 Started real-time sync manager for all devices")
}

// syncLoop runs the automatic sync check
func (rsm *RealtimeSyncManager) syncLoop() {
	ticker := time.NewTicker(rsm.syncInterval)
	defer ticker.Stop()
	
	// Initial sync for all devices
	rsm.syncAllDevices()
	
	for {
		select {
		case <-ticker.C:
			if config.WhatsappChatStorage {
				rsm.syncAllDevices()
			}
		}
	}
}

// syncAllDevices syncs all online devices
func (rsm *RealtimeSyncManager) syncAllDevices() {
	cm := GetClientManager()
	allClients := cm.GetAllClients()
	
	// logrus.Infof("🔄 Running real-time sync for %d devices", len(allClients))
	
	// Use goroutines for parallel processing (optimized for 3000 devices)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 50) // LIMIT concurrent syncs to 50
	userRepo := repository.GetUserRepository()
	
	for deviceID, client := range allClients {
		if client == nil {
			continue
		}
		
		// Check device status from database
		device, err := userRepo.GetDeviceByID(deviceID)
		if err != nil || device == nil || device.Status != "online" {
			continue
		}
		
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		
		go func(deviceID string, client *whatsmeow.Client) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore
			
			rsm.syncDevice(deviceID, client)
		}(deviceID, client)
	}
	
	wg.Wait()
	// logrus.Info("✅ Real-time sync completed for all devices")
}

// syncDevice syncs a single device
func (rsm *RealtimeSyncManager) syncDevice(deviceID string, client *whatsmeow.Client) {
	rsm.mu.Lock()
	lastSync, exists := rsm.lastSyncTime[deviceID]
	rsm.mu.Unlock()
	
	// Skip if synced recently (within 20 seconds)
	if exists && time.Since(lastSync) < 20*time.Second {
		return
	}
	
	// Update last sync time
	rsm.mu.Lock()
	rsm.lastSyncTime[deviceID] = time.Now()
	rsm.mu.Unlock()
	
	// No need to send presence for sync
	// Device is already connected and receiving messages
	
	logrus.Debugf("Synced device %s", deviceID)
}

// RegisterDevice registers a device for real-time message handling
func (rsm *RealtimeSyncManager) RegisterDevice(deviceID string, client *whatsmeow.Client) {
	rsm.mu.Lock()
	defer rsm.mu.Unlock()
	
	// Create message channel for this device
	msgChan := make(chan *events.Message, 100)
	rsm.messageHandlers[deviceID] = msgChan
	
	// Start message processor for this device
	go rsm.processDeviceMessages(deviceID, client, msgChan)
	
	logrus.Infof("Registered device %s for real-time sync", deviceID)
}

// UnregisterDevice removes a device from real-time sync
func (rsm *RealtimeSyncManager) UnregisterDevice(deviceID string) {
	rsm.mu.Lock()
	defer rsm.mu.Unlock()
	
	if ch, exists := rsm.messageHandlers[deviceID]; exists {
		close(ch)
		delete(rsm.messageHandlers, deviceID)
	}
	
	delete(rsm.lastSyncTime, deviceID)
	logrus.Infof("Unregistered device %s from real-time sync", deviceID)
}

// HandleRealtimeMessage processes a message in real-time
func (rsm *RealtimeSyncManager) HandleRealtimeMessage(deviceID string, evt *events.Message) {
	rsm.mu.RLock()
	msgChan, exists := rsm.messageHandlers[deviceID]
	rsm.mu.RUnlock()
	
	if exists {
		select {
		case msgChan <- evt:
			// Message queued for processing
		default:
			// Channel full, drop message to prevent blocking
			logrus.Warnf("Message channel full for device %s, dropping message", deviceID)
		}
	}
}

// processDeviceMessages processes messages for a specific device
func (rsm *RealtimeSyncManager) processDeviceMessages(deviceID string, client *whatsmeow.Client, msgChan chan *events.Message) {
	for evt := range msgChan {
		// Process the message in real-time
		if evt.Info.Chat.Server == types.DefaultUserServer {
			// Update chat info
			HandleMessageForChats(deviceID, client, evt)
			
			// Store the message
			HandleMessageForWebView(deviceID, evt)
			
			logrus.Debugf("Processed real-time message for device %s in chat %s", 
				deviceID, evt.Info.Chat.String())
		}
	}
}

// EnableRealtimeSync enables real-time sync when a message is received
func EnableRealtimeSync(deviceID string, client *whatsmeow.Client, evt *events.Message) {
	if !config.WhatsappChatStorage {
		return
	}
	
	// Register device if not already registered
	rsm := GetRealtimeSyncManager()
	rsm.mu.RLock()
	_, registered := rsm.messageHandlers[deviceID]
	rsm.mu.RUnlock()
	
	if !registered {
		rsm.RegisterDevice(deviceID, client)
	}
	
	// Handle the message in real-time
	rsm.HandleRealtimeMessage(deviceID, evt)
}

// InitializeRealtimeSync initializes the real-time sync system
func InitializeRealtimeSync() {
	if !config.WhatsappChatStorage {
		logrus.Info("WhatsApp chat storage disabled, skipping real-time sync")
		return
	}
	
	rsm := GetRealtimeSyncManager()
	rsm.StartRealtimeSync()
	
	logrus.Info("✅ Real-time sync initialized for WhatsApp Web")
}