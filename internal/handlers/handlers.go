package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"
	"nodepath-chat/internal/services"
	"nodepath-chat/internal/whatsapp"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// parseMySQLURI parses a MySQL URI and returns database configuration
func parseMySQLURI(uri string) (*DatabaseConfig, error) {
	// Remove mysql:// prefix
	if !strings.HasPrefix(uri, "mysql://") {
		return nil, fmt.Errorf("invalid MySQL URI format: missing mysql:// prefix")
	}
	uri = strings.TrimPrefix(uri, "mysql://")

	// Split by @ to separate credentials from host/database
	parts := strings.Split(uri, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid MySQL URI format: missing @ separator")
	}

	// Parse credentials (user:password)
	credentials := parts[0]
	credParts := strings.Split(credentials, ":")
	if len(credParts) != 2 {
		return nil, fmt.Errorf("invalid MySQL URI format: invalid credentials")
	}
	user := credParts[0]
	password := credParts[1]

	// Parse host:port/database
	hostDatabase := parts[1]

	// Split by / to separate host:port from database
	hostDbParts := strings.Split(hostDatabase, "/")
	if len(hostDbParts) != 2 {
		return nil, fmt.Errorf("invalid MySQL URI format: missing database")
	}

	hostPort := hostDbParts[0]
	database := hostDbParts[1]

	// Parse host and port
	hostPortParts := strings.Split(hostPort, ":")
	if len(hostPortParts) != 2 {
		return nil, fmt.Errorf("invalid MySQL URI format: missing port")
	}

	host := hostPortParts[0]
	portStr := hostPortParts[1]

	// Convert port to integer
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}

	return &DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}, nil
}

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
	authHandlers          *AuthHandlers
	wasapBotHandlers      *WasapBotHandlers
	profileHandlers       *ProfileHandlers
	billingHandlers       *BillingHandlers
	appDataHandlers       *AppDataHandlers // Optimized app data handlers
	executionProcessRepo  repository.ExecutionProcessRepository
	db                    *sql.DB // Add database field
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
	cfg *config.Config,
) *Handlers {
	// Initialize repositories
	aiRepo := repository.NewAIWhatsappRepository(db)
	deviceRepo := repository.NewDeviceSettingsRepository(db)
	wasapBotRepo := repository.NewWasapBotRepository(db)
	executionProcessRepo := repository.NewExecutionProcessRepository(db)

	// Initialize media detection service
	mediaDetectionService := services.NewMediaDetectionService()

	// Initialize AI WhatsApp service
	aiWhatsappService := services.NewAIWhatsappService(aiRepo, deviceRepo, flowService, mediaDetectionService, cfg)

	// Initialize AI WhatsApp handlers
	aiWhatsappHandlers := NewAIWhatsappHandlers(aiWhatsappService, aiRepo, deviceRepo)

	// Initialize WasapBot handlers
	wasapBotHandlers := NewWasapBotHandlers(wasapBotRepo)

	// Initialize authentication handlers
	authHandlers := NewAuthHandlers(db)

	// Initialize profile handlers
	var profileHandlers *ProfileHandlers
	if db != nil {
		profileHandlers = NewProfileHandlers(db)
	}

	// Initialize billing handlers
	orderRepo := repository.NewOrderRepository(db)
	billplzService := services.NewBillplzService()
	billingHandlers := NewBillingHandlers(orderRepo, billplzService, db)

	// Initialize optimized app data handlers
	appDataHandlers := NewAppDataHandlers(db)

	// Create main handlers instance
	mainHandlers := &Handlers{
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
		authHandlers:          authHandlers,
		wasapBotHandlers:      wasapBotHandlers,
		profileHandlers:       profileHandlers,
		billingHandlers:       billingHandlers,
		appDataHandlers:       appDataHandlers, // Add optimized app data handlers
		executionProcessRepo:  executionProcessRepo,
		db:                    db, // Store the database
	}

	// Set the reference to main handlers in AI WhatsApp handlers for flow routing
	aiWhatsappHandlers.SetMainHandlers(mainHandlers)

	return mainHandlers
}

