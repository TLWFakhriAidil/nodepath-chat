package handlers

import (
	"context"
	"database/sql"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"
	"nodepath-chat/internal/services"
	"nodepath-chat/internal/whatsapp"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	flowService           *services.FlowService
	aiService             *services.AIService
	queueService          *services.QueueService
	whatsappService       *whatsapp.Service
	deviceSettingsService *services.DeviceSettingsService
	websocketService      *services.WebSocketService
	mediaService          *services.MediaService
	mediaDetectionService *services.MediaDetectionService
	healthService         *services.HealthService
	aiWhatsappHandlers    *AIWhatsappHandlers
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	flowService *services.FlowService,
	aiService *services.AIService,
	queueService *services.QueueService,
	whatsappService *whatsapp.Service,
	deviceSettingsService *services.DeviceSettingsService,
	websocketService *services.WebSocketService,
	mediaService *services.MediaService,
	healthService *services.HealthService,
	db *sql.DB,
) *Handlers {
	// Initialize repositories
	aiRepo := repository.NewAIWhatsappRepository(db)
	deviceRepo := repository.NewDeviceSettingsRepository(db)
	
	// Initialize media detection service
	mediaDetectionService := services.NewMediaDetectionService()
	
	// Initialize AI WhatsApp service
	aiWhatsappService := services.NewAIWhatsappService(aiRepo, deviceRepo, flowService, mediaDetectionService)
	
	// Initialize AI WhatsApp handlers
	aiWhatsappHandlers := NewAIWhatsappHandlers(aiWhatsappService, aiRepo, deviceRepo)
	
	return &Handlers{
		flowService:           flowService,
		aiService:             aiService,
		queueService:          queueService,
		whatsappService:       whatsappService,
		deviceSettingsService: deviceSettingsService,
		websocketService:      websocketService,
		mediaService:          mediaService,
		mediaDetectionService: mediaDetectionService,
		healthService:         healthService,
		aiWhatsappHandlers:    aiWhatsappHandlers,
	}
}

// SetupRoutes sets up all API routes
func (h *Handlers) SetupRoutes(api fiber.Router) {
	// Flow routes
	flows := api.Group("/flows")
	flows.Get("/", h.GetFlows)
	flows.Post("/", h.CreateFlow)
	flows.Get("/:id", h.GetFlow)
	flows.Put("/:id", h.UpdateFlow)
	flows.Delete("/:id", h.DeleteFlow)

	// Test chat routes removed

	// Execution routes
	executions := api.Group("/executions")
	executions.Get("/", h.GetExecutions)
	executions.Get("/:id", h.GetExecution)
	executions.Put("/:id/complete", h.CompleteExecution)
	executions.Delete("/:id", h.DeleteExecution)

	// WhatsApp routes - simplified for webhook-based system
	whatsapp := api.Group("/whatsapp")
	whatsapp.Post("/send", h.SendWhatsAppMessage)

	// Queue management routes
	queue := api.Group("/queue")
	queue.Get("/stats", h.GetQueueStats)
	queue.Delete("/failed", h.ClearFailedQueue)

	// AI routes
	ai := api.Group("/ai")
	ai.Post("/validate-key", h.ValidateAPIKey)
	ai.Get("/models", h.GetSupportedModels)

	// Analytics routes
	analytics := api.Group("/analytics")
	analytics.Get("/overview", h.GetAnalyticsOverview)
	analytics.Get("/flows/:id/stats", h.GetFlowStats)

	// Health check routes for system monitoring
	health := api.Group("/health")
	health.Get("/", h.HandleHealthCheck)
	health.Get("/live", h.HandleLivenessProbe)
	health.Get("/ready", h.HandleReadinessProbe)
	health.Get("/components/:component", h.HandleComponentHealth)
	health.Get("/metrics", h.HandleHealthMetrics)
	health.Delete("/cache", h.HandleClearHealthCache)

	// Device settings routes
	deviceSettings := api.Group("/device-settings")
	deviceSettings.Get("/", h.GetDeviceSettings)
	deviceSettings.Get("/device-ids", h.GetDeviceIDs)
	deviceSettings.Post("/", h.CreateDeviceSettings)
	// Device status route - must be before /:id to avoid conflicts
	deviceSettings.Get("/:id/status", h.GetDeviceStatus)
	deviceSettings.Get("/:id", h.GetDeviceSettingsById)
	deviceSettings.Put("/:id", h.UpdateDeviceSettings)
	deviceSettings.Delete("/:id", h.DeleteDeviceSettings)
	// Device generation routes
	deviceSettings.Post("/generate-whacenter", h.GenerateWhacenterDevice)
	deviceSettings.Post("/generate-wablas", h.GenerateWablasDevice)

	// Webhook routes for receiving messages from providers
	webhook := api.Group("/webhook")
	webhook.Post("/:id_device/:instance", h.HandleWebhook)

	// AI WhatsApp routes - delegate to AIWhatsappHandlers
	h.aiWhatsappHandlers.SetupAIWhatsappRoutes(api)
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

	return h.successMessageResponse(c, "Flow created successfully", flow)
}

