package handlers

import (
	"strconv"
	"strings"

	"nodepath-chat/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Execution handlers

// GetExecutions returns all executions
func (h *Handlers) GetExecutions(c *fiber.Ctx) error {
	flowReference := c.Query("flow_reference")

	if flowReference != "" {
		// Use AI WhatsApp repository to get executions
		// Get userID from context
		userID, ok := c.Locals("user_id").(string)
		if !ok {
			return h.errorResponse(c, 401, "Authentication required")
		}
		userIDInt := 0
		if userID != "" {
			if id, err := strconv.Atoi(userID); err == nil {
				userIDInt = id
			}
		}
		executions, _, err := h.aiWhatsappHandlers.AIRepo.GetAllAIWhatsappData(100, 0, "", []string{}, flowReference, "", userIDInt)
		if err != nil {
			logrus.WithError(err).Error("Failed to get executions by flow")
			return h.errorResponse(c, 500, "Failed to retrieve executions")
		}
		return h.successResponse(c, executions)
	}

	// Return empty array for now
	return h.successResponse(c, []interface{}{})
}

// GetExecution returns a specific execution
func (h *Handlers) GetExecution(c *fiber.Ctx) error {
	executionID := c.Params("id")
	if executionID == "" {
		return h.errorResponse(c, 400, "Execution ID is required")
	}

	// Use AI WhatsApp service to get execution by prospect number
	execution, err := h.aiWhatsappHandlers.AIRepo.GetAIWhatsappByProspectNum(executionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get execution")
		return h.errorResponse(c, 500, "Failed to retrieve execution")
	}

	if execution == nil {
		return h.errorResponse(c, 404, "Execution not found")
	}

	return h.successResponse(c, execution)
}

// CompleteExecution marks an execution as completed
func (h *Handlers) CompleteExecution(c *fiber.Ctx) error {
	executionID := c.Params("id")
	if executionID == "" {
		return h.errorResponse(c, 400, "Execution ID is required")
	}

	// Use AI WhatsApp service to complete execution
	err := h.aiWhatsappHandlers.AIWhatsappService.CompleteFlowExecution(executionID, "")
	if err != nil {
		logrus.WithError(err).Error("Failed to complete execution")
		return h.errorResponse(c, 500, "Failed to complete execution")
	}

	return h.successMessageResponse(c, "Execution completed successfully", nil)
}

// DeleteExecution deletes an execution
func (h *Handlers) DeleteExecution(c *fiber.Ctx) error {
	executionID := c.Params("id")
	if executionID == "" {
		return h.errorResponse(c, 400, "Execution ID is required")
	}

	// Update execution status to failed/deleted
	err := h.aiWhatsappHandlers.AIWhatsappService.UpdateFlowExecution(executionID, "", "", nil, "failed")
	if err != nil {
		logrus.WithError(err).Error("Failed to delete execution")
		return h.errorResponse(c, 500, "Failed to delete execution")
	}

	return h.successMessageResponse(c, "Execution deleted successfully", nil)
}

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
	if req.MediaURL != "" {
		// For media messages, we need a device ID - using empty string as fallback
		err = h.whatsappService.SendMediaMessage("", req.PhoneNumber, req.MediaURL)
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
	// Get user ID from authentication context (integer version set by AuthMiddleware)
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return h.errorResponse(c, 401, "Authentication required")
	}
	
	// Analytics data filtered by user's devices
	overview := map[string]interface{}{
		"total_flows":      0,
		"active_executions": 0,
		"total_messages":   0,
		"success_rate":     0.0,
		"avg_response_time": 0.0,
	}

	// Get actual flow count filtered by user's devices
	flows, err := h.flowService.GetFlowsByUserDevices(userID)
	if err == nil {
		overview["total_flows"] = len(flows)
	}

	return h.successResponse(c, overview)
}

// GetFlowStats returns statistics for a specific flow
func (h *Handlers) GetFlowStats(c *fiber.Ctx) error {
	flowReference := c.Params("id")
	if flowReference == "" {
		return h.errorResponse(c, 400, "Flow reference is required")
	}

	// Get userID from context (integer version set by AuthMiddleware)
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return h.errorResponse(c, 401, "Authentication required")
	}
	
	// Verify that the flow belongs to user's devices
	userFlows, err := h.flowService.GetFlowsByUserDevices(userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user flows")
		return h.errorResponse(c, 500, "Failed to verify flow ownership")
	}
	
	// Check if the requested flow belongs to the user
	flowExists := false
	for _, flow := range userFlows {
		if flow.ID == flowReference {
			flowExists = true
			break
		}
	}
	
	if !flowExists {
		return h.errorResponse(c, 403, "Access denied: Flow not found or not owned by user")
	}
	
	// Get executions for the flow using AI WhatsApp repository
	executions, _, err := h.aiWhatsappHandlers.AIRepo.GetAllAIWhatsappData(100, 0, "", []string{}, flowReference, "", userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get flow executions")
		return h.errorResponse(c, 500, "Failed to get flow statistics")
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"total_executions": len(executions),
		"active_executions": 0,
		"completed_executions": 0,
		"failed_executions": 0,
	}

	for _, execution := range executions {
		// Use Human field to determine status: 0 = AI active, 1 = human takeover
		if execution.Human == 0 {
			stats["active_executions"] = stats["active_executions"].(int) + 1
		} else {
			stats["completed_executions"] = stats["completed_executions"].(int) + 1
		}
	}

	return h.successResponse(c, stats)
}

// Test chat processing functions removed

// Test chat node processing functions removed

// buildResponseFromParts constructs the final response string from AI response parts
func (h *Handlers) buildResponseFromParts(parts []models.AIResponsePart) string {
	var response strings.Builder
	
	for i, part := range parts {
		switch part.Type {
		case "text":
			if i > 0 {
				response.WriteString("\n")
			}
			response.WriteString(part.Content)
		case "image":
			if i > 0 {
				response.WriteString("\n")
			}
			response.WriteString("[Image: " + part.URL + "]")
		}
	}
	
	return response.String()
}

// Remaining test chat processing functions removed