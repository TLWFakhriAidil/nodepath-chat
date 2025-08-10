package rest

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	
	// Sequence endpoints
	app.Get("/api/sequences/:id/device-report", rest.GetSequenceDeviceReport)
	app.Get("/api/sequences/:id/device/:deviceId/leads", rest.GetSequenceDeviceLeads)
	app.Get("/api/sequences/:id/device/:deviceId/step/:stepId/leads", rest.GetSequenceStepLeads)
	app.Post("/api/sequences/:id/device/:deviceId/step/:stepId/resend-failed", rest.ResendFailedSequenceStep)
	app.Post("/api/sequences/:id/step/:stepId/resend", rest.ResendSequenceStepAllDevices)
	app.Get("/api/sequences/:id/update-pending-messages", rest.UpdateSequencePendingMessages)
	app.Post("/api/sequences/:id/update-pending-messages", rest.UpdateSequencePendingMessages)
	
	// Public routes (no auth required)
	app.Get("/:device_name", rest.PublicDeviceDashboard)
	app.Get("/api/public/device/:device_name/info", rest.GetPublicDeviceInfo)
	app.Get("/api/public/device/:device_name/sequences", rest.GetPublicDeviceSequences)
	app.Get("/sequence/:id/device-report", rest.PublicSequenceDeviceReport)
	
	// Device management endpoints
	app.Put("/api/devices/:id/update-jid", rest.UpdateDeviceJID)
	
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
	includeBroadcastHistory := c.Query("include_broadcast_history", "false") == "true"
	
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
	
	// If include_broadcast_history is true, add broadcast history for each lead
	if includeBroadcastHistory {
		db := database.GetDB()
		
		for i := range leads {
			// Get broadcast history for this lead
			historyQuery := `
				SELECT 
					bm.id,
					bm.sequence_id,
					s.name as sequence_name,
					bm.sequence_stepid,
					ss.step_name,
					ss.trigger_name,
					bm.status,
					bm.error_message,
					bm.sent_at,
					bm.scheduled_at
				FROM broadcast_messages bm
				LEFT JOIN sequences s ON s.id = bm.sequence_id
				LEFT JOIN sequence_steps ss ON ss.id = bm.sequence_stepid
				WHERE bm.recipient_phone = ? 
				AND bm.device_id = ?
				AND bm.user_id = ?
				AND bm.sequence_id IS NOT NULL
				ORDER BY bm.scheduled_at DESC
				LIMIT 10
			`
			
			rows, err := db.Query(historyQuery, leads[i].Phone, deviceId, session.UserID)
			if err != nil {
				log.Printf("Error getting broadcast history: %v", err)
				continue
			}
			
			history := []map[string]interface{}{}
			for rows.Next() {
				var item struct {
					ID             string
					SequenceID     sql.NullString
					SequenceName   sql.NullString
					SequenceStepID sql.NullString
					StepName       sql.NullString
					TriggerName    sql.NullString
					Status         string
					ErrorMessage   sql.NullString
					SentAt         sql.NullTime
					ScheduledAt    sql.NullTime
				}
				
				err := rows.Scan(
					&item.ID,
					&item.SequenceID,
					&item.SequenceName,
					&item.SequenceStepID,
					&item.StepName,
					&item.TriggerName,
					&item.Status,
					&item.ErrorMessage,
					&item.SentAt,
					&item.ScheduledAt,
				)
				
				if err != nil {
					continue
				}
				
				historyItem := map[string]interface{}{
					"id":            item.ID,
					"sequence_id":   item.SequenceID.String,
					"sequence_name": item.SequenceName.String,
					"step_id":       item.SequenceStepID.String,
					"step_name":     item.StepName.String,
					"trigger_name":  item.TriggerName.String,
					"status":        item.Status,
				}
				
				if item.ErrorMessage.Valid {
					historyItem["error_message"] = item.ErrorMessage.String
				}
				
				if item.SentAt.Valid {
					historyItem["sent_at"] = item.SentAt.Time
				} else if item.ScheduledAt.Valid {
					historyItem["sent_at"] = item.ScheduledAt.Time
				}
				
				history = append(history, historyItem)
			}
			rows.Close()
			
			// Add history to lead as a field
			leadMap := map[string]interface{}{
				"id":                 leads[i].ID,
				"user_id":           leads[i].UserID,
				"device_id":         leads[i].DeviceID,
				"name":              leads[i].Name,
				"phone":             leads[i].Phone,
				"email":             leads[i].Email,
				"niche":             leads[i].Niche,
				"source":            leads[i].Source,
				"status":            leads[i].Status,
				"target_status":     leads[i].TargetStatus,
				"trigger":           leads[i].Trigger,
				"notes":             leads[i].Notes,
				"journey":           leads[i].Notes,
				"created_at":        leads[i].CreatedAt,
				"broadcast_history": history,
			}
			
			// Convert to JSON and back to update the original slice
			jsonData, _ := json.Marshal(leadMap)
			json.Unmarshal(jsonData, &leads[i])
		}
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
	
	// Also update recipient_name in broadcast_messages table
	if request.Name != "" {
		db := database.GetDB()
		updateQuery := `
			UPDATE broadcast_messages 
			SET recipient_name = ?
			WHERE recipient_phone = ? 
			AND device_id = ?
			AND user_id = ?
		`
		result, err := db.Exec(updateQuery, request.Name, request.Phone, request.DeviceID, session.UserID)
		if err != nil {
			log.Printf("Error updating broadcast_messages recipient_name: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			log.Printf("Updated recipient_name in broadcast_messages for %d rows", rowsAffected)
		}
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
		log.Printf("GetCampaignSummary: Date filter - start=%s, end=%s, found %d campaigns", startDate, endDate, len(campaigns))
	} else {
		campaigns, err = campaignRepo.GetCampaignsByUser(session.UserID)
		log.Printf("GetCampaignSummary: No date filter, found %d campaigns for user %s", len(campaigns), session.UserID)
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
	
	log.Printf("Campaign Status Breakdown - Total: %d, Pending: %d, Triggered: %d, Processing: %d, Sent: %d, Failed: %d",
		totalCampaigns, pendingCampaigns, triggeredCampaigns, processingCampaigns, sentCampaigns, failedCampaigns)
	
	// Initialize totals based on broadcast_messages data
	totalShouldSend := 0
	totalDoneSend := 0
	totalFailedSend := 0
	totalPendingSend := 0
	
	// Get statistics FROM broadcast_messages table for filtered campaigns
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = os.Getenv("DB_URI")
	}
	
	// Convert mysql:// URL to DSN format if needed
	if strings.HasPrefix(mysqlURI, "mysql://") {
		mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		parts := strings.Split(mysqlURI, "@")
		if len(parts) == 2 {
			userPass := parts[0]
			hostDb := parts[1]
			mysqlURI = userPass + "@tcp(" + strings.Replace(hostDb, "/", ")/", 1) + "?parseTime=true&charset=utf8mb4"
		}
	}
	
	db, _ := sql.Open("mysql", mysqlURI)
	if db != nil {
		defer db.Close()
		
		// Get campaign IDs for queries
		var campaignIds []int
		for _, campaign := range campaigns {
			campaignIds = append(campaignIds, campaign.ID)
		}
		
		// Count total leads that should receive messages from all campaigns
		totalLeadsCount := 0
		for _, campaign := range campaigns {
			var leadCount int
			err := db.QueryRow(`
				SELECT COUNT(DISTINCT l.phone) 
				FROM leads l
				WHERE l.user_id = ? 
				AND l.niche LIKE CONCAT('%', ?, '%')
				AND (? = 'all' OR l.target_status = ?)
			`, campaign.UserID, campaign.Niche, campaign.TargetStatus, campaign.TargetStatus).Scan(&leadCount)
			
			if err == nil {
				totalLeadsCount += leadCount
			}
		}
		totalShouldSend = totalLeadsCount
		
		// Get actual broadcast message stats
		if len(campaignIds) > 0 {
			// Build query with placeholders
			placeholders := make([]string, len(campaignIds))
			args := make([]interface{}, len(campaignIds))
			for i, id := range campaignIds {
				placeholders[i] = "?"
				args[i] = id
			}
			
			query := fmt.Sprintf(`
				SELECT 
					COUNT(CASE WHEN status = 'sent' AND (error_message IS NULL OR error_message = '') THEN 1 END) as done_send,
					COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
					COUNT(CASE WHEN status IN ('pending', 'queued') THEN 1 END) as remaining
				FROM broadcast_messages
				WHERE campaign_id IN (%s)
			`, strings.Join(placeholders, ","))
			
			err := db.QueryRow(query, args...).Scan(&totalDoneSend, &totalFailedSend, &totalPendingSend)
			if err != nil {
				log.Printf("Error getting campaign broadcast stats: %v", err)
			}
		}
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
			
			// Get lead count for "should send"
			var leadCount int
			if db != nil {
				err := db.QueryRow(`
					SELECT COUNT(DISTINCT l.phone) 
					FROM leads l
					WHERE l.user_id = ? 
					AND l.niche LIKE CONCAT('%', ?, '%')
					AND (? = 'all' OR l.target_status = ?)
				`, campaign.UserID, campaign.Niche, campaign.TargetStatus, campaign.TargetStatus).Scan(&leadCount)
				
				if err != nil {
					leadCount = 0
					log.Printf("Error counting leads for campaign %d: %v", campaign.ID, err)
				}
			}
			
			// Get broadcast message stats
			var doneSend, failedSend, pendingSend int
			if db != nil {
				err := db.QueryRow(`
					SELECT 
						COUNT(CASE WHEN status = 'sent' AND (error_message IS NULL OR error_message = '') THEN 1 END) as done_send,
						COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
						COUNT(CASE WHEN status IN ('pending', 'queued') THEN 1 END) as remaining
					FROM broadcast_messages
					WHERE campaign_id = ?
				`, campaign.ID).Scan(&doneSend, &failedSend, &pendingSend)
				
				if err != nil {
					doneSend, failedSend, pendingSend = 0, 0, 0
				}
			}
			
			// Calculate remaining based on leads that haven't been sent to
			remainingSend := leadCount - doneSend - failedSend
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
				"should_send":      leadCount,      // Actual lead count
				"done_send":        doneSend,       // FROM broadcast_messages
				"failed_send":      failedSend,     // FROM broadcast_messages
				"remaining_send":   remainingSend,  // Calculated
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
	
	// Get date filter from query parameters
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")
	showToday := c.Query("today", "")
	
	// If no date filter provided and not explicitly showing today, default to today
	if startDate == "" && endDate == "" && showToday == "" {
		showToday = "true" // Default to showing today's data
	}
	
	// If showToday is true, override date filters
	if showToday == "true" {
		today := time.Now().Format("2006-01-02")
		startDate = today
		endDate = today
		log.Printf("Sequence Summary: Filtering by today: %s", today)
	} else if startDate != "" || endDate != "" {
		log.Printf("Sequence Summary: Date range filter - start=%s, end=%s", startDate, endDate)
	}
	
	// Get sequence statistics
	sequenceRepo := repository.GetSequenceRepository()
	sequences, err := sequenceRepo.GetSequences(session.UserID)
	if err != nil {
		log.Printf("Error getting sequences for user %s: %v", session.UserID, err)
		sequences = []models.Sequence{}
	}
	
	log.Printf("Found %d sequences for user %s", len(sequences), session.UserID)
	
	// Calculate statistics
	totalSequences := len(sequences)
	activeSequences := 0
	pausedSequences := 0
	draftSequences := 0
	totalContacts := 0
	totalFlows := 0
	
	// Get total flows from database
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = os.Getenv("DB_URI")
	}
	
	// Convert mysql:// URL to DSN format if needed
	if strings.HasPrefix(mysqlURI, "mysql://") {
		mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		parts := strings.Split(mysqlURI, "@")
		if len(parts) == 2 {
			userPass := parts[0]
			hostDb := parts[1]
			mysqlURI = userPass + "@tcp(" + strings.Replace(hostDb, "/", ")/", 1) + "?parseTime=true&charset=utf8mb4"
		}
	}
	
	db, err := sql.Open("mysql", mysqlURI)
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
				INNER JOIN sequences s ON s.id = ss.sequence_id
				WHERE s.user_id = ?
			`, session.UserID).Scan(&flowCount)
			
			if err != nil {
				fmt.Printf("Error counting user sequence flows: %v\n", err)
				// UUID casting not needed for MySQL
				totalFlows = 0 // Default to 0 if query fails
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
	}
	
	// Get total contacts FROM broadcast_messages for all sequences
	if db != nil {
		var totalSequenceMessages int
		
		// Build query with optional date filter
		query := `
			SELECT COUNT(DISTINCT recipient_phone) 
			FROM broadcast_messages
			WHERE sequence_id IS NOT NULL
			AND user_id = ?`
		
		args := []interface{}{session.UserID}
		
		if startDate != "" && endDate != "" {
			query += ` AND DATE(scheduled_at) BETWEEN ? AND ?`
			args = append(args, startDate, endDate)
		} else if startDate != "" {
			query += ` AND DATE(scheduled_at) >= ?`
			args = append(args, startDate)
		} else if endDate != "" {
			query += ` AND DATE(scheduled_at) <= ?`
			args = append(args, endDate)
		}
		
		err := db.QueryRow(query, args...).Scan(&totalSequenceMessages)
		
		if err == nil {
			totalContacts = totalSequenceMessages
		}
	}
	
	// Get total message counts across all sequences
	var totalShouldSend, totalDoneSend, totalFailedSend, totalRemainingSend, totalLeadsSum int
	
	// We'll calculate these totals after processing individual sequences
	
	// Add flow counts to each sequence
	sequencesWithFlows := []map[string]interface{}{}
	if db != nil {
		// Process ALL sequences, not just first 5
		for _, sequence := range sequences {
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
			
			// Get contact statistics for this sequence FROM broadcast_messages table
			var shouldSend, doneSend, failedSend, remainingSend, totalLeads int
			
			// Build query with date filter - UPDATED to use proper DISTINCT logic
			query := `
				SELECT 
					COUNT(DISTINCT CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id)) AS total,
					COUNT(DISTINCT CASE WHEN status = 'sent' AND (error_message IS NULL OR error_message = '') 
						THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) AS done_send,
					COUNT(DISTINCT CASE WHEN status = 'failed' 
						THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) AS failed,
					COUNT(DISTINCT CASE WHEN status IN ('pending', 'queued') 
						THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) AS remaining,
					COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id)) AS total_leads
				FROM broadcast_messages
				WHERE sequence_id = ?`
			
			args := []interface{}{sequence.ID}
			
			if startDate != "" && endDate != "" {
			query += ` AND DATE(scheduled_at) BETWEEN ? AND ?`
			args = append(args, startDate, endDate)
		} else if startDate != "" {
			query += ` AND DATE(scheduled_at) >= ?`
			args = append(args, startDate)
		} else if endDate != "" {
			query += ` AND DATE(scheduled_at) <= ?`
			args = append(args, endDate)
		}
			
			err = db.QueryRow(query, args...).Scan(&shouldSend, &doneSend, &failedSend, &remainingSend, &totalLeads)
			
			if err != nil {
				log.Printf("Error getting sequence stats for %s: %v", sequence.ID, err)
				shouldSend, doneSend, failedSend, remainingSend, totalLeads = 0, 0, 0, 0, 0
			} else {
				// Ensure shouldSend is the sum of all statuses for consistency
				shouldSend = doneSend + failedSend + remainingSend
			}
			
			sequenceData["should_send"] = shouldSend
			sequenceData["done_send"] = doneSend
			sequenceData["failed_send"] = failedSend
			sequenceData["remaining_send"] = remainingSend
			sequenceData["total_leads"] = totalLeads
			
			sequencesWithFlows = append(sequencesWithFlows, sequenceData)
		}
		
		// Calculate totals from individual sequences
		for _, seq := range sequencesWithFlows {
			if done, ok := seq["done_send"].(int); ok {
				totalDoneSend += done
			}
			if failed, ok := seq["failed_send"].(int); ok {
				totalFailedSend += failed
			}
			if remaining, ok := seq["remaining_send"].(int); ok {
				totalRemainingSend += remaining
			}
			if should, ok := seq["should_send"].(int); ok {
				totalShouldSend += should
			}
			// NOTE: We don't sum total_leads anymore - will calculate unique leads below
		}
		
		// Get UNIQUE total leads across all sequences (not sum)
		if db != nil {
			var uniqueTotalLeads int
			
			// Build query with optional date filter
			uniqueLeadsQuery := `
				SELECT COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id)) 
				FROM broadcast_messages
				WHERE user_id = ?
				AND sequence_id IS NOT NULL`
			
			uniqueLeadsArgs := []interface{}{session.UserID}
			
			if startDate != "" && endDate != "" {
				uniqueLeadsQuery += ` AND DATE(scheduled_at) BETWEEN ? AND ?`
				uniqueLeadsArgs = append(uniqueLeadsArgs, startDate, endDate)
			} else if startDate != "" {
				uniqueLeadsQuery += ` AND DATE(scheduled_at) >= ?`
				uniqueLeadsArgs = append(uniqueLeadsArgs, startDate)
			} else if endDate != "" {
				uniqueLeadsQuery += ` AND DATE(scheduled_at) <= ?`
				uniqueLeadsArgs = append(uniqueLeadsArgs, endDate)
			}
			
			err := db.QueryRow(uniqueLeadsQuery, uniqueLeadsArgs...).Scan(&uniqueTotalLeads)
			if err == nil {
				totalLeadsSum = uniqueTotalLeads
				log.Printf("Unique total leads across all sequences: %d", totalLeadsSum)
			} else {
				log.Printf("Error getting unique total leads: %v", err)
				totalLeadsSum = 0
			}
		}
		
		log.Printf("Processed %d sequences with flows", len(sequencesWithFlows))
		log.Printf("Calculated totals - Should: %d, Done: %d, Failed: %d, Remaining: %d, Unique Leads: %d", 
			totalShouldSend, totalDoneSend, totalFailedSend, totalRemainingSend, totalLeadsSum)
	} else {
		log.Printf("Database connection is nil")
	}
	
	summary := map[string]interface{}{
		"sequences": map[string]interface{}{
			"total": totalSequences,
			"active": activeSequences,
			"paused": pausedSequences,
			"draft": draftSequences,
		},
		"total_flows": totalFlows,
		"total_should_send": totalShouldSend,
		"total_done_send": totalDoneSend,
		"total_failed_send": totalFailedSend,
		"total_remaining_send": totalRemainingSend,
		"total_leads": totalLeadsSum,
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
	
	// Get campaign details first
	campaignRepo := repository.GetCampaignRepository()
	campaign, err := campaignRepo.GetCampaignByID(campaignId)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Campaign not found",
		})
	}
	
	// Count total leads that match campaign criteria
	var totalLeadCount int
	db := database.GetDB()
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT l.phone) 
		FROM leads l
		WHERE l.user_id = ? 
		AND l.niche LIKE CONCAT('%', ?, '%')
		AND (? = 'all' OR l.target_status = ?)
	`, session.UserID, campaign.Niche, campaign.TargetStatus, campaign.TargetStatus).Scan(&totalLeadCount)
	
	if err != nil {
		totalLeadCount = 0
	}
	
	// Get broadcast message stats from GetCampaignBroadcastStats
	_, doneSend, failedSend, _ := campaignRepo.GetCampaignBroadcastStats(campaignId)
	remainingSend := totalLeadCount - doneSend - failedSend
	if remainingSend < 0 {
		remainingSend = 0
	}
	
	// Get user devices - use direct query
	query := `
		SELECT id, device_name, phone, status, jid, created_at, last_seen, COALESCE(platform, '') as platform
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
			&device.JID, &device.CreatedAt, &device.LastSeen, &device.Platform)
		if err != nil {
			continue
		}
		device.UserID = session.UserID
		devices = append(devices, device)
		
		// Initialize device report
		log.Printf("Device Report - Initializing device: ID=%s, Name=%s, Platform=%s, JID=%s", device.ID, device.DeviceName, device.Platform, device.JID)
		deviceMap[device.ID] = &DeviceReport{
			ID:       device.ID,
			Name:     device.DeviceName,
			Status:   device.Status,
			Platform: device.Platform,
			JID:      device.JID,
		}
	}
	
	// Get broadcast message statistics grouped by device for this campaign
	messageQuery := `
		SELECT 
			bm.device_id,
			ud.device_name,
			ud.status as device_status,
			COUNT(*) as total_messages,
			COUNT(CASE WHEN bm.status = 'sent' AND (bm.error_message IS NULL OR bm.error_message = '') THEN 1 END) as success_count,
			COUNT(CASE WHEN bm.status = 'failed' THEN 1 END) as failed_count,
			COUNT(CASE WHEN bm.status = 'pending' THEN 1 END) as pending_count
		FROM broadcast_messages bm
		LEFT JOIN user_devices ud ON ud.id = bm.device_id
		WHERE bm.campaign_id = ? 
		AND bm.user_id = ?
		GROUP BY bm.device_id, ud.device_name, ud.status
		ORDER BY total_messages DESC
	`
	
	msgRows, err := db.Query(messageQuery, campaignId, session.UserID)
	if err == nil {
		defer msgRows.Close()
		
		for msgRows.Next() {
			var deviceId, deviceName, deviceStatus string
			var totalMessages, successCount, failedCount, pendingCount int
			
			// Handle null device_name and device_status
			var deviceNameNull, deviceStatusNull sql.NullString
			
			if err := msgRows.Scan(&deviceId, &deviceNameNull, &deviceStatusNull, 
				&totalMessages, &successCount, &failedCount, &pendingCount); err != nil {
				log.Printf("Device Report - Error scanning row: %v", err)
				continue
			}
			
			// Set device name and status with defaults
			if deviceNameNull.Valid {
				deviceName = deviceNameNull.String
			} else {
				deviceName = "Unknown Device"
			}
			
			if deviceStatusNull.Valid {
				deviceStatus = deviceStatusNull.String
			} else {
				deviceStatus = "unknown"
			}
			
			log.Printf("Device Report - Device: %s (%s), Total: %d, Success: %d, Failed: %d, Pending: %d", 
				deviceName, deviceId, totalMessages, successCount, failedCount, pendingCount)
			
			// Update or create device report
			if report, exists := deviceMap[deviceId]; exists {
				report.TotalLeads = totalMessages
				report.SuccessLeads = successCount
				report.FailedLeads = failedCount
				report.PendingLeads = pendingCount
				report.ShouldSend = totalMessages // For broadcast messages, should send = total messages
			} else {
				// Device not in user's device list but has messages
				deviceMap[deviceId] = &DeviceReport{
					ID:           deviceId,
					Name:         deviceName,
					Status:       deviceStatus,
					TotalLeads:   totalMessages,
					SuccessLeads: successCount,
					FailedLeads:  failedCount,
					PendingLeads: pendingCount,
					ShouldSend:   totalMessages,
				}
			}
		}
	} else {
		log.Printf("Device Report - Error querying broadcast messages: %v", err)
	}
	
	// If no broadcast messages exist, show lead distribution instead
	if len(deviceMap) == 0 || totalLeadCount > 0 {
		log.Printf("Device Report - No broadcast messages found, showing lead distribution")
		
		// Get lead distribution by device
		leadQuery := `
			SELECT 
				l.device_id,
				ud.device_name,
				ud.status as device_status,
				COUNT(*) as lead_count
			FROM leads l
			LEFT JOIN user_devices ud ON ud.id = l.device_id
			WHERE l.user_id = ? 
			AND l.niche LIKE CONCAT('%', ?, '%')
			AND (? = 'all' OR l.target_status = ?)
			GROUP BY l.device_id, ud.device_name, ud.status
			ORDER BY lead_count DESC
		`
		
		leadRows, err := db.Query(leadQuery, session.UserID, campaign.Niche, campaign.TargetStatus, campaign.TargetStatus)
		if err == nil {
			defer leadRows.Close()
			
			for leadRows.Next() {
				var deviceId string
				var deviceNameNull, deviceStatusNull sql.NullString
				var leadCount int
				
				if err := leadRows.Scan(&deviceId, &deviceNameNull, &deviceStatusNull, &leadCount); err != nil {
					continue
				}
				
				deviceName := "Unknown Device"
				deviceStatus := "unknown"
				
				if deviceNameNull.Valid {
					deviceName = deviceNameNull.String
				}
				if deviceStatusNull.Valid {
					deviceStatus = deviceStatusNull.String
				}
				
				// Only update devices that don't have broadcast data
				if report, exists := deviceMap[deviceId]; exists && report.TotalLeads == 0 {
					report.ShouldSend = leadCount
					report.PendingLeads = leadCount // All are pending since not sent
				} else if !exists {
					deviceMap[deviceId] = &DeviceReport{
						ID:           deviceId,
						Name:         deviceName,
						Status:       deviceStatus,
						ShouldSend:   leadCount,
						PendingLeads: leadCount,
						TotalLeads:   0,
						SuccessLeads: 0,
						FailedLeads:  0,
					}
				}
			}
		}
	}
	
	// Convert map to slice and calculate totals
	deviceReports := make([]DeviceReport, 0, len(deviceMap))
	totalMessages := 0
	pendingMessages := 0
	successMessages := 0
	failedMessages := 0
	totalDevicesWithData := 0
	onlineDevicesWithData := 0
	offlineDevicesWithData := 0
	
	for _, report := range deviceMap {
		// Only include devices that have data (either messages or leads)
		if report.TotalLeads == 0 && report.ShouldSend == 0 {
			continue
		}
		
		// For display purposes
		report.DoneSend = report.SuccessLeads
		report.FailedSend = report.FailedLeads
		report.RemainingSend = report.PendingLeads
		
		deviceReports = append(deviceReports, *report)
		
		// Add to totals
		totalMessages += report.TotalLeads
		pendingMessages += report.PendingLeads
		successMessages += report.SuccessLeads
		failedMessages += report.FailedLeads
		
		// Count devices
		totalDevicesWithData++
		if report.Status == "online" {
			onlineDevicesWithData++
		} else {
			offlineDevicesWithData++
		}
		
		log.Printf("Device %s (%s): Status=%s, Should=%d, Total=%d, Success=%d, Failed=%d, Pending=%d", 
			report.Name, report.ID, report.Status, report.ShouldSend, report.TotalLeads, 
			report.SuccessLeads, report.FailedLeads, report.PendingLeads)
	}
	
	log.Printf("Device Report Summary: Total=%d, Online=%d, Offline=%d", 
		totalDevicesWithData, onlineDevicesWithData, offlineDevicesWithData)
	
	result := map[string]interface{}{
		"totalDevices":        totalDevicesWithData,
		"activeDevices":       onlineDevicesWithData,
		"disconnectedDevices": offlineDevicesWithData,
		"totalLeads":          totalLeadCount,      // From campaign criteria
		"shouldSend":          totalLeadCount,      // Same as totalLeads
		"doneSend":            successMessages,     // FROM broadcast_messages
		"failedSend":          failedMessages,      // FROM broadcast_messages
		"remainingSend":       remainingSend,       // Calculated earlier
		"pendingLeads":        pendingMessages,     // FROM broadcast_messages
		"successLeads":        successMessages,     // FROM broadcast_messages
		"failedLeads":         failedMessages,      // FROM broadcast_messages
		"devices":             deviceReports,
		"campaign": map[string]interface{}{
			"id":            campaign.ID,
			"title":         campaign.Title,
			"niche":         campaign.Niche,
			"target_status": campaign.TargetStatus,
			"status":        campaign.Status,
		},
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
	
	// Get campaign details to know the criteria
	campaignRepo := repository.GetCampaignRepository()
	campaign, err := campaignRepo.GetCampaignByID(campaignId)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Campaign not found",
		})
	}
	
	var leads []map[string]interface{}
	
	// If status is "pending" or "all" and no broadcast messages exist, show leads that match campaign criteria
	if status == "pending" || status == "all" {
		// First check if there are any broadcast messages for this campaign
		var messageCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM broadcast_messages 
			WHERE campaign_id = ? AND device_id = ?
		`, campaignId, deviceId).Scan(&messageCount)
		
		if err == nil && messageCount == 0 {
			// No messages sent yet, show all matching leads for this device
			query := `
				SELECT l.phone, l.name, 'pending' as status, NULL as sent_at
				FROM leads l
				WHERE l.device_id = ? 
				AND l.user_id = ?
				AND l.niche LIKE CONCAT('%', ?, '%')
				AND (? = 'all' OR l.target_status = ?)
			`
			
			rows, err := db.Query(query, deviceId, session.UserID, campaign.Niche, campaign.TargetStatus, campaign.TargetStatus)
			if err == nil {
				defer rows.Close()
				
				for rows.Next() {
					var phone, name, status string
					var sentAt sql.NullTime
					
					if err := rows.Scan(&phone, &name, &status, &sentAt); err == nil {
						lead := map[string]interface{}{
							"phone":  phone,
							"name":   name,
							"status": status,
						}
						if sentAt.Valid {
							lead["sent_at"] = sentAt.Time
						}
						leads = append(leads, lead)
					}
				}
			}
			
			return c.JSON(utils.ResponseData{
				Status:  200,
				Code:    "SUCCESS",
				Message: "Lead details",
				Results: leads,
			})
		}
	}
	
	// Otherwise, get leads from broadcast messages
	var query string
	if aiType.Valid && aiType.String == "ai" {
		// For AI campaigns - use MySQL-compatible query
		query = `
			SELECT bm.recipient_phone, bm.status, bm.sent_at, lai.name
			FROM (
				SELECT recipient_phone, status, sent_at, created_at
				FROM broadcast_messages
				WHERE campaign_id = ? AND device_id = ? AND user_id = ?
				ORDER BY created_at DESC
			) bm
			LEFT JOIN leads_ai lai ON lai.phone = bm.recipient_phone AND lai.user_id = ?
			GROUP BY bm.recipient_phone, lai.name
		`
	} else {
		// For regular campaigns - use MySQL-compatible query
		query = `
			SELECT bm.recipient_phone, bm.status, bm.sent_at, l.name
			FROM (
				SELECT recipient_phone, status, sent_at, created_at
				FROM broadcast_messages
				WHERE campaign_id = ? AND device_id = ? AND user_id = ?
				ORDER BY created_at DESC
			) bm
			LEFT JOIN leads l ON l.phone = bm.recipient_phone AND l.user_id = ?
			GROUP BY bm.recipient_phone, l.name
		`
	}
	
	// Add status filter if not "all"
	if status != "all" {
		if status == "success" {
			query += ` HAVING bm.status IN ('sent', 'delivered', 'success')`
		} else if status == "pending" {
			query += ` HAVING bm.status IN ('pending', 'queued')`
		} else if status == "failed" {
			query += ` HAVING bm.status IN ('failed', 'error')`
		}
	}
	
	query += ` ORDER BY bm.sent_at DESC`
	
	// Execute query based on campaign type
	var rows *sql.Rows
	if aiType.Valid && aiType.String == "ai" {
		rows, err = db.Query(query, campaignId, deviceId, session.UserID, session.UserID)
	} else {
		rows, err = db.Query(query, campaignId, deviceId, session.UserID, session.UserID)
	}
	if err != nil {
		log.Printf("Error executing lead details query: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get lead details",
		})
	}
	defer rows.Close()
	
	if leads == nil {
		leads = []map[string]interface{}{}
	}
	
	for rows.Next() {
		var phone, msgStatus string
		var sentAt sql.NullTime
		var name sql.NullString
		
		err := rows.Scan(&phone, &msgStatus, &sentAt, &name)
		if err != nil {
			log.Printf("Error scanning lead row: %v", err)
			continue
		}
		
		leadName := "Unknown"
		if name.Valid && name.String != "" {
			leadName = name.String
		}
		
		lead := map[string]interface{}{
			"name":   leadName,
			"phone":  phone,
			"status": msgStatus,
		}
		
		if sentAt.Valid {
			lead["sent_at"] = sentAt.Time.Format("2006-01-02 03:04 PM")
		} else {
			lead["sent_at"] = "-"
		}
		
		leads = append(leads, lead)
	}
	
	log.Printf("GetCampaignDeviceLeads - Found %d leads for campaign %d, device %s, status %s", 
		len(leads), campaignId, deviceId, status)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Lead details retrieved successfully",
		Results: leads,
	})
}

