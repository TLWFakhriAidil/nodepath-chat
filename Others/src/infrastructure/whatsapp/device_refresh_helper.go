package whatsapp

import (
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// RefreshDeviceConnection attempts to refresh a device connection and ensures proper status
func RefreshDeviceConnection(deviceID string, client *whatsmeow.Client) {
	logrus.Infof("🔄 Starting device refresh for %s", deviceID)
	
	// Get user repo to update status
	userRepo := repository.GetUserRepository()
	
	// Get device info
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get device info: %v", err)
		userRepo.UpdateDeviceStatus(deviceID, "offline", "", "")
		return
	}
	
	// Try to connect if client exists
	if client != nil {
		if client.IsConnected() {
			logrus.Infof("✅ Device %s is already connected", device.DeviceName)
			userRepo.UpdateDeviceStatus(deviceID, "online", device.Phone, device.JID)
			return
		}
		
		// Try to connect
		logrus.Infof("Attempting to connect device %s", device.DeviceName)
		err := client.Connect()
		if err != nil {
			logrus.Warnf("Failed to connect device %s: %v", device.DeviceName, err)
			userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
			return
		}
		
		// Wait for connection
		time.Sleep(3 * time.Second)
		
		// Check final status
		if client.IsConnected() {
			logrus.Infof("✅ Device %s successfully connected", device.DeviceName)
			
			// Update with new info if available
			if client.Store != nil && client.Store.ID != nil {
				newJID := client.Store.ID.String()
				newPhone := client.Store.ID.User
				userRepo.UpdateDeviceStatus(deviceID, "online", newPhone, newJID)
			} else {
				userRepo.UpdateDeviceStatus(deviceID, "online", device.Phone, device.JID)
			}
		} else {
			logrus.Warnf("❌ Device %s failed to connect", device.DeviceName)
			userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
		}
	} else {
		// No client available
		logrus.Warnf("No client available for device %s", device.DeviceName)
		userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
	}
}

// EnsureDeviceStatus makes sure device has proper online/offline status
func EnsureDeviceStatus(deviceID string) {
	userRepo := repository.GetUserRepository()
	
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		return
	}
	
	// If status is not online or offline, set to offline
	if device.Status != "online" && device.Status != "offline" {
		logrus.Infof("Normalizing device %s status from '%s' to 'offline'", device.DeviceName, device.Status)
		userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
	}
}
