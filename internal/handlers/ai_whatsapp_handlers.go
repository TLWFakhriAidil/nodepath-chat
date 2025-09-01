package handlers

import (
	"strconv"
	"strings"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"
	"nodepath-chat/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// AIWhatsappHandlers contains all AI WhatsApp webhook handlers
type AIWhatsappHandlers struct {
	AIWhatsappService services.AIWhatsappService
	AIRepo            repository.AIWhatsappRepository
	DeviceRepo        repository.DeviceSettingsRepository
}

// NewAIWhatsappHandlers creates a new AI WhatsApp handlers instance
func NewAIWhatsappHandlers(
	aiWhatsappService services.AIWhatsappService,
	aiRepo repository.AIWhatsappRepository,
	deviceRepo repository.DeviceSettingsRepository,
) *AIWhatsappHandlers {
	return &AIWhatsappHandlers{
		AIWhatsappService: aiWhatsappService,
		AIRepo:            aiRepo,
		DeviceRepo:        deviceRepo,
	}
}

// SetupAIWhatsappRoutes sets up AI WhatsApp webhook routes
func (h *AIWhatsappHandlers) SetupAIWhatsappRoutes(api fiber.Router) {
	// Webhook endpoints for receiving WhatsApp messages
	api.Post("/webhook/whatsapp/:device_id", h.HandleWhatsappWebhook)
	api.Post("/webhook/wablas/:device_id", h.HandleWablasWebhook)
	api.Post("/webhook/whacenter/:device_id", h.HandleWhacenterWebhook)
	api.Post("/webhook/waha/:device_id", h.HandleWahaWebhook)

	// AI conversation management endpoints
	api.Post("/ai/conversation/start", h.StartAIConversation)
	api.Post("/ai/conversation/process", h.ProcessAIMessage)
	api.Post("/ai/conversation/toggle-human", h.ToggleHumanTakeover)
	api.Get("/ai/conversation/history/:prospect_num", h.GetConversationHistory)
	api.Get("/ai/conversation/status/:prospect_num", h.GetConversationStatus)

	// AI settings management
	api.Get("/ai/settings/:staff_id", h.GetAISettings)
	api.Post("/ai/settings", h.CreateAISettings)
	api.Put("/ai/settings/:staff_id", h.UpdateAISettings)
	api.Delete("/ai/settings/:staff_id", h.DeleteAISettings)

	// Device command processing
	api.Post("/ai/device/command", h.ProcessDeviceCommand)

	// Analytics endpoints
	api.Get("/ai/analytics", h.GetAnalytics)
	api.Post("/ai/analytics", h.GetAnalytics)

	// Data table endpoints
	api.Get("/ai/whatsapp/data", h.GetAllAIWhatsappData)
}

// WhatsappWebhookRequest represents incoming WhatsApp webhook data
type WhatsappWebhookRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	DeviceID  string `json:"device_id"`
}

// WablasWebhookRequest represents incoming Wablas webhook data
type WablasWebhookRequest struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
	Device  string `json:"device"`
	Time    string `json:"time"`
}

// WhacenterWebhookRequest represents incoming Whacenter webhook data
type WhacenterWebhookRequest struct {
	Number  string `json:"number"`
	Text    string `json:"text"`
	Device  string `json:"device"`
	Date    string `json:"date"`
}

// WahaWebhookRequest represents incoming WAHA webhook data
// WAHA uses nested payload structure with _data containing message info
type WahaWebhookRequest struct {
	Event   string `json:"event"`
	Session string `json:"session"`
	Payload struct {
		Data struct {
			From string `json:"from"`
			Body string `json:"body"`
			Info struct {
				IsGroup bool `json:"IsGroup"`
			} `json:"Info"`
		} `json:"_data"`
	} `json:"payload"`
}

// StartAIConversationRequest represents request to start AI conversation
type StartAIConversationRequest struct {
	ProspectNum string `json:"prospect_num"`
	IDDevice    string `json:"id_device"`
	Niche       string `json:"niche"`
	Stage       string `json:"stage"`
}

