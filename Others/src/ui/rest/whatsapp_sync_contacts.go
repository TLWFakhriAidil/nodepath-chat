package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// SyncWhatsAppContacts syncs WhatsApp contacts to leads table
func (handler *App) SyncWhatsAppContacts(c *fiber.Ctx) error {
	deviceId := c.Params("id")
	
	// Get session from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token found",
		})
	}
	
	// Get session from database
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil || session == nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}
	
	// Verify device belongs to user
	devices, err := userRepo.GetUserDevices(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	deviceBelongsToUser := false
	for _, device := range devices {
		if device.ID == deviceId {
			deviceBelongsToUser = true
			break
		}
	}
	
	if !deviceBelongsToUser {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "Device does not belong to user",
		})
	}
	
	// Run auto-save in background
	go func() {
		logrus.Infof("Starting WhatsApp contacts sync for device %s", deviceId)
		err := whatsapp.AutoSaveChatsToLeads(deviceId, session.UserID)
		if err != nil {
			logrus.Errorf("Failed to sync contacts: %v", err)
		} else {
			logrus.Infof("Successfully synced contacts for device %s", deviceId)
		}
	}()
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Contact sync started. Check leads page in a few seconds.",
	})
}

// MergeDeviceContacts merges contacts from old device to new device
func (handler *App) MergeDeviceContacts(c *fiber.Ctx) error {
	// Get session from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token found",
		})
	}
	
	// Get session from database
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil || session == nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}
	
	var request struct {
		OldDeviceID string `json:"old_device_id"`
		NewDeviceID string `json:"new_device_id"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	// Verify both devices belong to user
	devices, err := userRepo.GetUserDevices(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	oldDeviceBelongs := false
	newDeviceBelongs := false
	
	for _, device := range devices {
		if device.ID == request.OldDeviceID {
			oldDeviceBelongs = true
		}
		if device.ID == request.NewDeviceID {
			newDeviceBelongs = true
		}
	}
	
	if !oldDeviceBelongs || !newDeviceBelongs {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "One or both devices do not belong to user",
		})
	}
	
	// Handle device change
	err = whatsapp.HandleDeviceChange(request.OldDeviceID, request.NewDeviceID, session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to merge device data: " + err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Successfully merged device data. All contacts preserved.",
	})
}
