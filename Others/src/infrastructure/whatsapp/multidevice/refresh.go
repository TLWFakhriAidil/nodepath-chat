package multidevice

import (
	"context"
	"fmt"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	log "github.com/sirupsen/logrus"
)

// GetOrRefreshClient gets a healthy client or refreshes if needed (SELF-HEALING)
// This is the ONLY method that should be used for getting clients for message sending
func (dm *DeviceManager) GetOrRefreshClient(deviceID string) (*whatsmeow.Client, error) {
	// Get device info first
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device %s not found: %v", deviceID, err)
	}
	
	// Platform devices don't need WhatsApp client
	if device.Platform != "" {
		return nil, fmt.Errorf("device %s is platform device (%s)", deviceID, device.Platform)
	}
	
	// First, try to get existing healthy client
	dm.mu.RLock()
	conn, exists := dm.devices[deviceID]
	dm.mu.RUnlock()
	
	if exists && conn != nil && conn.Client != nil && 
	   conn.Client.IsConnected() && conn.Client.IsLoggedIn() {
		log.Debugf("✅ Device %s client is healthy", deviceID)
		return conn.Client, nil
	}
	
	// Check if device has JID (required for refresh)
	if device.JID == "" {
		// For recently connected devices, retry a few times
		for i := 0; i < 3; i++ {
			time.Sleep(1 * time.Second)
			
			// Re-fetch device to get updated JID
			device, err = userRepo.GetDeviceByID(deviceID)
			if err != nil {
				return nil, fmt.Errorf("device %s not found: %v", deviceID, err)
			}
			
			if device.JID != "" {
				log.Infof("Device %s JID found after retry %d", deviceID, i+1)
				break
			}
		}
		
		if device.JID == "" {
			return nil, fmt.Errorf("device %s has no JID, QR scan required", deviceID)
		}
	}
	
	// Perform refresh
	log.Infof("🔄 Refreshing device %s connection...", deviceID)
	
	// Parse the JID
	jid, err := types.ParseJID(device.JID)
	if err != nil {
		return nil, fmt.Errorf("invalid JID format: %v", err)
	}
	
	// Get device from store
	waDevice, err := dm.storeContainer.GetDevice(context.Background(), jid)
	if err != nil || waDevice == nil {
		return nil, fmt.Errorf("device not found in store")
	}
	
	// Create new client
	client := whatsmeow.NewClient(waDevice, dm.dbLog)
	client.EnableAutoReconnect = false // We handle reconnection
	client.AutoTrustIdentity = true
	
	// Add event handlers - using the existing infrastructure handler
	client.AddEventHandler(func(evt interface{}) {
		// Import the handler from parent package to avoid circular dependency
		// This should be handled by the existing event system
		log.Debugf("Device %s event: %T", deviceID, evt)
	})
	
	// Connect
	err = client.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	
	// Wait for connection
	for i := 0; i < 10; i++ {
		if client.IsConnected() {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	
	if !client.IsConnected() {
		return nil, fmt.Errorf("failed to establish connection")
	}
	
	// Update device connection in manager
	dm.mu.Lock()
	dm.devices[deviceID] = &DeviceConnection{
		DeviceID:    deviceID,
		UserID:      device.UserID,
		Phone:       device.Phone,
		Client:      client,
		Store:       waDevice,
		Connected:   true,
		ConnectedAt: time.Now().Unix(),
	}
	dm.mu.Unlock()
	
	// Update database status
	userRepo.UpdateDeviceStatus(deviceID, "online", device.Phone, device.JID)
	
	log.Infof("✅ Successfully refreshed device %s", deviceID)
	return client, nil
}

// IsClientHealthy checks if a client is healthy
func (dm *DeviceManager) IsClientHealthy(client *whatsmeow.Client) bool {
	return client != nil && client.IsConnected() && client.IsLoggedIn()
}