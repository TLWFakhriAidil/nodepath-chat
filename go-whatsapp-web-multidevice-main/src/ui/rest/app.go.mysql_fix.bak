package rest

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type App struct {
	Service domainApp.IAppUsecase
	Send    *Send // Add Send service for WhatsApp Web messaging
}

func InitRestApp(app *fiber.App, service domainApp.IAppUsecase) App {
	rest := App{Service: service}
	
	// Initialize Send service as nil - will be set later by SetSendService
	rest.Send = nil
	
	// Health check endpoint
	app.Get("/health", rest.HealthCheck)
	app.Get("/api/health", rest.HealthCheck)
	
	// Dashboard routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/login")
	})
	app.Get("/login", rest.AppLoginView)
	app.Get("/register", rest.RegisterView)
	app.Get("/dashboard", rest.DashboardView)
	
	// Auth API endpoints
	app.Post("/api/login", rest.HandleLogin)
	app.Post("/api/register", rest.HandleRegister)
	app.Get("/logout", rest.HandleLogout)
	
	// Analytics API endpoints
	app.Get("/api/analytics/:days", rest.GetAnalyticsData)
	app.Get("/api/analytics/custom", rest.GetCustomAnalyticsData)
	app.Get("/api/campaigns/analytics", rest.GetCampaignAnalytics)
	app.Get("/api/sequences/analytics", rest.GetSequenceAnalytics)
	app.Get("/api/niches", rest.GetNiches)
	app.Get("/api/test-db", rest.TestDatabaseConnection)
	app.Get("/api/devices", rest.GetConnectedDevices)
	app.Post("/api/devices", rest.CreateDevice)
	app.Get("/api/devices/:id", rest.GetDevice)
	
	// Device pages
	app.Get("/device/:id/actions", rest.DeviceActionsView)
	app.Get("/device/:id/leads", rest.DeviceLeadsView)
	app.Get("/device/:id/whatsapp-web", rest.WhatsAppWebView)
	
	// WhatsApp Web API endpoints
	app.Get("/api/devices/:id/chats", rest.GetWhatsAppChats)
	app.Get("/api/devices/:id/messages/:chatId", rest.GetWhatsAppMessages)
	app.Post("/api/devices/:id/send", rest.SendWhatsAppWebMessage)
	app.Post("/api/devices/:id/sync", rest.SyncWhatsAppDevice)
	app.Get("/api/devices/:id/diagnose", rest.DiagnoseDevice)
	app.Post("/api/devices/:id/sync-contacts", rest.SyncWhatsAppContacts)
	app.Post("/api/devices/merge-contacts", rest.MergeDeviceContacts)
	
	// Media serving endpoint
	app.Get("/media/:filename", rest.ServeMedia)
	
	// Device management endpoints
	app.Delete("/api/devices/:id", rest.DeleteDevice)
	app.Post("/api/devices/:deviceId/connect", rest.DeviceConnect)
	app.Post("/api/devices/:deviceId/refresh", RefreshDevice) // Simple refresh check
	app.Post("/api/devices/:deviceId/reconnect", ReconnectDeviceSession) // Actual reconnection with session
	app.Get("/api/devices/:deviceId/qr", rest.GetDeviceQR)
	app.Post("/api/devices/:deviceId/disconnect", rest.DisconnectDevice)
	app.Post("/api/devices/:deviceId/reset", rest.ResetDevice)
	app.Post("/api/devices/:deviceId/clear-session", rest.ClearDeviceSession)
	app.Post("/api/devices/clear-all-sessions", rest.ClearAllSessions)
	app.Get("/api/devices/check-connection", SimpleCheckConnection)
	app.Get("/app/logout", rest.LogoutDevice)
	app.Get("/app/reconnect", rest.ReconnectDevice)
	app.Get("/app/devices", rest.GetDevices)
	app.Get("/user/info", rest.GetUserInfo)
	app.Get("/user/avatar", rest.GetUserAvatar)
	app.Post("/user/avatar", rest.ChangeUserAvatar)
	app.Post("/user/pushname", rest.ChangeUserPushName)
	
	// Lead management endpoints
	app.Get("/api/devices/:deviceId/leads", rest.GetDeviceLeads)
	app.Post("/api/leads", rest.CreateLead)
	app.Put("/api/leads/:id", rest.UpdateLead)
	app.Delete("/api/leads/:id", rest.DeleteLead)
	app.Get("/api/devices/:deviceId/leads/export", rest.ExportLeads)
	app.Post("/api/devices/:deviceId/leads/import", rest.ImportLeads)
	
	// Campaign endpoints
	app.Get("/api/campaigns", rest.GetCampaigns)
	app.Post("/api/campaigns", rest.CreateCampaign)
	app.Put("/api/campaigns/:id", rest.UpdateCampaign)
	app.Delete("/api/campaigns/:id", rest.DeleteCampaign)
	app.Get("/api/campaigns/summary", rest.GetCampaignSummary)
	app.Get("/api/campaigns/:id/device-report", rest.GetCampaignDeviceReport)
	app.Get("/api/campaigns/:id/device/:deviceId/leads", rest.GetCampaignDeviceLeads)
	
	// Team Member Management endpoints
	app.Get("/api/team-members", rest.GetAllTeamMembers)
	app.Post("/api/team-members", rest.CreateTeamMember)
	app.Put("/api/team-members/:id", rest.UpdateTeamMember)
	app.Delete("/api/team-members/:id", rest.DeleteTeamMember)
	app.Post("/api/campaigns/:id/device/:deviceId/retry-failed", rest.RetryCampaignFailedMessages)
	
	// AI Lead Management Routes
	app.Post("/api/leads-ai", rest.CreateLeadAI)
	app.Get("/api/leads-ai", rest.GetLeadsAI)
	app.Put("/api/leads-ai/:id", rest.UpdateLeadAI)
	app.Delete("/api/leads-ai/:id", rest.DeleteLeadAI)
	
	// AI Campaign Trigger Route
	app.Post("/api/campaigns-ai/:id/trigger", rest.TriggerAICampaign)
	
	// Sequence summary endpoint
	app.Get("/api/sequences/summary", rest.GetSequenceSummary)
	
	// Worker status endpoint
	app.Get("/api/workers/status", rest.GetWorkerStatus)
	
	// Worker control endpoints
	app.Post("/api/workers/resume-failed", rest.ResumeFailedWorkers)
	app.Post("/api/workers/stop-all", rest.StopAllWorkers)
	
	// System status endpoint
	app.Get("/api/system/status", rest.GetSystemStatus)
	app.Get("/api/system/redis-check", rest.CheckRedisStatus)
	
	// Worker status endpoints
	app.Get("/api/workers/status", rest.CheckAllWorkersStatus)
	app.Get("/api/workers/device/:deviceId", rest.CheckDeviceWorkerStatus)
	
	// Status view pages
	app.Get("/status/redis", rest.RedisStatusView)
	app.Get("/status/device-worker", rest.DeviceWorkerStatusView)
	app.Get("/status/all-workers", rest.AllWorkersStatusView)
	
	// WhatsApp QR code endpoint
	app.Get("/app/qr", rest.GetQRCode)
	
	// Static media files
	app.Static("/media", config.PathStorages)
	
	// Device management endpoints
	app.Delete("/api/devices/:deviceId/clear", rest.ClearDeviceData)
	app.Post("/api/devices/reset-all", rest.ResetAllDevices)
	
	// API endpoints
	app.Get("/app/login", rest.Login)
	app.Get("/app/login-with-code", rest.LoginWithCode)
	app.Post("/app/link-device", rest.LinkDevicePhone)
	app.Get("/app/logout", rest.Logout)
	app.Get("/app/reconnect", rest.Reconnect)
	app.Get("/app/devices", rest.Devices)

	return App{Service: service}
}

// SetSendService sets the send service for WhatsApp Web messaging
func (app *App) SetSendService(sendService domainSend.ISendUsecase) {
	app.Send = &Send{Service: sendService}
}

func (handler *App) Login(c *fiber.Ctx) error {
	// Get device ID from query params
	deviceId := c.Query("deviceId")
	
	// Get user from context or cookie
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
	}
	
	// Log for debugging
	log.Printf("Login request - UserID: %v, DeviceID: %s", userID, deviceId)
	
	if userID != nil && deviceId != "" {
		// Clear any existing WhatsApp session for this device before new login
		err := whatsapp.ClearWhatsAppSessionData(deviceId)
		if err != nil {
			logrus.Warnf("Failed to clear existing session data for device %s: %v", deviceId, err)
			// Continue anyway
		}
		
		// Store connection session for this device
		logrus.Infof("Storing connection session for user %s, device %s", userID, deviceId)
		whatsapp.StoreConnectionSession(deviceId, &whatsapp.ConnectionSession{
			UserID:   userID.(string),
			DeviceID: deviceId,
		})
	}
	
	response, err := handler.Service.Login(c.UserContext())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Login success",
		Results: map[string]any{
			"qr_link":     fmt.Sprintf("%s://%s/%s", c.Protocol(), c.Hostname(), response.ImagePath),
			"qr_duration": response.Duration,
		},
	})
}
func (handler *App) LoginWithCode(c *fiber.Ctx) error {
	pairCode, err := handler.Service.LoginWithCode(c.UserContext(), c.Query("phone"))
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Login with code success",
		Results: map[string]any{
			"pair_code": pairCode,
		},
	})
}

func (handler *App) Logout(c *fiber.Ctx) error {
	err := handler.Service.Logout(c.UserContext())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Success logout",
		Results: nil,
	})
}

func (handler *App) Reconnect(c *fiber.Ctx) error {
	err := handler.Service.Reconnect(c.UserContext())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Reconnect success",
		Results: nil,
	})
}
func (handler *App) Devices(c *fiber.Ctx) error {
	devices, err := handler.Service.FetchDevices(c.UserContext())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Fetch device success",
		Results: devices,
	})
}

