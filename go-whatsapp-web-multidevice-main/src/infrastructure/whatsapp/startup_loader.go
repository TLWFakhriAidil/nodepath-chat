package whatsapp

import (
	"context"
	"database/sql"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/multidevice"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// LoadAllDevicesOnStartup loads all WhatsApp devices from database on server startup
func LoadAllDevicesOnStartup() {
	logrus.Info("=== LOADING ALL WHATSAPP DEVICES ON STARTUP ===")
	
	db := database.GetDB()
	
	// Get all non-platform devices with JID
	rows, err := db.Query(`
		SELECT id, user_id, device_name, phone, jid, status
		FROM user_devices 
		WHERE (platform IS NULL OR platform = '')
		AND jid IS NOT NULL AND jid != ''
		ORDER BY device_name
	`)
	if err != nil {
		logrus.Errorf("Failed to get devices: %v", err)
		return
	}
	defer rows.Close()
	
	// Initialize WhatsApp database
	dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)
	container, err := sqlstore.New(context.Background(), "postgres", config.DBURI, dbLog)
	if err != nil {
		logrus.Errorf("Failed to initialize WhatsApp database: %v", err)
		return
	}
	
	deviceManager := multidevice.GetDeviceManager()
	clientManager := GetClientManager()
	
	loadedCount := 0
	failedCount := 0
	
	for rows.Next() {
		var deviceID, userID, deviceName, phone, jid, status sql.NullString
		
		err := rows.Scan(&deviceID, &userID, &deviceName, &phone, &jid, &status)
		if err != nil {
			logrus.Errorf("Failed to scan device row: %v", err)
			continue
		}
		
		if !deviceID.Valid || !userID.Valid || !jid.Valid {
			continue
		}
		
		logrus.Infof("Loading device: %s (JID: %s)", deviceName.String, jid.String)
		
		// Try to get existing device from WhatsApp database
		devices, err := container.GetAllDevices(context.Background())
		if err != nil {
			logrus.Errorf("Failed to get WhatsApp devices: %v", err)
			continue
		}
		
		var waDevice *store.Device
		for _, dev := range devices {
			if dev.ID != nil && dev.ID.String() == jid.String {
				waDevice = dev
				break
			}
		}
		
		if waDevice == nil {
			logrus.Warnf("No WhatsApp session found for device %s", deviceName.String)
			failedCount++
			continue
		}
		
		// Create WhatsApp client
		client := whatsmeow.NewClient(waDevice, dbLog)
		
		// Disable auto-reconnect to prevent issues
		client.EnableAutoReconnect = false
		client.AutoTrustIdentity = true
		
		// Register device with manager
		deviceManager.RegisterDevice(deviceID.String, userID.String, phone.String, client)
		
		// Add to client manager
		clientManager.AddClient(deviceID.String, client)
		
		// Try to connect if logged in
		if client.IsLoggedIn() {
			logrus.Infof("Device %s is logged in, attempting to connect...", deviceName.String)
			
			err := client.Connect()
			if err != nil {
				logrus.Errorf("Failed to connect device %s: %v", deviceName.String, err)
				// Update status to offline
				db.Exec("UPDATE user_devices SET status = 'offline' WHERE id = ?", deviceID.String)
			} else {
				// Wait a bit for connection
				time.Sleep(3 * time.Second)
				
				if client.IsConnected() {
					logrus.Infof("✓ Device %s connected successfully", deviceName.String)
					// Update status to online
					db.Exec("UPDATE user_devices SET status = 'online' WHERE id = ?", deviceID.String)
					loadedCount++
				} else {
					logrus.Warnf("Device %s failed to establish connection", deviceName.String)
					db.Exec("UPDATE user_devices SET status = 'offline' WHERE id = ?", deviceID.String)
					failedCount++
				}
			}
		} else {
			logrus.Infof("Device %s needs QR code scan", deviceName.String)
			db.Exec("UPDATE user_devices SET status = 'offline' WHERE id = ?", deviceID.String)
			failedCount++
		}
	}
	
	logrus.Info("=== DEVICE LOADING COMPLETED ===")
	logrus.Infof("Successfully loaded: %d devices", loadedCount)
	logrus.Infof("Failed/Need QR: %d devices", failedCount)
}