// ProcessAIMessageRequest represents request to process AI message
type ProcessAIMessageRequest struct {
	ProspectNum string `json:"prospect_num"`
	IDDevice    string `json:"id_device"`
	Message     string `json:"message"`
	Stage       string `json:"stage"`
}

// ToggleHumanTakeoverRequest represents request to toggle human takeover
type ToggleHumanTakeoverRequest struct {
	ProspectNum string `json:"prospect_num"`
	Human       bool   `json:"human"`
}

// ProcessDeviceCommandRequest represents request to process device command
type ProcessDeviceCommandRequest struct {
	ProspectNum string `json:"prospect_num"`
	Command     string `json:"command"`
	IDDevice    string `json:"id_device"`
}

// AnalyticsRequest represents the request structure for analytics endpoint
type AnalyticsRequest struct {
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
	DeviceID  string `json:"device_id" form:"device_id"`
}

// AnalyticsResponse represents the response structure for analytics endpoint
type AnalyticsResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// HandleWhatsappWebhook handles generic WhatsApp webhook messages
func (h *AIWhatsappHandlers) HandleWhatsappWebhook(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	if deviceID == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Device ID is required")
	}

	var req WhatsappWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse WhatsApp webhook request")
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	req.DeviceID = deviceID

	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"from":      req.From,
		"message":   req.Message,
	}).Info("Received WhatsApp webhook")

	// Process the message asynchronously
	go h.processIncomingMessage(req.From, req.Message, deviceID, "whatsapp")

	return h.successResponse(c, map[string]string{"status": "received"})
}

// HandleWablasWebhook handles Wablas provider webhook messages
func (h *AIWhatsappHandlers) HandleWablasWebhook(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	if deviceID == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Device ID is required")
	}

	var req WablasWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse Wablas webhook request")
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"phone":     req.Phone,
		"message":   req.Message,
	}).Info("Received Wablas webhook")

	// Process the message asynchronously
	go h.processIncomingMessage(req.Phone, req.Message, deviceID, "wablas")

	return h.successResponse(c, map[string]string{"status": "received"})
}

// HandleWhacenterWebhook handles Whacenter provider webhook messages
func (h *AIWhatsappHandlers) HandleWhacenterWebhook(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	if deviceID == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Device ID is required")
	}

	var req WhacenterWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse Whacenter webhook request")
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"number":    req.Number,
		"text":      req.Text,
	}).Info("Received Whacenter webhook")

	// Process the message asynchronously
	go h.processIncomingMessage(req.Number, req.Text, deviceID, "whacenter")

	return h.successResponse(c, map[string]string{"status": "received"})
}

// HandleWahaWebhook handles incoming WAHA webhook requests
// WAHA provider uses nested payload structure with specific event types
func (h *AIWhatsappHandlers) HandleWahaWebhook(c *fiber.Ctx) error {
	var req WahaWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse WAHA webhook request")
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format",
		})
	}

	// Only process message events, ignore other events like status updates
	if req.Event != "message" {
		logrus.WithFields(logrus.Fields{
			"event":   req.Event,
			"session": req.Session,
		}).Debug("Ignoring non-message WAHA event")
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Event ignored",
		})
	}

	// Skip group messages as per business logic
	if req.Payload.Data.Info.IsGroup {
		logrus.WithFields(logrus.Fields{
			"from":    req.Payload.Data.From,
			"session": req.Session,
		}).Debug("Ignoring group message from WAHA")
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Group message ignored",
		})
	}

	logrus.WithFields(logrus.Fields{
		"from":    req.Payload.Data.From,
		"body":    req.Payload.Data.Body,
		"session": req.Session,
		"event":   req.Event,
	}).Info("Received WAHA webhook")

	// Process the incoming message asynchronously
	// Use session as device identifier for WAHA
	go h.processIncomingMessage(req.Payload.Data.From, req.Payload.Data.Body, req.Session, "waha")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Webhook received",
	})
}

