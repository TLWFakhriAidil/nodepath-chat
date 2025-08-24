package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Execution handlers have been removed as they depended on the deprecated ChatService

// WhatsApp handlers









type SendWhatsAppMessageRequest struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
	MediaURL    string `json:"media_url,omitempty"`
	MediaType   string `json:"media_type,omitempty"`
}

// SendWhatsAppMessage sends a WhatsApp message
func (h *Handlers) SendWhatsAppMessage(c *fiber.Ctx) error {
	var req SendWhatsAppMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	if req.PhoneNumber == "" || req.Message == "" {
		return h.errorResponse(c, 400, "Phone number and message are required")
	}

	if h.whatsappService == nil {
		return h.errorResponse(c, 500, "WhatsApp service not available")
	}

	var err error
	if req.MediaURL != "" && req.MediaType != "" {
		err = h.whatsappService.SendMediaMessage(req.PhoneNumber, req.Message, req.MediaURL, req.MediaType)
	} else {
		err = h.whatsappService.SendMessage(req.PhoneNumber, req.Message)
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to send WhatsApp message")
		return h.errorResponse(c, 500, "Failed to send message")
	}

	return h.successMessageResponse(c, "Message sent successfully", nil)
}

// Queue handlers

// GetQueueStats returns queue statistics
func (h *Handlers) GetQueueStats(c *fiber.Ctx) error {
	if h.queueService == nil {
		return h.successResponse(c, map[string]int64{
			"outbound": 0,
			"failed":   0,
			"delayed":  0,
		})
	}

	stats, err := h.queueService.GetQueueStats()
	if err != nil {
		logrus.WithError(err).Error("Failed to get queue stats")
		return h.errorResponse(c, 500, "Failed to get queue statistics")
	}

	return h.successResponse(c, stats)
}

// ClearFailedQueue clears the failed message queue
func (h *Handlers) ClearFailedQueue(c *fiber.Ctx) error {
	if h.queueService == nil {
		return h.successMessageResponse(c, "Queue service not available", nil)
	}

	err := h.queueService.ClearFailedMessages()
	if err != nil {
		logrus.WithError(err).Error("Failed to clear failed queue")
		return h.errorResponse(c, 500, "Failed to clear failed queue")
	}

	return h.successMessageResponse(c, "Failed queue cleared successfully", nil)
}

// AI handlers

type ValidateAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}

// ValidateAPIKey validates an OpenRouter API key
func (h *Handlers) ValidateAPIKey(c *fiber.Ctx) error {
	var req ValidateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	if req.APIKey == "" {
		return h.errorResponse(c, 400, "API key is required")
	}

	err := h.aiService.ValidateAPIKey(req.APIKey)
	if err != nil {
		logrus.WithError(err).Error("API key validation failed")
		return h.errorResponse(c, 400, "Invalid API key")
	}

	return h.successMessageResponse(c, "API key is valid", nil)
}

// GetSupportedModels returns supported AI models
func (h *Handlers) GetSupportedModels(c *fiber.Ctx) error {
	models := h.aiService.GetSupportedModels()
	return h.successResponse(c, models)
}

// Analytics handlers

// GetAnalyticsOverview returns analytics overview
func (h *Handlers) GetAnalyticsOverview(c *fiber.Ctx) error {
	// Mock analytics data - in a real implementation, you'd calculate these from the database
	overview := map[string]interface{}{
		"total_flows":      0,
		"active_executions": 0,
		"total_messages":   0,
		"success_rate":     0.0,
		"avg_response_time": 0.0,
	}

	// Get actual flow count
	flows, err := h.flowService.GetAllFlows()
	if err == nil {
		overview["total_flows"] = len(flows)
	}

	return h.successResponse(c, overview)
}

// GetFlowStats has been removed as it depended on the deprecated ChatService
// Flow statistics are no longer available due to removal of execution tracking

// Test chat processing functions have been removed as they depended on the deprecated ChatService

// All remaining test processing functions have been removed as they depended on the deprecated ChatService