// DeviceReport structure for device-wise report
type DeviceReport struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Platform     string `json:"platform"`
	JID          string `json:"jid"`
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


// GetSequenceDeviceReport gets device-wise report for a sequence broken down by steps
func (handler *App) GetSequenceDeviceReport(c *fiber.Ctx) error {
	sequenceId := c.Params("id") // Sequence ID is already a string UUID
	log.Printf("GetSequenceDeviceReport called for sequence: %s", sequenceId)
	
	// Get date filters from query parameters
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")
	
	log.Printf("Date filters - start: %s, end: %s", startDate, endDate)
	
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
		log.Printf("Invalid session: %v", err)
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED", 
			Message: "Invalid session",
		})
	}
	
	// Get sequence details
	sequenceRepo := repository.GetSequenceRepository()
	sequence, err := sequenceRepo.GetSequenceByID(sequenceId) // Already a string
	if err != nil {
		log.Printf("Error getting sequence %s: %v", sequenceId, err)
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Sequence not found",
		})
	}
	
	log.Printf("Found sequence: %s - %s", sequence.ID, sequence.Name)
	
	// Get MySQL connection directly like in sequence summary
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = os.Getenv("DB_URI")
	}
	
	// Convert mysql:// URL to DSN format if needed
	if strings.HasPrefix(mysqlURI, "mysql://") {
		mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		parts := strings.Split(mysqlURI, "@")
		if len(parts) == 2 {
			userPass := parts[0]
			hostDb := parts[1]
			mysqlURI = userPass + "@tcp(" + strings.Replace(hostDb, "/", ")/", 1) + "?parseTime=true&charset=utf8mb4"
		}
	}
	
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Printf("Error opening MySQL connection: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Database connection error",
		})
	}
	defer db.Close()
	
	// Get sequence steps first
	stepsQuery := `
		SELECT id, COALESCE(day_number, day, 1) as step_order, message_type, content, COALESCE(day_number, day, 1) as day_num
		FROM sequence_steps
		WHERE sequence_id = ?
		ORDER BY COALESCE(day_number, day, 1)
	`
	
	// Use string sequence ID for query
	log.Printf("Getting steps for sequence ID: %s", sequenceId)
	
	stepRows, err := db.Query(stepsQuery, sequenceId)
	if err != nil {
		log.Printf("Error getting sequence steps: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get sequence steps",
		})
	}
	defer stepRows.Close()
	
	// Store steps info
	type StepInfo struct {
		ID          string
		Order       int
		MessageType string
		Content     string
		DayNumber   int
	}
	
	steps := []StepInfo{}
	stepMap := make(map[string]StepInfo)
	
	for stepRows.Next() {
		var step StepInfo
		var contentNull sql.NullString
		var dayNumberNull sql.NullInt64
		
		err := stepRows.Scan(&step.ID, &step.Order, &step.MessageType, &contentNull, &dayNumberNull)
		if err != nil {
			log.Printf("Error scanning step row: %v", err)
			continue
		}
		
		if contentNull.Valid {
			step.Content = contentNull.String
			if len(step.Content) > 50 {
				step.Content = step.Content[:50] + "..."
			}
		}
		
		if dayNumberNull.Valid {
			step.DayNumber = int(dayNumberNull.Int64)
		}
		
		steps = append(steps, step)
		stepMap[step.ID] = step
	}
	
	// Get devices that have messages for this sequence
	deviceQuery := `
		SELECT DISTINCT 
			bm.device_id,
			COALESCE(ud.device_name, 'Unknown Device') as device_name,
			COALESCE(ud.status, 'unknown') as device_status,
			COALESCE(ud.platform, '') as platform,
			COALESCE(ud.jid, '') as jid
		FROM broadcast_messages bm
		LEFT JOIN user_devices ud ON ud.id = bm.device_id
		WHERE bm.sequence_id = ? 
		AND bm.user_id = ?`
	
	args := []interface{}{sequenceId, session.UserID}
	
	// Add date filters if provided
	if startDate != "" && endDate != "" {
		deviceQuery += ` AND DATE(bm.scheduled_at) BETWEEN ? AND ?`
		args = append(args, startDate, endDate)
	} else if startDate != "" {
		deviceQuery += ` AND DATE(bm.scheduled_at) >= ?`
		args = append(args, startDate)
	} else if endDate != "" {
		deviceQuery += ` AND DATE(bm.scheduled_at) <= ?`
		args = append(args, endDate)
	}
	
	deviceRows, err := db.Query(deviceQuery, args...)
	if err != nil {
		log.Printf("Error getting devices for sequence %s: %v", sequenceId, err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get devices",
		})
	}
	defer deviceRows.Close()
	
	// Device report structure with steps
	type StepReport struct {
		StepID        string `json:"step_id"`
		StepOrder     int    `json:"step_order"`
		StepName      string `json:"step_name"`
		DayNumber     int    `json:"day_number"`
		ShouldSend    int    `json:"should_send"`
		DoneSend      int    `json:"done_send"`
		FailedSend    int    `json:"failed_send"`
		RemainingSend int    `json:"remaining_send"`
		TotalLeads    int    `json:"total_leads"`
	}
	
	type DeviceStepReport struct {
		ID            string       `json:"id"`
		Name          string       `json:"name"`
		Status        string       `json:"status"`
		Platform      string       `json:"platform"`
		JID           string       `json:"jid"`
		TotalMessages int          `json:"total_messages"`
		Steps         []StepReport `json:"steps"`
	}
	
	deviceReports := []DeviceStepReport{}
	totalDevicesWithData := 0
	onlineDevicesWithData := 0
	offlineDevicesWithData := 0
	
	// Process each device
	for deviceRows.Next() {
		var deviceId, deviceName, deviceStatus, platform, jid string
		err := deviceRows.Scan(&deviceId, &deviceName, &deviceStatus, &platform, &jid)
		if err != nil {
			continue
		}
		
		deviceReport := DeviceStepReport{
			ID:       deviceId,
			Name:     deviceName,
			Status:   deviceStatus,
			Platform: platform,
			JID:      jid,
			Steps:    []StepReport{},
		}
		
		// Get stats for each step for this device - UPDATED to use new counting logic
		stepStatsQuery := `
			SELECT 
				bm.sequence_stepid,
				COUNT(DISTINCT CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id)) as total,
				COUNT(DISTINCT CASE WHEN bm.status = 'sent' AND (bm.error_message IS NULL OR bm.error_message = '') 
					THEN CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id) END) as done_send,
				COUNT(DISTINCT CASE WHEN bm.status = 'failed' 
					THEN CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id) END) as failed_send,
				COUNT(DISTINCT CASE WHEN bm.status IN ('pending', 'queued') 
					THEN CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id) END) as remaining_send,
				COUNT(DISTINCT CONCAT(bm.recipient_phone, '|', bm.device_id)) as total_leads
			FROM broadcast_messages bm
			WHERE bm.sequence_id = ? 
			AND bm.device_id = ?
			AND bm.user_id = ?
			AND bm.sequence_stepid IS NOT NULL`
		
		statsArgs := []interface{}{sequenceId, deviceId, session.UserID}
		
		// Add date filters if provided
		if startDate != "" && endDate != "" {
			stepStatsQuery += ` AND DATE(bm.scheduled_at) BETWEEN ? AND ?`
			statsArgs = append(statsArgs, startDate, endDate)
		} else if startDate != "" {
			stepStatsQuery += ` AND DATE(bm.scheduled_at) >= ?`
			statsArgs = append(statsArgs, startDate)
		} else if endDate != "" {
			stepStatsQuery += ` AND DATE(bm.scheduled_at) <= ?`
			statsArgs = append(statsArgs, endDate)
		}
		
		stepStatsQuery += ` GROUP BY bm.sequence_stepid`
		
		statsRows, err := db.Query(stepStatsQuery, statsArgs...)
		if err != nil {
			log.Printf("Error getting step stats for device %s: %v", deviceId, err)
			continue
		}
		
		totalDeviceMessages := 0
		
		for statsRows.Next() {
			var stepId string
			var total, doneSend, failedSend, remainingSend, totalLeads int
			
			err := statsRows.Scan(&stepId, &total, &doneSend, &failedSend, &remainingSend, &totalLeads)
			if err != nil {
				continue
			}
			
			// Get step info
			stepInfo, exists := stepMap[stepId]
			if !exists {
				continue
			}
			
			// Calculate should_send using the same logic as summary page
			shouldSend := doneSend + failedSend + remainingSend
			
			stepReport := StepReport{
				StepID:        stepId,
				StepOrder:     stepInfo.Order,
				StepName:      fmt.Sprintf("Step %d: %s", stepInfo.Order, stepInfo.MessageType),
				DayNumber:     stepInfo.DayNumber,
				ShouldSend:    shouldSend,
				DoneSend:      doneSend,
				FailedSend:    failedSend,
				RemainingSend: remainingSend,
				TotalLeads:    totalLeads,
			}
			
			deviceReport.Steps = append(deviceReport.Steps, stepReport)
			totalDeviceMessages += shouldSend
		}
		statsRows.Close()
		
		// Only include devices with messages
		if totalDeviceMessages > 0 {
			deviceReport.TotalMessages = totalDeviceMessages
			deviceReports = append(deviceReports, deviceReport)
			
			totalDevicesWithData++
			if deviceStatus == "online" || deviceStatus == "connected" {
				onlineDevicesWithData++
			} else {
				offlineDevicesWithData++
			}
		}
	}
	
	// Calculate overall totals using the same logic as summary page - UPDATED
	overallQuery := `
		SELECT 
			COUNT(DISTINCT CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id)) as total,
			COUNT(DISTINCT CASE WHEN status = 'sent' AND (error_message IS NULL OR error_message = '') 
				THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) as done_send,
			COUNT(DISTINCT CASE WHEN status = 'failed' 
				THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) as failed_send,
			COUNT(DISTINCT CASE WHEN status IN ('pending', 'queued') 
				THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) as remaining_send,
			COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id)) as total_leads
		FROM broadcast_messages
		WHERE sequence_id = ?`
	
	overallArgs := []interface{}{sequenceId}
	
	// Add date filters if provided
	if startDate != "" && endDate != "" {
		overallQuery += ` AND DATE(scheduled_at) BETWEEN ? AND ?`
		overallArgs = append(overallArgs, startDate, endDate)
	} else if startDate != "" {
		overallQuery += ` AND DATE(scheduled_at) >= ?`
		overallArgs = append(overallArgs, startDate)
	} else if endDate != "" {
		overallQuery += ` AND DATE(scheduled_at) <= ?`
		overallArgs = append(overallArgs, endDate)
	}
	
	var totalMessages, totalDoneSend, totalFailedSend, totalRemainingSend, totalLeadsCount int
	err = db.QueryRow(overallQuery, overallArgs...).Scan(
		&totalMessages, &totalDoneSend, &totalFailedSend, &totalRemainingSend, &totalLeadsCount)
	
	if err != nil {
		totalMessages, totalDoneSend, totalFailedSend, totalRemainingSend, totalLeadsCount = 0, 0, 0, 0, 0
	}
	
	// Calculate total using the same logic as summary page
	totalShouldSend := totalDoneSend + totalFailedSend + totalRemainingSend
	
	log.Printf("Sequence Device Report - Total devices: %d, Online: %d, Offline: %d", 
		totalDevicesWithData, onlineDevicesWithData, offlineDevicesWithData)
	log.Printf("Overall stats - Should: %d, Done: %d, Failed: %d, Remaining: %d, TotalLeads: %d",
		totalShouldSend, totalDoneSend, totalFailedSend, totalRemainingSend, totalLeadsCount)
	
	// Special logging for COLD Sequence
	if sequence.Name == "COLD Sequence" {
		log.Printf("COLD Sequence Device Report - Date filter: start=%s, end=%s", startDate, endDate)
		log.Printf("COLD Sequence Device Report - Final totals: Should=%d (Done=%d + Failed=%d + Remaining=%d)", 
			totalShouldSend, totalDoneSend, totalFailedSend, totalRemainingSend)
	}
	
	// Calculate step-wise totals across ALL devices
	stepTotalsQuery := `
		SELECT 
			sequence_stepid,
			COUNT(DISTINCT CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id)) as total_messages,
			COUNT(DISTINCT CASE WHEN status = 'sent' AND (error_message IS NULL OR error_message = '') 
				THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) as done_send,
			COUNT(DISTINCT CASE WHEN status = 'failed' 
				THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) as failed_send,
			COUNT(DISTINCT CASE WHEN status IN ('pending', 'queued') 
				THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) as remaining_send,
			COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id)) as total_leads
		FROM broadcast_messages
		WHERE sequence_id = ? 
		AND user_id = ?
		AND sequence_stepid IS NOT NULL`
	
	stepTotalsArgs := []interface{}{sequenceId, session.UserID}
	
	// Add date filters if provided
	if startDate != "" && endDate != "" {
		stepTotalsQuery += ` AND DATE(scheduled_at) BETWEEN ? AND ?`
		stepTotalsArgs = append(stepTotalsArgs, startDate, endDate)
	} else if startDate != "" {
		stepTotalsQuery += ` AND DATE(scheduled_at) >= ?`
		stepTotalsArgs = append(stepTotalsArgs, startDate)
	} else if endDate != "" {
		stepTotalsQuery += ` AND DATE(scheduled_at) <= ?`
		stepTotalsArgs = append(stepTotalsArgs, endDate)
	}
	
	stepTotalsQuery += ` GROUP BY sequence_stepid`
	
	stepTotals := make(map[string]map[string]interface{})
	
	stepTotalsRows, err := db.Query(stepTotalsQuery, stepTotalsArgs...)
	if err == nil {
		defer stepTotalsRows.Close()
		
		for stepTotalsRows.Next() {
			var stepId string
			var total, doneSend, failedSend, remainingSend, totalLeads int
			
			err := stepTotalsRows.Scan(&stepId, &total, &doneSend, &failedSend, &remainingSend, &totalLeads)
			if err == nil {
				shouldSend := doneSend + failedSend + remainingSend
				stepTotals[stepId] = map[string]interface{}{
					"should_send":    shouldSend,
					"done_send":      doneSend,
					"failed_send":    failedSend,
					"remaining_send": remainingSend,
					"total_leads":    totalLeads,
				}
			}
		}
	}
	
	result := map[string]interface{}{
		"totalDevices":        totalDevicesWithData,
		"activeDevices":       onlineDevicesWithData,
		"disconnectedDevices": offlineDevicesWithData,
		"totalLeads":          totalLeadsCount,
		"shouldSend":          totalShouldSend,
		"doneSend":            totalDoneSend,
		"failedSend":          totalFailedSend,
		"remainingSend":       totalRemainingSend,
		"devices":             deviceReports,
		"steps":               steps,
		"stepTotals":          stepTotals,  // NEW: Add step-wise totals
		"sequence": map[string]interface{}{
			"id":      sequence.ID,
			"name":    sequence.Name,
			"niche":   sequence.Niche,
			"trigger": sequence.Trigger,
			"status":  sequence.Status,
		},
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequence device report",
		Results: result,
	})
}

