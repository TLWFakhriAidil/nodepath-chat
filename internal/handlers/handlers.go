package handlers

import (
	"strconv"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"
	"nodepath-chat/internal/whatsapp"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	flowService           *services.FlowService
	chatService           *services.ChatService
	aiService             *services.AIService
	queueService          *services.QueueService
	whatsappService       *whatsapp.Service
	deviceSettingsService *services.DeviceSettingsService
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	flowService *services.FlowService,
	chatService *services.ChatService,
	aiService *services.AIService,
	queueService *services.QueueService,
	whatsappService *whatsapp.Service,
	deviceSettingsService *services.DeviceSettingsService,
) *Handlers {
	return &Handlers{
		flowService:           flowService,
		chatService:           chatService,
		aiService:             aiService,
		queueService:          queueService,
		whatsappService:       whatsappService,
		deviceSettingsService: deviceSettingsService,
	}
}

// SetupRoutes sets up all API routes
func (h *Handlers) SetupRoutes(api fiber.Router) {
	// Flow management routes
	flows := api.Group("/flows")
	flows.Get("/", h.GetFlows)
	flows.Post("/", h.CreateFlow)
	flows.Get("/:id", h.GetFlow)
	flows.Put("/:id", h.UpdateFlow)
	flows.Delete("/:id", h.DeleteFlow)

	// Test chat routes
	testChat := api.Group("/test-chat")
	testChat.Post("/start/:flowId", h.StartTestChat)
	testChat.Post("/send/:executionId", h.SendTestMessage)
	testChat.Get("/history/:executionId", h.GetTestChatHistory)
	testChat.Post("/reset/:executionId", h.ResetTestChat)

	// Execution routes
	executions := api.Group("/executions")
	executions.Get("/", h.GetExecutions)
	executions.Get("/:id", h.GetExecution)
	executions.Post("/:id/complete", h.CompleteExecution)
	executions.Delete("/:id", h.DeleteExecution)

	// WhatsApp routes
	whatsapp := api.Group("/whatsapp")
	whatsapp.Get("/status", h.GetWhatsAppStatus)
	whatsapp.Post("/connect", h.ConnectWhatsApp)
	whatsapp.Post("/disconnect", h.DisconnectWhatsApp)
	whatsapp.Get("/qr", h.GetWhatsAppQR)
	whatsapp.Post("/send", h.SendWhatsAppMessage)

	// Queue routes
	queue := api.Group("/queue")
	queue.Get("/stats", h.GetQueueStats)
	queue.Post("/clear-failed", h.ClearFailedQueue)

	// AI routes
	ai := api.Group("/ai")
	ai.Post("/validate-key", h.ValidateAPIKey)
	ai.Get("/models", h.GetSupportedModels)

	// Analytics routes
	analytics := api.Group("/analytics")
	analytics.Get("/overview", h.GetAnalyticsOverview)
	analytics.Get("/flows", h.GetFlowStats)

	// Device settings routes
	deviceSettings := api.Group("/device-settings")
	deviceSettings.Get("/", h.GetDeviceSettings)
	deviceSettings.Get("/:id", h.GetDeviceSettingsById)
	deviceSettings.Post("/", h.CreateDeviceSettings)
	deviceSettings.Put("/:id", h.UpdateDeviceSettings)
	deviceSettings.Delete("/:id", h.DeleteDeviceSettings)
	deviceSettings.Post("/generate-whacenter", h.GenerateWhacenterDevice)
	deviceSettings.Post("/generate-wablas", h.GenerateWablasDevice)

	// Webhook routes for receiving messages from providers
	webhook := api.Group("/webhook")
	webhook.Post("/:id_device/:instance", h.HandleWebhook)
}

// Response helpers
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func (h *Handlers) successResponse(c *fiber.Ctx, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handlers) successMessageResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func (h *Handlers) errorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}

// Flow handlers

// GetFlows returns all flows
func (h *Handlers) GetFlows(c *fiber.Ctx) error {
	flows, err := h.flowService.GetAllFlows()
	if err != nil {
		logrus.WithError(err).Error("Failed to get flows")
		return h.errorResponse(c, 500, "Failed to retrieve flows")
	}

	return h.successResponse(c, flows)
}

// CreateFlow creates a new flow
func (h *Handlers) CreateFlow(c *fiber.Ctx) error {
	var flow models.ChatbotFlow
	if err := c.BodyParser(&flow); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	if err := h.flowService.CreateFlow(&flow); err != nil {
		logrus.WithError(err).Error("Failed to create flow")
		return h.errorResponse(c, 500, "Failed to create flow")
	}

	return h.successResponse(c, flow)
}

