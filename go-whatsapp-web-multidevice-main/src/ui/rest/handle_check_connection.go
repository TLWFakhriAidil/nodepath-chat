package rest

import (
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
)

// Add this function to your rest package to handle the check-connection endpoint
func HandleCheckConnection(c *fiber.Ctx) error {
	// This is a simple health check endpoint
	// The actual device status is handled by GetDevices
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "API is running",
		Results: map[string]interface{}{
			"status":    "online",
			"timestamp": time.Now().Unix(),
			"note":      "Device status is checked in /api/devices endpoint",
		},
	})
}