// GetSequenceDeviceLeads gets lead details for a specific device in a sequence
func (handler *App) GetSequenceDeviceLeads(c *fiber.Ctx) error {
	sequenceId := c.Params("id") // Already a string UUID
	
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
	
	// Get leads from broadcast messages for this sequence
	db := database.GetDB()
	
	query := `
		SELECT bm.recipient_phone, bm.status, bm.sent_at, l.name
		FROM broadcast_messages bm
		LEFT JOIN leads l ON l.phone = bm.recipient_phone AND l.user_id = bm.user_id
		WHERE bm.sequence_id = ? AND bm.device_id = ? AND bm.user_id = ?
	`
	
	// Add status filter if not "all"
	if status != "all" {
		if status == "success" {
			query += ` AND bm.status IN ('sent', 'delivered', 'success')`
		} else if status == "pending" {
			query += ` AND bm.status IN ('pending', 'queued')`
		} else if status == "failed" {
			query += ` AND bm.status IN ('failed', 'error')`
		}
	}
	
	query += ` ORDER BY bm.sent_at DESC`
	
	rows, err := db.Query(query, sequenceId, deviceId, session.UserID)
	if err != nil {
		log.Printf("Error executing sequence lead details query: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get lead details",
		})
	}
	defer rows.Close()
	
	leads := []map[string]interface{}{}
	
	for rows.Next() {
		var phone, msgStatus string
		var sentAt sql.NullTime
		var name sql.NullString
		
		err := rows.Scan(&phone, &msgStatus, &sentAt, &name)
		if err != nil {
			log.Printf("Error scanning lead row: %v", err)
			continue
		}
		
		leadName := "Unknown"
		if name.Valid && name.String != "" {
			leadName = name.String
		}
		
		lead := map[string]interface{}{
			"name":   leadName,
			"phone":  phone,
			"status": msgStatus,
		}
		
		if sentAt.Valid {
			lead["sent_at"] = sentAt.Time.Format("2006-01-02 03:04 PM")
		} else {
			lead["sent_at"] = "-"
		}
		
		leads = append(leads, lead)
	}
	
	log.Printf("GetSequenceDeviceLeads - Found %d leads for sequence %d, device %s, status %s", 
		len(leads), sequenceId, deviceId, status)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Lead details retrieved successfully",
		Results: leads,
	})
}


