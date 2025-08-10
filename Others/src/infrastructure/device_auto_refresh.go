package infrastructure

import (
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// DeviceAutoRefreshHandler handles device auto-refresh globally
type DeviceAutoRefreshHandler struct {
	refreshing map[string]bool
	userRepo   *repository.UserRepository
}

var globalRefreshHandler *DeviceAutoRefreshHandler

// InitDeviceAutoRefresh initializes the global refresh handler
func InitDeviceAutoRefresh() {
	globalRefreshHandler = &DeviceAutoRefreshHandler{
		refreshing: make(map[string]bool),
		userRepo:   repository.GetUserRepository(),
	}
	logrus.Info("Device auto-refresh handler initialized")
}

// TriggerDeviceRefresh triggers auto-refresh for a device
func TriggerDeviceRefresh(deviceID string) {
	if globalRefreshHandler == nil {
		InitDeviceAutoRefresh()
	}
	
	// Check if already refreshing
	if globalRefreshHandler.refreshing[deviceID] {
		logrus.Debugf("Device %s is already being refreshed", deviceID)
		return
	}
	
	globalRefreshHandler.refreshing[deviceID] = true
	
	// Run refresh in background
	go func() {
		defer func() {
			delete(globalRefreshHandler.refreshing, deviceID)
		}()
		
		logrus.Infof("🔄 Auto-refresh triggered for device %s", deviceID)
		
		// Get device info
		device, err := globalRefreshHandler.userRepo.GetDeviceByID(deviceID)
		if err != nil {
			logrus.Errorf("Failed to get device info for auto-refresh: %v", err)
			return
		}
		
		// Log the refresh attempt
		logrus.Infof("Attempting to refresh device %s (%s)", device.DeviceName, device.JID)
		
		// The actual reconnection will be handled by the connection logic
		// We just need to ensure the device ends up with proper status
		
		// Wait for reconnection attempt
		time.Sleep(10 * time.Second)
		
		// Check final status and ensure it's either online or offline
		updatedDevice, err := globalRefreshHandler.userRepo.GetDeviceByID(deviceID)
		if err != nil {
			logrus.Errorf("Failed to check device status after refresh: %v", err)
			globalRefreshHandler.userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
			return
		}
		
		// If status is still "refreshing" or anything other than online/offline, set to offline
		if updatedDevice.Status != "online" && updatedDevice.Status != "offline" {
			logrus.Warnf("Device %s refresh completed - setting status to offline", device.DeviceName)
			globalRefreshHandler.userRepo.UpdateDeviceStatus(deviceID, "offline", device.Phone, device.JID)
		} else {
			logrus.Infof("Device %s refresh completed - status: %s", device.DeviceName, updatedDevice.Status)
		}
	}()
}

// IsDeviceRefreshing checks if a device is currently being refreshed
func IsDeviceRefreshing(deviceID string) bool {
	if globalRefreshHandler == nil {
		return false
	}
	return globalRefreshHandler.refreshing[deviceID]
}
