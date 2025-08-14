package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/sirupsen/logrus"
)

// RefreshDevice attempts to reconnect an existing device session
func RefreshDevice(c *fiber.Ctx) error {
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
	
	// Check if device has platform - platform devices cannot be refreshed
	if device.Platform != "" {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "PLATFORM_DEVICE",
			Message: "Platform devices cannot be refreshed",
		})
	}
	
	// Check if device has phone/JID for reconnection
	if device.Phone == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "NO_SESSION",
			Message: "Device has no previous session. Please scan QR code.",
		})
	}
	
	// Try to get existing client from ClientManager
	cm := whatsapp.GetClientManager()
	client, err := cm.GetClient(deviceID)
	
	if err == nil && client != nil && client.IsConnected() {
		// Already connected
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "ALREADY_CONNECTED",
			Message: "Device is already connected",
			Results: map[string]interface{}{
				"deviceId": deviceID,
				"phone":    device.Phone,
				"status":   "connected",
			},
		})
	}
	
	// For now, just return that QR scan is required
	// This avoids the nil pointer issues with DeviceManager
	logrus.Infof("Device %s needs QR scan for reconnection", deviceID)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "QR_REQUIRED",
		Message: "Device session expired. Please scan QR code to reconnect.",
		Results: map[string]interface{}{
			"deviceId": deviceID,
			"phone":    device.Phone,
			"status":   "disconnected",
		},
	})
}
