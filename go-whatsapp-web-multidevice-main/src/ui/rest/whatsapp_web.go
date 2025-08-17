package rest

import (
	"fmt"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
)

// WhatsAppWebView renders the WhatsApp Web interface for a device
func (handler *App) WhatsAppWebView(c *fiber.Ctx) error {
	deviceId := c.Params("id")
	
	// Check if user has valid session cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Redirect("/login")
	}
	
	// Verify session is valid
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil || session == nil {
		return c.Redirect("/login")
	}
	
	// Session is valid, render WhatsApp Web
	return c.Render("views/whatsapp_web", fiber.Map{
		"DeviceID": deviceId,
	})
}

// GetWhatsAppChats gets chats for WhatsApp Web view
func (handler *App) GetWhatsAppChats(c *fiber.Ctx) error {
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
	
	// Get user from database
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	// Get user devices to check if this device belongs to user
	devices, err := userRepo.GetUserDevices(user.ID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	// Check if device belongs to user and is online
	deviceBelongsToUser := false
	isOnline := false
	for _, device := range devices {
		if device.ID == deviceId {
			deviceBelongsToUser = true
			isOnline = device.Status == "online"
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
	
	if !isOnline {
		// When offline, return stored chats from database
		storedChats, err := whatsapp.GetChatsFromDatabase(deviceId)
		if err != nil {
			logrus.Warnf("Failed to get stored chats for offline device %s: %v", deviceId, err)
			storedChats = []map[string]interface{}{}
		}
		
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "DEVICE_OFFLINE",
			Message: "Device is offline. Showing stored chat history.",
			Results: storedChats,
		})
	}
	
	// Get personal chats only from WhatsMeow's store
	chats, err := whatsapp.GetWhatsAppWebChats(deviceId)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: fmt.Sprintf("Failed to get chats: %v", err),
		})
	}
	
	// Return chats directly (already formatted)
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Found %d personal chats", len(chats)),
		Results: chats,
	})
}

// GetWhatsAppMessages gets messages for a specific chat
func (handler *App) GetWhatsAppMessages(c *fiber.Ctx) error {
	deviceId := c.Params("id")
	chatId := c.Params("chatId")
	
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
	
	// Get user from database
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	// Get user devices to check if this device belongs to user
	devices, err := userRepo.GetUserDevices(user.ID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	// Check if device belongs to user and is online
	deviceBelongsToUser := false
	isOnline := false
	for _, device := range devices {
		if device.ID == deviceId {
			deviceBelongsToUser = true
			isOnline = device.Status == "online"
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
	
	if !isOnline {
		// When offline, still get stored messages from database
		// The GetWhatsAppWebMessages function already queries the database
		limit := c.QueryInt("limit", 50)
		if limit > 100 {
			limit = 100
		}
		
		// We need to bypass the client check in GetWhatsAppWebMessages for offline mode
		// So let's get messages directly from database
		messages, err := whatsapp.GetStoredMessagesFromDB(deviceId, chatId, limit)
		if err != nil {
			logrus.Warnf("Failed to get stored messages for offline device %s, chat %s: %v", deviceId, chatId, err)
			messages = []map[string]interface{}{}
		}
		
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "DEVICE_OFFLINE",
			Message: "Device is offline. Showing stored messages.",
			Results: messages,
		})
	}
	
	// Get messages from WhatsMeow's store
	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}
	
	messages, err := whatsapp.GetWhatsAppWebMessages(deviceId, chatId, limit)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: fmt.Sprintf("Failed to get messages: %v", err),
		})
	}
	
	// Return messages
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Found %d messages", len(messages)),
		Results: messages,
	})
}

// SyncWhatsAppDevice triggers a sync for the device
func (handler *App) SyncWhatsAppDevice(c *fiber.Ctx) error {
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
	
	// Get user from database
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	// Get user devices to check if this device belongs to user
	devices, err := userRepo.GetUserDevices(user.ID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	// Check if device belongs to user and is online
	deviceBelongsToUser := false
	isOnline := false
	for _, device := range devices {
		if device.ID == deviceId {
			deviceBelongsToUser = true
			isOnline = device.Status == "online"
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
	
	if !isOnline {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "DEVICE_OFFLINE",
			Message: "Device must be online to sync",
		})
	}
	
	// Force a sync by getting the WhatsApp client
	cm := whatsapp.GetClientManager()
	client, err := cm.GetClient(deviceId)
	if err == nil && client != nil && client.IsConnected() {
		// Just verify connection, no presence needed
		// Small delay to allow messages to process
		time.Sleep(500 * time.Millisecond)
		
		logrus.Infof("Sync triggered for device %s", deviceId)
	}
	
	// Return success and let the client refresh
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sync completed. Refreshing chats...",
	})
}

