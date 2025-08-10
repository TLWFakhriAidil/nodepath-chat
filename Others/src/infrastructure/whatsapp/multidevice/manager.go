package multidevice

import (
	"context"
	"fmt"
	"sync"
	"strings"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	log "github.com/sirupsen/logrus"
)

// DeviceManager manages multiple WhatsApp device connections
type DeviceManager struct {
	mu             sync.RWMutex
	devices        map[string]*DeviceConnection // deviceID -> connection
	storeContainer *sqlstore.Container
	dbLog          waLog.Logger
}

// DeviceConnection represents a single device's WhatsApp connection
type DeviceConnection struct {
	DeviceID     string
	UserID       string
	Phone        string
	Client       *whatsmeow.Client
	Store        *store.Device
	Connected    bool
	ConnectedAt  int64
}

var (
	manager     *DeviceManager
	managerOnce sync.Once
)

// GetDeviceManager returns singleton instance of DeviceManager
func GetDeviceManager() *DeviceManager {
	managerOnce.Do(func() {
		ctx := context.Background()
		dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)
		
		// Initialize the store container
		storeContainer, err := initDatabase(ctx, dbLog)
		if err != nil {
			log.Errorf("Failed to initialize device manager database: %v", err)
			panic(err)
		}
		
		manager = &DeviceManager{
			devices:        make(map[string]*DeviceConnection),
			storeContainer: storeContainer,
			dbLog:          dbLog,
		}
		
		log.Infof("Device Manager initialized for multi-device support")
	})
	return manager
}

// initDatabase creates the WhatsApp store
func initDatabase(ctx context.Context, dbLog waLog.Logger) (*sqlstore.Container, error) {
	// Use PostgreSQL for WhatsApp sessions
	// Check if we have a configured DB URI
	dbURI := config.DBURI
	
	// If the main DB is PostgreSQL, create a separate schema for WhatsApp
	if strings.HasPrefix(dbURI, "postgres:") || strings.HasPrefix(dbURI, "postgresql:") {
		// For production, you might want to use a separate database
		// For now, we'll use the same database with a whatsapp_ prefix for tables
		if strings.HasPrefix(dbURI, "postgresql://") {
			dbURI = strings.Replace(dbURI, "postgresql://", "postgres://", 1)
		}
		return sqlstore.New(ctx, "postgres", dbURI, dbLog)
	}
	
	// Fallback to file-based SQLite if not using PostgreSQL
	return sqlstore.New(ctx, "sqlite3", "file:storages/whatsapp_multidevice.db?_foreign_keys=on", dbLog)
}

// CreateDeviceSession creates a new WhatsApp session for a device
func (dm *DeviceManager) CreateDeviceSession(deviceID, userID, phone string) (*DeviceConnection, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	// Check if device already exists
	if conn, exists := dm.devices[deviceID]; exists {
		return conn, fmt.Errorf("device session already exists for device %s", deviceID)
	}
	
	// Create a new device in the WhatsApp store
	// Use deviceID as the device identifier
	device := dm.storeContainer.NewDevice()
	
	// Store device metadata
	err := dm.storeDeviceMapping(deviceID, userID, phone, device.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to store device mapping: %w", err)
	}
	
	// Create WhatsApp client
	client := whatsmeow.NewClient(device, dm.dbLog)
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true
	
	// Create device connection
	conn := &DeviceConnection{
		DeviceID:  deviceID,
		UserID:    userID,
		Phone:     phone,
		Client:    client,
		Store:     device,
		Connected: false,
	}
	
	dm.devices[deviceID] = conn
	
	log.Infof("Created new device session for device %s (user: %s, phone: %s)", deviceID, userID, phone)
	
	return conn, nil
}