// GetFlow returns a specific flow
func (h *Handlers) GetFlow(c *fiber.Ctx) error {
	flowID := c.Params("id")
	if flowID == "" {
		return h.errorResponse(c, 400, "Flow ID is required")
	}

	flow, err := h.flowService.GetFlow(flowID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get flow")
		return h.errorResponse(c, 500, "Failed to retrieve flow")
	}

	if flow == nil {
		return h.errorResponse(c, 404, "Flow not found")
	}

	return h.successResponse(c, flow)
}

// UpdateFlow updates an existing flow
func (h *Handlers) UpdateFlow(c *fiber.Ctx) error {
	flowID := c.Params("id")
	if flowID == "" {
		return h.errorResponse(c, 400, "Flow ID is required")
	}

	var flow models.ChatbotFlow
	if err := c.BodyParser(&flow); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	flow.ID = flowID
	if err := h.flowService.UpdateFlow(&flow); err != nil {
		logrus.WithError(err).Error("Failed to update flow")
		return h.errorResponse(c, 500, "Failed to update flow")
	}

	return h.successMessageResponse(c, "Flow updated successfully", flow)
}

// DeleteFlow deletes a flow
func (h *Handlers) DeleteFlow(c *fiber.Ctx) error {
	flowID := c.Params("id")
	if flowID == "" {
		return h.errorResponse(c, 400, "Flow ID is required")
	}

	if err := h.flowService.DeleteFlow(flowID); err != nil {
		logrus.WithError(err).Error("Failed to delete flow")
		return h.errorResponse(c, 500, "Failed to delete flow")
	}

	return h.successMessageResponse(c, "Flow deleted successfully", nil)
}

// Health Check handlers

// HandleHealthCheck returns overall system health status
func (h *Handlers) HandleHealthCheck(c *fiber.Ctx) error {
	ctx := context.Background()
	health := h.healthService.GetSystemHealth(ctx)

	status := fiber.StatusOK
	if !h.healthService.IsHealthy(ctx) {
		status = fiber.StatusServiceUnavailable
	}

	return c.Status(status).JSON(health)
}

// HandleLivenessProbe returns simple liveness status for Kubernetes
func (h *Handlers) HandleLivenessProbe(c *fiber.Ctx) error {
	ctx := context.Background()
	isAlive := h.healthService.IsHealthy(ctx)
	
	if !isAlive {
		return c.Status(503).JSON(fiber.Map{"status": "unhealthy"})
	}
	
	return c.JSON(fiber.Map{"status": "healthy"})
}

// HandleReadinessProbe returns readiness probe for Kubernetes
func (h *Handlers) HandleReadinessProbe(c *fiber.Ctx) error {
	ctx := context.Background()
	isReady := h.healthService.IsHealthy(ctx)
	
	if !isReady {
		return c.Status(503).JSON(fiber.Map{"status": "unhealthy"})
	}
	
	return c.JSON(fiber.Map{"status": "healthy"})
}

// HandleComponentHealth returns health status for a specific component
func (h *Handlers) HandleComponentHealth(c *fiber.Ctx) error {
	component := c.Params("component")
	if component == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Component name is required",
		})
	}

	ctx := context.Background()
	health := h.healthService.GetComponentHealth(ctx, component)
	if health == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Component not found",
		})
	}

	status := fiber.StatusOK
	if health.Status != "healthy" {
		status = fiber.StatusServiceUnavailable
	}

	return c.Status(status).JSON(health)
}

// HandleHealthMetrics returns health metrics for monitoring systems
func (h *Handlers) HandleHealthMetrics(c *fiber.Ctx) error {
	ctx := context.Background()
	health := h.healthService.GetSystemHealth(ctx)
	
	// Create metrics from health data
	metrics := fiber.Map{
		"status":     health.Status,
		"timestamp":  health.Timestamp,
		"uptime":     health.Uptime.Seconds(),
		"version":    health.Version,
		"components": health.Components,
	}
	
	return c.JSON(metrics)
}

// HandleClearHealthCache clears the health check cache
func (h *Handlers) HandleClearHealthCache(c *fiber.Ctx) error {
	h.healthService.ClearCache()
	return c.JSON(fiber.Map{
		"message": "Health check cache cleared successfully",
	})
}