// GetFlow returns a specific flow
func (h *Handlers) GetFlow(c *fiber.Ctx) error {
	id := c.Params("id")
	flowID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return h.errorResponse(c, 400, "Invalid flow ID")
	}

	flow, err := h.flowService.GetFlowByID(uint(flowID))
	if err != nil {
		logrus.WithError(err).Error("Failed to get flow")
		return h.errorResponse(c, 404, "Flow not found")
	}

	return h.successResponse(c, flow)
}

// UpdateFlow updates an existing flow
func (h *Handlers) UpdateFlow(c *fiber.Ctx) error {
	id := c.Params("id")
	flowID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return h.errorResponse(c, 400, "Invalid flow ID")
	}

	var flow models.ChatbotFlow
	if err := c.BodyParser(&flow); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	flow.ID = uint(flowID)
	if err := h.flowService.UpdateFlow(&flow); err != nil {
		logrus.WithError(err).Error("Failed to update flow")
		return h.errorResponse(c, 500, "Failed to update flow")
	}

	return h.successResponse(c, flow)
}

// DeleteFlow deletes a flow
func (h *Handlers) DeleteFlow(c *fiber.Ctx) error {
	id := c.Params("id")
	flowID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return h.errorResponse(c, 400, "Invalid flow ID")
	}

	if err := h.flowService.DeleteFlow(uint(flowID)); err != nil {
		logrus.WithError(err).Error("Failed to delete flow")
		return h.errorResponse(c, 500, "Failed to delete flow")
	}

	return h.successMessageResponse(c, "Flow deleted successfully", nil)
}

// Test chat handlers

// StartTestChat starts a new test chat session
func (h *Handlers) StartTestChat(c *fiber.Ctx) error {
	flowID := c.Params("flowId")
	parsedFlowID, err := strconv.ParseUint(flowID, 10, 32)
	if err != nil {
		return h.errorResponse(c, 400, "Invalid flow ID")
	}

	// Get the flow to validate it exists
	flow, err := h.flowService.GetFlowByID(uint(parsedFlowID))
	if err != nil {
		logrus.WithError(err).Error("Failed to get flow for test chat")
		return h.errorResponse(c, 404, "Flow not found")
	}

	// Create a new execution for the test chat
	execution := &models.ChatbotExecution{
		FlowID:      flow.ID,
		CurrentStep: "start",
		Status:      "active",
		StartedAt:   time.Now(),
	}

	if err := h.chatService.CreateExecution(execution); err != nil {
		logrus.WithError(err).Error("Failed to create test chat execution")
		return h.errorResponse(c, 500, "Failed to start test chat")
	}

	return h.successResponse(c, map[string]interface{}{
		"executionId": execution.ID,
		"flow":        flow,
		"message":     "Test chat started successfully",
	})
}

// SendTestMessage sends a message in the test chat
func (h *Handlers) SendTestMessage(c *fiber.Ctx) error {
	executionID := c.Params("executionId")
	if executionID == "" {
		return h.errorResponse(c, 400, "Execution ID is required")
	}

	var request struct {
		Message string `json:"message"`
	}

	if err := c.BodyParser(&request); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	if request.Message == "" {
		return h.errorResponse(c, 400, "Message is required")
	}

	// Get the execution
	execution, err := h.chatService.GetExecution(executionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get execution")
		return h.errorResponse(c, 404, "Execution not found")
	}

	// Process the message
	response, err := h.processTestChatMessage(execution, request.Message)
	if err != nil {
		logrus.WithError(err).Error("Failed to process test chat message")
		return h.errorResponse(c, 500, "Failed to process message")
	}

	return h.successResponse(c, map[string]interface{}{
		"response": response,
		"execution": execution,
	})
}

// GetTestChatHistory returns the chat history for a test session
func (h *Handlers) GetTestChatHistory(c *fiber.Ctx) error {
	executionID := c.Params("executionId")
	if executionID == "" {
		return h.errorResponse(c, 400, "Execution ID is required")
	}

	// Get the execution with its history
	execution, err := h.chatService.GetExecution(executionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get execution")
		return h.errorResponse(c, 404, "Execution not found")
	}

	// Get chat history
	history, err := h.chatService.GetChatHistory(executionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get chat history")
		return h.errorResponse(c, 500, "Failed to retrieve chat history")
	}

	return h.successResponse(c, map[string]interface{}{
		"execution": execution,
		"history":   history,
	})
}

// ResetTestChat resets a test chat session
func (h *Handlers) ResetTestChat(c *fiber.Ctx) error {
	executionID := c.Params("executionId")
	if executionID == "" {
		return h.errorResponse(c, 400, "Execution ID is required")
	}

	// Complete the current execution
	err := h.chatService.CompleteExecution(executionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to complete execution")
		return h.errorResponse(c, 500, "Failed to reset test chat")
	}

	return h.successMessageResponse(c, "Test chat reset successfully", nil)
}