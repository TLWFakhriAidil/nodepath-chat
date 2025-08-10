package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/multidevice"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "github.com/lib/pq"
)

// WorkerClientManager provides self-healing client retrieval for workers
type WorkerClientManager struct {
	mu            sync.RWMutex
	refreshing    map[string]bool // Track which devices are being refreshed
	refreshMutex  map[string]*sync.Mutex // Per-device refresh locks
}

var (
	workerManager *WorkerClientManager
	workerOnce    sync.Once
)

// GetWorkerClientManager returns singleton instance
func GetWorkerClientManager() *WorkerClientManager {
	workerOnce.Do(func() {
		workerManager = &WorkerClientManager{
			refreshing:   make(map[string]bool),
			refreshMutex: make(map[string]*sync.Mutex),
		}
	})
	return workerManager
}

// GetOrRefreshClient gets client or refreshes if needed - CORE FUNCTION FOR WORKERS
func (wcm *WorkerClientManager) GetOrRefreshClient(deviceID string) (*whatsmeow.Client, error) {
	// Get device info first
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device %s not found: %v", deviceID, err)
	}
	
	// Platform devices don't need refresh - they use external APIs
	if device.Platform != "" {
		return nil, fmt.Errorf("device %s is platform device (%s) - no WhatsApp client needed", deviceID, device.Platform)
	}
	
	// For WhatsApp devices only (platform is null/empty)
	// First, try to get existing healthy client
	cm := GetClientManager()
	client, err := cm.GetClient(deviceID)
	if err == nil && client != nil && client.IsConnected() && client.IsLoggedIn() {
		logrus.Debugf("✅ Device %s client is healthy", deviceID)
		return client, nil
	}
	
	// Check if device has JID (previous session)
	if device.JID == "" {
		return nil, fmt.Errorf("device %s has no JID, QR scan required", deviceID)
	}
	
	// Get or create per-device mutex to prevent concurrent refreshes
	wcm.mu.Lock()
	if wcm.refreshMutex[deviceID] == nil {
		wcm.refreshMutex[deviceID] = &sync.Mutex{}
	}
	deviceMutex := wcm.refreshMutex[deviceID]
	wcm.mu.Unlock()
	
	// Lock this device's refresh
	deviceMutex.Lock()
	defer deviceMutex.Unlock()
	
	// Check if already being refreshed by another goroutine
	wcm.mu.RLock()
	isRefreshing := wcm.refreshing[deviceID]
	wcm.mu.RUnlock()
	
	if isRefreshing {
		// Wait a bit and try again
		time.Sleep(2 * time.Second)
		client, err := cm.GetClient(deviceID)
		if err == nil && client != nil && client.IsConnected() {
			return client, nil
		}
		return nil, fmt.Errorf("device %s refresh in progress by another worker", deviceID)
	}
	
	// Mark as refreshing
	wcm.mu.Lock()
	wcm.refreshing[deviceID] = true
	wcm.mu.Unlock()
	
	defer func() {
		wcm.mu.Lock()
		wcm.refreshing[deviceID] = false
		wcm.mu.Unlock()
	}()
	
	logrus.Infof("🔄 Refreshing WhatsApp device %s for worker message sending...", deviceID)
	
	// Attempt refresh using same logic as refresh button
	refreshedClient, err := wcm.performRefresh(deviceID, device)
	if err != nil {
		logrus.Warnf("❌ Failed to refresh device %s: %v", deviceID, err)
		return nil, fmt.Errorf("refresh failed: %v", err)
	}
	
	logrus.Infof("✅ Successfully refreshed device %s", deviceID)
	return refreshedClient, nil
}

// performRefresh performs the actual device refresh (based on working device_reconnect.go)
func (wcm *WorkerClientManager) performRefresh(deviceID string, device *models.UserDevice) (*whatsmeow.Client, error) {
	// Check if session exists in database
	db, err := sql.Open("postgres", config.DBURI)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %v", err)
	}
	defer db.Close()
	
	// Query whatsmeow_sessions table using the JID
	var sessionData []byte
	err = db.QueryRow(`
		SELECT session 
		FROM whatsmeow_sessions 
		WHERE our_jid = ?
		LIMIT 1
	`, device.JID).Scan(&sessionData)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no session found for JID %s", device.JID)
		}
		return nil, fmt.Errorf("error querying session: %v", err)
	}
	
	// Initialize WhatsApp store container
	ctx := context.Background()
	dbLog := waLog.Stdout("Worker_"+deviceID, config.WhatsappLogLevel, false) // Reduce logging
	
	// Convert postgresql:// to postgres:// for whatsmeow
	dbURI := config.DBURI
	if strings.HasPrefix(dbURI, "postgresql://") {
		dbURI = strings.Replace(dbURI, "postgresql://", "postgres://", 1)
	}
	
	container, err := sqlstore.New(ctx, "postgres", dbURI, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to create store container: %v", err)
	}
	
	// Parse the JID
	jid, err := types.ParseJID(device.JID)
	if err != nil {
		return nil, fmt.Errorf("invalid JID format: %v", err)
	}
	
	// Get device from store
	waDevice, err := container.GetDevice(ctx, jid)
	if err != nil || waDevice == nil {
		return nil, fmt.Errorf("device not found in store")
	}
	
	// Create client
	client := whatsmeow.NewClient(waDevice, dbLog)
	client.EnableAutoReconnect = false // DISABLED - we handle reconnection manually
	client.AutoTrustIdentity = true
	
	// Register with DeviceManager BEFORE connecting
	dm := multidevice.GetDeviceManager()
	dm.RegisterDevice(deviceID, device.UserID, device.Phone, client)
	
	// Add minimal event handlers (no heavy processing)
	client.AddEventHandler(func(evt interface{}) {
		// Minimal event handling for workers
		HandleDeviceEvent(context.Background(), deviceID, evt)
	})
	
	// Try to connect
	err = client.Connect()
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}
	
	// Wait for connection with timeout
	connected := false
	for i := 0; i < 20; i++ { // 10 seconds total
		if client.IsConnected() && client.IsLoggedIn() {
			connected = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	
	if !connected {
		client.Disconnect()
		return nil, fmt.Errorf("connection timeout")
	}
	
	// Register with ClientManager (this makes it available to other workers)
	cm := GetClientManager()
	cm.AddClient(deviceID, client)
	
	// Update device status in database
	userRepo := repository.GetUserRepository()
	newJID := ""
	if client.Store.ID != nil {
		newJID = client.Store.ID.String()
	}
	userRepo.UpdateDeviceStatus(deviceID, "online", device.Phone, newJID)
	
	return client, nil
}

// IsClientHealthy checks if a client is healthy for sending messages
func (wcm *WorkerClientManager) IsClientHealthy(client *whatsmeow.Client) bool {
	return client != nil && client.IsConnected() && client.IsLoggedIn()
}

// GetRefreshStatus returns current refresh status for debugging
func (wcm *WorkerClientManager) GetRefreshStatus() map[string]bool {
	wcm.mu.RLock()
	defer wcm.mu.RUnlock()
	
	status := make(map[string]bool)
	for k, v := range wcm.refreshing {
		status[k] = v
	}
	return status
}
