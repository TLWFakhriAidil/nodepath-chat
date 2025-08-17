package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/sirupsen/logrus"
)

// ClearDeviceSession clears WhatsApp session for a specific device
func (handler *App) ClearDeviceSession(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device ID is required",
		})
	}
	
	// Get session from cookie
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
	
	// Get device details
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Verify device belongs to user
	if device.UserID != session.UserID {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "Device does not belong to this user",
		})
	}
	
	logrus.Infof("Clearing WhatsApp session for device %s (%s)", device.ID, device.DeviceName)
	
	// Disconnect WhatsApp client
	cm := whatsapp.GetClientManager()
	var jid string
	var phone string
	
	// Get phone and JID BEFORE doing anything
	if client, err := cm.GetClient(deviceID); err == nil && client != nil {
		// Get JID and phone before any disconnect/logout
		if client.Store != nil && client.Store.ID != nil {
			jid = client.Store.ID.String()
			phone = client.Store.ID.User
			logrus.Infof("Got from client - Phone: %s, JID: %s", phone, jid)
		}
		
		// Logout from WhatsApp to remove from linked devices
		if client.IsConnected() {
			err = client.Logout(c.UserContext())
			if err != nil {
				logrus.Errorf("Error logging out from WhatsApp: %v", err)
			} else {
				logrus.Info("Successfully logged out from WhatsApp (removed from linked devices)")
			}
		}
		
		// Disconnect client
		client.Disconnect()
		
		// Remove from client manager
		cm.RemoveClient(deviceID)
		
		logrus.Info("WhatsApp client disconnected and removed from manager")
	}
	
	// If we couldn't get from client, get from database
	if phone == "" && device.Phone != "" {
		phone = device.Phone
	}
	if jid == "" && device.JID != "" {
		jid = device.JID
	}
	
	// Clear WhatsApp session tables for this device
	// This should happen AFTER logout to ensure clean state
	err = whatsapp.ClearWhatsAppSessionData(deviceID)
	if err != nil {
		logrus.Errorf("Error clearing WhatsApp session: %v", err)
	}
	
	// Log current state
	logrus.Infof("Device before update - DB Phone: %s, DB JID: %s", device.Phone, device.JID)
	logrus.Infof("Will update with - Phone: %s, JID: %s", phone, jid)
	
	// IMPORTANT: Even after logout, we keep phone/JID for easier reconnection
	// The user can scan QR with same device and it will update to online
	
	// Update device status to offline but KEEP phone and JID
	err = userRepo.UpdateDeviceStatus(deviceID, "offline", phone, jid)
	if err != nil {
		logrus.Errorf("Error updating device status: %v", err)
	}
	
	logrus.Infof("Device logout complete - preserved Phone: %s, JID: %s for reconnection", phone, jid)
	
	// Clear any connection session
	whatsapp.ClearConnectionSession(session.UserID)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device session cleared successfully",
		Results: map[string]interface{}{
			"device_id": deviceID,
			"device_name": device.DeviceName,
			"status": "offline",
		},
	})
}