package handlers

import (
	"strconv"
	"strings"

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
		aiWhatsappService: aiWhatsappService,
		aiRepo:            aiRepo,
		deviceRepo:        deviceRepo,
	}
}

// SetupAIWhatsappRoutes sets up AI WhatsApp webhook routes
func (h *AIWhatsappHandlers) SetupAIWhatsappRoutes(api fiber.Router) {
	// Webhook endpoints for receiving WhatsApp messages
	api.Post("/webhook/whatsapp/:device_id", h.HandleWhatsappWebhook)
	api.Post("/webhook/wablas/:device_id", h.HandleWablasWebhook)
	api.Post("/webhook/whacenter/:device_id", h.HandleWhacenterWebhook)

	// AI conversation management endpoints
	api.Post("/ai/conversation/start", h.StartAIConversation)
	api.Post("/ai/conversation/process", h.ProcessAIMessage)
	api.Post("/ai/conversation/toggle-human", h.ToggleHumanTakeover)
	api.Get("/ai/conversation/history/:prospect_num", h.GetConversationHistory)
	api.Get("/ai/conversation/status/:prospect_num", h.GetConversationStatus)

	// AI settings management
	api.Get("/ai/settings/:staff_id", h.GetAISettings)
	api.Post("/ai/settings", h.CreateAISettings)
	api.Put("/ai/settings/:id", h.UpdateAISettings)
	api.Delete("/ai/settings/:id", h.DeleteAISettings)

	// Device command processing
	api.Post("/ai/device/command", h.ProcessDeviceCommand)
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

// StartAIConversationRequest represents request to start AI conversation
type StartAIConversationRequest struct {
	ProspectNum string `json:"prospect_num"`
	IDStaff     string `json:"id_staff"`
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

// StartAIConversation starts a new AI conversation
func (h *AIWhatsappHandlers) StartAIConversation(c *fiber.Ctx) error {
	var req StartAIConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.ProspectNum == "" || req.IDStaff == "" || req.IDDevice == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Missing required fields")
	}

	// Create AI WhatsApp conversation record
	aiWhatsapp := &models.AIWhatsapp{
		ProspectNum: req.ProspectNum,
		IDStaff:     req.IDStaff,
		Stage:       req.Stage,
		Human:       0, // AI active by default
		Niche:       req.Niche,
	}

	err := h.aiRepo.CreateAIWhatsapp(aiWhatsapp)
	if err != nil {
		logrus.WithError(err).Error("Failed to create AI conversation")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to start AI conversation")
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num": req.ProspectNum,
		"id_staff":     req.IDStaff,
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
	response, err := h.aiWhatsappService.ProcessAIConversation(req.ProspectNum, req.IDDevice, req.Message, req.Stage)
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

	err := h.aiWhatsappService.ToggleHumanTakeover(req.ProspectNum, req.Human)
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

	history, err := h.aiRepo.GetConversationHistory(prospectNum, limit)
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

	aiConv, err := h.aiRepo.GetAIWhatsappByProspectNum(prospectNum)
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

	settings, err := h.aiWhatsappService.GetAISettings(staffID)
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
	if settings.IDStaff == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Staff ID is required")
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

	err := h.aiWhatsappService.ProcessDeviceCommand(req.ProspectNum, req.Command, req.IDDevice)
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
		err := h.aiWhatsappService.ProcessDeviceCommand(prospectNum, message, deviceID)
		if err != nil {
			logrus.WithError(err).Error("Failed to process device command")
		}
		return
	}

	// Get current conversation stage
	aiConv, err := h.aiRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI conversation")
		return
	}

	var stage string
	if aiConv != nil {
		stage = aiConv.Stage
	}

	// Process AI conversation
	response, err := h.aiWhatsappService.ProcessAIConversation(prospectNum, deviceID, message, stage)
	if err != nil {
		logrus.WithError(err).Error("Failed to process AI conversation")
		return
	}

	// Send response back to WhatsApp (this would integrate with your WhatsApp sending service)
	if response != nil {
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