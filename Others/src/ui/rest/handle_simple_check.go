package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// SimpleCheckConnection - Fixed version of device connection check
func SimpleCheckConnection(c *fiber.Ctx) error {
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

	// Build response with device statuses
	deviceStatuses := make([]map[string]interface{}, 0)
	
	for _, device := range devices {
		deviceStatuses = append(deviceStatuses, map[string]interface{}{
			"id":           device.ID,
			"device_name":  device.DeviceName,
			"phone":        device.Phone,
			"jid":          device.JID,
			"status":       device.Status,
			"last_seen":    device.LastSeen,
		})
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Device connection status",
		"data":    deviceStatuses,
	})
}