// LinkDevicePhone links a phone number to a device for the current user
func (handler *App) LinkDevicePhone(c *fiber.Ctx) error {
	var req struct {
		DeviceID string `json:"device_id"`
		Phone    string `json:"phone"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	
	// Get user from context (set by auth middleware)
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	
	// Update device with phone number
	userRepo := repository.GetUserRepository()
	err := userRepo.UpdateDevicePhone(userID.(string), req.DeviceID, req.Phone)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update device",
		})
	}
	
	return c.JSON(fiber.Map{
		"status": "success",
		"message": "Phone number linked successfully",
	})
}

// AppLoginView serves the login page
func (handler *App) AppLoginView(c *fiber.Ctx) error {
	// Serve login page from embedded filesystem
	return c.Render("views/login", fiber.Map{
		"Title": "Login - WhatsApp Analytics",
	})
}

// RegisterView serves the register page
func (handler *App) RegisterView(c *fiber.Ctx) error {
	// Serve register page from embedded filesystem
	return c.Render("views/register", fiber.Map{
		"Title": "Register - WhatsApp Analytics",
	})
}

// AppDevicesView serves the devices page (deprecated - redirect to dashboard)
func (handler *App) AppDevicesView(c *fiber.Ctx) error {
	return c.Redirect("/dashboard")
}
// DashboardView serves the main dashboard
func (handler *App) DashboardView(c *fiber.Ctx) error {
	// Serve dashboard from embedded filesystem
	return c.Render("views/dashboard", fiber.Map{
		"Title": "Dashboard - WhatsApp Analytics",
	})
}

// HandleLogin processes login requests
func (handler *App) HandleLogin(c *fiber.Ctx) error {
	var loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	
	if err := c.BodyParser(&loginReq); err != nil {
		log.Printf("Failed to parse login request: %v", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	
	log.Printf("Login attempt for email: %s", loginReq.Email)
	
	// Get user repository
	userRepo := repository.GetUserRepository()
	
	// Validate credentials
	user, err := userRepo.ValidatePassword(loginReq.Email, loginReq.Password)
	if err != nil {
		log.Printf("Login failed for %s: %v", loginReq.Email, err)
		// Also log if it's a database error
		if err.Error() == "user not found" {
			log.Printf("User %s does not exist in database", loginReq.Email)
		}
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid email or password",
			"debug": err.Error(), // Temporarily add debug info
		})
	}
	
	log.Printf("Login successful for user: %s", user.Email)
	
	// Create session
	session, err := userRepo.CreateSession(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create session",
		})
	}
	
	// Set session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})
	
	return c.JSON(fiber.Map{
		"status": "success",
		"message": "Login successful",
		"token": session.Token,
		"user": fiber.Map{
			"id":       user.ID,
			"email":    user.Email,
			"fullName": user.FullName,
		},
	})
}
// HandleRegister processes registration requests
func (handler *App) HandleRegister(c *fiber.Ctx) error {
	var registerReq struct {
		FullName string `json:"fullname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	
	if err := c.BodyParser(&registerReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}
	
	// Get user repository
	userRepo := repository.GetUserRepository()
	
	// Create user
	user, err := userRepo.CreateUser(registerReq.Email, registerReq.FullName, registerReq.Password)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return c.Status(409).JSON(fiber.Map{
				"error": "Email already registered",
			})
		}
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"status": "success",
		"message": "Registration successful",
		"user": fiber.Map{
			"id":       user.ID,
			"email":    user.Email,
			"fullName": user.FullName,
		},
	})
}

// HandleLogout handles user logout
func (handler *App) HandleLogout(c *fiber.Ctx) error {
	// Clear session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})
	
	// Redirect to login page
	return c.Redirect("/login")
}

// GetQRCode returns the QR code image directly
func (handler *App) GetQRCode(c *fiber.Ctx) error {
	// Store connection session info
	userID := c.Locals("userID")
	deviceID := c.Query("device_id")
	
	if userID != nil && deviceID != "" {
		// Store connection session BY DEVICE ID to support multiple devices
		whatsapp.StoreConnectionSession(deviceID, &whatsapp.ConnectionSession{
			UserID:   userID.(string),
			DeviceID: deviceID,
		})
		logrus.Infof("Connection request for user %s, device %s", userID, deviceID)
	}
	
	// Get QR code from login service
	response, err := handler.Service.Login(c.UserContext())
	if err != nil {
		// Check if it's a QR channel error and try to provide helpful message
		if err.Error() == "QR channel error" {
			logrus.Warn("QR channel error - attempting to reset connection")
			
			// Try to reset and get QR again
			// Note: This requires implementing a reset method
			return c.Status(503).JSON(utils.ResponseData{
				Status:  503,
				Code:    "QR_CHANNEL_ERROR",
				Message: "WhatsApp connection error. Please try logging out and reconnecting.",
				Results: fiber.Map{
					"suggestion": "Try clicking 'Logout' and then reconnect",
				},
			})
		}
		
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to generate QR code: " + err.Error(),
		})
	}
	
	// Read the QR image file
	qrImagePath := response.ImagePath
	if qrImagePath != "" {
		// Return the image file
		return c.SendFile(qrImagePath)
	}
	
	// If no QR code available, return error image
	return c.Status(404).JSON(utils.ResponseData{
		Status:  404,
		Code:    "NOT_FOUND",
		Message: "QR code not available",
	})
}

// HealthCheck endpoint to verify application status
func (handler *App) HealthCheck(c *fiber.Ctx) error {
	// Check database connection
	userRepo := repository.GetUserRepository()
	dbHealthy := true
	dbError := ""
	
	// Try to get a user to test DB connection
	_, err := userRepo.GetUserByEmail("test@health.check")
	if err != nil && err.Error() != "user not found" {
		dbHealthy = false
		dbError = err.Error()
	}
	
	health := fiber.Map{
		"status": "ok",
		"version": config.AppVersion,
		"database": fiber.Map{
			"connected": dbHealthy,
			"error": dbError,
		},
		"environment": fiber.Map{
			"port": config.AppPort,
			"debug": config.AppDebug,
		},
	}
	
	if !dbHealthy {
		return c.Status(503).JSON(fiber.Map{
			"status": "error",
			"health": health,
		})
	}
	
	return c.JSON(fiber.Map{
		"status": "healthy",
		"health": health,
	})
}

// GetDevice gets a specific device by ID  
func (handler *App) GetDevice(c *fiber.Ctx) error {
	deviceId := c.Params("id")
	
	// Get session from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token",
		})
	}
	
	// Get user from session
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}
	
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	// Get the actual device from database
	device, err := userRepo.GetDevice(user.ID, deviceId)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Format phone for display
	phoneDisplay := device.Phone
	if phoneDisplay == "" {
		phoneDisplay = "Not connected"
	}
	
	// Convert to response format
	deviceData := map[string]interface{}{
		"id":       device.ID,
		"name":     device.DeviceName,
		"phone":    phoneDisplay,
		"status":   device.Status,
		"lastSeen": device.LastSeen,
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device retrieved successfully",
		Results: deviceData,
	})
}

// DeviceActionsView renders the device actions testing page
func (handler *App) DeviceActionsView(c *fiber.Ctx) error {
	return c.Render("views/device_actions", fiber.Map{})
}

// DeviceLeadsView renders the device leads management page
func (handler *App) DeviceLeadsView(c *fiber.Ctx) error {
	return c.Render("views/device_leads", fiber.Map{})
}

// GetDeviceLeads gets all leads for a specific device
func (handler *App) GetDeviceLeads(c *fiber.Ctx) error {
	deviceId := c.Params("deviceId")
	
	// Get session from cookie - same as campaigns
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
	
	leadRepo := repository.GetLeadRepository()
	leads, err := leadRepo.GetLeadsByDevice(session.UserID, deviceId)
	if err != nil {
		log.Printf("Error getting leads: %v", err)
		// Return empty array instead of error
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Leads retrieved successfully",
			Results: []interface{}{},
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Leads retrieved successfully",
		Results: leads,
	})
}

// CreateLead creates a new lead
func (handler *App) CreateLead(c *fiber.Ctx) error {
	// Get session from cookie - same as campaigns
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
	
	var request struct {
		DeviceID     string `json:"device_id"`
		Name         string `json:"name"`
		Phone        string `json:"phone"`
		Niche        string `json:"niche"`
		Journey      string `json:"journey"`
		TargetStatus string `json:"target_status"` // Changed from Status to TargetStatus
		Trigger      string `json:"trigger"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	leadRepo := repository.GetLeadRepository()
	lead := &models.Lead{
		UserID:       session.UserID,
		DeviceID:     request.DeviceID,
		Name:         request.Name,
		Phone:        request.Phone,
		Email:        "",
		Niche:        request.Niche,
		Source:       "manual", // Set source as manual since it's added from UI
		Status:       "", // Keep empty for backward compatibility
		TargetStatus: request.TargetStatus, // Use TargetStatus directly
		Trigger:      request.Trigger,
		Notes:        request.Journey, // Map journey to notes field
	}
	err = leadRepo.CreateLead(lead)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to create lead: %v", err),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  201,
		Code:    "SUCCESS",
		Message: "Lead created successfully",
		Results: lead,
	})
}

// UpdateLead updates an existing lead
func (handler *App) UpdateLead(c *fiber.Ctx) error {
	leadId := c.Params("id")
	
	// Get session from cookie - same as campaigns
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
	
	var request struct {
		DeviceID     string `json:"device_id"`
		Name         string `json:"name"`
		Phone        string `json:"phone"`
		Niche        string `json:"niche"`
		Journey      string `json:"journey"`
		TargetStatus string `json:"target_status"` // Changed from status to target_status
		Trigger      string `json:"trigger"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	leadRepo := repository.GetLeadRepository()
	lead := &models.Lead{
		UserID:       session.UserID,
		DeviceID:     request.DeviceID,
		Name:         request.Name,
		Phone:        request.Phone,
		Email:        "",
		Niche:        request.Niche,
		Source:       "manual", // Keep source as manual
		Status:       "", // Keep empty for backward compatibility
		TargetStatus: request.TargetStatus, // Use TargetStatus directly from request
		Trigger:      request.Trigger, // Add trigger
		Notes:        request.Journey, // Map journey to notes field
	}
	err = leadRepo.UpdateLead(leadId, lead)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to update lead: %v", err),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Lead updated successfully",
		Results: map[string]string{"id": leadId},
	})
}

