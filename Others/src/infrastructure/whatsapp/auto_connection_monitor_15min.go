package whatsapp

import (
	"context"
	"sync"
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// AutoConnectionMonitor monitors and reconnects devices every 15 minutes
type AutoConnectionMonitor struct {
	mu              sync.RWMutex
	isRunning       bool
	checkInterval   time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
	userRepo        *repository.UserRepository
	clientManager   *ClientManager
}

var (
	autoMonitor     *AutoConnectionMonitor
	autoMonitorOnce sync.Once
)

// GetAutoConnectionMonitor returns singleton instance
func GetAutoConnectionMonitor() *AutoConnectionMonitor {
	autoMonitorOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		autoMonitor = &AutoConnectionMonitor{
			checkInterval:  15 * time.Minute, // Check every 15 minutes
			ctx:           ctx,
			cancel:        cancel,
			userRepo:      repository.GetUserRepository(),
			clientManager: GetClientManager(),
		}
	})
	return autoMonitor
}

// Start begins automatic connection monitoring
func (acm *AutoConnectionMonitor) Start() error {
	logrus.Info("Auto connection monitor DISABLED - no auto reconnect")
	return nil // DISABLED - No auto reconnect
	/*
	acm.mu.Lock()
	defer acm.mu.Unlock()
	
	if acm.isRunning {
		return nil
	}
	
	acm.isRunning = true
	go acm.monitorLoop()
	
	logrus.Info("Auto connection monitor started - checking every 15 minutes")
	return nil
	*/
}

// Stop stops the monitoring
func (acm *AutoConnectionMonitor) Stop() {
	acm.mu.Lock()
	defer acm.mu.Unlock()
	
	if !acm.isRunning {
		return
	}
	
	acm.cancel()
	acm.isRunning = false
	logrus.Info("Auto connection monitor stopped")
}

// monitorLoop is the main monitoring loop
func (acm *AutoConnectionMonitor) monitorLoop() {
	// Initial check after 1 minute (let system stabilize first)
	time.Sleep(1 * time.Minute)
	acm.checkAndReconnectDevices()
	
	ticker := time.NewTicker(acm.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-acm.ctx.Done():
			return
		case <-ticker.C:
			acm.checkAndReconnectDevices()
		}
	}
}

// checkAndReconnectDevices checks all devices and attempts ONE reconnection for offline devices
func (acm *AutoConnectionMonitor) checkAndReconnectDevices() {
	startTime := time.Now()
	logrus.Info("Starting 15-minute device check and reconnection...")
	
	// Get all devices from database
	devices, err := acm.userRepo.GetAllDevices()
	if err != nil {
		logrus.Errorf("Failed to get all devices: %v", err)
		return
	}
	
	totalDevices := len(devices)
	onlineCount := 0
	offlineCount := 0
	reconnectAttempted := 0
	reconnectSuccess := 0
	skippedPlatform := 0
	
	// Process each device
	for _, device := range devices {
		// Skip devices with platform value
		if device.Platform != "" {
			skippedPlatform++
			logrus.Debugf("Skipping device %s - has platform: %s", device.DeviceName, device.Platform)
			continue
		}
		// Get WhatsApp client
		client, err := acm.clientManager.GetClient(device.ID)
		
		if err != nil || client == nil {
			// No client exists - mark as offline
			if device.Status != "offline" {
				acm.userRepo.UpdateDeviceStatus(device.ID, "offline", device.Phone, device.JID)
				logrus.Debugf("Device %s has no client - marked offline", device.DeviceName)
			}
			offlineCount++
			continue
		}
		
		// Check if connected
		if client.IsConnected() {
			// Device is online
			onlineCount++
			if device.Status != "online" {
				acm.userRepo.UpdateDeviceStatus(device.ID, "online", device.Phone, device.JID)
				logrus.Infof("Device %s is online - status updated", device.DeviceName)
			}
			continue
		}
		
		// Device is offline
		offlineCount++
		
		// Check if we can reconnect (client exists, is logged in, but not connected)
		if client.IsLoggedIn() {
			logrus.Infof("Device %s is offline but logged in - attempting ONE reconnection...", device.DeviceName)
			reconnectAttempted++
			
			// Try to reconnect ONCE
			err := client.Connect()
			if err != nil {
				logrus.Warnf("Failed to reconnect device %s: %v", device.DeviceName, err)
				// Update status to offline
				acm.userRepo.UpdateDeviceStatus(device.ID, "offline", device.Phone, device.JID)
			} else {
				// Wait a bit for connection to establish
				time.Sleep(3 * time.Second)
				
				// Check if now connected
				if client.IsConnected() {
					reconnectSuccess++
					logrus.Infof("✅ Successfully reconnected device %s", device.DeviceName)
					
					// Update phone/JID if available
					if client.Store != nil && client.Store.ID != nil {
						newJID := client.Store.ID.String()
						newPhone := client.Store.ID.User
						acm.userRepo.UpdateDeviceStatus(device.ID, "online", newPhone, newJID)
					} else {
						acm.userRepo.UpdateDeviceStatus(device.ID, "online", device.Phone, device.JID)
					}
				} else {
					logrus.Warnf("Device %s reconnection failed - still offline", device.DeviceName)
					acm.userRepo.UpdateDeviceStatus(device.ID, "offline", device.Phone, device.JID)
				}
			}
		} else {
			// Not logged in - can't auto-reconnect
			logrus.Debugf("Device %s is logged out - needs QR scan", device.DeviceName)
			if device.Status != "offline" {
				acm.userRepo.UpdateDeviceStatus(device.ID, "offline", device.Phone, device.JID)
			}
		}
		
		// Small delay between devices to avoid overwhelming system
		time.Sleep(100 * time.Millisecond)
	}
	
	duration := time.Since(startTime)
	
	// Log summary
	logrus.Infof("15-minute check complete: %d devices, %d online, %d offline, %d reconnect attempted, %d reconnect success, %d skipped (platform) (took %v)",
		totalDevices, onlineCount, offlineCount, reconnectAttempted, reconnectSuccess, skippedPlatform, duration)
	
	// Log detailed stats if there were changes
	if reconnectAttempted > 0 {
		successRate := float64(reconnectSuccess) / float64(reconnectAttempted) * 100
		logrus.Infof("Reconnection success rate: %.1f%% (%d/%d)", successRate, reconnectSuccess, reconnectAttempted)
	}
}

// GetStatus returns current monitor status
func (acm *AutoConnectionMonitor) GetStatus() map[string]interface{} {
	acm.mu.RLock()
	defer acm.mu.RUnlock()
	
	return map[string]interface{}{
		"is_running":     acm.isRunning,
		"check_interval": acm.checkInterval.String(),
		"next_check":     time.Now().Add(acm.checkInterval).Format(time.RFC3339),
	}
}

// ForceCheck triggers an immediate check (for testing/manual trigger)
func (acm *AutoConnectionMonitor) ForceCheck() {
	go acm.checkAndReconnectDevices()
}