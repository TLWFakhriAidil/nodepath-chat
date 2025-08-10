package rest

import (
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/gofiber/fiber/v2"
)

// GetAnalyticsData returns analytics data for the dashboard
func (handler *App) GetAnalyticsData(c *fiber.Ctx) error {
	days := c.Params("days", "7")
	deviceFilter := c.Query("device", "all")
	
	// Get session from cookie
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
	
	daysInt := 7
	switch days {
	case "today":
		daysInt = 1
	case "7":
		daysInt = 7
	case "30":
		daysInt = 30
	case "90":
		daysInt = 90
	}
	
	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -daysInt)
	
	// Get real analytics from database
	analyticsRepo := repository.NewMessageAnalyticsRepository()
	analytics, err := analyticsRepo.GetUserAnalytics(session.UserID, startDate, endDate, deviceFilter)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get analytics",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Analytics data retrieved",
		Results: analytics,
	})
}

// GetCustomAnalyticsData returns analytics for custom date range
func (handler *App) GetCustomAnalyticsData(c *fiber.Ctx) error {
	startStr := c.Query("start")
	endStr := c.Query("end")
	deviceFilter := c.Query("device", "all")
	
	// Get session from cookie
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
	
	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid start date format",
		})
	}
	
	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST", 
			Message: "Invalid end date format",
		})
	}
	
	// Get real analytics from database
	analyticsRepo := repository.NewMessageAnalyticsRepository()
	analytics, err := analyticsRepo.GetUserAnalytics(session.UserID, startDate, endDate, deviceFilter)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get analytics",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Custom analytics data retrieved",
		Results: analytics,
	})
}

// GetConnectedDevices returns real connected devices
func (handler *App) GetConnectedDevices(c *fiber.Ctx) error {
	// Get user from context (set by auth middleware)
	userID := c.Locals("userID")
	if userID == nil {
		// Fallback: try to get from session cookie
		token := c.Cookies("session_token")
		if token == "" {
			// Try headers as last resort
			token = c.Get("Authorization")
			if token == "" {
				token = c.Get("X-Auth-Token")
			}
		}
		
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
	
	// Get user devices from database
	userRepo := repository.GetUserRepository()
	devices, err := userRepo.GetUserDevices(userID.(string))
	if err != nil {
		// Return empty array if no devices found
		if err.Error() == "no devices found" {
			return c.JSON(utils.ResponseData{
				Status:  200,
				Code:    "SUCCESS",
				Message: "No devices found",
				Results: []fiber.Map{},
			})
		}
		
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	// Convert to response format
	deviceList := make([]fiber.Map, 0, len(devices))
	for _, device := range devices {
		deviceList = append(deviceList, fiber.Map{
			"id":       device.ID,
			"name":     device.DeviceName,
			"phone":    device.Phone,
			"status":   device.Status,
			"lastSeen": device.LastSeen.Format("2006-01-02 15:04:05"),
			"jid":      device.JID,
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Devices retrieved",
		Results: deviceList,
	})
}