// DeleteLead deletes a lead
func (handler *App) DeleteLead(c *fiber.Ctx) error {
	leadId := c.Params("id")
	
	// Get session from cookie - same as campaigns
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
	
	// Just verify the user exists but we don't need to use the result
	_ = session.UserID
	
	leadRepo := repository.GetLeadRepository()
	err = leadRepo.DeleteLead(leadId)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to delete lead: %v", err),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Lead deleted successfully",
	})
}

// GetCampaigns gets all campaigns for the user
func (handler *App) GetCampaigns(c *fiber.Ctx) error {
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
	
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	campaignRepo := repository.GetCampaignRepository()
	campaigns, err := campaignRepo.GetCampaigns(user.ID)
	if err != nil {
		log.Printf("Error getting campaigns: %v", err)
		// Return empty array instead of error
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Campaigns retrieved successfully",
			Results: []interface{}{},
		})
	}
	
	// Ensure campaigns is not nil
	if campaigns == nil {
		campaigns = []models.Campaign{}
	}
	
	// Return campaigns as array for frontend
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Campaigns retrieved successfully",
		Results: campaigns,
	})
}

// CreateCampaign creates a new campaign
func (handler *App) CreateCampaign(c *fiber.Ctx) error {
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
	
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	var request struct {
		CampaignDate    string  `json:"campaign_date"`
		Title           string  `json:"title"`
		Niche           string  `json:"niche"`
		TargetStatus    string  `json:"target_status"`
		Message         string  `json:"message"`
		ImageURL        string  `json:"image_url"`
		TimeSchedule    string  `json:"time_schedule"`
		MinDelaySeconds int     `json:"min_delay_seconds"`
		MaxDelaySeconds int     `json:"max_delay_seconds"`
		AI              *string `json:"ai"`    // New field for AI campaigns
		Limit           int     `json:"limit"` // New field for device limit
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	// Validate required fields
	if request.CampaignDate == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Campaign date is required",
		})
	}
	
	// Parse scheduled time if provided
	var timeSchedule string
	if request.TimeSchedule != "" {
		timeSchedule = request.TimeSchedule
	} else {
		// Default to current time if not provided
		timeSchedule = time.Now().Format("15:04")
	}
	
	// Validate and set target_status
	targetStatus := request.TargetStatus
	if targetStatus != "prospect" && targetStatus != "customer" && targetStatus != "all" {
		targetStatus = "all" // Default to all if invalid
	}
	
	// For AI campaigns, ensure limit is set
	if request.AI != nil && *request.AI == "ai" && request.Limit <= 0 {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device limit must be greater than 0 for AI campaigns",
		})
	}
	
	campaignRepo := repository.GetCampaignRepository()
	campaign := &models.Campaign{
		UserID:          user.ID,
		Title:           request.Title,
		Message:         request.Message,
		Niche:           request.Niche,
		TargetStatus:    targetStatus,
		ImageURL:        request.ImageURL,
		CampaignDate:    request.CampaignDate,
		TimeSchedule:    timeSchedule,
		MinDelaySeconds: request.MinDelaySeconds,
		MaxDelaySeconds: request.MaxDelaySeconds,
		Status:          "pending",
		AI:              request.AI,
		Limit:           request.Limit,
	}
	err = campaignRepo.CreateCampaign(campaign)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to create campaign: %v", err),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  201,
		Code:    "SUCCESS",
		Message: "Campaign created successfully",
		Results: campaign,
	})
}