// GetSequenceStepLeads gets lead details for a specific step in a sequence on a device
func (handler *App) GetSequenceStepLeads(c *fiber.Ctx) error {
	sequenceId := c.Params("id") // Already a string UUID
	
	deviceId := c.Params("deviceId")
	stepId := c.Params("stepId")
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
	
	// Get leads from broadcast messages for this specific step
	db := database.GetDB()
	
	query := `
		SELECT bm.recipient_phone, bm.status, bm.sent_at, l.name, bm.error_message
		FROM broadcast_messages bm
		LEFT JOIN leads l ON l.phone = bm.recipient_phone AND l.user_id = bm.user_id
		WHERE bm.sequence_id = ? 
		AND bm.device_id = ? 
		AND bm.sequence_stepid = ?
		AND bm.user_id = ?
	`
	
	// Use sequence ID directly as string
	args := []interface{}{sequenceId, deviceId, stepId, session.UserID}
	
	// Add status filter if not "all"
	if status != "all" {
		if status == "success" {
			query += ` AND bm.status IN ('sent', 'delivered', 'success')`
		} else if status == "pending" {
			query += ` AND bm.status IN ('pending', 'queued')`
		} else if status == "failed" {
			query += ` AND bm.status IN ('failed', 'error')`
		}
	}
	
	query += ` ORDER BY bm.sent_at DESC`
	
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error executing sequence step lead details query: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get lead details",
		})
	}
	defer rows.Close()
	
	leads := []map[string]interface{}{}
	
	for rows.Next() {
		var phone, msgStatus string
		var sentAt sql.NullTime
		var name sql.NullString
		var errorMessage sql.NullString
		
		err := rows.Scan(&phone, &msgStatus, &sentAt, &name, &errorMessage)
		if err != nil {
			log.Printf("Error scanning lead row: %v", err)
			continue
		}
		
		leadName := "Unknown"
		if name.Valid && name.String != "" {
			leadName = name.String
		}
		
		lead := map[string]interface{}{
			"name":   leadName,
			"phone":  phone,
			"status": msgStatus,
		}
		
		// Add error message if exists
		if errorMessage.Valid && errorMessage.String != "" {
			lead["error_message"] = errorMessage.String
		} else {
			lead["error_message"] = "-"
		}
		
		if sentAt.Valid {
			lead["sent_at"] = sentAt.Time.Format("2006-01-02 03:04 PM")
		} else {
			lead["sent_at"] = "-"
		}
		
		leads = append(leads, lead)
	}
	
	log.Printf("GetSequenceStepLeads - Found %d leads for sequence %d, device %s, step %s, status %s", 
		len(leads), sequenceId, deviceId, stepId, status)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Lead details retrieved successfully",
		Results: leads,
	})
}


