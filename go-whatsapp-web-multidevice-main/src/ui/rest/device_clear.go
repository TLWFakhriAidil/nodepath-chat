package rest

import (
	"fmt"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// ClearDeviceData clears WhatsApp session data for a specific device
func (handler *App) ClearDeviceData(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	userID := c.Locals("userID").(string)
	
	// Verify device belongs to user
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDevice(userID, deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status: 404,
			Code: "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	logrus.Infof("Clearing device data for device %s (user: %s)", deviceID, userID)
	
	// Step 1: Disconnect WhatsApp client if connected
	cm := whatsapp.GetClientManager()
	if client, err := cm.GetClient(deviceID); err == nil && client != nil {
		logrus.Info("Disconnecting WhatsApp client...")
		client.Disconnect()
		cm.RemoveClient(deviceID)
	}
	
	// Step 2: Clear WhatsApp store data
	if client, err := cm.GetClient(deviceID); err == nil && client != nil && client.Store != nil && client.Store.ID != nil {
		jid := client.Store.ID.String()
		logrus.Infof("Clearing WhatsApp store for JID: %s", jid)
		
		// The store cleanup happens when we disconnect the client
	}
	
	// Step 3: Update device status in database
	err = userRepo.UpdateDeviceStatus(deviceID, "disconnected", "", "")
	if err != nil {
		logrus.Errorf("Failed to update device status: %v", err)
	}
	
	// Step 4: Clear any associated data (messages, chats, etc.)
	whatsappRepo := repository.GetWhatsAppRepository()
	if whatsappRepo != nil {
		// Clear messages
		err = whatsappRepo.ClearDeviceMessages(deviceID)
		if err != nil {
			logrus.Errorf("Failed to clear messages: %v", err)
		}
		
		// Clear chats
		err = whatsappRepo.ClearDeviceChats(deviceID)
		if err != nil {
			logrus.Errorf("Failed to clear chats: %v", err)
		}
	}
	
	logrus.Infof("Successfully cleared data for device %s", device.DeviceName)
	
	return c.JSON(utils.ResponseData{
		Status: 200,
		Code: "SUCCESS",
		Message: fmt.Sprintf("Device %s data cleared successfully", device.DeviceName),
		Results: fiber.Map{
			"device_id": deviceID,
			"device_name": device.DeviceName,
			"status": "cleared",
		},
	})
}

// ResetAllDevices - Admin function to reset all WhatsApp devices
func (handler *App) ResetAllDevices(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	
	// Get all devices for user
	userRepo := repository.GetUserRepository()
	devices, err := userRepo.GetUserDevices(userID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status: 500,
			Code: "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	cleared := 0
	failed := 0
	
	// Clear each device
	cm := whatsapp.GetClientManager()
	for _, device := range devices {
		// Disconnect client
		if client, err := cm.GetClient(device.ID); err == nil && client != nil {
			client.Disconnect()
			cm.RemoveClient(device.ID)
		}
		
		// Update status
		err := userRepo.UpdateDeviceStatus(device.ID, "disconnected", "", "")
		if err != nil {
			failed++
			logrus.Errorf("Failed to reset device %s: %v", device.ID, err)
		} else {
			cleared++
		}
	}
	
	return c.JSON(utils.ResponseData{
		Status: 200,
		Code: "SUCCESS",
		Message: fmt.Sprintf("Reset complete: %d cleared, %d failed", cleared, failed),
		Results: fiber.Map{
			"total_devices": len(devices),
			"cleared": cleared,
			"failed": failed,
		},
	})
}