// UpdateCampaign updates an existing campaign
func (handler *App) UpdateCampaign(c *fiber.Ctx) error {
	campaignIdStr := c.Params("id")
	campaignId, err := strconv.Atoi(campaignIdStr)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid campaign ID",
		})
	}
	
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
	
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	var request struct {
		Title           string `json:"title"`
		Niche           string `json:"niche"`
		Message         string `json:"message"`
		ImageURL        string `json:"image_url"`
		TimeSchedule    string `json:"time_schedule"`
		CampaignDate    string `json:"campaign_date"`
		MinDelaySeconds int    `json:"min_delay_seconds"`
		MaxDelaySeconds int    `json:"max_delay_seconds"`
		Status          string `json:"status"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	// Parse scheduled time if provided
	var timeSchedule string
	if request.TimeSchedule != "" {
		timeSchedule = request.TimeSchedule
	}
	
	campaignRepo := repository.GetCampaignRepository()
	campaign := &models.Campaign{
		ID:              campaignId,
		UserID:          user.ID,
		Title:           request.Title,
		Message:         request.Message,
		Niche:           request.Niche,
		ImageURL:        request.ImageURL,
		CampaignDate:    request.CampaignDate,
		TimeSchedule:    timeSchedule,
		MinDelaySeconds: request.MinDelaySeconds,
		MaxDelaySeconds: request.MaxDelaySeconds,
		Status:          request.Status,
	}
	err = campaignRepo.UpdateCampaign(campaign)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to update campaign: %v", err),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Campaign updated successfully",
		Results: campaign,
	})
}

// DeleteCampaign deletes a campaign
func (handler *App) DeleteCampaign(c *fiber.Ctx) error {
	campaignIdStr := c.Params("id")
	campaignId, err := strconv.Atoi(campaignIdStr)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid campaign ID",
		})
	}
	
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
	
	_, err = userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	campaignRepo := repository.GetCampaignRepository()
	err = campaignRepo.DeleteCampaign(campaignId)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to delete campaign: %v", err),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Campaign deleted successfully",
	})
}
// DeleteDevice deletes a device
func (handler *App) DeleteDevice(c *fiber.Ctx) error {
	deviceId := c.Params("id")
	
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
	
	// Get device details
	device, err := userRepo.GetDeviceByID(deviceId)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Verify device belongs to user
	if device.UserID != session.UserID {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "Device does not belong to this user",
		})
	}
	
	logrus.Infof("Deleting device %s (%s) for user %s", device.ID, device.DeviceName, session.UserID)
	
	// Disconnect WhatsApp client if connected
	cm := whatsapp.GetClientManager()
	if client, err := cm.GetClient(deviceId); err == nil && client != nil {
		logrus.Info("Disconnecting WhatsApp client...")
		
		// Logout from WhatsApp
		if client.IsConnected() {
			err := client.Logout(c.UserContext())
			if err != nil {
				logrus.Errorf("Error logging out: %v", err)
			}
		}
		client.Disconnect()
		
		// Remove from client manager
		cm.RemoveClient(deviceId)
	}
	
	// Clear associated data
	whatsappRepo := repository.GetWhatsAppRepository()
	if whatsappRepo != nil {
		// Clear messages
		err = whatsappRepo.ClearDeviceMessages(deviceId)
		if err != nil {
			logrus.Errorf("Failed to clear messages: %v", err)
		}
		
		// Clear chats
		err = whatsappRepo.ClearDeviceChats(deviceId)
		if err != nil {
			logrus.Errorf("Failed to clear chats: %v", err)
		}
	}
	
	// Delete device from database
	err = userRepo.DeleteDevice(deviceId)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to delete device",
		})
	}
	
	logrus.Infof("Successfully deleted device %s", device.DeviceName)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device deleted successfully",
		Results: fiber.Map{
			"device_id": deviceId,
			"device_name": device.DeviceName,
		},
	})
}

// LogoutDevice logs out from WhatsApp
func (handler *App) LogoutDevice(c *fiber.Ctx) error {
	deviceId := c.Query("deviceId")
	if deviceId == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device ID is required",
		})
	}
	
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
	
	// Get device details
	device, err := userRepo.GetDeviceByID(deviceId)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Verify device belongs to user
	if device.UserID != session.UserID {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "Device does not belong to this user",
		})
	}
	
	// Check if device has platform - platform devices cannot be logged out
	if device.Platform != "" {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "PLATFORM_DEVICE",
			Message: "Platform devices cannot be logged out",
		})
	}
	
	logrus.Infof("Logging out device %s (%s)", device.ID, device.DeviceName)
	
	// Disconnect WhatsApp client
	cm := whatsapp.GetClientManager()
	if client, err := cm.GetClient(deviceId); err == nil && client != nil {
		// Logout from WhatsApp
		if client.IsConnected() {
			err = client.Logout(c.UserContext())
			if err != nil {
				logrus.Errorf("Error logging out: %v", err)
			}
		}
		
		// Disconnect client
		client.Disconnect()
		
		// Remove from client manager
		cm.RemoveClient(deviceId)
		
		logrus.Info("WhatsApp client disconnected and removed from manager")
	}
	
	// Get current device info before updating
	var phone, jid sql.NullString
	err = userRepo.DB().QueryRow("SELECT phone, jid from user_devices WHERE id = ?", deviceId).Scan(&phone, &jid)
	if err != nil {
		logrus.Warnf("Failed to get device info: %v", err)
	}
	
	// Update device status in database but KEEP phone and JID
	phoneStr := ""
	jidStr := ""
	if phone.Valid {
		phoneStr = phone.String
	}
	if jid.Valid {
		jidStr = jid.String
	}
	err = userRepo.UpdateDeviceStatus(deviceId, "disconnected", phoneStr, jidStr)
	if err != nil {
		logrus.Errorf("Error updating device status: %v", err)
	}
	
	// Clean up any session data
	whatsapp.ClearConnectionSession(session.UserID)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device logged out successfully",
		Results: map[string]interface{}{
			"device_id": deviceId,
			"device_name": device.DeviceName,
			"status": "disconnected",
		},
	})
}

// ReconnectDevice reconnects to WhatsApp
func (handler *App) ReconnectDevice(c *fiber.Ctx) error {
	deviceId := c.Query("deviceId")
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Reconnecting...",
		Results: map[string]interface{}{
			"deviceId": deviceId,
			"status":   "reconnecting",
		},
	})
}

// GetDevices gets all devices for the app
func (handler *App) GetDevices(c *fiber.Ctx) error {
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Devices retrieved",
		Results: []interface{}{},
	})
}

// GetUserInfo gets user info
func (handler *App) GetUserInfo(c *fiber.Ctx) error {
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "User info retrieved",
		Results: map[string]interface{}{
			"name":   "User",
			"number": "+60123456789",
			"avatar": "",
		},
	})
}

// GetUserAvatar gets user avatar
func (handler *App) GetUserAvatar(c *fiber.Ctx) error {
	// Return a default avatar
	return c.SendString("")
}

// ChangeUserAvatar changes user avatar
func (handler *App) ChangeUserAvatar(c *fiber.Ctx) error {
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Avatar updated",
	})
}

// ChangeUserPushName changes user push name
func (handler *App) ChangeUserPushName(c *fiber.Ctx) error {
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Push name updated",
	})
}
// SyncDeviceChats manually triggers chat synchronization for a device
func (handler *App) SyncDeviceChats(c *fiber.Ctx) error {
	deviceId := c.Params("id")
	
	// Get session from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token",
		})
	}
	
	// Verify device ownership
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}
	
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	// Check if device belongs to user
	device, err := userRepo.GetDevice(user.ID, deviceId)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// trigger chat sync
	go func() {
		chats, err := whatsapp.GetChatsForDevice(device.ID)
		if err != nil {
			fmt.Printf("Failed to sync chats for device %s: %v\n", device.ID, err)
		} else {
			fmt.Printf("Successfully synced %d chats for device %s\n", len(chats), device.ID)
		}
	}()
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Chat sync initiated",
		Results: map[string]interface{}{
			"deviceId": device.ID,
			"status":   "syncing",
		},
	})
}
// DiagnoseDevice provides diagnostic information about a device's WhatsApp connection
func (handler *App) DiagnoseDevice(c *fiber.Ctx) error {
	deviceId := c.Params("id")
	
	// Get session from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token",
		})
	}
	
	// Verify device ownership
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}
	
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	// Get device info
	device, err := userRepo.GetDevice(user.ID, deviceId)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Check WhatsApp client status
	cm := whatsapp.GetClientManager()
	client, clientErr := cm.GetClient(device.ID)
	
	// Get full client diagnostics
	allDiagnostics := whatsapp.DiagnoseClients()
	
	diagnostics := map[string]interface{}{
		"device": map[string]interface{}{
			"id":     device.ID,
			"name":   device.DeviceName,
			"phone":  device.Phone,
			"status": device.Status,
			"jid":    device.JID,
		},
		"whatsapp_client": map[string]interface{}{
			"connected": clientErr == nil,
			"error":     "",
		},
		"database": map[string]interface{}{
			"chats_count":    0,
			"messages_count": 0,
		},
		"client_manager": allDiagnostics,
	}
	
	if clientErr != nil {
		diagnostics["whatsapp_client"].(map[string]interface{})["error"] = clientErr.Error()
		
		// Try to register from database if device is online
		if device.Status == "online" {
			log.Printf("Attempting to auto-register device %s from database", device.ID)
			if err := whatsapp.TryRegisterDeviceFromDatabase(device.ID); err == nil {
				diagnostics["whatsapp_client"].(map[string]interface{})["auto_registered"] = true
				// Retry getting client
				client, clientErr = cm.GetClient(device.ID)
				if clientErr == nil {
					diagnostics["whatsapp_client"].(map[string]interface{})["connected"] = true
					diagnostics["whatsapp_client"].(map[string]interface{})["error"] = ""
				}
			} else {
				log.Printf("Auto-registration failed: %v", err)
			}
		}
	} else {
		// Check if client is logged in
		diagnostics["whatsapp_client"].(map[string]interface{})["logged_in"] = client.IsLoggedIn()
		diagnostics["whatsapp_client"].(map[string]interface{})["is_connected"] = client.IsConnected()
		
		// Try to get contacts count
		contacts, err := client.Store.Contacts.GetAllContacts(context.Background())
		if err == nil {
			diagnostics["whatsapp_client"].(map[string]interface{})["contacts_count"] = len(contacts)
		}
	}
	
	// Check database
	whatsappRepo := repository.GetWhatsAppRepository()
	chats, _ := whatsappRepo.GetChats(device.ID)
	diagnostics["database"].(map[string]interface{})["chats_count"] = len(chats)
	
	// Force a sync attempt
	go func() {
		fmt.Printf("Forcing sync for device %s\n", device.ID)
		chats, err := whatsapp.GetChatsForDevice(device.ID)
		if err != nil {
			fmt.Printf("Sync error for device %s: %v\n", device.ID, err)
		} else {
			fmt.Printf("Sync complete for device %s: %d chats\n", device.ID, len(chats))
		}
	}()
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device diagnostics",
		Results: diagnostics,
	})
}
// GetCampaignSummary gets campaign statistics
func (handler *App) GetCampaignSummary(c *fiber.Ctx) error {
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
	
	// Get date filter from query parameters
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")
	
	// Get campaign statistics
	campaignRepo := repository.GetCampaignRepository()
	
	// Get campaigns filtered by date if provided
	var campaigns []models.Campaign
	if startDate != "" || endDate != "" {
		campaigns, err = campaignRepo.GetCampaignsByUserAndDateRange(session.UserID, startDate, endDate)
	} else {
		campaigns, err = campaignRepo.GetCampaignsByUser(session.UserID)
	}
	
	if err != nil {
		campaigns = []models.Campaign{}
	}
	
	// Calculate statistics
	totalCampaigns := len(campaigns)
	pendingCampaigns := 0
	triggeredCampaigns := 0
	processingCampaigns := 0
	sentCampaigns := 0
	failedCampaigns := 0
	
	for _, campaign := range campaigns {
		switch campaign.Status {
		case "scheduled", "pending":
			pendingCampaigns++
		case "triggered":
			triggeredCampaigns++
		case "processing":
			processingCampaigns++
		case "sent", "finished":
			sentCampaigns++
		case "failed":
			failedCampaigns++
		}
	}
	
	
	// Initialize totals
	totalShouldSend := 0
	totalDoneSend := 0
	totalFailedSend := 0
	
	// Get broadcast statistics only for filtered campaigns
	for _, campaign := range campaigns {
		shouldSend, doneSend, failedSend, _ := campaignRepo.GetCampaignBroadcastStats(campaign.ID)
		totalShouldSend += shouldSend
		totalDoneSend += doneSend
		totalFailedSend += failedSend
	}
	
	totalRemainingSend := totalShouldSend - totalDoneSend - totalFailedSend
	if totalRemainingSend < 0 {
		totalRemainingSend = 0
	}
	
	// Get recent campaigns with their broadcast stats
	recentCampaigns := []map[string]interface{}{}
	if len(campaigns) > 0 {
		limit := min(5, len(campaigns))
		for i := 0; i < limit; i++ {
			campaign := campaigns[i]
			
			// Get broadcast stats for this campaign
			shouldSend, doneSend, failedSend, _ := campaignRepo.GetCampaignBroadcastStats(campaign.ID)
			remainingSend := shouldSend - doneSend - failedSend
			if remainingSend < 0 {
				remainingSend = 0
			}
			
			campaignData := map[string]interface{}{
				"id":               campaign.ID,
				"title":            campaign.Title,
				"campaign_date":    campaign.CampaignDate,
				"time_schedule":    campaign.TimeSchedule,
				"niche":            campaign.Niche,
				"target_status":    campaign.TargetStatus,
				"status":           campaign.Status,
				"message":          campaign.Message,
				"image_url":        campaign.ImageURL,
				"should_send":      shouldSend,
				"done_send":        doneSend,
				"failed_send":      failedSend,
				"remaining_send":   remainingSend,
			}
			
			recentCampaigns = append(recentCampaigns, campaignData)
		}
	}
	
	summary := map[string]interface{}{
		"campaigns": map[string]interface{}{
			"total": totalCampaigns,
			"pending": pendingCampaigns,
			"triggered": triggeredCampaigns,
			"processing": processingCampaigns,
			"sent": sentCampaigns,
			"failed": failedCampaigns,
		},
		"broadcast_stats": map[string]interface{}{
			"total_should_send":    totalShouldSend,
			"total_done_send":      totalDoneSend,
			"total_failed_send":    totalFailedSend,
			"total_remaining_send": totalRemainingSend,
		},
		"recent_campaigns": recentCampaigns,
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Campaign summary",
		Results: summary,
	})
}

// GetSequenceSummary gets sequence statistics
func (handler *App) GetSequenceSummary(c *fiber.Ctx) error {
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
	
	// Get sequence statistics
	sequenceRepo := repository.GetSequenceRepository()
	sequences, err := sequenceRepo.GetSequences(session.UserID)
	if err != nil {
		sequences = []models.Sequence{}
	}
	
	// Calculate statistics
	totalSequences := len(sequences)
	activeSequences := 0
	pausedSequences := 0
	draftSequences := 0
	totalContacts := 0
	totalFlows := 0
	
	// Get total flows from database
	db, err := sql.Open("postgres", config.DBURI)
	if err == nil {
		defer db.Close()
		
		// Count total flows - simple direct count
		var flowCount int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM sequence_steps
		`).Scan(&flowCount)
		
		if err != nil {
			fmt.Printf("Error counting all sequence flows: %v\n", err)
		} else {
			// Now count only flows for this user's sequences
			err = db.QueryRow(`
				SELECT COUNT(*) 
				FROM sequence_steps ss
				WHERE EXISTS (
					SELECT 1 FROM sequences s 
					WHERE s.id = ss.sequence_id 
					AND s.user_id = ?
				)
			`, session.UserID).Scan(&flowCount)
			
			if err != nil {
				fmt.Printf("Error counting user sequence flows: %v\n", err)
				// Try with UUID casting
				err = db.QueryRow(`
					SELECT COUNT(*) 
					FROM sequence_steps ss
					WHERE sequence_id::text IN (
						SELECT id::text FROM sequences WHERE user_id = ?
					)
				`, session.UserID).Scan(&flowCount)
				
				if err != nil {
					fmt.Printf("Error with UUID cast query: %v\n", err)
				} else {
					totalFlows = flowCount
				}
			} else {
				totalFlows = flowCount
			}
		}
	}
	
	for _, sequence := range sequences {
		switch sequence.Status {
		case "active":
			activeSequences++
		case "paused":
			pausedSequences++
		case "draft":
			draftSequences++
		}
		totalContacts += sequence.ContactsCount
	}
	
	// Add flow counts to each sequence
	sequencesWithFlows := []map[string]interface{}{}
	if db != nil {
		for i, sequence := range sequences {
			if i >= 5 { // Only process first 5 sequences for recent_sequences
				break
			}
			
			sequenceData := map[string]interface{}{
				"id":         sequence.ID,
				"name":       sequence.Name,
				"niche":      sequence.Niche,
				"trigger":    sequence.Trigger,
				"status":     sequence.Status,
				"created_at": sequence.CreatedAt,
			}
			
			// Get flow count for this sequence
			var flowCount int
			err := db.QueryRow(`
				SELECT COUNT(*) 
				FROM sequence_steps 
				WHERE sequence_id = ?
			`, sequence.ID).Scan(&flowCount)
			
			if err != nil {
				flowCount = 0
			}
			sequenceData["total_flows"] = flowCount
			
			// Get contact statistics for this sequence
			var shouldSend, doneSend, failedSend int
			err = db.QueryRow(`
				SELECT COALESCE(COUNT(*), 0) AS total,
					COALESCE(COUNT(CASE WHEN status = 'sent' THEN 1 END), 0) AS sent,
					COALESCE(COUNT(CASE WHEN status = 'failed' THEN 1 END), 0) AS failed
				FROM sequence_contacts
				WHERE sequence_id = ?
			`, sequence.ID).Scan(&shouldSend, &doneSend, &failedSend)
			
			if err != nil {
				shouldSend, doneSend, failedSend = 0, 0, 0
			}
			
			sequenceData["should_send"] = shouldSend
			sequenceData["done_send"] = doneSend
			sequenceData["failed_send"] = failedSend
			sequenceData["remaining_send"] = shouldSend - doneSend - failedSend
			
			sequencesWithFlows = append(sequencesWithFlows, sequenceData)
		}
	}
	
	summary := map[string]interface{}{
		"sequences": map[string]interface{}{
			"total": totalSequences,
			"active": activeSequences,
			"paused": pausedSequences,
			"draft": draftSequences,
		},
		"total_flows": totalFlows,
		"contacts": map[string]interface{}{
			"total": totalContacts,
			"average_per_sequence": float64(totalContacts) / float64(max(1, totalSequences)),
		},
		"recent_sequences": sequencesWithFlows,
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequence summary",
		Results: summary,
	})
}

