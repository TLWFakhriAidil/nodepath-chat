package handlers

import (
	"encoding/json"
	"fmt"
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

	// Test endpoints for webhook data extraction
	api.Post("/test/waha/extraction", h.TestWahaExtraction)

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
// Processes WhatsApp messages and triggers AI responses based on device settings
// Implements standardized WAHA webhook data extraction and processing logic
func (h *AIWhatsappHandlers) HandleWahaWebhook(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	body := c.Body()
	
	// Log raw payload for debugging production issues
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"payload_size": len(body),
		"raw_payload": string(body),
	}).Info("🔍 WAHA: Raw webhook payload received")

	// Parse as generic map first for flexible extraction
	var rawPayload map[string]interface{}
	if err := json.Unmarshal(body, &rawPayload); err != nil {
		logrus.WithError(err).Error("Failed to parse WAHA webhook JSON")
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid JSON format",
		})
	}

	// Extract standardized webhook data according to requirements
	extractedData := h.extractWahaWebhookData(rawPayload)

	// Logic 1: If is_group = true → ignore (do not execute)
	if extractedData.IsGroup {
		logrus.WithFields(logrus.Fields{
			"sender_phone": extractedData.SenderPhone,
			"sender_name": extractedData.SenderName,
			"device_id": deviceID,
		}).Info("⏭️ WAHA: Ignoring group message as per requirements")
		return c.JSON(fiber.Map{
			"status": "ignored", 
			"reason": "group message",
			"extracted_data": extractedData,
		})
	}

	// Validate required fields
	if extractedData.SenderPhone == "" || extractedData.Message == "" {
		logrus.WithFields(logrus.Fields{
			"sender_phone": extractedData.SenderPhone,
			"message": truncateString(extractedData.Message, 100),
			"sender_name": extractedData.SenderName,
			"is_from_me": extractedData.IsFromMe,
			"is_group": extractedData.IsGroup,
		}).Warn("⚠️ WAHA: Missing required fields in extracted data")
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing required fields",
			"missing": map[string]bool{
				"sender_phone": extractedData.SenderPhone == "",
				"message": extractedData.Message == "",
			},
			"extracted_data": extractedData,
		})
	}

	// Clean phone number format (remove @c.us suffix if present)
	if strings.HasSuffix(extractedData.SenderPhone, "@c.us") {
		extractedData.SenderPhone = strings.TrimSuffix(extractedData.SenderPhone, "@c.us")
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"cleaned_from": extractedData.SenderPhone,
		}).Info("🔧 WAHA: Phone number cleaned - stripped @c.us suffix")
	}

	// Logic 2: If is_from_me = true → check message starting with %, #, cmd, or & (system/device commands)
	if extractedData.IsFromMe {
		message := strings.TrimSpace(extractedData.Message)
		isSystemCommand := strings.HasPrefix(message, "%") || 
						 strings.HasPrefix(message, "#") || 
						 strings.HasPrefix(message, "cmd") || 
						 strings.HasPrefix(message, "&")
		
		if isSystemCommand {
			logrus.WithFields(logrus.Fields{
				"sender_phone": extractedData.SenderPhone,
				"command": truncateString(message, 50),
				"device_id": deviceID,
			}).Info("🔧 WAHA: Processing system/device command from sender")
			
			// Process system command using existing device command logic
			go h.processIncomingMessage(extractedData.SenderPhone, message, deviceID, "waha")
			
			return c.JSON(fiber.Map{
				"status": "success",
				"type": "system_command",
				"extracted_data": extractedData,
			})
		} else {
			logrus.WithFields(logrus.Fields{
				"sender_phone": extractedData.SenderPhone,
				"message": truncateString(message, 100),
				"device_id": deviceID,
			}).Info("⏭️ WAHA: Ignoring non-system message from sender (is_from_me=true)")
			
			return c.JSON(fiber.Map{
				"status": "ignored", 
				"reason": "non-system message from sender",
				"extracted_data": extractedData,
			})
		}
	}

	// Logic 3: Otherwise → treat as normal customer message
	logrus.WithFields(logrus.Fields{
		"sender_phone": extractedData.SenderPhone,
		"sender_name": extractedData.SenderName,
		"message": truncateString(extractedData.Message, 100),
		"device_id": deviceID,
	}).Info("💬 WAHA: Processing normal customer message")

	// Process the customer message through the AI system
	go h.processIncomingMessage(extractedData.SenderPhone, extractedData.Message, deviceID, "waha")

	return c.JSON(fiber.Map{
		"status": "success",
		"type": "customer_message",
		"extracted_data": extractedData,
	})
}

