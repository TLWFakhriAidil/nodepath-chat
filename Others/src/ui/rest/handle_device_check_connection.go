package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// HandleDeviceCheckConnection handles real-time device connection status checks
func HandleDeviceCheckConnection(c *fiber.Ctx) error {
	// Get session from cookie
	cookie := c.Cookies("session_token")
	if cookie == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Validate session
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(cookie)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid session",
		})
	}

	// Get user devices
	devices, err := userRepo.GetUserDevices(session.UserID)
	if err != nil {
		logrus.Errorf("Failed to get user devices: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get devices",
		})
	}

	// Get client manager
	clientMgr := whatsapp.GetClientManager()
	
	// Check each device connection
	deviceStatuses := make([]map[string]interface{}, 0)
	
	for _, device := range devices {
		// Get client for device
		client, err := clientMgr.GetClient(device.ID)
		
		status := device.Status
		isConnected := false
		needsQR := false
		
		if err != nil || client == nil {
			// No client exists
			status = "offline"
			needsQR = true
		} else {
			// Check actual connection
			if client.IsConnected() {
				status = "online"
				isConnected = true
				
				// Update phone/JID if available
				if client.Store != nil && client.Store.ID != nil {
					device.JID = client.Store.ID.String()
					device.Phone = client.Store.ID.User
					
					// Update in database
					userRepo.UpdateDeviceStatus(device.ID, "online", device.Phone, device.JID)
				}
			} else if client.IsLoggedIn() {
				// Logged in but not connected - try to reconnect
				status = "offline"
				needsQR = false
				
				// Trigger reconnection in background
				go func(deviceID string, cli whatsmeow.Client) {
					logrus.Infof("Attempting to reconnect device %s", device.DeviceName)
					if err := cli.Connect(); err == nil {
						logrus.Infof("Device %s reconnected successfully", device.DeviceName)
					}
				}(device.ID, *client)
			} else {
				// Not logged in
				status = "offline"
				needsQR = true
			}
		}
		
		// Update status if changed
		if status != device.Status {
			userRepo.UpdateDeviceStatus(device.ID, status, device.Phone, device.JID)
		}
		
		deviceStatuses = append(deviceStatuses, map[string]interface{}{
			"id":           device.ID,
			"device_name":  device.DeviceName,
			"phone":        device.Phone,
			"jid":          device.JID,
			"status":       status,
			"is_connected": isConnected,
			"needs_qr":     needsQR,
			"last_seen":    device.LastSeen,
		})
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Device connection status",
		"data":    deviceStatuses,
	})
}