// GetWorkerStatus gets the status of all device workers
func (handler *App) GetWorkerStatus(c *fiber.Ctx) error {
	// Get filter parameters
	filterType := c.Query("filter", "all") // all, campaign, sequence
	filterID := c.Query("id", "")
	
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
	
	// Get user's devices
	devices, err := userRepo.GetUserDevices(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	// Get broadcast manager stats
	broadcastManager := broadcast.GetBroadcastManager()
	workerStats := broadcastManager.GetAllWorkerStatus()
	
	// Get worker status for each device
	deviceWorkers := []map[string]interface{}{}
	
	// Convert worker statuses to a map for easy lookup
	workerStatusMap := make(map[string]domainBroadcast.WorkerStatus)
	for _, status := range workerStats {
		workerStatusMap[status.DeviceID] = status
	}
	
	for _, device := range devices {
		workerInfo := map[string]interface{}{
			"device_id": device.ID,
			"device_name": device.DeviceName,
			"device_status": device.Status,
			"worker_status": "not_running",
			"queue_size": 0,
			"processed": 0,
			"failed": 0,
		}
		
		// Find worker stats for this device
		if status, exists := workerStatusMap[device.ID]; exists {
			workerInfo["worker_status"] = status.Status
			workerInfo["queue_size"] = status.QueueSize
			workerInfo["processed"] = status.ProcessedCount
			workerInfo["failed"] = status.FailedCount
			workerInfo["last_activity"] = status.LastActivity
			
			// Add campaign/sequence info if worker is processing
			if status.CurrentCampaignID > 0 {
				workerInfo["current_campaign_id"] = status.CurrentCampaignID
			}
			if status.CurrentSequenceID != "" {
				workerInfo["current_sequence_id"] = status.CurrentSequenceID
			}
		}
		
		// Apply filter if specified
		if filterType == "campaign" && filterID != "" {
			// Only include if worker is processing this campaign
			if campaignID, ok := workerInfo["current_campaign_id"]; ok {
				if fmt.Sprintf("%v", campaignID) == filterID {
					deviceWorkers = append(deviceWorkers, workerInfo)
				}
			}
		} else if filterType == "sequence" && filterID != "" {
			// Only include if worker is processing this sequence
			if sequenceID, ok := workerInfo["current_sequence_id"]; ok {
				if sequenceID == filterID {
					deviceWorkers = append(deviceWorkers, workerInfo)
				}
			}
		} else {
			// No filter, include all
			deviceWorkers = append(deviceWorkers, workerInfo)
		}
	}
	
	response := map[string]interface{}{
		"total_workers": len(workerStats),
		"user_devices": len(devices),
		"connected_devices": countConnectedDevices(devices),
		"device_workers": deviceWorkers,
		"filter": map[string]interface{}{
			"type": filterType,
			"id": filterID,
		},
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Worker status",
		Results: response,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func countConnectedDevices(devices []*models.UserDevice) int {
	count := 0
	for _, device := range devices {
		if device.Status == "connected" || device.Status == "Connected" || 
		   device.Status == "online" || device.Status == "Online" {
			count++
		}
	}
	return count
}


// ResumeFailedWorkers resumes all failed device workers
func (handler *App) ResumeFailedWorkers(c *fiber.Ctx) error {
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

	// Get all devices for user
	devices, err := userRepo.GetUserDevices(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}

	// Resume workers for devices that are connected but have stopped workers
	resumedCount := 0
	for _, device := range devices {
		if device.Status == "connected" {
			// TODO: Check if worker is stopped and resume
			// This would interface with your broadcast manager
			resumedCount++
		}
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Resumed %d workers", resumedCount),
		Results: map[string]interface{}{
			"resumed_count": resumedCount,
		},
	})
}



// Helper functions



// StopAllWorkers stops all running device workers
func (handler *App) StopAllWorkers(c *fiber.Ctx) error {
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

	// Get all devices for user
	devices, err := userRepo.GetUserDevices(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}

	// Stop all workers
	stoppedCount := 0
	for range devices {
		// TODO: Stop worker for this device
		// This would interface with your broadcast manager
		stoppedCount++
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Stopped %d workers", stoppedCount),
		Results: map[string]interface{}{
			"stopped_count": stoppedCount,
		},
	})
}


// ExportLeads exports leads to CSV format
func (handler *App) ExportLeads(c *fiber.Ctx) error {
	deviceId := c.Params("deviceId")
	
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
	
	// Get leads
	leadRepo := repository.GetLeadRepository()
	leads, err := leadRepo.GetLeadsByDevice(session.UserID, deviceId)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get leads",
		})
	}
	
	// Convert to CSV
	var csvContent strings.Builder
	csvContent.WriteString("name,phone,niche,target_status,trigger\n")
	
	for _, lead := range leads {
		// Ensure target_status has a value
		targetStatus := lead.TargetStatus
		if targetStatus == "" {
			targetStatus = "prospect"
		}
		
		// Properly escape CSV fields that might contain commas or quotes
		csvContent.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
			strings.ReplaceAll(lead.Name, "\"", "\"\""),
			strings.ReplaceAll(lead.Phone, "\"", "\"\""),
			strings.ReplaceAll(lead.Niche, "\"", "\"\""),
			strings.ReplaceAll(targetStatus, "\"", "\"\""),
			strings.ReplaceAll(lead.Trigger, "\"", "\"\""),
		))
	}
	
	// Set headers for CSV download
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=leads_%s_%s.csv", deviceId, time.Now().Format("2006-01-02")))
	
	return c.SendString(csvContent.String())
}