// ResendFailedSequenceStep resets failed messages to pending for a specific sequence step
func (handler *App) ResendFailedSequenceStep(c *fiber.Ctx) error {
	sequenceId := c.Params("id")
	deviceId := c.Params("deviceId")
	stepId := c.Params("stepId")
	
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
	
	// Update failed messages to pending and clear error_message
	db := database.GetDB()
	query := `
		UPDATE broadcast_messages
		SET status = 'pending', error_message = NULL
		WHERE sequence_id = ? 
		AND device_id = ? 
		AND sequence_stepid = ?
		AND user_id = ?
		AND status = 'failed'
	`
	
	result, err := db.Exec(query, sequenceId, deviceId, stepId, session.UserID)
	if err != nil {
		log.Printf("Error resending failed messages: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to resend messages",
		})
	}
	
	rowsAffected, _ := result.RowsAffected()
	
	log.Printf("Resend failed messages - Sequence: %s, Device: %s, Step: %s, Updated: %d", 
		sequenceId, deviceId, stepId, rowsAffected)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Failed messages reset to pending",
		Results: map[string]interface{}{
			"updated": rowsAffected,
		},
	})
}

// UpdateDeviceJID updates the JID for a device
func (handler *App) UpdateDeviceJID(c *fiber.Ctx) error {
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
	
	// Parse request body
	var request struct {
		JID string `json:"jid"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	// Update device JID
	db := database.GetDB()
	query := `
		UPDATE user_devices
		SET jid = ?
		WHERE id = ? AND user_id = ?
	`
	
	result, err := db.Exec(query, request.JID, deviceId, session.UserID)
	if err != nil {
		log.Printf("Error updating device JID: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update JID",
		})
	}
	
	rowsAffected, _ := result.RowsAffected()
	
	if rowsAffected == 0 {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	log.Printf("Updated device JID - Device: %s, New JID: %s", deviceId, request.JID)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device JID updated successfully",
	})
}


// ResendSequenceStepAllDevices resends failed/remaining messages for a sequence step across all devices
func (handler *App) ResendSequenceStepAllDevices(c *fiber.Ctx) error {
	sequenceId := c.Params("id")
	stepId := c.Params("stepId")
	
	// Get request body
	var request struct {
		ResendFailed    bool `json:"resend_failed"`
		ResendRemaining bool `json:"resend_remaining"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
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
	
	// Get MySQL connection
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = os.Getenv("DB_URI")
	}
	
	// Convert mysql:// URL to DSN format if needed
	if strings.HasPrefix(mysqlURI, "mysql://") {
		mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		parts := strings.Split(mysqlURI, "@")
		if len(parts) == 2 {
			userPass := parts[0]
			hostDb := parts[1]
			mysqlURI = userPass + "@tcp(" + strings.Replace(hostDb, "/", ")/", 1) + "?parseTime=true&charset=utf8mb4"
		}
	}
	
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Printf("Error opening MySQL connection: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Database connection error",
		})
	}
	defer db.Close()
	
	// Build status filter
	statusFilter := []string{}
	if request.ResendFailed {
		statusFilter = append(statusFilter, "'failed'")
	}
	if request.ResendRemaining {
		statusFilter = append(statusFilter, "'pending'", "'queued'")
	}
	
	if len(statusFilter) == 0 {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "No status selected for resend",
		})
	}
	
	statusCondition := fmt.Sprintf("status IN (%s)", strings.Join(statusFilter, ", "))
	
	// Update messages to pending for resend
	updateQuery := fmt.Sprintf(`
		UPDATE broadcast_messages 
		SET status = 'pending',
		    error_message = NULL,
		    updated_at = NOW()
		WHERE sequence_id = ?
		AND sequence_stepid = ?
		AND user_id = ?
		AND %s
	`, statusCondition)
	
	result, err := db.Exec(updateQuery, sequenceId, stepId, session.UserID)
	if err != nil {
		log.Printf("Error updating messages for resend: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to queue messages for resend",
		})
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("Queued %d messages for resend - Sequence: %s, Step: %s", rowsAffected, sequenceId, stepId)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Successfully queued %d messages for resend", rowsAffected),
		Results: map[string]interface{}{
			"queued_count": rowsAffected,
		},
	})
}


