package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/sirupsen/logrus"
)

// CheckDeviceConnectionStatus checks real-time connection status of all devices
func CheckDeviceConnectionStatus(c *fiber.Ctx) error {
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
	
	// Get all devices for this user
	devices, err := userRepo.GetUserDevices(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	// Check connection status for each device
	cm := whatsapp.GetClientManager()
	deviceStatuses := make([]map[string]interface{}, 0)
	
	for _, device := range devices {
		status := "offline"
		phone := ""
		jid := ""
		
		// Check if client exists and is connected
		if client, err := cm.GetClient(device.ID); err == nil && client != nil {
			if client.IsConnected() {
				status = "connected"
				
				// Get phone and JID info
				if client.Store != nil && client.Store.ID != nil {
					jid = client.Store.ID.String()
					phone = client.Store.ID.User
				}
			}
		}
		
		// Update database if status changed
		if device.Status != status || device.Phone != phone || device.JID != jid {
			err = userRepo.UpdateDeviceStatus(device.ID, status, phone, jid)
			if err != nil {
				logrus.Errorf("Failed to update device status for %s: %v", device.ID, err)
			}
		}
		
		deviceStatuses = append(deviceStatuses, map[string]interface{}{
			"device_id":   device.ID,
			"device_name": device.DeviceName,
			"status":      status,
			"phone":       phone,
			"jid":         jid,
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device connection status checked",
		Results: deviceStatuses,
	})
}