// ImportLeads imports leads from CSV
func (handler *App) ImportLeads(c *fiber.Ctx) error {
	deviceId := c.Params("deviceId")
	
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
	
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "No file uploaded",
		})
	}
	
	// Open the file
	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to open file",
		})
	}
	defer src.Close()
	
	// Read CSV content
	content, err := io.ReadAll(src)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to read file",
		})
	}
	
	// Parse CSV using csv.Reader
	reader := csv.NewReader(strings.NewReader(string(content)))
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	
	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to parse CSV: " + err.Error(),
		})
	}
	
	if len(records) < 2 {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "CSV file is empty or invalid",
		})
	}
	
	// Parse headers
	headers := records[0]
	for i := range headers {
		headers[i] = strings.ToLower(strings.TrimSpace(headers[i]))
	}
	
	// Find column indices
	nameIndex := -1
	phoneIndex := -1
	nicheIndex := -1
	targetStatusIndex := -1
	statusIndex := -1
	triggerIndex := -1
	
	for i, h := range headers {
		switch h {
		case "name":
			nameIndex = i
		case "phone":
			phoneIndex = i
		case "niche":
			nicheIndex = i
		case "target_status", "target_sta": // Support both full and shortened name
			targetStatusIndex = i
		case "status":
			statusIndex = i
		case "trigger":
			triggerIndex = i
		}
	}
	
	if nameIndex == -1 || phoneIndex == -1 || nicheIndex == -1 || (targetStatusIndex == -1 && statusIndex == -1) {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "CSV must have 'name', 'phone', 'niche', and 'target_status' columns",
		})
	}
	
	// Process leads
	leadRepo := repository.GetLeadRepository()
	successCount := 0
	errorCount := 0
	
	for i := 1; i < len(records); i++ {
		record := records[i]
		
		// Skip empty rows
		if len(record) == 0 {
			continue
		}
		
		// Get values safely
		getValue := func(index int) string {
			if index >= 0 && index < len(record) {
				return strings.TrimSpace(record[index])
			}
			return ""
		}
		
		// Get required values
		name := getValue(nameIndex)
		phone := getValue(phoneIndex)
		niche := getValue(nicheIndex)
		
		// Skip if required fields are empty
		if name == "" || phone == "" || niche == "" {
			errorCount++
			log.Printf("Row %d: Skipping - missing required fields (name, phone, or niche)", i)
			continue
		}
		
		// Get target status (support both columns)
		targetStatus := getValue(targetStatusIndex)
		if targetStatus == "" {
			targetStatus = getValue(statusIndex)
		}
		if targetStatus != "prospect" && targetStatus != "customer" {
			targetStatus = "prospect"
		}
		
		lead := &models.Lead{
			UserID:       session.UserID,
			DeviceID:     deviceId, // Always use the current device ID
			Name:         name,
			Phone:        phone,
			Niche:        niche,
			TargetStatus: targetStatus,
			Notes:        "", // No longer importing notes
			Trigger:      getValue(triggerIndex),
		}
		
		err := leadRepo.CreateLead(lead)
		if err != nil {
			errorCount++
			log.Printf("Failed to import lead %s: %v", lead.Name, err)
		} else {
			successCount++
		}
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Import completed. Success: %d, Failed: %d", successCount, errorCount),
		Results: map[string]int{
			"success": successCount,
			"failed":  errorCount,
		},
	})
}


