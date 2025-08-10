package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
)

// CreateDevice creates a new device for the user
func (handler *App) CreateDevice(c *fiber.Ctx) error {
	// Get user from context
	userID := c.Locals("userID")
	if userID == nil {
		// Fallback: try to get from session cookie
		token := c.Cookies("session_token")
		if token == "" {
			return c.Status(401).JSON(utils.ResponseData{
				Status:  401,
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			})
		}
		
		userRepo := repository.GetUserRepository()
		session, err := userRepo.GetSession(token)
		if err != nil {
			return c.Status(401).JSON(utils.ResponseData{
				Status:  401,
				Code:    "UNAUTHORIZED",
				Message: "Invalid session",
			})
		}
		userID = session.UserID
	}
	
	// Parse request body
	var req struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	if req.Name == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device name is required",
		})
	}
	
	// Create device in database
	userRepo := repository.GetUserRepository()
	
	var device *models.UserDevice
	var err error
	
	if req.Phone != "" {
		device, err = userRepo.AddUserDeviceWithPhone(userID.(string), req.Name, req.Phone)
	} else {
		device, err = userRepo.AddUserDevice(userID.(string), req.Name)
	}
	
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to create device",
		})
	}
	
	// Format phone for display
	phoneDisplay := "Not connected"
	if device.Phone != "" {
		phoneDisplay = device.Phone
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device created successfully",
		Results: fiber.Map{
			"success": true,
			"device": fiber.Map{
				"id":       device.ID,
				"name":     device.DeviceName,
				"phone":    phoneDisplay,
				"status":   device.Status,
				"lastSeen": device.LastSeen.Format("2006-01-02 15:04:05"),
			},
		},
	})
}