// PublicDeviceDashboard renders the public device dashboard page
func (handler *App) PublicDeviceDashboard(c *fiber.Ctx) error {
	deviceName := c.Params("device_name")
	
	// Check if this is actually a device name (not other routes)
	// Skip if it's a known route like "login", "api", etc.
	knownRoutes := []string{"login", "api", "app", "device", "sequence", "status", "media"}
	for _, route := range knownRoutes {
		if deviceName == route {
			return c.Next()
		}
	}
	
	// Render the public device dashboard template
	return c.Render("public_device_dashboard", fiber.Map{
		"DeviceName": deviceName,
	})
}

// GetPublicDeviceInfo returns public device information
func (handler *App) GetPublicDeviceInfo(c *fiber.Ctx) error {
	deviceName := c.Params("device_name")
	
	// Get MySQL connection
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = os.Getenv("DB_URI")
	}
	
	// Convert mysql:// URL to DSN format if needed
	if strings.HasPrefix(mysqlURI, "mysql://") {
		mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		parts := strings.Split(mysqlURI, "@")
		if len(parts) == 2 {
			userPass := parts[0]
			hostDb := parts[1]
			mysqlURI = userPass + "@tcp(" + strings.Replace(hostDb, "/", ")/", 1) + "?parseTime=true&charset=utf8mb4"
		}
	}
	
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Database connection error",
		})
	}
	defer db.Close()
	
	// Get device info
	var device struct {
		ID          string         `json:"id"`
		DeviceName  string         `json:"device_name"`
		JID         sql.NullString `json:"-"`
		Platform    sql.NullString `json:"-"`
		Status      string         `json:"status"`
		CreatedAt   time.Time      `json:"created_at"`
	}
	
	err = db.QueryRow(`
		SELECT id, device_name, jid, platform, status, created_at
		FROM user_devices
		WHERE device_name = ?
	`, deviceName).Scan(&device.ID, &device.DeviceName, &device.JID, &device.Platform, &device.Status, &device.CreatedAt)
	
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Get lead count
	var leadCount int
	db.QueryRow("SELECT COUNT(*) FROM leads WHERE device_id = ?", device.ID).Scan(&leadCount)
	
	// Get messages sent count
	var messagesSent int
	db.QueryRow("SELECT COUNT(*) FROM broadcast_messages WHERE device_id = ? AND status = 'sent'", device.ID).Scan(&messagesSent)
	
	result := map[string]interface{}{
		"id":           device.ID,
		"device_name":  device.DeviceName,
		"jid":          device.JID.String,
		"platform":     device.Platform.String,
		"status":       device.Status,
		"created_at":   device.CreatedAt,
		"total_leads":  leadCount,
		"messages_sent": messagesSent,
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device info retrieved successfully",
		Results: result,
	})
}