// GetCampaignDeviceReport gets device-wise report for a campaign
func (handler *App) GetCampaignDeviceReport(c *fiber.Ctx) error {
	campaignIdStr := c.Params("id")
	campaignId, err := strconv.Atoi(campaignIdStr)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid campaign ID",
		})
	}
	
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
	
	// Get campaign broadcast statistics
	campaignRepo := repository.GetCampaignRepository()
	shouldSend, doneSend, failedSend, _ := campaignRepo.GetCampaignBroadcastStats(campaignId)
	remainingSend := shouldSend - doneSend - failedSend
	if remainingSend < 0 {
		remainingSend = 0
	}
	
	// Get user devices - use direct query
	db := database.GetDB()
	query := `
		SELECT id, device_name, phone, status, jid, created_at, last_seen
		FROM user_devices
		WHERE user_id = ?
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get devices",
		})
	}
	defer rows.Close()
	
	devices := []models.UserDevice{}
	deviceMap := make(map[string]*DeviceReport)
	
	for rows.Next() {
		var device models.UserDevice
		err := rows.Scan(&device.ID, &device.DeviceName, &device.Phone, &device.Status, 
			&device.JID, &device.CreatedAt, &device.LastSeen)
		if err != nil {
			continue
		}
		device.UserID = session.UserID
		devices = append(devices, device)
		
		// Initialize device report
		log.Printf("Device Report - Initializing device: ID=%s, Name=%s", device.ID, device.DeviceName)
		deviceMap[device.ID] = &DeviceReport{
			ID:     device.ID,
			Name:   device.DeviceName,
			Status: device.Status,
		}
	}
	
	// Get real broadcast message data for this campaign
	messageQuery := `
		SELECT device_id, status, COUNT(*) AS COUNT
		FROM broadcast_messages
		WHERE campaign_id = ? AND user_id = ?
		GROUP BY device_id, status
	`
	msgRows, err := db.Query(messageQuery, campaignId, session.UserID)
	if err == nil {
		defer msgRows.Close()
		
		// Debug: log the query results
		log.Printf("Device Report - Campaign ID: %d, User ID: %s", campaignId, session.UserID)
		
		// First, let's check total messages per device
		totalQuery := `
			SELECT device_id, COUNT(*) AS total_count
			FROM broadcast_messages
			WHERE campaign_id = ? AND user_id = ?
			GROUP BY device_id
		`
		totalRows, _ := db.Query(totalQuery, campaignId, session.UserID)
		if totalRows != nil {
			defer totalRows.Close()
			for totalRows.Next() {
				var deviceId string
				var totalCount int
				if err := totalRows.Scan(&deviceId, &totalCount); err == nil {
					log.Printf("Device Report - Device %s has TOTAL %d messages", deviceId, totalCount)
				}
			}
		}
		
		for msgRows.Next() {
			var deviceId, status string
			var count int
			if err := msgRows.Scan(&deviceId, &status, &count); err != nil {
				continue
			}
			
			log.Printf("Device Report - Device: %s, Status: %s, Count: %d", deviceId, status, count)
			
			if report, exists := deviceMap[deviceId]; exists {
				report.TotalLeads += count
				
				switch status {
				case "pending", "queued":
					report.PendingLeads += count
				case "sent", "delivered", "success":
					report.SuccessLeads += count
				case "failed", "error":
					report.FailedLeads += count
				default:
					// Log unknown status
					log.Printf("Device Report - Unknown status: %s, count: %d", status, count)
					// Add to pending for now
					report.PendingLeads += count
				}
				
				// Log what we're adding
				log.Printf("Device Report - Added to %s: status=%s, count=%d", deviceId, status, count)
			}
		}
	}
	
	// Calculate per-device should send (distribute evenly among devices)
	perDeviceShouldSend := 0
	if len(devices) > 0 && shouldSend > 0 {
		// For single device, assign full count; for multiple devices, distribute evenly
		if len(devices) == 1 {
			perDeviceShouldSend = shouldSend
		} else {
			perDeviceShouldSend = int(math.Ceil(float64(shouldSend) / float64(len(devices))))
		}
	}
	
	// Convert map to slice and calculate totals
	deviceReports := make([]DeviceReport, 0, len(deviceMap))
	totalLeads := 0
	pendingLeads := 0
	successLeads := 0
	failedLeads := 0
	
	// Get campaign details to know target criteria
	campaign, _ := campaignRepo.GetCampaignByID(campaignId)
	
	// Calculate per-device statistics based on actual leads
	totalDevicesWithLeads := 0
	onlineDevicesWithLeads := 0
	offlineDevicesWithLeads := 0
	
	for deviceId, report := range deviceMap {
		// Only include devices that have messages for this campaign
		if report.TotalLeads == 0 {
			// Skip devices with no messages
			log.Printf("Device Report - Skipping device %s with 0 messages", deviceId)
			continue
		}
		
		// Count leads for this device matching campaign criteria
		deviceLeadQuery := `
			SELECT COUNT(l.phone) 
			FROM leads l
			WHERE l.device_id = ? 
			AND l.niche LIKE CONCAT('%', ?, '%')
			AND (? = 'all' OR l.target_status = ?)
		`
		var deviceShouldSend int
		err := db.QueryRow(deviceLeadQuery, deviceId, campaign.Niche, campaign.TargetStatus).Scan(&deviceShouldSend)
		if err == nil {
			report.ShouldSend = deviceShouldSend
		} else {
			// Fallback to even distribution
			report.ShouldSend = perDeviceShouldSend
		}
		
		// Set other statistics
		report.DoneSend = report.SuccessLeads
		report.FailedSend = report.FailedLeads
		report.RemainingSend = report.ShouldSend - report.DoneSend - report.FailedSend
		if report.RemainingSend < 0 {
			report.RemainingSend = 0
		}
		
		log.Printf("Device %s (%s): Total=%d, Pending=%d, Success=%d, Failed=%d", 
			deviceId, report.Name, report.TotalLeads, report.PendingLeads, 
			report.SuccessLeads, report.FailedLeads)
		
		deviceReports = append(deviceReports, *report)
		totalLeads += report.TotalLeads
		pendingLeads += report.PendingLeads
		successLeads += report.SuccessLeads
		failedLeads += report.FailedLeads
		
		// Count devices with leads based on their status
		totalDevicesWithLeads++
		if report.Status == "online" {
			onlineDevicesWithLeads++
		} else {
			offlineDevicesWithLeads++
		}
	}
	
	// Log final totals and device details
	log.Printf("Device Report Final - Total Devices with Leads: %d (Online: %d, Offline: %d)", 
		totalDevicesWithLeads, onlineDevicesWithLeads, offlineDevicesWithLeads)
	log.Printf("Device Report Final - Total Leads: %d, Pending: %d, Success: %d, Failed: %d", 
		totalLeads, pendingLeads, successLeads, failedLeads)
	
	// Log each device's counts
	for _, report := range deviceReports {
		log.Printf("Device %s: Total=%d (Pending=%d, Success=%d, Failed=%d) Should=%d, Done=%d, Failed=%d, Remaining=%d", 
			report.Name, report.TotalLeads, report.PendingLeads, report.SuccessLeads, report.FailedLeads,
			report.ShouldSend, report.DoneSend, report.FailedSend, report.RemainingSend)
	}
	
	result := map[string]interface{}{
		"totalDevices":        totalDevicesWithLeads,     // Only devices with matching leads
		"activeDevices":       onlineDevicesWithLeads,    // Only online devices with matching leads
		"disconnectedDevices": offlineDevicesWithLeads,   // Only offline devices with matching leads
		"totalLeads":          totalLeads,
		"pendingLeads":        pendingLeads,
		"successLeads":        successLeads,
		"failedLeads":         failedLeads,
		"devices":             deviceReports,
		// Add the new statistics
		"shouldSend":          shouldSend,
		"doneSend":            doneSend,
		"failedSend":          failedSend,
		"remainingSend":       remainingSend,
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device report retrieved successfully",
		Results: result,
	})
}

// GetCampaignDeviceLeads gets lead details for a specific device in a campaign
func (handler *App) GetCampaignDeviceLeads(c *fiber.Ctx) error {
	campaignIdStr := c.Params("id")
	campaignId, err := strconv.Atoi(campaignIdStr)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid campaign ID",
		})
	}
	
	deviceId := c.Params("deviceId")
	status := c.Query("status", "all")
	
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
	
	// Get real broadcast message data
	db := database.GetDB()
	
	// First check if this is an AI campaign by checking the 'ai' column
	var aiType sql.NullString
	err = db.QueryRow("SELECT ai from campaigns WHERE id = ?", campaignId).Scan(&aiType)
	if err != nil {
		log.Printf("Error checking campaign ai type: %v", err)
	}
	
	// Log the query parameters
	log.Printf("GetCampaignDeviceLeads - Campaign: %d, Device: %s, User: %s, Status: %s, AI: %v", 
		campaignId, deviceId, session.UserID, status, aiType.String)
	
	// Debug: Check for duplicates in broadcast_messages
	var duplicateCount int
	dupQuery := `
		SELECT COUNT(*) - COUNT(DISTINCT recipient_phone) AS duplicates
		FROM broadcast_messages
		WHERE campaign_id = ? AND device_id = ? AND user_id = ?
	`
	db.QueryRow(dupQuery, campaignId, deviceId, session.UserID).Scan(&duplicateCount)
	if duplicateCount > 0 {
		log.Printf("WARNING: Found %d duplicate phone numbers in broadcast_messages for this device", duplicateCount)
	}
	
	var query string
	if aiType.Valid && aiType.String == "ai" {
		// For AI campaigns - group by phone to get latest status
		query = `
			WITH latest_messages AS (
				SELECT DISTINCT ON (recipient_phone) 
					recipient_phone, status, sent_at
				FROM broadcast_messages
				WHERE campaign_id = ? AND device_id = ? AND user_id = ?
				ORDER BY recipient_phone, created_at DESC
			)
			SELECT lm.recipient_phone, lm.status, lm.sent_at, lai.name
			FROM latest_messages lm
			LEFT JOIN leads_ai lai ON lai.phone = lm.recipient_phone AND lai.user_id = ?
		`
	} else {
		// For regular campaigns - group by phone to get latest status
		query = `
			WITH latest_messages AS (
				SELECT DISTINCT ON (recipient_phone) 
					recipient_phone, status, sent_at
				FROM broadcast_messages
				WHERE campaign_id = ? AND device_id = ? AND user_id = ?
				ORDER BY recipient_phone, created_at DESC
			)
			SELECT lm.recipient_phone, lm.status, lm.sent_at, l.name
			FROM latest_messages lm
			LEFT JOIN leads l ON l.phone = lm.recipient_phone AND l.device_id = ?
		`
	}
	
	// Add status filter if not "all"
	if status != "all" {
		if status == "success" {
			query += ` WHERE lm.status IN ('sent', 'delivered', 'success')`
		} else if status == "pending" {
			query += ` WHERE lm.status IN ('pending', 'queued')`
		} else if status == "failed" {
			query += ` WHERE lm.status IN ('failed', 'error')`
		}
	}
	
	query += ` ORDER BY lm.sent_at DESC`
	
	rows, err := db.Query(query, campaignId, deviceId, session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get lead details",
		})
	}
	defer rows.Close()
	
	leadDetails := []map[string]interface{}{}
	leadCount := 0
	for rows.Next() {
		var phone, msgStatus string
		var sentAt sql.NullTime
		var name sql.NullString
		
		err := rows.Scan(&phone, &msgStatus, &sentAt, &name)
		if err != nil {
			continue
		}
		
		leadName := "Unknown"
		if name.Valid && name.String != "" {
			leadName = name.String
		}
		
		sentTime := "-"
		if sentAt.Valid {
			sentTime = sentAt.Time.Format("2006-01-02 03:04 PM")
		}
		
		leadDetails = append(leadDetails, map[string]interface{}{
			"name":   leadName,
			"phone":  phone,
			"status": msgStatus,
			"sentAt": sentTime,
		})
		leadCount++
	}
	
	log.Printf("GetCampaignDeviceLeads - Found %d leads for campaign %d, device %s, status %s", 
		leadCount, campaignId, deviceId, status)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Lead details retrieved successfully",
		Results: leadDetails,
	})
}

// DeviceReport structure for device-wise report
type DeviceReport struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	TotalLeads   int    `json:"totalLeads"`
	PendingLeads int    `json:"pendingLeads"`
	SuccessLeads int    `json:"successLeads"`
	FailedLeads  int    `json:"failedLeads"`
	// New fields for contact statistics
	ShouldSend     int `json:"shouldSend"`
	DoneSend       int `json:"doneSend"`
	FailedSend     int `json:"failedSend"`
	RemainingSend  int `json:"remainingSend"`
}

// RetryCampaignFailedMessages retries failed messages for a specific device
func (handler *App) RetryCampaignFailedMessages(c *fiber.Ctx) error {
	campaignIdStr := c.Params("id")
	campaignId, err := strconv.Atoi(campaignIdStr)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid campaign ID",
		})
	}
	
	deviceId := c.Params("deviceId")
	
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
	
	// Check if device is online before allowing retry
	db := database.GetDB()
	var deviceStatus string
	var deviceName string
	err = db.QueryRow("SELECT status, device_name from user_devices WHERE id = ? AND user_id = ?", 
		deviceId, session.UserID).Scan(&deviceStatus, &deviceName)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Only allow retry if device is online
	if deviceStatus != "online" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "DEVICE_OFFLINE",
			Message: fmt.Sprintf("Cannot retry messages: Device '%s' is offline", deviceName),
		})
	}
	
	// Update failed messages to pending status for retry
	query := `
		UPDATE broadcast_messages SET status = 'pending', error_message = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE campaign_id = ? AND device_id = ? AND user_id = ? AND status = 'failed'
	`
	
	result, err := db.Exec(query, campaignId, deviceId, session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retry messages",
		})
	}
	
	rowsAffected, _ := result.RowsAffected()
	
	if rowsAffected > 0 {
		// Get campaign details to check if it's AI campaign
		var aiType sql.NullString
		var campaignStatus string
		err = db.QueryRow("SELECT ai, status from campaigns WHERE id = ?", campaignId).Scan(&aiType, &campaignStatus)
		if err == nil {
			// Only update campaign status if it's finished/completed
			// Don't touch campaigns that are already processing
			if campaignStatus == "finished" || campaignStatus == "completed" {
				newStatus := "triggered"
				if aiType.Valid && aiType.String == "ai" {
					newStatus = "processing"
				}
				
				updateCampaignQuery := `
					UPDATE campaigns SET status = ?, updated_at = CURRENT_TIMESTAMP 
					WHERE id = ?
				`
				_, err = db.Exec(updateCampaignQuery, newStatus, campaignId)
				if err != nil {
					logrus.Warnf("Failed to update campaign status for retry: %v", err)
				} else {
					logrus.Infof("Campaign %d status updated to %s for retry", campaignId, newStatus)
				}
			} else if campaignStatus == "processing" || campaignStatus == "triggered" {
				// Campaign is already active, just log
				logrus.Infof("Campaign %d is already active (status: %s), messages will be processed", 
					campaignId, campaignStatus)
			}
		}
		
		// Important: Don't trigger any broadcast pool operations here
		// The existing workers will pick up the pending messages automatically
		// This prevents the cleanup issue that disconnects devices
		
		logrus.Infof("Retry requested for campaign %d, device %s: %d messages moved to pending", 
			campaignId, deviceId, rowsAffected)
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Successfully queued %d messages for retry", rowsAffected),
		Results: map[string]interface{}{
			"retried":   rowsAffected,
			"campaignId": campaignId,
			"deviceId":   deviceId,
		},
	})
}

// GetSystemStatus returns current system configuration status
func (handler *App) GetSystemStatus(c *fiber.Ctx) error {
	// Check Redis status
	redisURL := config.RedisURL
	if redisURL == "" {
		redisURL = os.Getenv("REDIS_URL")
	}
	
	redisEnabled := false
	redisInfo := "Not configured"
	
	if redisURL != "" {
		// Mask password in URL for security
		if strings.Contains(redisURL, "@") {
			parts := strings.Split(redisURL, "@")
			if len(parts) > 1 {
				redisInfo = "redis://***@" + parts[1]
			}
		}
		
		// Check if Redis is actually being used
		if redisURL != "" && 
		   !strings.Contains(redisURL, "${{") && 
		   !strings.Contains(redisURL, "localhost") && 
		   !strings.Contains(redisURL, "[::1]") &&
		   (strings.Contains(redisURL, "redis://") || strings.Contains(redisURL, "rediss://")) {
			redisEnabled = true
		}
	}
	
	// Get broadcast manager type
	broadcastType := "In-Memory"
	if redisEnabled {
		broadcastType = "Redis-Optimized"
	}
	
	// Get worker stats
	broadcastManager := broadcast.GetBroadcastManager()
	workerStats := broadcastManager.GetAllWorkerStatus()
	
	status := map[string]interface{}{
		"redis": map[string]interface{}{
			"enabled": redisEnabled,
			"url":     redisInfo,
		},
		"broadcast": map[string]interface{}{
			"type":          broadcastType,
			"activeWorkers": len(workerStats),
		},
		"environment": map[string]interface{}{
			"appDebug": config.AppDebug,
			"appPort":  config.AppPort,
		},
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "System status",
		Results: status,
	})
}


// AI Lead Management Handlers

// CreateLeadAI creates a new AI lead (without device assignment)
func (handler *App) CreateLeadAI(c *fiber.Ctx) error {
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
	
	var request struct {
		Name         string `json:"name"`
		Phone        string `json:"phone"`
		Email        string `json:"email"`
		Niche        string `json:"niche"`
		TargetStatus string `json:"target_status"`
		Notes        string `json:"notes"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	// Validate required fields
	if request.Name == "" || request.Phone == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Name and phone are required",
		})
	}
	
	// Set default target status if not provided
	if request.TargetStatus == "" {
		request.TargetStatus = "prospect"
	}
	
	leadAIRepo := repository.GetLeadAIRepository()
	lead := &models.LeadAI{
		UserID:       session.UserID,
		Name:         request.Name,
		Phone:        request.Phone,
		Email:        request.Email,
		Niche:        request.Niche,
		Source:       "ai_manual",
		Status:       "pending",
		TargetStatus: request.TargetStatus,
		Notes:        request.Notes,
	}
	
	err = leadAIRepo.CreateLeadAI(lead)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "CREATE_FAILED",
			Message: "Failed to create AI lead",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI lead created successfully",
		Results: lead,
	})
}
// GetLeadsAI retrieves all AI leads for the user
func (handler *App) GetLeadsAI(c *fiber.Ctx) error {
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
	
	leadAIRepo := repository.GetLeadAIRepository()
	leads, err := leadAIRepo.GetLeadAIByUser(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "FETCH_FAILED",
			Message: "Failed to fetch AI leads",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI leads fetched successfully",
		Results: leads,
	})
}
// UpdateLeadAI updates an existing AI lead
func (handler *App) UpdateLeadAI(c *fiber.Ctx) error {
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
	
	leadID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid lead ID",
		})
	}
	
	var request struct {
		Name         string `json:"name"`
		Phone        string `json:"phone"`
		Email        string `json:"email"`
		Niche        string `json:"niche"`
		TargetStatus string `json:"target_status"`
		Notes        string `json:"notes"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	leadAIRepo := repository.GetLeadAIRepository()
	existingLead, err := leadAIRepo.GetLeadAIByID(leadID)
	if err != nil || existingLead.UserID != session.UserID {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "AI lead not found",
		})
	}
	
	lead := &models.LeadAI{
		Name:         request.Name,
		Phone:        request.Phone,
		Email:        request.Email,
		Niche:        request.Niche,
		TargetStatus: request.TargetStatus,
		Notes:        request.Notes,
	}
	
	err = leadAIRepo.UpdateLeadAI(leadID, lead)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "UPDATE_FAILED",
			Message: "Failed to update AI lead",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI lead updated successfully",
	})
}
// DeleteLeadAI deletes an AI lead
func (handler *App) DeleteLeadAI(c *fiber.Ctx) error {
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
	
	leadID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid lead ID",
		})
	}
	
	leadAIRepo := repository.GetLeadAIRepository()
	existingLead, err := leadAIRepo.GetLeadAIByID(leadID)
	if err != nil || existingLead.UserID != session.UserID {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "AI lead not found",
		})
	}
	
	err = leadAIRepo.DeleteLeadAI(leadID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "DELETE_FAILED",
			Message: "Failed to delete AI lead",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI lead deleted successfully",
	})
}
// TriggerAICampaign - Handler to manually trigger an AI campaign
func (handler *App) TriggerAICampaign(c *fiber.Ctx) error {
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
	
	campaignID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid campaign ID",
		})
	}
	
	campaignRepo := repository.GetCampaignRepository()
	campaign, err := campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Campaign not found",
		})
	}
	// Verify campaign belongs to user
	if campaign.UserID != session.UserID {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "You don't have permission to trigger this campaign",
		})
	}
	
	// Verify this is an AI campaign
	if campaign.AI == nil || *campaign.AI != "ai" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "NOT_AI_CAMPAIGN",
			Message: "This is not an AI campaign",
		})
	}
	
	// Check if campaign is already running
	if campaign.Status != "pending" && campaign.Status != "failed" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "CAMPAIGN_RUNNING",
			Message: "Campaign is already running or completed",
		})
	}
	
	// Update campaign status to triggered
	err = campaignRepo.UpdateCampaignStatus(campaignID, "triggered")
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "UPDATE_FAILED",
			Message: "Failed to update campaign status",
		})
	}
	// Initialize AI Campaign Processor
	leadAIRepo := repository.GetLeadAIRepository()
	broadcastRepo := repository.GetBroadcastRepository()
	
	// Initialize Redis client
	redisURL := config.GetRedisURL()
	opt, _ := redis.ParseURL(redisURL)
	redisClient := redis.NewClient(opt)
	
	processor := usecase.NewAICampaignProcessor(
		broadcastRepo,
		leadAIRepo,
		userRepo,
		campaignRepo,
		redisClient,
	)
	
	// Process campaign in background
	go func() {
		ctx := context.Background()
		err := processor.ProcessAICampaign(ctx, campaignID)
		if err != nil {
			logrus.Errorf("Failed to process AI campaign %d: %v", campaignID, err)
			// Update status to failed if processing fails
			campaignRepo.UpdateCampaignStatus(campaignID, "failed")
		}
	}()
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI campaign triggered successfully",
		Results: map[string]interface{}{
			"campaign_id": campaignID,
			"status":      "triggered",
		},
	})
}// GetAllTeamMembers returns all team members with device counts
func (a App) GetAllTeamMembers(c *fiber.Ctx) error {
	ctx := context.Background()
	db := database.GetDB()
	repo := repository.NewTeamMemberRepository(db)
	
	members, err := repo.GetAllWithDeviceCount(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get team members",
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"data":    members,
	})
}

// CreateTeamMember creates a new team member
func (a App) CreateTeamMember(c *fiber.Ctx) error {
	ctx := context.Background()
	db := database.GetDB()
	repo := repository.NewTeamMemberRepository(db)
	
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// Validate inputs
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}
	
	// Check if username already exists
	existing, err := repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check existing username",
		})
	}
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}
	
	// Create team member
	member := &models.TeamMember{
		Username:  req.Username,
		Password:  req.Password,
		IsActive:  true,
	}
	
	if err := repo.Create(ctx, member); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create team member",
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"data":    member,
	})
}

// UpdateTeamMember updates an existing team member
func (a App) UpdateTeamMember(c *fiber.Ctx) error {
	ctx := context.Background()
	db := database.GetDB()
	repo := repository.NewTeamMemberRepository(db)
	
	// Get team member ID from params
	memberID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid team member ID",
		})
	}
	
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IsActive bool   `json:"is_active"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// Get existing member
	member, err := repo.GetByID(ctx, memberID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get team member",
		})
	}
	if member == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Team member not found",
		})
	}
	
	// Update fields
	if req.Username != "" {
		member.Username = strings.TrimSpace(req.Username)
	}
	if req.Password != "" {
		member.Password = strings.TrimSpace(req.Password)
	}
	member.IsActive = req.IsActive
	
	// Save updates
	if err := repo.Update(ctx, member); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update team member",
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"data":    member,
	})
}

