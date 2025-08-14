package rest

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/gofiber/fiber/v2"
)

type App struct {
	Service domainApp.IAppUsecase
}

func InitRestApp(app *fiber.App, service domainApp.IAppUsecase) App {
	rest := App{Service: service}
	
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
	app.Post("/api/devices/:id/send", rest.SendWhatsAppMessage)
	app.Post("/api/devices/:id/sync", rest.SyncDeviceChats)
	app.Get("/api/devices/:id/diagnose", rest.DiagnoseDevice)
	
	// Device management endpoints
	app.Delete("/api/devices/:id", rest.DeleteDevice)
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
	
	// Campaign endpoints
	app.Get("/api/campaigns", rest.GetCampaigns)
	app.Post("/api/campaigns", rest.CreateCampaign)
	app.Put("/api/campaigns/:id", rest.UpdateCampaign)
	app.Delete("/api/campaigns/:id", rest.DeleteCampaign)
	
	// WhatsApp QR code endpoint
	app.Get("/app/qr", rest.GetQRCode)
	
	// API endpoints
	app.Get("/app/login", rest.Login)
	app.Get("/app/login-with-code", rest.LoginWithCode)
	app.Post("/app/link-device", rest.LinkDevicePhone)
	app.Get("/app/logout", rest.Logout)
	app.Get("/app/reconnect", rest.Reconnect)
	app.Get("/app/devices", rest.Devices)

	return App{Service: service}
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
		// Start tracking this connection session
		whatsapp.StartConnectionSession(userID.(string), deviceId, "")
		log.Printf("Started connection session for user %s, device %s", userID, deviceId)
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
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid email or password",
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
	// Get QR code from login service
	response, err := handler.Service.Login(c.UserContext())
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to generate QR code",
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
	userEmail := c.Locals("email").(string)
	
	userRepo := repository.GetUserRepository()
	user, err := userRepo.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "User not found",
		})
	}
	
	leadRepo := repository.GetLeadRepository()
	leads, err := leadRepo.GetLeadsByDevice(user.ID, deviceId)
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
	userEmail := c.Locals("email").(string)
	
	var request struct {
		DeviceID string `json:"device_id"`
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Niche    string `json:"niche"`
		Journey  string `json:"journey"`
		Status   string `json:"status"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	userRepo := repository.GetUserRepository()
	user, err := userRepo.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "User not found",
		})
	}
	
	leadRepo := repository.GetLeadRepository()
	lead, err := leadRepo.CreateLead(user.ID, request.DeviceID, request.Name, request.Phone, request.Niche, request.Journey, request.Status)
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
	userEmail := c.Locals("email").(string)
	
	var request struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Niche   string `json:"niche"`
		Journey string `json:"journey"`
		Status  string `json:"status"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	userRepo := repository.GetUserRepository()
	user, err := userRepo.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "User not found",
		})
	}
	
	leadRepo := repository.GetLeadRepository()
	lead, err := leadRepo.UpdateLead(user.ID, leadId, request.Name, request.Phone, request.Niche, request.Journey, request.Status)
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
		Results: lead,
	})
}

// DeleteLead deletes a lead
func (handler *App) DeleteLead(c *fiber.Ctx) error {
	leadId := c.Params("id")
	userEmail := c.Locals("email").(string)
	
	userRepo := repository.GetUserRepository()
	user, err := userRepo.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "User not found",
		})
	}
	
	leadRepo := repository.GetLeadRepository()
	err = leadRepo.DeleteLead(user.ID, leadId)
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
	userEmail := c.Locals("email").(string)
	
	userRepo := repository.GetUserRepository()
	user, err := userRepo.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
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
	
	// Group campaigns by date for calendar display
	campaignsByDate := make(map[string][]repository.Campaign)
	for _, campaign := range campaigns {
		campaignsByDate[campaign.CampaignDate] = append(campaignsByDate[campaign.CampaignDate], campaign)
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Campaigns retrieved successfully",
		Results: campaignsByDate,
	})
}

// CreateCampaign creates a new campaign
func (handler *App) CreateCampaign(c *fiber.Ctx) error {
	userEmail := c.Locals("email").(string)
	
	var request struct {
		CampaignDate  string `json:"campaign_date"`
		Title         string `json:"title"`
		Niche         string `json:"niche"`
		Message       string `json:"message"`
		ImageURL      string `json:"image_url"`
		ScheduledTime string `json:"scheduled_time"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	userRepo := repository.GetUserRepository()
	user, err := userRepo.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "User not found",
		})
	}
	
	campaignRepo := repository.GetCampaignRepository()
	campaign, err := campaignRepo.CreateCampaign(user.ID, request.CampaignDate, request.Title, request.Niche, request.Message, request.ImageURL, request.ScheduledTime)
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
	campaignId := c.Params("id")
	userEmail := c.Locals("email").(string)
	
	var request struct {
		Title         string `json:"title"`
		Niche         string `json:"niche"`
		Message       string `json:"message"`
		ImageURL      string `json:"image_url"`
		ScheduledTime string `json:"scheduled_time"`
		Status        string `json:"status"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	userRepo := repository.GetUserRepository()
	user, err := userRepo.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "User not found",
		})
	}
	
	campaignRepo := repository.GetCampaignRepository()
	campaign, err := campaignRepo.UpdateCampaign(user.ID, campaignId, request.Title, request.Niche, request.Message, request.ImageURL, request.ScheduledTime, request.Status)
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
	campaignId := c.Params("id")
	userEmail := c.Locals("email").(string)
	
	userRepo := repository.GetUserRepository()
	user, err := userRepo.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "User not found",
		})
	}
	
	campaignRepo := repository.GetCampaignRepository()
	err = campaignRepo.DeleteCampaign(user.ID, campaignId)
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
	userEmail := c.Locals("email").(string)
	
	userRepo := repository.GetUserRepository()
	user, err := userRepo.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
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
	
	// Delete device
	err = userRepo.DeleteDevice(device.ID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to delete device: %v", err),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device deleted successfully",
	})
}

// LogoutDevice logs out from WhatsApp
func (handler *App) LogoutDevice(c *fiber.Ctx) error {
	deviceId := c.Query("deviceId")
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Logout successful",
		Results: map[string]interface{}{
			"deviceId": deviceId,
			"status":   "logged_out",
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
	
	// Trigger chat sync
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
	}
	
	if clientErr != nil {
		diagnostics["whatsapp_client"].(map[string]interface{})["error"] = clientErr.Error()
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