// SetupRoutes sets up all API routes
func (h *Handlers) SetupRoutes(api fiber.Router) {
	// Flow routes - protected with device requirement
	flows := api.Group("/flows")
	flows.Use(h.authHandlers.AuthMiddleware())
	flows.Use(h.authHandlers.DeviceRequiredMiddleware())
	flows.Get("/", h.GetFlows)
	flows.Post("/", h.CreateFlow)
	flows.Get("/:id", h.GetFlow)
	flows.Put("/:id", h.UpdateFlow)
	flows.Delete("/:id", h.DeleteFlow)

	// Test chat routes removed

	// Execution routes - protected with device requirement
	executions := api.Group("/executions")
	executions.Use(h.authHandlers.AuthMiddleware())
	executions.Use(h.authHandlers.DeviceRequiredMiddleware())
	executions.Get("/", h.GetExecutions)
	executions.Get("/:id", h.GetExecution)
	executions.Put("/:id/complete", h.CompleteExecution)
	executions.Delete("/:id", h.DeleteExecution)

	// WhatsApp routes - simplified for webhook-based system
	whatsapp := api.Group("/whatsapp")
	whatsapp.Post("/send", h.SendWhatsAppMessage)

	// Queue management routes - protected with device requirement
	queue := api.Group("/queue")
	queue.Use(h.authHandlers.AuthMiddleware())
	queue.Use(h.authHandlers.DeviceRequiredMiddleware())
	queue.Get("/stats", h.GetQueueStats)
	queue.Delete("/failed", h.ClearFailedQueue)

	// AI routes
	ai := api.Group("/ai")
	ai.Post("/validate-key", h.ValidateAPIKey)
	ai.Get("/models", h.GetSupportedModels)

	// Analytics routes - protected with device requirement
	analytics := api.Group("/analytics")
	analytics.Use(h.authHandlers.AuthMiddleware())
	analytics.Use(h.authHandlers.DeviceRequiredMiddleware())
	analytics.Get("/overview", h.GetAnalyticsOverview)
	analytics.Get("/flows/:id/stats", h.GetFlowStats)

	// Dashboard routes - protected with authentication
	dashboard := api.Group("/dashboard")
	dashboard.Use(h.authHandlers.AuthMiddleware())
	dashboard.Get("/chart-data", h.GetDashboardChartData)

	// Health check routes for system monitoring
	health := api.Group("/health")
	health.Get("/", h.HandleHealthCheck)
	health.Get("/live", h.HandleLivenessProbe)
	health.Get("/ready", h.HandleReadinessProbe)
	health.Get("/components/:component", h.HandleComponentHealth)
	health.Get("/metrics", h.HandleHealthMetrics)
	health.Delete("/cache", h.HandleClearHealthCache)

	// Config routes
	config := api.Group("/config")
	config.Get("/database", h.GetDatabaseConfig)

	// Device settings routes (protected with authentication middleware)
	deviceSettings := api.Group("/device-settings")
	deviceSettings.Use(h.authHandlers.AuthMiddleware())
	deviceSettings.Get("/", h.GetDeviceSettings)
	deviceSettings.Get("/device-ids", h.GetDeviceIDs)
	deviceSettings.Post("/", h.CreateDeviceSettings)
	// Device status route - must be before /:id to avoid conflicts
	deviceSettings.Get("/:id/status", h.GetDeviceStatus)
	deviceSettings.Get("/:id/waha-status", h.GetWahaDeviceStatus)
	deviceSettings.Get("/:id", h.GetDeviceSettingsById)
	deviceSettings.Put("/:id", h.UpdateDeviceSettings)
	deviceSettings.Delete("/:id", h.DeleteDeviceSettings)
	// Device generation routes
	deviceSettings.Post("/generate-whacenter", h.GenerateWhacenterDevice)
	deviceSettings.Post("/generate-wablas", h.GenerateWablasDevice)
	deviceSettings.Post("/generate-waha", h.GenerateWahaDevice)

	// AI WhatsApp routes - delegate to AIWhatsappHandlers (must be registered before generic webhook routes)
	aiWhatsapp := api.Group("/ai-whatsapp")
	h.aiWhatsappHandlers.SetupAIWhatsappRoutes(aiWhatsapp)

	// WasapBot routes
	wasapBot := api.Group("/wasapbot")
	wasapBot.Use(h.authHandlers.AuthMiddleware())
	wasapBot.Use(h.authHandlers.DeviceRequiredMiddleware())
	wasapBot.Get("/data", h.wasapBotHandlers.GetWasapBotData)
	wasapBot.Get("/stats", h.wasapBotHandlers.GetWasapBotStats)
	wasapBot.Delete("/data/:id", h.wasapBotHandlers.DeleteWasapBotRecord)

	// Stage Values routes (protected with authentication)
	stageValues := api.Group("/stage-values")
	stageValues.Use(h.authHandlers.AuthMiddleware())
	stageValues.Get("/", h.GetAllStageValues)
	stageValues.Post("/", h.CreateStageValue)
	stageValues.Put("/:id", h.UpdateStageValue)
	stageValues.Delete("/:id", h.DeleteStageValue)

	// Authentication routes
	h.authHandlers.SetupAuthRoutes(api)

	// Profile routes (protected with authentication)
	if h.profileHandlers != nil {
		profile := api.Group("/profile")
		profile.Use(h.authHandlers.AuthMiddleware())
		profile.Get("/", h.profileHandlers.GetProfile)
		profile.Put("/", h.profileHandlers.UpdateProfile)
		profile.Get("/status", h.profileHandlers.GetUserStatus)
	}

	// Optimized app data routes (protected with authentication)
	if h.appDataHandlers != nil {
		h.appDataHandlers.SetupAppDataRoutes(api.Group("/", h.authHandlers.AuthMiddleware()))
	}

	// Billing routes
	billing := api.Group("/billing")
	billing.Post("/callback", h.billingHandlers.BillplzCallback) // Billplz callback endpoint (public)
	// Protected billing routes
	billing.Use(h.authHandlers.AuthMiddleware())
	billing.Post("/create-order", h.billingHandlers.CreateOrder) // Create order (protected)
	billing.Get("/orders/:id", h.billingHandlers.GetOrder)       // Get specific order
	billing.Get("/orders", h.billingHandlers.GetOrders)          // Get user's orders
	billing.Get("/all-orders", h.billingHandlers.GetAllOrders)   // Admin: Get all orders

	// Webhook routes for receiving messages from providers
	webhook := api.Group("/webhook")
	webhook.Post("/:id_device/:instance", h.HandleWebhook)
}