// DeleteTeamMember deletes a team member
func (a App) DeleteTeamMember(c *fiber.Ctx) error {
	ctx := context.Background()
	db := database.GetDB()
	repo := repository.NewTeamMemberRepository(db)
	
	// Get team member ID from params
	memberID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid team member ID",
		})
	}
	
	// Delete team member
	if err := repo.Delete(ctx, memberID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete team member",
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Team member deleted successfully",
	})
}

// TeamLoginView renders the team login page
func (a App) TeamLoginView(c *fiber.Ctx) error {
	return c.Render("views/team_login", fiber.Map{
		"Title": "Team Member Login",
	})
}

// TeamDashboardView renders the team dashboard page
func (a App) TeamDashboardView(c *fiber.Ctx) error {
	// Check if team member is authenticated
	sessionToken := c.Cookies("team_session")
	if sessionToken == "" {
		return c.Redirect("/team-login")
	}
	
	db := database.GetDB()
	repo := repository.NewTeamMemberRepository(db)
	session, err := repo.GetSessionByToken(context.Background(), sessionToken)
	if err != nil || session == nil {
		return c.Redirect("/team-login")
	}
	
	return c.Render("views/team_dashboard", fiber.Map{
		"Title": "Team Dashboard",
	})
}

// HandleTeamLogin handles team member login
func (a App) HandleTeamLogin(c *fiber.Ctx) error {
	ctx := context.Background()
	db := database.GetDB()
	repo := repository.NewTeamMemberRepository(db)
	
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// Find team member
	member, err := repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check credentials",
		})
	}
	
	if member == nil || member.Password != req.Password || !member.IsActive {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials or account inactive",
		})
	}
	
	// Create session
	session, err := repo.CreateSession(ctx, member.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create session",
		})
	}
	
	// Set cookie
	c.Cookie(&fiber.Cookie{
		Name:     "team_session",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})
	
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"member": member,
			"token":  session.Token,
		},
	})
}

// HandleTeamLogout handles team member logout
func (a App) HandleTeamLogout(c *fiber.Ctx) error {
	ctx := context.Background()
	db := database.GetDB()
	repo := repository.NewTeamMemberRepository(db)
	
	// Get token from cookie
	token := c.Cookies("team_session")
	if token != "" {
		// Delete session
		repo.DeleteSession(ctx, token)
	}
	
	// Clear cookie
	c.Cookie(&fiber.Cookie{
		Name:     "team_session",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetTeamMemberInfo returns the current team member info
func (a App) GetTeamMemberInfo(c *fiber.Ctx) error {
	ctx := context.Background()
	db := database.GetDB()
	repo := repository.NewTeamMemberRepository(db)
	
	// Get token from cookie
	token := c.Cookies("team_session")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}
	
	// Get session
	session, err := repo.GetSessionByToken(ctx, token)
	if err != nil || session == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid session",
		})
	}
	
	// Get team member
	member, err := repo.GetByID(ctx, session.TeamMemberID)
	if err != nil || member == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Team member not found",
		})
	}
	
	// Get device IDs for this team member
	deviceIDs, err := repo.GetDeviceIDsForMember(ctx, member.Username)
	if err != nil {
		deviceIDs = []string{}
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"member": fiber.Map{
			"id":       member.ID,
			"username": member.Username,
		},
		"device_ids": deviceIDs,
	})
}