// extractWahaFields extracts fields from WAHA webhook payload using multiple fallback methods
// Handles different WAHA payload structures that may vary in production
func (h *AIWhatsappHandlers) extractWahaFields(payload map[string]interface{}) (from, message, event, session string, isGroup bool) {
	// Use the new standardized extraction function
	extractedData := h.extractWahaWebhookData(payload)
	
	// Map to old function signature for backward compatibility
	from = extractedData.SenderPhone
	message = extractedData.Message
	isGroup = extractedData.IsGroup
	
	// Extract event and session from top level for backward compatibility
	if eventVal, ok := payload["event"].(string); ok {
		event = eventVal
	}
	if sessionVal, ok := payload["session"].(string); ok {
		session = sessionVal
	}
	
	return from, message, event, session, isGroup
}

// WahaWebhookData represents the standardized extracted data from WAHA webhook
type WahaWebhookData struct {
	SenderPhone string `json:"sender_phone"`
	SenderName  string `json:"sender_name"`
	Message     string `json:"message"`
	IsFromMe    bool   `json:"is_from_me"`
	IsGroup     bool   `json:"is_group"`
}

// extractWahaWebhookData extracts WAHA webhook data according to standardized requirements
// Only uses data inside the "payload" object as specified
// Returns extracted fields in standardized JSON format
func (h *AIWhatsappHandlers) extractWahaWebhookData(webhookPayload map[string]interface{}) WahaWebhookData {
	var result WahaWebhookData
	
	// Log payload structure for debugging
	logrus.WithFields(logrus.Fields{
		"payload_keys": getMapKeys(webhookPayload),
		"has_payload": webhookPayload["payload"] != nil,
	}).Debug("🔍 WAHA: Analyzing webhook payload structure")
	
	// Extract from payload object only as specified in requirements
	payloadObj, ok := webhookPayload["payload"].(map[string]interface{})
	if !ok {
		logrus.Warn("⚠️ WAHA: No payload object found in webhook data")
		return result
	}
	
	// Extract sender_phone = payload.from
	if fromVal, ok := payloadObj["from"].(string); ok {
		result.SenderPhone = fromVal
		logrus.WithField("extraction_method", "payload_from").Debug("🔍 WAHA: Sender phone extracted from payload.from")
	}
	
	// Extract message = payload.body
	if bodyVal, ok := payloadObj["body"].(string); ok {
		result.Message = bodyVal
		logrus.WithField("extraction_method", "payload_body").Debug("🔍 WAHA: Message extracted from payload.body")
	}
	
	// Extract is_from_me = payload.fromMe (true/false)
	if fromMeVal, ok := payloadObj["fromMe"].(bool); ok {
		result.IsFromMe = fromMeVal
		logrus.WithField("extraction_method", "payload_fromMe").Debug("🔍 WAHA: IsFromMe extracted from payload.fromMe")
	}
	
	// Extract sender_name = payload.media.Info.PushName
	if mediaObj, ok := payloadObj["media"].(map[string]interface{}); ok {
		if infoObj, ok := mediaObj["Info"].(map[string]interface{}); ok {
			if pushNameVal, ok := infoObj["PushName"].(string); ok {
				result.SenderName = pushNameVal
				logrus.WithField("extraction_method", "payload_media_info_pushname").Debug("🔍 WAHA: Sender name extracted from payload.media.Info.PushName")
			}
		}
	}
	
	// Extract is_group = payload.media.Info.IsGroup (true/false)
	if mediaObj, ok := payloadObj["media"].(map[string]interface{}); ok {
		if infoObj, ok := mediaObj["Info"].(map[string]interface{}); ok {
			if isGroupVal, ok := infoObj["IsGroup"].(bool); ok {
				result.IsGroup = isGroupVal
				logrus.WithField("extraction_method", "payload_media_info_isgroup").Debug("🔍 WAHA: IsGroup extracted from payload.media.Info.IsGroup")
			}
		}
	}
	
	// Log extraction results
	logrus.WithFields(logrus.Fields{
		"sender_phone": result.SenderPhone,
		"sender_name": result.SenderName,
		"message": truncateString(result.Message, 100),
		"is_from_me": result.IsFromMe,
		"is_group": result.IsGroup,
		"extraction_success": result.SenderPhone != "" && result.Message != "",
	}).Info("🔍 WAHA: Standardized field extraction completed")
	
	// Log warning if critical fields are missing
	if result.SenderPhone == "" || result.Message == "" {
		logrus.WithFields(logrus.Fields{
			"missing_sender_phone": result.SenderPhone == "",
			"missing_message": result.Message == "",
			"payload_keys": getMapKeys(payloadObj),
		}).Warn("⚠️ WAHA: Critical field extraction failed - check payload structure")
	}
	
	return result
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

// TestWahaExtraction tests the WAHA webhook data extraction
// Returns extracted fields in the standardized JSON format for testing
func (h *AIWhatsappHandlers) TestWahaExtraction(c *fiber.Ctx) error {
	// Parse the incoming webhook request
	var req WahaWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("❌ WAHA TEST: Failed to parse webhook request")
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request format",
			"details": err.Error(),
		})
	}

	// Extract standardized webhook data
	extractedData := h.extractWahaWebhookData(req.Payload)

	// Log the test extraction
	logrus.WithFields(logrus.Fields{
		"sender_phone": extractedData.SenderPhone,
		"sender_name": extractedData.SenderName,
		"message": truncateString(extractedData.Message, 100),
		"is_from_me": extractedData.IsFromMe,
		"is_group": extractedData.IsGroup,
	}).Info("🧪 WAHA TEST: Extraction completed")

	// Return extracted fields in standardized JSON format as specified
	return c.JSON(extractedData)
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

	// Get user ID from authentication context
	userID, ok := c.Locals("userID").(int)
	if !ok {
		logrus.Error("User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(AnalyticsResponse{
			Success: false,
			Message: "Authentication required",
		})
	}

	// Get analytics data from repository with user-specific filtering
	analyticsData, err := h.AIRepo.GetAnalyticsData(startDate, endDate, req.DeviceID, userID)
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

	// Get user ID from authentication context
	userID, ok := c.Locals("userID").(int)
	if !ok {
		logrus.Error("User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Authentication required",
		})
	}

	// Get data from repository with user-specific filtering
	data, total, err := h.AIRepo.GetAllAIWhatsappData(limit, offset, deviceFilter, stageFilter, search, userID)
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

// Helper functions for comprehensive WAHA webhook debugging

// getMapKeys returns all keys from a map for debugging payload structure
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// analyzePayloadDepth analyzes the depth and structure of nested payload
func analyzePayloadDepth(payload map[string]interface{}) map[string]interface{} {
	analysis := make(map[string]interface{})
	
	for key, value := range payload {
		switch v := value.(type) {
		case map[string]interface{}:
			analysis[key] = map[string]interface{}{
				"type": "object",
				"keys": getMapKeys(v),
				"key_count": len(v),
			}
		case []interface{}:
			analysis[key] = map[string]interface{}{
				"type": "array",
				"length": len(v),
			}
		case string:
			analysis[key] = map[string]interface{}{
				"type": "string",
				"length": len(v),
			}
		default:
			analysis[key] = map[string]interface{}{
				"type": fmt.Sprintf("%T", v),
			}
		}
	}
	
	return analysis
}

// truncateString truncates a string to specified length for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}