// StartAIConversation starts a new AI conversation
func (h *AIWhatsappHandlers) StartAIConversation(c *fiber.Ctx) error {
	var req StartAIConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.ProspectNum == "" || req.IDDevice == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Missing required fields")
	}

	// Create AI WhatsApp conversation record
	aiWhatsapp := &models.AIWhatsapp{
		ProspectNum: req.ProspectNum,
		IDDevice:    req.IDDevice,
		Stage:       req.Stage,
		Human:       0, // AI active by default
		Niche:       req.Niche,
	}

	err := h.AIRepo.CreateAIWhatsapp(aiWhatsapp)
	if err != nil {
		logrus.WithError(err).Error("Failed to create AI conversation")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to start AI conversation")
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num": req.ProspectNum,
		"id_device":    req.IDDevice,
	}).Info("AI conversation started")

	return h.successResponse(c, aiWhatsapp)
}

// ProcessAIMessage processes an AI message manually
func (h *AIWhatsappHandlers) ProcessAIMessage(c *fiber.Ctx) error {
	var req ProcessAIMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.ProspectNum == "" || req.IDDevice == "" || req.Message == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Missing required fields")
	}

	// Process AI conversation
	response, err := h.AIWhatsappService.ProcessAIConversation(req.ProspectNum, req.IDDevice, req.Message, req.Stage)
	if err != nil {
		logrus.WithError(err).Error("Failed to process AI conversation")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to process AI message")
	}

	return h.successResponse(c, response)
}

// ToggleHumanTakeover toggles human takeover for a conversation
func (h *AIWhatsappHandlers) ToggleHumanTakeover(c *fiber.Ctx) error {
	var req ToggleHumanTakeoverRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	if req.ProspectNum == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Prospect number is required")
	}

	err := h.AIWhatsappService.ToggleHumanTakeover(req.ProspectNum, req.Human)
	if err != nil {
		logrus.WithError(err).Error("Failed to toggle human takeover")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to toggle human takeover")
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num": req.ProspectNum,
		"human":        req.Human,
	}).Info("Human takeover toggled")

	return h.successResponse(c, map[string]interface{}{
		"prospect_num": req.ProspectNum,
		"human":        req.Human,
		"status":       "updated",
	})
}

// GetConversationHistory retrieves conversation history for a prospect
func (h *AIWhatsappHandlers) GetConversationHistory(c *fiber.Ctx) error {
	prospectNum := c.Params("prospect_num")
	if prospectNum == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Prospect number is required")
	}

	// Get limit from query parameter
	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	history, err := h.AIRepo.GetConversationHistory(prospectNum, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation history")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get conversation history")
	}

	return h.successResponse(c, history)
}

// GetConversationStatus retrieves conversation status for a prospect
func (h *AIWhatsappHandlers) GetConversationStatus(c *fiber.Ctx) error {
	prospectNum := c.Params("prospect_num")
	if prospectNum == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Prospect number is required")
	}

	aiConv, err := h.AIRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation status")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get conversation status")
	}

	if aiConv == nil {
		return h.errorResponse(c, fiber.StatusNotFound, "Conversation not found")
	}

	return h.successResponse(c, aiConv)
}

// GetAISettings retrieves AI settings for a staff member
func (h *AIWhatsappHandlers) GetAISettings(c *fiber.Ctx) error {
	staffID := c.Params("staff_id")
	if staffID == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid staff ID")
	}

	settings, err := h.AIWhatsappService.GetAISettings(staffID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI settings")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get AI settings")
	}

	if settings == nil {
		return h.errorResponse(c, fiber.StatusNotFound, "AI settings not found")
	}

	return h.successResponse(c, settings)
}

// CreateAISettings creates new AI settings
func (h *AIWhatsappHandlers) CreateAISettings(c *fiber.Ctx) error {
	var settings models.AISettings
	if err := c.BodyParser(&settings); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if settings.IDDevice == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Device ID is required")
	}

	// TODO: Implement CreateAISettings method in repository
	logrus.Info("AI settings creation requested but not implemented yet")
	return h.errorResponse(c, fiber.StatusNotImplemented, "AI settings creation not implemented yet")
}

