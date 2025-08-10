package whatsapp

import (
	"context"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// ReconnectDeviceByJID attempts to reconnect a device using its stored JID
func ReconnectDeviceByJID(deviceID string) error {
	userRepo := repository.GetUserRepository()
	
	// Get device details
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get device %s: %v", deviceID, err)
		return err
	}
	
	// Skip if no JID
	if device.JID == "" {
		logrus.Warnf("Device %s has no JID - needs QR scan", device.DeviceName)
		return nil
	}
	
	// Skip platform devices
	if device.Platform != "" {
		logrus.Debugf("Skipping platform device %s", device.DeviceName)
		return nil
	}
	
	// Check if already connected
	cm := GetClientManager()
	if existingClient, err := cm.GetClient(deviceID); err == nil && existingClient != nil {
		if existingClient.IsConnected() {
			logrus.Infof("Device %s is already connected", device.DeviceName)
			return nil
		}
		// Remove old client
		cm.RemoveClient(deviceID)
	}
	
	// Don't try to reconnect if marked offline recently (within 5 minutes)
	if !device.LastSeen.IsZero() && time.Since(device.LastSeen) < 5*time.Minute {
		logrus.Debugf("Device %s was offline recently, skipping reconnect", device.DeviceName)
		userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
		return nil
	}
	
	// Initialize WhatsApp container
	waLogger := waLog.Stdout("Database", config.WhatsappLogLevel, true)
	container, err := sqlstore.New(context.Background(), "postgres", config.DBURI, waLogger)
	if err != nil {
		logrus.Errorf("Failed to create WhatsApp container: %v", err)
		return err
	}
	
	// Try to get the device from store
	jid := parseJID(device.JID)
	waDevice, err := container.GetDevice(context.Background(), jid)
	if err != nil {
		logrus.Errorf("Failed to get device from store: %v", err)
		userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
		return err
	}
	
	// Create client
	client := whatsmeow.NewClient(waDevice, waLogger)
	
	// Add event handler
	client.AddEventHandler(func(evt interface{}) {
		// Process asynchronously
		go HandleDeviceEvent(context.Background(), deviceID, evt)
	})
	
	// Try to connect
	logrus.Infof("Connecting device %s...", device.DeviceName)
	err = client.Connect()
	if err != nil {
		logrus.Errorf("Failed to connect device %s: %v", device.DeviceName, err)
		userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
		return err
	}
	
	// Wait for connection
	time.Sleep(3 * time.Second)
	
	// Verify connection status
	if !client.IsConnected() {
		logrus.Warnf("Device %s failed to establish connection", device.DeviceName)
		client.Disconnect()
		userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
		return nil
	}
	
	// Check if logged in
	if !client.IsLoggedIn() {
		logrus.Warnf("Device %s connected but not logged in - session expired", device.DeviceName)
		client.Disconnect()
		userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
		return nil
	}
	
	// Success! Register with client manager
	cm.AddClient(deviceID, client)
	
	// Update device status
	actualPhone := device.Phone
	actualJID := device.JID
	if client.Store.ID != nil {
		actualPhone = client.Store.ID.User
		actualJID = client.Store.ID.String()
	}
	
	err = userRepo.UpdateDeviceStatus(deviceID, "online", actualPhone, actualJID)
	if err != nil {
		logrus.Errorf("Failed to update device status: %v", err)
	}
	
	logrus.Infof("✅ Successfully reconnected device %s (%s)", device.DeviceName, deviceID)
	
	return nil
}

// parseJID converts a string JID to types.JID
func parseJID(jid string) types.JID {
	parsed, _ := types.ParseJID(jid)
	return parsed
}

// RefreshDeviceConnectionByID refreshes a device connection by ID
// This is an alias for ReconnectDeviceByJID to maintain compatibility
func RefreshDeviceConnectionByID(deviceID string) error {
	return ReconnectDeviceByJID(deviceID)
}