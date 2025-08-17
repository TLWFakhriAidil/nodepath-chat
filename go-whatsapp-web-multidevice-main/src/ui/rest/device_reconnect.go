package rest

import (
	"context"
	"database/sql"
	"strings"
	"time"
	
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/multidevice"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "github.com/lib/pq"
)

// ReconnectDeviceSession attempts to reconnect using existing WhatsApp session
func ReconnectDeviceSession(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device ID is required",
		})
	}
	
	// Get user from session
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token",
		})
	}
	
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}
	
	// Verify device ownership
	device, err := userRepo.GetUserDevice(session.UserID, deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Check if device has JID for reconnection
	if device.JID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "NO_SESSION",
			Message: "Device has no JID. Please scan QR code to connect.",
		})
	}
	
	// First check if already connected in ClientManager
	cm := whatsapp.GetClientManager()
	existingClient, _ := cm.GetClient(deviceID)
	if existingClient != nil && existingClient.IsConnected() {
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "ALREADY_CONNECTED",
			Message: "Device is already connected",
			Results: map[string]interface{}{
				"deviceId": deviceID,
				"phone":    device.Phone,
				"jid":      device.JID,
				"status":   "connected",
			},
		})
	}
	
	logrus.Infof("Attempting to reconnect device %s with JID %s...", deviceID, device.JID)
	
	// Check if we have WhatsApp session data in the database
	db, err := sql.Open("postgres", config.DBURI)
	if err != nil {
		logrus.Errorf("Failed to open database: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Database connection failed",
		})
	}
	defer db.Close()
	
	// Query whatsmeow_sessions table using the JID
	var sessionData []byte
	err = db.QueryRow(`
		SELECT session 
		FROM whatsmeow_sessions 
		WHERE our_jid = ?
		LIMIT 1
	`, device.JID).Scan(&sessionData)
	
	if err != nil {
		if err == sql.ErrNoRows {
			logrus.Infof("No session found in whatsmeow_sessions for JID %s", device.JID)
			return c.JSON(utils.ResponseData{
				Status:  200,
				Code:    "QR_REQUIRED",
				Message: "Session not found in database. Please scan QR code.",
				Results: map[string]interface{}{
					"deviceId": deviceID,
					"jid":      device.JID,
					"reason":   "no_session_in_database",
				},
			})
		}
		logrus.Errorf("Error querying session: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to query session data",
		})
	}
	
	logrus.Infof("Found session data in database for JID %s", device.JID)
	
	// Initialize WhatsApp store container
	ctx := context.Background()
	dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)
	
	// Convert postgresql:// to postgres:// for whatsmeow
	dbURI := config.DBURI
	if strings.HasPrefix(dbURI, "postgresql://") {
		dbURI = strings.Replace(dbURI, "postgresql://", "postgres://", 1)
	}
	
	container, err := sqlstore.New(ctx, "postgres", dbURI, dbLog)
	if err != nil {
		logrus.Errorf("Failed to create store container: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to initialize WhatsApp store",
		})
	}
	
	// Parse the JID to get the proper device
	jid, err := types.ParseJID(device.JID)
	if err != nil {
		logrus.Errorf("Failed to parse JID %s: %v", device.JID, err)
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "QR_REQUIRED",
			Message: "Invalid JID format. Please scan QR code.",
			Results: map[string]interface{}{
				"deviceId": deviceID,
				"jid":      device.JID,
				"reason":   "invalid_jid",
			},
		})
	}
	
	// Get device by JID from store
	waDevice, err := container.GetDevice(ctx, jid)
	if err != nil {
		logrus.Errorf("Failed to get device from store: %v", err)
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "QR_REQUIRED",
			Message: "Device not found in store. Please scan QR code.",
			Results: map[string]interface{}{
				"deviceId": deviceID,
				"jid":      device.JID,
				"reason":   "device_not_in_store",
			},
		})
	}
	
	if waDevice == nil {
		logrus.Warnf("Device is nil for JID %s", device.JID)
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "QR_REQUIRED",
			Message: "Session data missing. Please scan QR code.",
			Results: map[string]interface{}{
				"deviceId": deviceID,
				"jid":      device.JID,
				"reason":   "null_device",
			},
		})
	}
	
	logrus.Infof("Found WhatsApp device for JID %s", device.JID)
	
	// Create client with stored device
	client := whatsmeow.NewClient(waDevice, waLog.Stdout("Client", config.WhatsappLogLevel, true))
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true
	
	// IMPORTANT: Register with DeviceManager BEFORE adding event handlers
	// This ensures the event handler can find the device
	dm := multidevice.GetDeviceManager()
	dm.RegisterDevice(deviceID, device.UserID, device.Phone, client)
	
	// Add event handlers AFTER registration
	client.AddEventHandler(func(evt interface{}) {
		whatsapp.HandleDeviceEvent(context.Background(), deviceID, evt)
	})
	
	// Try to connect
	err = client.Connect()
	if err != nil {
		logrus.Errorf("Failed to connect: %v", err)
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "QR_REQUIRED",
			Message: "Failed to reconnect. Please scan QR code.",
			Results: map[string]interface{}{
				"deviceId": deviceID,
				"phone":    device.Phone,
				"reason":   "connection_failed",
				"error":    err.Error(),
			},
		})
	}
	
	// Wait a bit for connection to establish
	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	connected := false
	for i := 0; i < 20; i++ {
		if client.IsConnected() && client.IsLoggedIn() {
			connected = true
			break
		}
		select {
		case <-ctx2.Done():
			break
		case <-time.After(500 * time.Millisecond):
			continue
		}
	}
	
	if connected {
		// Don't register here - already done by event handler
		// Just update status and return success
		
		// But ensure it's registered in MultiDeviceManager as backup
		mdm := whatsapp.GetMultiDeviceManager()
		mdm.EnsureDeviceRegistered(deviceID, client)
		
		// Update device status
		jidStr := ""
		if client.Store.ID != nil {
			jidStr = client.Store.ID.String()
		}
		userRepo.UpdateDeviceStatus(deviceID, "online", device.Phone, jidStr)
		
		logrus.Infof("✅ Successfully reconnected device %s", deviceID)
		
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Device reconnected successfully!",
			Results: map[string]interface{}{
				"deviceId": deviceID,
				"phone":    device.Phone,
				"jid":      jidStr,
				"status":   "connected",
			},
		})
	}
	
	// Not connected after timeout
	client.Disconnect()
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "QR_REQUIRED",
		Message: "Session expired. Please scan QR code.",
		Results: map[string]interface{}{
			"deviceId": deviceID,
			"phone":    device.Phone,
			"reason":   "login_timeout",
		},
	})
}
