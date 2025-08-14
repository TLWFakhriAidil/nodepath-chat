package whatsapp

import (
	"sync"
	"time"
	
	"github.com/sirupsen/logrus"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/multidevice"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/websocket"
)

// MultiDeviceAutoReconnect attempts to reconnect devices after server restart
// Optimized for 3000+ devices with proper throttling
func MultiDeviceAutoReconnect() {
	logrus.Info("Starting multi-device auto-reconnect (optimized for 3000+ devices)...")
	
	userRepo := repository.GetUserRepository()
	db := userRepo.DB()
	
	// Don't set devices offline - just find devices with JID
	// They should remain "online" in database if they have valid sessions
	
	// Find all devices that have a JID (were previously connected)
	rows, err := db.Query(`
		SELECT id, device_name, phone, jid 
		FROM user_devices 
		WHERE jid IS NOT NULL AND jid != ''
		ORDER BY last_seen DESC
		LIMIT 100  -- Process only 100 devices at a time to avoid overwhelming
	`)
	if err != nil {
		logrus.Errorf("Failed to query devices for auto-reconnect: %v", err)
		return
	}
	defer rows.Close()
	
	// Use worker pool pattern for 3000+ devices
	const maxWorkers = 10 // Only 10 concurrent reconnections
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	
	deviceCount := 0
	reconnectCount := 0
	
	for rows.Next() {
		var deviceID, deviceName, phone, jid string
		err := rows.Scan(&deviceID, &deviceName, &phone, &jid)
		if err != nil {
			logrus.Warnf("Failed to scan device row: %v", err)
			continue
		}
		
		deviceCount++
		wg.Add(1)
		
		// Acquire semaphore (wait if all workers are busy)
		semaphore <- struct{}{}
		
		go func(devID, devName, devJID, devPhone string, index int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore
			
			// Stagger reconnections to avoid thundering herd
			time.Sleep(time.Duration(index*2) * time.Second)
			
			if reconnectDevice(devID, devName, devJID, devPhone) {
				reconnectCount++
			}
		}(deviceID, deviceName, jid, phone, deviceCount)
		
		// Additional throttling - wait every 20 devices
		if deviceCount%20 == 0 {
			logrus.Infof("Processed %d devices, waiting before next batch...", deviceCount)
			time.Sleep(5 * time.Second)
		}
	}
	
	wg.Wait()
	logrus.Infof("Auto-reconnect completed: %d/%d devices reconnected", reconnectCount, deviceCount)
}

// reconnectDevice attempts to reconnect a single device using DeviceManager
func reconnectDevice(deviceID, deviceName, jid, phone string) bool {
	logrus.Infof("Checking device %s (%s) - JID: %s", deviceName, deviceID, jid)
	
	// Get DeviceManager instance
	dm := multidevice.GetDeviceManager()
	
	// Try to get or create device connection
	conn, err := dm.GetOrCreateDeviceConnection(deviceID, "", phone)
	if err != nil {
		logrus.Warnf("Failed to get/create connection for device %s: %v", deviceName, err)
		// Don't set offline - device might still have valid session
		return false
	}
	
	// Check if already connected
	if conn.Client.IsConnected() {
		logrus.Infof("✅ Device %s is already connected", deviceName)
		
		// Make sure it's registered with ClientManager
		cm := GetClientManager()
		cm.AddClient(deviceID, conn.Client)
		
		// Update status to online (in case it was marked offline)
		userRepo := repository.GetUserRepository()
		userRepo.UpdateDeviceStatus(deviceID, "online", phone, jid)
		
		// Send success notification
		websocket.Broadcast <- websocket.BroadcastMessage{
			Code:    "DEVICE_RECONNECTED",
			Message: "Device reconnected successfully",
			Result: map[string]interface{}{
				"deviceId": deviceID,
				"phone":    phone,
				"name":     deviceName,
				"status":   "online",
			},
		}
		
		return true
	}
	
	// Try to connect if not connected
	logrus.Infof("Device %s not connected, attempting to connect...", deviceName)
	err = conn.Client.Connect()
	if err != nil {
		logrus.Warnf("Failed to connect device %s: %v", deviceName, err)
		// Don't set offline - keep trying in future attempts
		return false
	}
	
	// Wait a bit for connection to establish
	time.Sleep(3 * time.Second)
	
	// Check if connected and logged in
	if conn.Client.IsConnected() && conn.Client.IsLoggedIn() {
		logrus.Infof("✅ Successfully reconnected device %s", deviceName)
		
		// Register with ClientManager
		cm := GetClientManager()
		cm.AddClient(deviceID, conn.Client)
		
		// Update status
		userRepo := repository.GetUserRepository()
		userRepo.UpdateDeviceStatus(deviceID, "online", phone, jid)
		
		// Send success notification
		websocket.Broadcast <- websocket.BroadcastMessage{
			Code:    "DEVICE_RECONNECTED",
			Message: "Device reconnected successfully",
			Result: map[string]interface{}{
				"deviceId": deviceID,
				"phone":    phone,
				"name":     deviceName,
				"status":   "online",
			},
		}
		
		return true
	}
	
	logrus.Warnf("Device %s failed to establish connection or login", deviceName)
	// Don't set offline - session might still be valid for next attempt
	return false
}

// StartMultiDeviceAutoReconnect starts the auto-reconnect process with proper delays
func StartMultiDeviceAutoReconnect() {
	go func() {
		// Wait for server to fully initialize (60 seconds for 3000 devices)
		time.Sleep(60 * time.Second)
		
		// Run initial reconnect
		MultiDeviceAutoReconnect()
		
		// Run periodic checks every 30 minutes (not too frequent for 3000 devices)
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			logrus.Info("Running periodic multi-device reconnect check...")
			MultiDeviceAutoReconnect()
		}
	}()
	
	logrus.Info("Multi-device auto-reconnect scheduled (60s delay, 30min intervals)")
}