// SetupTemplateRoutes configures template serving routes
func (h *Handlers) SetupTemplateRoutes(app *fiber.App) {
	h.authHandlers.SetupTemplateRoutes(app)
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

// GetFlows returns flows filtered by authenticated user's devices
func (h *Handlers) GetFlows(c *fiber.Ctx) error {
	// Get user ID from authentication context (string UUID)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return h.errorResponse(c, 401, "Authentication required")
	}

	// Get flows filtered by user's devices using string UUID
	flows, err := h.flowService.GetFlowsByUserDevicesString(userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get flows for user devices")
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
	logrus.Info("Health check endpoint called")

	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).Error("Panic in health check handler")
		}
	}()

	ctx := context.Background()

	// Check if health service is nil
	if h.healthService == nil {
		logrus.Error("Health service is nil")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Health service not initialized",
		})
	}

	health := h.healthService.GetSystemHealth(ctx)

	status := fiber.StatusOK
	if !h.healthService.IsHealthy(ctx) {
		status = fiber.StatusServiceUnavailable
	}

	logrus.WithField("health_status", health.Status).Info("Health check completed")
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

// GetDatabaseConfig returns database configuration for frontend
func (h *Handlers) GetDatabaseConfig(c *fiber.Ctx) error {
	// Get MYSQL_URI from Railway environment variables
	mysqlURI := os.Getenv("MYSQL_URI")
	logrus.WithField("mysql_uri", mysqlURI).Info("Database config endpoint called")

	if mysqlURI == "" {
		logrus.Error("MYSQL_URI environment variable not set")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "MYSQL_URI environment variable not set",
		})
	}

	// Parse the MySQL URI
	config, err := parseMySQLURI(mysqlURI)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse MYSQL_URI")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse MYSQL_URI: " + err.Error(),
		})
	}

	logrus.WithField("config", config).Info("Database config parsed successfully")
	return c.JSON(config)
}
