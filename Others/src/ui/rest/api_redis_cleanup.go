package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// InitRedisCleanupAPI initializes Redis cleanup endpoints
func InitRedisCleanupAPI(app *fiber.App) {
	app.Post("/api/redis/cleanup-device/:deviceId", CleanupDeviceFromRedis)
	app.Post("/api/redis/cleanup-old-devices", CleanupAllOldDevices)
}

// CleanupDeviceFromRedis removes all Redis data for a specific device
func CleanupDeviceFromRedis(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	
	// Get broadcast manager
	bm := broadcast.GetBroadcastManager()
	
	// Check if Redis cleanup is available
	if !isRedisManager(bm) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Redis cleanup not available - not using Redis manager",
		})
	}
	
	// For now, just log the cleanup request
	logrus.Infof("Device cleanup requested for %s", deviceID)
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Device cleanup request logged",
		"deviceId": deviceID,
	})
}

// CleanupAllOldDevices cleans up all non-existent devices from Redis
func CleanupAllOldDevices(c *fiber.Ctx) error {
	// List of known old devices to clean up
	oldDevices := []string{
		"3472b8c5-974b-4deb-bab9-792cc5a09c57", // Your old device
		// Add more as needed
	}
	
	bm := broadcast.GetBroadcastManager()
	
	// Check if Redis cleanup is available
	if !isRedisManager(bm) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Redis cleanup not available - not using Redis manager",
		})
	}
	
	// Log cleanup requests
	for _, deviceID := range oldDevices {
		logrus.Infof("Cleanup requested for device %s", deviceID)
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Old devices cleanup requests logged",
		"devices": oldDevices,
	})
}