// GetPublicDeviceSequences returns sequences for a specific device
func (handler *App) GetPublicDeviceSequences(c *fiber.Ctx) error {
	deviceName := c.Params("device_name")
	
	// Get MySQL connection
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = os.Getenv("DB_URI")
	}
	
	// Convert mysql:// URL to DSN format if needed
	if strings.HasPrefix(mysqlURI, "mysql://") {
		mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		parts := strings.Split(mysqlURI, "@")
		if len(parts) == 2 {
			userPass := parts[0]
			hostDb := parts[1]
			mysqlURI = userPass + "@tcp(" + strings.Replace(hostDb, "/", ")/", 1) + "?parseTime=true&charset=utf8mb4"
		}
	}
	
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Database connection error",
		})
	}
	defer db.Close()
	
	// First get device ID
	var deviceID string
	err = db.QueryRow("SELECT id FROM user_devices WHERE device_name = ?", deviceName).Scan(&deviceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Get sequences that this device is involved in
	query := `
		SELECT DISTINCT
			s.id,
			s.name,
			s.trigger,
			(SELECT COUNT(DISTINCT ss.id) FROM sequence_steps ss WHERE ss.sequence_id = s.id) as total_flows,
			COUNT(DISTINCT CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id)) as total_contacts,
			COUNT(DISTINCT CASE WHEN bm.status = 'sent' AND (bm.error_message IS NULL OR bm.error_message = '') 
				THEN CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id) END) as contacts_done,
			COUNT(DISTINCT CASE WHEN bm.status IN ('failed', 'error') OR (bm.status = 'sent' AND bm.error_message IS NOT NULL AND bm.error_message != '') 
				THEN CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id) END) as contacts_failed
		FROM sequences s
		INNER JOIN broadcast_messages bm ON bm.sequence_id = s.id
		WHERE bm.device_id = ?
		GROUP BY s.id, s.name, s.trigger
		ORDER BY s.created_at DESC
	`
	
	rows, err := db.Query(query, deviceID)
	if err != nil {
		log.Printf("Error querying sequences: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to fetch sequences",
		})
	}
	defer rows.Close()
	
	sequences := []map[string]interface{}{}
	for rows.Next() {
		var seq struct {
			ID             string
			Name           string
			Trigger        sql.NullString
			TotalFlows     int
			TotalContacts  int
			ContactsDone   int
			ContactsFailed int
		}
		
		err := rows.Scan(&seq.ID, &seq.Name, &seq.Trigger, &seq.TotalFlows, 
			&seq.TotalContacts, &seq.ContactsDone, &seq.ContactsFailed)
		if err != nil {
			continue
		}
		
		successRate := 0.0
		if seq.TotalContacts > 0 {
			successRate = float64(seq.ContactsDone) / float64(seq.TotalContacts) * 100
		}
		
		sequences = append(sequences, map[string]interface{}{
			"id":               seq.ID,
			"name":             seq.Name,
			"trigger":          seq.Trigger.String,
			"total_flows":      seq.TotalFlows,
			"total_contacts":   seq.TotalContacts,
			"contacts_done":    seq.ContactsDone,
			"contacts_failed":  seq.ContactsFailed,
			"success_rate":     fmt.Sprintf("%.1f", successRate),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequences retrieved successfully",
		Results: sequences,
	})
}