// GetDeviceConnection gets an existing device connection
func (dm *DeviceManager) GetDeviceConnection(deviceID string) (*DeviceConnection, error) {
	// First check if this is a platform device
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(deviceID)
	if err == nil && device.Platform != "" {
		// Return a dummy connection for platform devices
		return &DeviceConnection{
			DeviceID:  deviceID,
			UserID:    device.UserID,
			Phone:     device.Phone,
			Connected: true, // Always consider platform devices as connected
		}, nil
	}
	
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	conn, exists := dm.devices[deviceID]
	if !exists {
		// Auto-refresh disabled - using auto-reconnect on startup instead
		log.Warnf("No device connection found for %s", deviceID)
		
		// Don't trigger auto-refresh anymore
		// infrastructure.TriggerDeviceRefresh(deviceID)
		
		// Return error immediately
		return nil, fmt.Errorf("no device connection found for device %s", deviceID)
	}
	
	return conn, nil
}

// GetOrCreateDeviceConnection gets existing or creates new device connection
func (dm *DeviceManager) GetOrCreateDeviceConnection(deviceID, userID, phone string) (*DeviceConnection, error) {
	// Try to get existing first
	conn, err := dm.GetDeviceConnection(deviceID)
	if err == nil {
		return conn, nil
	}
	
	// Check if we have a stored session
	storedDevice, err := dm.getStoredDevice(deviceID)
	if err == nil && storedDevice != nil {
		// Restore from stored session
		client := whatsmeow.NewClient(storedDevice, dm.dbLog)
		client.EnableAutoReconnect = true
		client.AutoTrustIdentity = true
		
		conn := &DeviceConnection{
			DeviceID:  deviceID,
			UserID:    userID,
			Phone:     phone,
			Client:    client,
			Store:     storedDevice,
			Connected: false,
		}
		
		dm.mu.Lock()
		dm.devices[deviceID] = conn
		dm.mu.Unlock()
		
		log.Infof("Restored device session for device %s", deviceID)
		return conn, nil
	}
	
	// Create new session
	return dm.CreateDeviceSession(deviceID, userID, phone)
}

// RemoveDeviceSession removes a device session
func (dm *DeviceManager) RemoveDeviceSession(deviceID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	conn, exists := dm.devices[deviceID]
	if !exists {
		return fmt.Errorf("no device session found for device %s", deviceID)
	}
	
	// Disconnect if connected
	if conn.Client != nil && conn.Client.IsConnected() {
		conn.Client.Disconnect()
	}
	
	// Delete from WhatsApp store
	if conn.Store != nil {
		ctx := context.Background()
		err := conn.Store.Delete(ctx)
		if err != nil {
			log.Errorf("Failed to delete device from store: %v", err)
		}
	}
	
	// Remove device mapping
	err := dm.removeDeviceMapping(deviceID)
	if err != nil {
		log.Errorf("Failed to remove device mapping: %v", err)
	}
	
	// Remove from active connections
	delete(dm.devices, deviceID)
	
	log.Infof("Removed device session for device %s", deviceID)
	
	return nil
}

// Device mapping functions to link database device IDs with WhatsApp device IDs

func (dm *DeviceManager) storeDeviceMapping(deviceID, userID, phone, waDeviceID string) error {
	// Store mapping in a separate table or file
	// For now, we'll use a simple approach with the device ID as the identifier
	return nil
}

func (dm *DeviceManager) getStoredDevice(deviceID string) (*store.Device, error) {
	// Try to find device by matching JID pattern or stored mapping
	// For now, return nil to force new device creation
	return nil, fmt.Errorf("no stored device found")
}

func (dm *DeviceManager) removeDeviceMapping(deviceID string) error {
	// Remove mapping from storage
	return nil
}

// GetAllDeviceConnections returns all active device connections
func (dm *DeviceManager) GetAllDeviceConnections() map[string]*DeviceConnection {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	devices := make(map[string]*DeviceConnection)
	for k, v := range dm.devices {
		devices[k] = v
	}
	
	return devices
}

// UpdateDeviceStatus updates the connection status of a device
func (dm *DeviceManager) UpdateDeviceStatus(deviceID string, connected bool, phone string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	conn, exists := dm.devices[deviceID]
	if !exists {
		return fmt.Errorf("no device connection found for device %s", deviceID)
	}
	
	conn.Connected = connected
	if connected {
		conn.Phone = phone
	}
	
	return nil
}
