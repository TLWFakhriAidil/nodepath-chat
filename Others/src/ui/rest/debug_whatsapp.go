package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/gofiber/fiber/v2"
)

// DebugWhatsAppClients returns all registered WhatsApp clients
func DebugWhatsAppClients(c *fiber.Ctx) error {
	clientManager := whatsapp.GetClientManager()
	
	// Get all registered clients
	clients := clientManager.GetAllClients()
	
	clientInfo := make([]map[string]interface{}, 0)
	for deviceID, client := range clients {
		info := map[string]interface{}{
			"device_id": deviceID,
			"connected": client != nil && client.IsConnected(),
			"logged_in": client != nil && client.IsLoggedIn(),
		}
		clientInfo = append(clientInfo, info)
	}
	
	return c.JSON(fiber.Map{
		"total_clients": len(clients),
		"clients": clientInfo,
		"looking_for": "2de48db2-f1ab-4d81-8a26-58b01df75bdf",
	})
}

// TestDeviceClient tests if a specific device client exists
func TestDeviceClient(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	
	clientManager := whatsapp.GetClientManager()
	client, err := clientManager.GetClient(deviceID)
	
	if err != nil {
		return c.JSON(fiber.Map{
			"error": true,
			"message": err.Error(),
			"device_id": deviceID,
		})
	}
	
	return c.JSON(fiber.Map{
		"error": false,
		"device_id": deviceID,
		"client_exists": client != nil,
		"is_connected": client != nil && client.IsConnected(),
		"is_logged_in": client != nil && client.IsLoggedIn(),
	})
}