// UpdateAISettings updates existing AI settings
func (h *AIWhatsappHandlers) UpdateAISettings(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid settings ID")
	}

	var settings models.AISettings
	if err := c.BodyParser(&settings); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	settings.ID = idStr
	// TODO: Implement UpdateAISettings method in repository
	logrus.Info("AI settings update requested but not implemented yet")
	return h.errorResponse(c, fiber.StatusNotImplemented, "AI settings update not implemented yet")
}

// DeleteAISettings deletes AI settings
func (h *AIWhatsappHandlers) DeleteAISettings(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid settings ID")
	}

	// TODO: Implement DeleteAISettings method in repository
	logrus.Info("AI settings deletion requested but not implemented yet")
	return h.errorResponse(c, fiber.StatusNotImplemented, "AI settings deletion not implemented yet")
}

// ProcessDeviceCommand processes device-specific commands
func (h *AIWhatsappHandlers) ProcessDeviceCommand(c *fiber.Ctx) error {
	var req ProcessDeviceCommandRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	if req.ProspectNum == "" || req.Command == "" || req.IDDevice == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Missing required fields")
	}

	err := h.AIWhatsappService.ProcessDeviceCommand(req.ProspectNum, req.Command, req.IDDevice)
	if err != nil {
		logrus.WithError(err).Error("Failed to process device command")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to process device command")
	}

	return h.successResponse(c, map[string]string{"status": "processed"})
}

// processIncomingMessage processes incoming WhatsApp messages asynchronously
func (h *AIWhatsappHandlers) processIncomingMessage(prospectNum, message, deviceID, provider string) {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"device_id":    deviceID,
		"provider":     provider,
		"message":      message,
	}).Info("Processing incoming message")

	// Check if this is a device command
	if strings.HasPrefix(message, "%") || strings.HasPrefix(message, "#") || strings.ToLower(message) == "cmd" {
		err := h.AIWhatsappService.ProcessDeviceCommand(prospectNum, message, deviceID)
		if err != nil {
			logrus.WithError(err).Error("Failed to process device command")
		}
		return
	}

	// Get current conversation stage
	aiConv, err := h.AIRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI conversation")
		return
	}

	var stage string
	if aiConv != nil {
		stage = aiConv.Stage
	}

	// Process AI conversation
	response, err := h.AIWhatsappService.ProcessAIConversation(prospectNum, deviceID, message, stage)
	if err != nil {
		logrus.WithError(err).Error("Failed to process AI conversation")
		return
	}

	// Send response if we have a response
	// Note: ProcessAIConversation already handles conversation logging internally via LogConversation
	// Removed duplicate SaveConversationHistory call to prevent duplicate saves
	if response != nil {
		// Send response back to WhatsApp
		h.sendWhatsappResponse(prospectNum, deviceID, provider, response)
	}
}

// sendWhatsappResponse sends AI response back to WhatsApp
func (h *AIWhatsappHandlers) sendWhatsappResponse(prospectNum, deviceID, provider string, response *services.AIWhatsappResponse) {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"device_id":    deviceID,
		"provider":     provider,
		"stage":        response.Stage,
	}).Info("Sending AI response to WhatsApp")

	// TODO: Implement actual WhatsApp sending logic based on provider
	// This would integrate with your existing WhatsApp service
	for _, item := range response.Response {
		if item.Type == "text" {
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"content":      item.Content,
			}).Info("Sending text message")
			// Send text message
		} else if item.Type == "image" {
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"image_url":    item.Content,
			}).Info("Sending image message")
			// Send image message
		}
	}
}

