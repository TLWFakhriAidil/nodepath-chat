package handlers

import (
	"encoding/json"
	"fmt"
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
		executions, err := h.chatService.GetExecutionsByFlow(flowReference)
		if err != nil {
			logrus.WithError(err).Error("Failed to get executions by flow")
			return h.errorResponse(c, 500, "Failed to retrieve executions")
		}
		return h.successResponse(c, executions)
	}

	// For now, return empty array - in a full implementation, you'd have a GetAllExecutions method
	return h.successResponse(c, []interface{}{})
}

// GetExecution returns a specific execution
func (h *Handlers) GetExecution(c *fiber.Ctx) error {
	executionID := c.Params("id")
	if executionID == "" {
		return h.errorResponse(c, 400, "Execution ID is required")
	}

	execution, err := h.chatService.GetExecution(executionID)
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

	err := h.chatService.CompleteExecution(executionID)
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

	// For now, just mark as failed - in a full implementation, you'd have a DeleteExecution method
	err := h.chatService.FailExecution(executionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete execution")
		return h.errorResponse(c, 500, "Failed to delete execution")
	}

	return h.successMessageResponse(c, "Execution deleted successfully", nil)
}

// WhatsApp handlers

// GetWhatsAppStatus returns WhatsApp connection status
func (h *Handlers) GetWhatsAppStatus(c *fiber.Ctx) error {
	if h.whatsappService == nil {
		return h.successResponse(c, map[string]interface{}{
			"connected": false,
			"status":    "not_initialized",
		})
	}

	return h.successResponse(c, map[string]interface{}{
		"connected": h.whatsappService.IsConnected(),
		"status":    "initialized",
	})
}

// ConnectWhatsApp connects to WhatsApp
func (h *Handlers) ConnectWhatsApp(c *fiber.Ctx) error {
	if h.whatsappService == nil {
		return h.errorResponse(c, 500, "WhatsApp service not available")
	}

	err := h.whatsappService.Connect()
	if err != nil {
		logrus.WithError(err).Error("Failed to connect to WhatsApp")
		return h.errorResponse(c, 500, "Failed to connect to WhatsApp")
	}

	return h.successMessageResponse(c, "Connected to WhatsApp successfully", nil)
}

// DisconnectWhatsApp disconnects from WhatsApp
func (h *Handlers) DisconnectWhatsApp(c *fiber.Ctx) error {
	if h.whatsappService == nil {
		return h.errorResponse(c, 500, "WhatsApp service not available")
	}

	h.whatsappService.Disconnect()
	return h.successMessageResponse(c, "Disconnected from WhatsApp successfully", nil)
}

// GetWhatsAppQR returns QR code for WhatsApp pairing
func (h *Handlers) GetWhatsAppQR(c *fiber.Ctx) error {
	if h.whatsappService == nil {
		return h.errorResponse(c, 500, "WhatsApp service not available")
	}

	qrCode, err := h.whatsappService.GetQRCode()
	if err != nil {
		logrus.WithError(err).Error("Failed to get QR code")
		return h.errorResponse(c, 500, "Failed to get QR code")
	}

	return h.successResponse(c, map[string]string{
		"qr_code": qrCode,
	})
}

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

// GetFlowStats returns statistics for a specific flow
func (h *Handlers) GetFlowStats(c *fiber.Ctx) error {
	flowReference := c.Params("id")
	if flowReference == "" {
		return h.errorResponse(c, 400, "Flow reference is required")
	}

	// Get executions for the flow
	executions, err := h.chatService.GetExecutionsByFlow(flowReference)
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
		switch execution.Status {
		case models.ExecutionStatusActive:
			stats["active_executions"] = stats["active_executions"].(int) + 1
		case models.ExecutionStatusCompleted:
			stats["completed_executions"] = stats["completed_executions"].(int) + 1
		case models.ExecutionStatusFailed:
			stats["failed_executions"] = stats["failed_executions"].(int) + 1
		}
	}

	return h.successResponse(c, stats)
}

// processTestChatMessage processes a message in test chat mode
func (h *Handlers) processTestChatMessage(execution *models.ChatbotExecution, userInput string) (string, error) {
	// Get the flow
	flow, err := h.flowService.GetFlow(execution.FlowReference)
	if err != nil {
		return "", fmt.Errorf("failed to get flow: %w", err)
	}

	if flow == nil {
		return "", fmt.Errorf("flow not found")
	}

	// Get current node
	currentNode, err := h.flowService.FindNodeByID(flow, execution.CurrentNode)
	if err != nil {
		// If no current node, start from the beginning
		currentNode, err = h.flowService.GetStartNode(flow)
		if err != nil {
			return "", fmt.Errorf("failed to get start node: %w", err)
		}
		execution.CurrentNode = currentNode.ID
	}

	// Process based on node type
	switch currentNode.Type {
	case models.NodeTypeAIPrompt:
		return h.processTestAIPromptNode(flow, execution, currentNode, userInput)
	case models.NodeTypeAdvancedAIPrompt:
		return h.processTestAdvancedAIPromptNode(flow, execution, currentNode, userInput)
	case models.NodeTypeManual:
		return h.processTestManualNode(flow, execution, currentNode, userInput)
	default:
		return h.processTestDefaultNode(flow, execution, currentNode, userInput)
	}
}

