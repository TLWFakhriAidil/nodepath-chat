package whatsapp

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/stability"
	"github.com/sirupsen/logrus"
)

// InitializeUltraStableMode sets up ultra stable connections for all devices
func InitializeUltraStableMode() {
	stabilityConfig := config.GetStabilityConfig()
	
	if !stabilityConfig.UltraStableMode {
		logrus.Info("Ultra Stable Mode is DISABLED")
		return
	}
	
	logrus.Info("🚀🚀🚀 ULTRA STABLE MODE ACTIVATED 🚀🚀🚀")
	logrus.Info("Devices will NEVER disconnect")
	logrus.Info("Rate limits are IGNORED")
	logrus.Info("Maximum speed messaging ENABLED")
	
	// Get ultra stable instance
	ultraStable := stability.GetUltraStableConnection()
	
	// Get all existing clients
	clientManager := GetClientManager()
	allClients := clientManager.GetAllClients()
	
	// Register all existing clients for ultra stable
	for deviceID, client := range allClients {
		if client != nil {
			logrus.Infof("Registering device %s for ULTRA STABLE connection", deviceID)
			ultraStable.RegisterClient(deviceID, client)
			ultraStable.DisableDisconnection(deviceID)
		}
	}
	
	// Force all devices online
	ultraStable.ForceAllOnline()
	
	logrus.Infof("✅ %d devices registered for ULTRA STABLE mode", len(allClients))
}

// EnsureDeviceUltraStable ensures a specific device is in ultra stable mode
func EnsureDeviceUltraStable(deviceID string) {
	stabilityConfig := config.GetStabilityConfig()
	
	if !stabilityConfig.UltraStableMode {
		return
	}
	
	ultraStable := stability.GetUltraStableConnection()
	clientManager := GetClientManager()
	
	// Check if already registered
	if _, err := ultraStable.GetStableClient(deviceID); err == nil {
		// Already registered
		return
	}
	
	// Get client from manager
	client, err := clientManager.GetClient(deviceID)
	if err != nil {
		logrus.Warnf("Device %s not found in client manager for ultra stable registration", deviceID)
		return
	}
	
	// Register for ultra stable
	logrus.Infof("Registering device %s for ULTRA STABLE connection", deviceID)
	ultraStable.RegisterClient(deviceID, client)
	ultraStable.DisableDisconnection(deviceID)
}
