package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/multidevice"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"context"
	"fmt"
)

// DeviceConnect initiates WhatsApp connection for a specific device
func (handler *App) DeviceConnect(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device ID is required",
		})
	}
	
	// Get user from context
	userID := c.Locals("userID")
	if userID == nil {
		// Try to get from session cookie as fallback
		token := c.Cookies("session_token")
		if token != "" {
			userRepo := repository.GetUserRepository()
			session, err := userRepo.GetSession(token)
			if err == nil && session != nil {
				userID = session.UserID
			}
		}
		
		// Still no user ID?
		if userID == nil {
			return c.Status(401).JSON(utils.ResponseData{
				Status:  401,
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			})
		}
	}
	
	// Get device from database to verify ownership and get phone number
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetUserDevice(userID.(string), deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Get or create device connection using multi-device manager
	dm := multidevice.GetDeviceManager()
	if dm == nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Device manager not initialized",
		})
	}
	
	conn, err := dm.GetOrCreateDeviceConnection(deviceID, userID.(string), device.Phone)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: fmt.Sprintf("Failed to create device connection: %v", err),
		})
	}
	
	// Check if already logged in
	if conn.Client.IsLoggedIn() {
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
	
	// Store connection session for tracking BY DEVICE ID
	whatsapp.StoreConnectionSession(deviceID, &whatsapp.ConnectionSession{
		DeviceID: deviceID,
		UserID:   userID.(string),
	})
	
	// Add event handlers specific to this device
	conn.Client.AddEventHandler(func(evt interface{}) {
		whatsapp.HandleDeviceEvent(context.Background(), deviceID, evt)
	})
	
	// Generate QR code
	qrChan, err := conn.Client.GetQRChannel(context.Background())
	if err != nil {
		// If error, try to disconnect and reconnect
		if conn.Client.IsConnected() {
			conn.Client.Disconnect()
		}
		
		// Try again
		err = conn.Client.Connect()
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: fmt.Sprintf("Failed to connect: %v", err),
			})
		}
		
		qrChan, err = conn.Client.GetQRChannel(context.Background())
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: fmt.Sprintf("Failed to get QR channel: %v", err),
			})
		}
	}
	
	// Connect the client
	err = conn.Client.Connect()
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: fmt.Sprintf("Failed to connect client: %v", err),
		})
	}
	
	// Store the QR channel for this device
	whatsapp.SetDeviceQRChannel(deviceID, qrChan)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device connection initiated. Please scan the QR code.",
		Results: map[string]interface{}{
			"deviceId": deviceID,
			"status":   "waiting_qr",
		},
	})
}

// GetDeviceQR gets the current QR code for a device
func (handler *App) GetDeviceQR(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device ID is required",
		})
	}
	
	// Get user from context
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}
	
	// Verify device ownership
	userRepo := repository.GetUserRepository()
	_, err := userRepo.GetUserDevice(userID.(string), deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Get QR code from channel
	qr, err := whatsapp.GetDeviceQR(deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NO_QR",
			Message: "No QR code available. Please initiate connection first.",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "QR code retrieved",
		Results: map[string]interface{}{
			"qr":       qr,
			"deviceId": deviceID,
		},
	})
}

// DisconnectDevice disconnects a specific device
func (handler *App) DisconnectDevice(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device ID is required",
		})
	}
	
	// Get user from context
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}
	
	// Verify device ownership
	userRepo := repository.GetUserRepository()
	_, err := userRepo.GetUserDevice(userID.(string), deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Get device connection
	dm := multidevice.GetDeviceManager()
	conn, err := dm.GetDeviceConnection(deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_CONNECTED",
			Message: "Device is not connected",
		})
	}
	
	// Disconnect
	if conn.Client != nil && conn.Client.IsConnected() {
		conn.Client.Disconnect()
	}
	
	// Update device status in database
	err = userRepo.UpdateDeviceStatus(deviceID, "offline", "", "")
	if err != nil {
		// Log but don't fail the request
		fmt.Printf("Failed to update device status: %v\n", err)
	}
	
	// Clear QR channel
	whatsapp.ClearDeviceQRChannel(deviceID)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device disconnected successfully",
		Results: map[string]interface{}{
			"deviceId": deviceID,
			"status":   "disconnected",
		},
	})
}

// ResetDevice completely resets a device (removes WhatsApp session)
func (handler *App) ResetDevice(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device ID is required",
		})
	}
	
	// Get user from context
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}
	
	// Verify device ownership
	userRepo := repository.GetUserRepository()
	_, err := userRepo.GetUserDevice(userID.(string), deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Remove device session completely
	dm := multidevice.GetDeviceManager()
	err = dm.RemoveDeviceSession(deviceID)
	if err != nil {
		// Log but continue
		fmt.Printf("Failed to remove device session: %v\n", err)
	}
	
	// Update device status in database
	err = userRepo.UpdateDeviceStatus(deviceID, "offline", "", "")
	if err != nil {
		fmt.Printf("Failed to update device status: %v\n", err)
	}
	
	// Clear any connection sessions
	whatsapp.ClearConnectionSession(userID.(string))
	whatsapp.ClearDeviceQRChannel(deviceID)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device reset successfully. You can now connect with a new WhatsApp account.",
		Results: map[string]interface{}{
			"deviceId": deviceID,
			"status":   "reset",
		},
	})
}