// GetAnalytics retrieves analytics data from ai_whatsapp_nodepath with date filtering
func (h *AIWhatsappHandlers) GetAnalytics(c *fiber.Ctx) error {
	var req AnalyticsRequest

	// Parse query parameters with frontend parameter names or JSON body
	if c.Method() == "GET" {
		// Handle GET request with query parameters
		req.StartDate = c.Query("startDate", "")
		req.EndDate = c.Query("endDate", "")
		req.DeviceID = c.Query("idDevice", "")
	} else {
		// Handle POST request with JSON body
		if err := c.BodyParser(&req); err != nil {
			logrus.WithError(err).Error("Failed to parse analytics request")
			return c.Status(fiber.StatusBadRequest).JSON(AnalyticsResponse{
				Success: false,
				Message: "Invalid request format",
			})
		}
	}

	// Set default date range if not provided (current month start to today)
	now := time.Now()
	var startDate, endDate time.Time
	var err error

	if req.StartDate == "" {
		// Default to first day of current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	} else {
		startDate, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			logrus.WithError(err).Error("Invalid start date format")
			return c.Status(fiber.StatusBadRequest).JSON(AnalyticsResponse{
				Success: false,
				Message: "Invalid start date format. Use YYYY-MM-DD",
			})
		}
	}

	if req.EndDate == "" {
		// Default to today
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	} else {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			logrus.WithError(err).Error("Invalid end date format")
			return c.Status(fiber.StatusBadRequest).JSON(AnalyticsResponse{
				Success: false,
				Message: "Invalid end date format. Use YYYY-MM-DD",
			})
		}
		// Set end time to end of day
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
	}

	// Validate date range
	if startDate.After(endDate) {
		return c.Status(fiber.StatusBadRequest).JSON(AnalyticsResponse{
			Success: false,
			Message: "Start date cannot be after end date",
		})
	}

	// Set default device filter to "all" if not provided
	if req.DeviceID == "" {
		req.DeviceID = "all"
	}

	// Get analytics data from repository
	analyticsData, err := h.AIRepo.GetAnalyticsData(startDate, endDate, req.DeviceID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get analytics data")
		return c.Status(fiber.StatusInternalServerError).JSON(AnalyticsResponse{
			Success: false,
			Message: "Failed to retrieve analytics data",
		})
	}

	// Log successful analytics request
	logrus.WithFields(logrus.Fields{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"device_id":  req.DeviceID,
	}).Info("Analytics data retrieved successfully")

	// Transform data to match frontend expectations
	summary := analyticsData["summary"].(map[string]interface{})
	
	// Transform stage distribution from array to object format
	stageDistributionArray := analyticsData["stage_distribution"].([]map[string]interface{})
	stageDistributionMap := make(map[string]interface{})
	for _, item := range stageDistributionArray {
		stage := item["stage"].(string)
		count := item["count"]
		stageDistributionMap[stage] = count
	}
	
	responseData := map[string]interface{}{
		"totalConversations":       summary["total_conversations"],
		"aiActiveConversations":    summary["ai_active"],
		"humanTakeovers":           summary["human_takeover"],
		"uniqueDevices":            summary["unique_devices"],
		"uniqueNiches":             summary["unique_niches"],
		"conversationsWithStages":  summary["conversations_with_stage"],
		"aiActivePercentage":       summary["ai_active_percentage"],
		"humanTakeoverPercentage":  summary["human_takeover_percentage"],
		"dailyBreakdown":           analyticsData["daily_data"],
		"stageDistribution":        stageDistributionMap,
		"dateRange":                analyticsData["date_range"],
	}

	return c.JSON(responseData)
}

// GetAllAIWhatsappData retrieves all AI WhatsApp conversation records for data table display
func (h *AIWhatsappHandlers) GetAllAIWhatsappData(c *fiber.Ctx) error {
	// Parse query parameters for pagination and filtering
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	deviceFilter := c.Query("device_id", "")
	stageFilter := c.Query("stage", "")
	search := c.Query("search", "")

	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Get data from repository
	data, total, err := h.AIRepo.GetAllAIWhatsappData(limit, offset, deviceFilter, stageFilter, search)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp data")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve AI WhatsApp data",
		})
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit

	// Return paginated response
	return c.JSON(fiber.Map{
		"success": true,
		"data": data,
		"pagination": fiber.Map{
			"current_page": page,
			"total_pages":  totalPages,
			"total_records": total,
			"limit":        limit,
		},
	})
}

// Helper methods for consistent response formatting
func (h *AIWhatsappHandlers) successResponse(c *fiber.Ctx, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

func (h *AIWhatsappHandlers) errorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}