// processTestAIPromptNode processes an AI prompt node in test chat
func (h *Handlers) processTestAIPromptNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get AI configuration from node data
	var systemPrompt, instance, apiProvider string

	// Check node data for configuration
	if sp, ok := node.Data["system_prompt"].(string); ok {
		systemPrompt = sp
	}
	if inst, ok := node.Data["instance"].(string); ok {
		instance = inst
	}
	if ap, ok := node.Data["apiprovider"].(string); ok {
		apiProvider = ap
	}

	// Use global settings as fallback
	if apiProvider == "" {
		apiProvider = flow.Niche
	}

	// Check if we have complete AI configuration
	if systemPrompt == "" || instance == "" || apiProvider == "" {
		// Fallback to manual response
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}

	// Get conversation history
	history, err := h.chatService.GetConversationHistory(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get conversation history")
		history = []models.ConversationMessage{}
	}

	// Get execution variables for prompt replacement
	variables, err := h.chatService.GetExecutionVariables(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in system prompt
	systemPrompt = h.flowService.ReplaceVariables(systemPrompt, variables)

	// Generate AI response
	response, err := h.aiService.GenerateResponse(systemPrompt, userInput, apiProvider, history)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate AI response")
		return "I'm sorry, I'm having trouble processing your request right now. Please try again later.", nil
	}

	// Move to next node
	nextNode, err := h.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		h.chatService.UpdateExecution(execution)
	} else {
		// End of flow
		h.chatService.CompleteExecution(execution.ID)
	}

	return response, nil
}

// processTestManualNode processes a manual node in test chat
func (h *Handlers) processTestManualNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, _ string) (string, error) {
	// Get manual response from node data
	response := "Thank you for your message."
	if msg, ok := node.Data["message"].(string); ok && msg != "" {
		response = msg
	}

	// Get execution variables for response replacement
	variables, err := h.chatService.GetExecutionVariables(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in response
	response = h.flowService.ReplaceVariables(response, variables)

	// Move to next node
	nextNode, err := h.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		h.chatService.UpdateExecution(execution)
	} else {
		// End of flow
		h.chatService.CompleteExecution(execution.ID)
	}

	return response, nil
}

// processTestAdvancedAIPromptNode processes an advanced AI prompt node in test chat
func (h *Handlers) processTestAdvancedAIPromptNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get AI configuration from node data
	var systemPrompt, instance, apiProvider, closingPrompt string

	// Check node data for configuration
	if sp, ok := node.Data["system_prompt"].(string); ok {
		systemPrompt = sp
	}
	if inst, ok := node.Data["instance"].(string); ok {
		instance = inst
	}
	if ap, ok := node.Data["apiprovider"].(string); ok {
		apiProvider = ap
	}
	if cp, ok := node.Data["closing_prompt"].(string); ok {
		closingPrompt = cp
	}

	// Use global settings as fallback
	if apiProvider == "" {
		apiProvider = flow.Niche
	}

	// Check if we have complete AI configuration
	if systemPrompt == "" || instance == "" || apiProvider == "" {
		// Fallback to manual response
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}

	// Get conversation history
	history, err := h.chatService.GetConversationHistory(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get conversation history")
		history = []models.ConversationMessage{}
	}

	// Get execution variables for prompt replacement
	variables, err := h.chatService.GetExecutionVariables(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in prompts
	systemPrompt = h.flowService.ReplaceVariables(systemPrompt, variables)
	if closingPrompt != "" {
		closingPrompt = h.flowService.ReplaceVariables(closingPrompt, variables)
	}

	// Generate AI response using advanced method
	aiResponse, err := h.aiService.GenerateAdvancedResponse(systemPrompt, userInput, apiProvider, history, closingPrompt)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate advanced AI response")
		return "I'm sorry, I'm having trouble processing your request right now. Please try again later.", nil
	}

	// Update execution stage if provided
	if aiResponse.Stage != "" {
		if execution.Variables == nil {
			execution.Variables = json.RawMessage("{}")
		}
		
		var vars map[string]interface{}
		if err := json.Unmarshal(execution.Variables, &vars); err != nil {
			vars = make(map[string]interface{})
		}
		
		vars["current_stage"] = aiResponse.Stage
		
		updatedVars, err := json.Marshal(vars)
		if err == nil {
			execution.Variables = updatedVars
		}
	}

	// Build response from parts
	response := h.buildResponseFromParts(aiResponse.Response)

	// Move to next node
	nextNode, err := h.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		h.chatService.UpdateExecution(execution)
	} else {
		// End of flow
		h.chatService.CompleteExecution(execution.ID)
	}

	return response, nil
}

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

// processTestDefaultNode processes other node types in test chat
func (h *Handlers) processTestDefaultNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// For other node types, just move to the next node
	nextNode, err := h.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		h.chatService.UpdateExecution(execution)
		return h.processTestChatMessage(execution, userInput)
	}

	// End of flow
	h.chatService.CompleteExecution(execution.ID)
	return "Thank you for using our service!", nil
}