// PublicSequenceDeviceReport renders the public sequence device report page
func (handler *App) PublicSequenceDeviceReport(c *fiber.Ctx) error {
	return c.Render("public_sequence_device_report", fiber.Map{})
}


// UpdateSequencePendingMessages updates pending broadcast messages with latest sequence step content
func (handler *App) UpdateSequencePendingMessages(c *fiber.Ctx) error {
	sequenceId := c.Params("id")
	
	// Get session
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
	
	// Get MySQL connection
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = os.Getenv("DB_URI")
	}
	
	if strings.HasPrefix(mysqlURI, "mysql://") {
		mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		parts := strings.Split(mysqlURI, "@")
		if len(parts) == 2 {
			userPass := parts[0]
			hostDb := parts[1]
			mysqlURI = userPass + "@tcp(" + strings.Replace(hostDb, "/", ")/", 1) + "?parseTime=true&charset=utf8mb4"
		}
	}
	
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Database connection error",
		})
	}
	defer db.Close()
	
	// Step 1: Find changed steps by comparing with latest pending messages
	detectQuery := `
		SELECT 
			ss.id as step_id,
			ss.content as current_content,
			ss.media_url as current_media,
			latest_msg.content as message_content,
			latest_msg.media_url as message_media
		FROM sequence_steps ss
		LEFT JOIN (
			SELECT DISTINCT ON (sequence_stepid) 
				sequence_stepid, content, media_url
			FROM broadcast_messages
			WHERE sequence_id = ? AND status = 'pending' AND user_id = ?
			ORDER BY sequence_stepid, created_at DESC
		) latest_msg ON latest_msg.sequence_stepid = ss.id
		WHERE ss.sequence_id = ?
	`
	
	// MySQL doesn't support DISTINCT ON, so we need to rewrite the query
	detectQuery = `
		SELECT 
			ss.id as step_id,
			ss.content as current_content,
			ss.media_url as current_media,
			latest_msg.content as message_content,
			latest_msg.media_url as message_media
		FROM sequence_steps ss
		LEFT JOIN (
			SELECT bm1.sequence_stepid, bm1.content, bm1.media_url
			FROM broadcast_messages bm1
			INNER JOIN (
				SELECT sequence_stepid, MAX(created_at) as max_created
				FROM broadcast_messages
				WHERE sequence_id = ? AND status = 'pending' AND user_id = ?
				GROUP BY sequence_stepid
			) bm2 ON bm1.sequence_stepid = bm2.sequence_stepid 
			AND bm1.created_at = bm2.max_created
			WHERE bm1.sequence_id = ? AND bm1.status = 'pending' AND bm1.user_id = ?
		) latest_msg ON latest_msg.sequence_stepid = ss.id
		WHERE ss.sequence_id = ?
	`
	
	rows, err := db.Query(detectQuery, sequenceId, session.UserID, sequenceId, session.UserID, sequenceId)
	if err != nil {
		log.Printf("Error detecting changed steps: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to detect changes",
		})
	}
	defer rows.Close()
	
	type ChangedStep struct {
		StepID       string
		NewContent   string
		NewMediaURL  string
		MessageCount int
	}
	
	var changedSteps []ChangedStep
	var previewInfo []map[string]interface{}
	
	for rows.Next() {
		var stepID, currentContent, currentMedia string
		var messageContent, messageMedia sql.NullString
		
		err := rows.Scan(&stepID, &currentContent, &currentMedia, &messageContent, &messageMedia)
		if err != nil {
			continue
		}
		
		// Check if content or media changed
		contentChanged := !messageContent.Valid || messageContent.String != currentContent
		mediaChanged := !messageMedia.Valid || messageMedia.String != currentMedia
		
		if contentChanged || mediaChanged {
			// Count how many messages will be updated
			var count int
			countQuery := `
				SELECT COUNT(*) 
				FROM broadcast_messages 
				WHERE sequence_stepid = ? AND status = 'pending' AND user_id = ?
			`
			db.QueryRow(countQuery, stepID, session.UserID).Scan(&count)
			
			if count > 0 {
				changedSteps = append(changedSteps, ChangedStep{
					StepID:       stepID,
					NewContent:   currentContent,
					NewMediaURL:  currentMedia,
					MessageCount: count,
				})
				
				// Get step details for preview
				var dayNum int
				stepQuery := `
					SELECT COALESCE(day_number, day, 1) 
					FROM sequence_steps 
					WHERE id = ?
				`
				db.QueryRow(stepQuery, stepID).Scan(&dayNum)
				
				previewInfo = append(previewInfo, map[string]interface{}{
					"step_id":       stepID,
					"day":           dayNum,
					"message_count": count,
					"changes": map[string]bool{
						"content": contentChanged,
						"media":   mediaChanged,
					},
				})
			}
		}
	}
	
	// If preview mode (no confirm parameter), return preview info
	if c.Query("confirm") != "true" {
		totalMessages := 0
		for _, step := range changedSteps {
			totalMessages += step.MessageCount
		}
		
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "PREVIEW",
			Message: fmt.Sprintf("%d steps have changes affecting %d messages", len(changedSteps), totalMessages),
			Results: map[string]interface{}{
				"total_messages": totalMessages,
				"changed_steps":  len(changedSteps),
				"steps":          previewInfo,
			},
		})
	}
	
	// Step 2: Update messages for changed steps
	totalUpdated := 0
	
	for _, step := range changedSteps {
		updateQuery := `
			UPDATE broadcast_messages 
			SET content = ?,
			    media_url = ?,
			    updated_at = NOW()
			WHERE sequence_stepid = ? 
			AND status = 'pending' 
			AND user_id = ?
		`
		
		result, err := db.Exec(updateQuery, step.NewContent, step.NewMediaURL, step.StepID, session.UserID)
		if err != nil {
			log.Printf("Error updating messages for step %s: %v", step.StepID, err)
			continue
		}
		
		rowsAffected, _ := result.RowsAffected()
		totalUpdated += int(rowsAffected)
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Successfully updated %d pending messages", totalUpdated),
		Results: map[string]interface{}{
			"total_updated": totalUpdated,
			"steps_updated": len(changedSteps),
		},
	})
}
