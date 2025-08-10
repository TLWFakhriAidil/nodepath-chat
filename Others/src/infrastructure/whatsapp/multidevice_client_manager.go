package whatsapp

import (
	"sync"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// MultiDeviceClientManager ensures multiple devices stay registered
type MultiDeviceClientManager struct {
	clients map[string]*whatsmeow.Client
	mu      sync.RWMutex
}

var (
	multiDeviceManager *MultiDeviceClientManager
	mdOnce            sync.Once
)

// GetMultiDeviceManager returns the singleton multi-device manager
func GetMultiDeviceManager() *MultiDeviceClientManager {
	mdOnce.Do(func() {
		multiDeviceManager = &MultiDeviceClientManager{
			clients: make(map[string]*whatsmeow.Client),
		}
		logrus.Info("MultiDeviceClientManager initialized")
	})
	return multiDeviceManager
}

// RegisterDevice registers a device ensuring it doesn't override others
func (mdm *MultiDeviceClientManager) RegisterDevice(deviceID string, client *whatsmeow.Client) {
	mdm.mu.Lock()
	defer mdm.mu.Unlock()
	
	// Check if we're replacing an existing client
	if existingClient, exists := mdm.clients[deviceID]; exists {
		if existingClient == client {
			logrus.Debugf("Device %s already registered with same client", deviceID)
			return
		}
		logrus.Warnf("Replacing client for device %s", deviceID)
	}
	
	mdm.clients[deviceID] = client
	logrus.Infof("Registered device %s in MultiDeviceManager (total: %d)", deviceID, len(mdm.clients))
	
	// Also register with the main ClientManager
	cm := GetClientManager()
	cm.AddClient(deviceID, client)
}

// GetDevice retrieves a client for a specific device
func (mdm *MultiDeviceClientManager) GetDevice(deviceID string) (*whatsmeow.Client, bool) {
	mdm.mu.RLock()
	defer mdm.mu.RUnlock()
	
	client, exists := mdm.clients[deviceID]
	return client, exists
}

// GetAllDevices returns all registered devices
func (mdm *MultiDeviceClientManager) GetAllDevices() map[string]*whatsmeow.Client {
	mdm.mu.RLock()
	defer mdm.mu.RUnlock()
	
	// Create a copy
	devices := make(map[string]*whatsmeow.Client)
	for k, v := range mdm.clients {
		devices[k] = v
	}
	return devices
}

// RemoveDevice removes a device from the manager
func (mdm *MultiDeviceClientManager) RemoveDevice(deviceID string) {
	mdm.mu.Lock()
	defer mdm.mu.Unlock()
	
	delete(mdm.clients, deviceID)
	logrus.Infof("Removed device %s from MultiDeviceManager", deviceID)
	
	// Also remove from ClientManager
	cm := GetClientManager()
	cm.RemoveClient(deviceID)
}

// EnsureDeviceRegistered ensures a device stays registered in both managers
func (mdm *MultiDeviceClientManager) EnsureDeviceRegistered(deviceID string, client *whatsmeow.Client) {
	// Register in MultiDeviceManager
	mdm.RegisterDevice(deviceID, client)
	
	// Double-check it's in ClientManager
	cm := GetClientManager()
	if _, err := cm.GetClient(deviceID); err != nil {
		logrus.Warnf("Device %s was not in ClientManager, re-registering", deviceID)
		cm.AddClient(deviceID, client)
	}
}
