package handlers

import (
	"context"
	"time"

	"nodepath-chat/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// HealthHandlers handles health check endpoints
type HealthHandlers struct {
	healthService *services.HealthService
}

// NewHealthHandlers creates new health handlers
func NewHealthHandlers(healthService *services.HealthService) *HealthHandlers {
	return &HealthHandlers{
		healthService: healthService,
	}
}

// HandleHealthCheck handles the main health check endpoint
// GET /health
func (h *HealthHandlers) HandleHealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	health := h.healthService.GetSystemHealth(ctx)

	// Set appropriate HTTP status based on health
	var statusCode int
	switch health.Status {
	case services.HealthStatusHealthy:
		statusCode = 200
	case services.HealthStatusDegraded:
		statusCode = 200 // Still operational but with issues
	case services.HealthStatusUnhealthy:
		statusCode = 503 // Service unavailable
	default:
		statusCode = 500
	}

	return c.Status(statusCode).JSON(health)
}

// HandleLivenessProbe handles Kubernetes liveness probe
// GET /health/live
func (h *HealthHandlers) HandleLivenessProbe(c *fiber.Ctx) error {
	// Liveness probe should only check if the application is running
	// It should not check external dependencies
	return c.JSON(fiber.Map{
		"status":    "alive",
		"timestamp": time.Now(),
		"message":   "Application is running",
	})
}

// HandleReadinessProbe handles Kubernetes readiness probe
// GET /health/ready
func (h *HealthHandlers) HandleReadinessProbe(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Readiness probe should check if the application is ready to serve traffic
	// This includes checking critical dependencies like database
	dbHealth := h.healthService.GetComponentHealth(ctx, "database")

	if dbHealth.Status == services.HealthStatusUnhealthy {
		return c.Status(503).JSON(fiber.Map{
			"status":    "not_ready",
			"timestamp": time.Now(),
			"message":   "Database is not available",
			"details":   dbHealth,
		})
	}

	return c.JSON(fiber.Map{
		"status":    "ready",
		"timestamp": time.Now(),
		"message":   "Application is ready to serve traffic",
	})
}

// HandleComponentHealth handles individual component health checks
// GET /health/component/:name
func (h *HealthHandlers) HandleComponentHealth(c *fiber.Ctx) error {
	componentName := c.Params("name")
	if componentName == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Component name is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	componentHealth := h.healthService.GetComponentHealth(ctx, componentName)

	// Set appropriate HTTP status based on component health
	var statusCode int
	switch componentHealth.Status {
	case services.HealthStatusHealthy:
		statusCode = 200
	case services.HealthStatusDegraded:
		statusCode = 200
	case services.HealthStatusUnhealthy:
		statusCode = 503
	default:
		statusCode = 404 // Component not found
	}

	return c.Status(statusCode).JSON(componentHealth)
}

// HandleHealthMetrics handles health metrics endpoint for monitoring systems
// GET /health/metrics
func (h *HealthHandlers) HandleHealthMetrics(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	health := h.healthService.GetSystemHealth(ctx)

	// Create metrics in a format suitable for monitoring systems
	metrics := fiber.Map{
		"system_health_status":  health.Status,
		"system_uptime_seconds": health.Uptime.Seconds(),
		"system_version":        health.Version,
		"timestamp":             health.Timestamp.Unix(),
		"components":            make(map[string]interface{}),
	}

	// Add component metrics
	for name, component := range health.Components {
		metrics["components"].(map[string]interface{})[name] = fiber.Map{
			"status":        component.Status,
			"response_time": component.ResponseTime.Milliseconds(),
			"last_checked":  component.LastChecked.Unix(),
			"message":       component.Message,
		}
	}

	return c.JSON(metrics)
}

// HandleClearHealthCache handles clearing the health check cache
// POST /health/cache/clear
func (h *HealthHandlers) HandleClearHealthCache(c *fiber.Ctx) error {
	h.healthService.ClearCache()

	logrus.Info("Health check cache cleared")

	return c.JSON(fiber.Map{
		"message":   "Health check cache cleared successfully",
		"timestamp": time.Now(),
	})
}

// RegisterHealthRoutes registers all health check routes
func (h *HealthHandlers) RegisterHealthRoutes(app *fiber.App) {
	health := app.Group("/health")

	// Main health check endpoint
	health.Get("/", h.HandleHealthCheck)

	// Kubernetes probes
	health.Get("/live", h.HandleLivenessProbe)
	health.Get("/ready", h.HandleReadinessProbe)

	// Component-specific health checks
	health.Get("/component/:name", h.HandleComponentHealth)

	// Metrics endpoint
	health.Get("/metrics", h.HandleHealthMetrics)

	// Cache management
	health.Post("/cache/clear", h.HandleClearHealthCache)

	logrus.Info("Health check routes registered")
}
