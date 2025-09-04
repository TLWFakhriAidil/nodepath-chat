package whatsapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"

	"github.com/sirupsen/logrus"
)

// QueuedMessage represents a message in the processing queue
type QueuedMessage struct {
	DeviceID  string
	Message   interface{} // Generic message data from webhook
	Retries   int
	Timestamp time.Time
}

// Service handles WhatsApp operations via webhook processing
// Simplified version focusing on message processing without whatsmeow client management
type Service struct {
	cfg *config.Config

	// Service dependencies
	queueService          *services.QueueService
	flowService           *services.FlowService
	aiService             *services.AIService
	aiWhatsappService     services.AIWhatsappService
	websocketService      *services.WebSocketService
	deviceSettingsService *services.DeviceSettingsService
	providerService       *services.ProviderService
	mediaDetectionService *services.MediaDetectionService

	// Message processing queue for performance
	messageQueue chan *WebhookMessage
	processingWG sync.WaitGroup
}

// WebhookMessage represents an incoming message from webhook
type WebhookMessage struct {
	PhoneNumber string
	Content     string
	DeviceID    string
	Provider    string
	Timestamp   time.Time
	Retries     int
}

// NewService creates a new simplified WhatsApp service for webhook-based system
func NewService(cfg *config.Config, queueService *services.QueueService, flowService *services.FlowService, aiService *services.AIService, aiWhatsappService services.AIWhatsappService, websocketService *services.WebSocketService, deviceSettingsService *services.DeviceSettingsService, providerService *services.ProviderService, mediaDetectionService *services.MediaDetectionService) (*Service, error) {
	service := &Service{
		cfg:                   cfg,
		queueService:          queueService,
		flowService:           flowService,
		aiService:             aiService,
		aiWhatsappService:     aiWhatsappService,
		websocketService:      websocketService,
		deviceSettingsService: deviceSettingsService,
		providerService:       providerService,
		mediaDetectionService: mediaDetectionService,
		messageQueue:          make(chan *WebhookMessage, 1000), // Buffered queue for performance
	}

	// Start message processing workers for high performance
	for i := 0; i < 10; i++ { // 10 worker goroutines for handling 3000+ devices
		go service.messageProcessor()
	}

	logrus.Info("ðŸš€ WHATSAPP: Simplified webhook-based service initialized")
	return service, nil
}

// messageProcessor processes incoming webhook messages from the queue
func (s *Service) messageProcessor() {
	for msg := range s.messageQueue {
		s.processingWG.Add(1)
		go func(webhookMsg *WebhookMessage) {
			defer s.processingWG.Done()
			if err := s.processWebhookMessageInternal(webhookMsg); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"device_id":    webhookMsg.DeviceID,
					"phone_number": webhookMsg.PhoneNumber,
					"retries":      webhookMsg.Retries,
				}).Error("Failed to process webhook message")

				// Retry logic for failed messages
				if webhookMsg.Retries < 3 {
					webhookMsg.Retries++
					time.Sleep(time.Second * time.Duration(webhookMsg.Retries))
					s.messageQueue <- webhookMsg
				}
			}
		}(msg)
	}
}

// processWebhookMessageInternal processes a single webhook message
func (s *Service) processWebhookMessageInternal(msg *WebhookMessage) error {
	return s.processIncomingMessage(msg.PhoneNumber, msg.Content, msg.DeviceID)
}

// SetServices updates service dependencies
func (s *Service) SetServices(flowService *services.FlowService, aiService *services.AIService) {
	s.flowService = flowService
	s.aiService = aiService
}

// ProcessIncomingMessageFromWebhook processes incoming messages from webhook providers
// This is the main entry point for webhook-based message processing
func (s *Service) ProcessIncomingMessageFromWebhook(phoneNumber, content, deviceID, provider string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"provider":     provider,
		"content":      content,
	}).Info("ðŸ“¨ WEBHOOK: Processing incoming message")

	// Add to processing queue for high performance
	webhookMsg := &WebhookMessage{
		PhoneNumber: phoneNumber,
		Content:     content,
		DeviceID:    deviceID,
		Provider:    provider,
		Timestamp:   time.Now(),
		Retries:     0,
	}

	select {
	case s.messageQueue <- webhookMsg:
		return nil
	default:
		return fmt.Errorf("message queue is full, dropping message")
	}
}

// SendMessage sends a message using the default device (for backward compatibility)
func (s *Service) SendMessage(phoneNumber, message string) error {
	// For now, just log the message sending attempt
	// Message sending would be implemented through the provider service
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"message":      message,
	}).Info("ðŸ“¤ MESSAGE: Sending message (not implemented)")
	return nil
}

// SendMessageFromDevice sends a message from a specific device through the appropriate provider
func (s *Service) SendMessageFromDevice(deviceID, phoneNumber, message string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"message":      message,
	}).Info("ðŸ“¤ MESSAGE: Sending message from device")

	// Get device settings by device_id
	deviceSettings, err := s.deviceSettingsService.GetByIDDevice(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device settings for %s: %w", deviceID, err)
	}

	// Send message through provider service
	err = s.providerService.SendMessage(deviceSettings, phoneNumber, message)
	if err != nil {
		return fmt.Errorf("failed to send message through provider: %w", err)
	}

	return nil
}

// SendMediaMessage sends a media message through the appropriate provider
func (s *Service) SendMediaMessage(deviceID, phoneNumber, mediaURL string) error {
	// Console log for tracing media URL extraction
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"media_url":    mediaURL,
		"media_url_length": len(mediaURL),
		"media_url_preview": func() string {
			if len(mediaURL) > 100 {
				return mediaURL[:100] + "..."
			}
			return mediaURL
		}(),
	}).Info("ðŸ“¤ MEDIA: Sending media message - URL EXTRACTED FOR TRACING")

	// Get device settings by device_id
	deviceSettings, err := s.deviceSettingsService.GetByIDDevice(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device settings for %s: %w", deviceID, err)
	}

	// Send media message through provider service
	err = s.providerService.SendMediaMessage(deviceSettings, phoneNumber, mediaURL)
	if err != nil {
		return fmt.Errorf("failed to send media message through provider: %w", err)
	}

	return nil
}

// processIncomingMessage processes incoming messages and handles flow/AI logic using ai_whatsapp_nodepath
func (s *Service) processIncomingMessage(phoneNumber, content string, deviceID string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"content":      content,
	}).Info("ðŸ” FLOW: Checking for active execution in ai_whatsapp_nodepath")

	// Check for personal commands (%, #, cmd)
	if strings.HasPrefix(content, "%") || strings.HasPrefix(content, "#") || strings.HasPrefix(content, "cmd") {
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"command":   content,
		}).Info("ðŸ”§ COMMAND: Personal command detected")
		return s.handlePersonalCommand(phoneNumber, content, deviceID)
	}

	// Get or create active execution from ai_whatsapp_nodepath
	aiExecution, err := s.aiWhatsappService.GetActiveFlowExecution(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).Error("âŒ FLOW: Failed to get active execution from ai_whatsapp_nodepath")
		return err
	}

	if aiExecution == nil {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("ðŸ†• FLOW: No active execution found, checking for default flow")

		// Get default flow for device
		defaultFlow, err := s.flowService.GetDefaultFlowForDevice(deviceID)
		if err != nil {
			logrus.WithError(err).Error("âŒ FLOW: Failed to get default flow for device")
			return err
		}

		if defaultFlow == nil {
			logrus.WithFields(logrus.Fields{
				"phone_number": phoneNumber,
				"device_id":    deviceID,
			}).Info("âš ï¸ FLOW: No default flow found for device, falling back to AI conversation")
			
			// Fallback to AI conversation when no flow is configured
			return s.processAIConversation(phoneNumber, content, deviceID)
		}

		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
			"flow_id":      defaultFlow.ID,
			"flow_name":    defaultFlow.Name,
		}).Info("ðŸš€ FLOW: Starting new execution with default flow in ai_whatsapp_nodepath")

		// Start new execution with default flow in ai_whatsapp_nodepath
		// StartFlowExecution call removed ok {
		systemPrompt = sp
	} else if sp, ok := node.Data["systemPrompt"].(string); ok {
		systemPrompt = sp
	}
	
	if inst, ok := node.Data["instance"].(string); ok {
		instance = inst
	}
	if ap, ok := node.Data["apiprovider"].(string); ok {
		apiProvider = ap
	} else if ap, ok := node.Data["apiProvider"].(string); ok {
		apiProvider = ap
	}

	// ðŸ” DEBUG TRACE: Log extracted node data for debugging
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"extracted_system_prompt_length": len(systemPrompt),
		"extracted_system_prompt_preview": func() string {
			if len(systemPrompt) > 200 {
				return systemPrompt[:200] + "..."
			}
			return systemPrompt
		}(),
		"extracted_instance": instance,
		"extracted_api_provider": apiProvider,
		"node_data_keys": func() []string {
			keys := make([]string, 0, len(node.Data))
			for k := range node.Data {
				keys = append(keys, k)
			}
			return keys
		}(),
	}).Info("ðŸ” AI_PROMPT_DEBUG: Extracted node configuration")

	// Get device settings for fallback values
	deviceSettings, err := s.deviceSettingsService.GetByIDDevice(execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get device settings for AI prompt")
	}

	// Use device settings as fallback
	if instance == "" && deviceSettings != nil {
		if deviceSettings.Instance.Valid {
			instance = deviceSettings.Instance.String
		}
	}
	if apiProvider == "" && deviceSettings != nil {
		apiProvider = deviceSettings.Provider
	}

	// ðŸ” DEBUG TRACE: Log device settings and final values
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"device_settings_found": deviceSettings != nil,
		"device_instance_valid": func() bool {
			if deviceSettings != nil {
				return deviceSettings.Instance.Valid
			}
			return false
		}(),
		"device_instance_value": func() string {
			if deviceSettings != nil && deviceSettings.Instance.Valid {
				return deviceSettings.Instance.String
			}
			return "null"
		}(),
		"device_provider": func() string {
			if deviceSettings != nil {
				return deviceSettings.Provider
			}
			return "null"
		}(),
		"final_instance": instance,
		"final_api_provider": apiProvider,
		"final_system_prompt_length": len(systemPrompt),
	}).Info("ðŸ” AI_PROMPT_DEBUG: Device settings and final configuration")
	// Use global settings as final fallback
	if apiProvider == "" {
		apiProvider = flow.Niche
	}

	logrus.WithFields(logrus.Fields{
		"system_prompt_length": len(systemPrompt),
		"instance": instance,
		"api_provider": apiProvider,
	}).Info("ðŸ¤– AI_PROMPT: Configuration loaded")

	// Check if we have complete AI configuration
	if systemPrompt == "" {
		logrus.Error("ðŸ¤– AI_PROMPT: No system prompt configured")
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}
	if instance == "" {
		logrus.Error("ðŸ¤– AI_PROMPT: No instance configured")
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}
	if apiProvider == "" {
		logrus.Error("ðŸ¤– AI_PROMPT: No API provider configured")
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}

	// Get execution variables for prompt replacement
	// GetFlowExecutionVariables removed`nvariables := make(map[string]interface{}))
	}

	// Replace variables in system prompt
	originalSystemPrompt := systemPrompt
	systemPrompt = s.flowService.ReplaceVariables(systemPrompt, variables)

	// ðŸ” DEBUG TRACE: Log variable replacement
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"variables_count": len(variables),
		"original_prompt_length": len(originalSystemPrompt),
		"final_prompt_length": len(systemPrompt),
		"prompt_changed": originalSystemPrompt != systemPrompt,
	}).Info("ðŸ” AI_PROMPT_DEBUG: Variable replacement completed")

	// ðŸ” DEBUG TRACE: Log final AI service call parameters
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"system_prompt_length": len(systemPrompt),
		"system_prompt_preview": func() string {
			if len(systemPrompt) > 300 {
				return systemPrompt[:300] + "..."
			}
			return systemPrompt
		}(),
		"user_input": userInput,
		"instance": instance,
		"api_provider": apiProvider,
		"device_id": execution.IDDevice,
		"prospect_num": execution.ProspectNum,
	}).Info("ðŸ” AI_PROMPT_DEBUG: Final parameters for AI service call")

	// Get actual API key from device settings
	var actualAPIKey string
	if deviceSettings != nil && deviceSettings.APIKey.Valid {
		actualAPIKey = deviceSettings.APIKey.String
	}

	// Generate AI response
	logrus.WithFields(logrus.Fields{
		"id_device": execution.IDDevice,
		"api_provider": apiProvider,
		"api_key_provided": actualAPIKey != "",
		"user_input": userInput,
	}).Info("ðŸ¤– AI_PROMPT: Generating AI response")
	
	response, err := s.aiService.GenerateResponse(systemPrompt, userInput, actualAPIKey, execution.IDDevice, []models.ConversationMessage{})
	if err != nil {
		logrus.WithError(err).Error("ðŸ¤– AI_PROMPT: Failed to generate AI response")
		return "I'm sorry, I'm having trouble processing your request right now. Please try again later.", nil
	}
	
	logrus.WithFields(logrus.Fields{
		"response_length": len(response),
		"node_id": node.ID,
		"ai_response": response,
	}).Info("ðŸ¤– AI_PROMPT: AI response generated successfully")

	// Check if next node exists and advance to it
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
				"ai_response": response,
			}).Info("ðŸ¤– AI_PROMPT: AI response generated, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return response, err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return response, err
			}
			
			return response, nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after AI prompt node")
			return response, err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after AI prompt")
			return response, err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return response + "\n" + nextResponse, nil
		}
	} else {
		// End of flow
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	return response, nil
}

// processAdvancedAIPromptNode processes an advanced AI prompt node with JSON response parsing
func (s *Service) processAdvancedAIPromptNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
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

	// ðŸ” DEBUG TRACE: Log extracted node configuration for advanced AI prompt
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"node_type": "advanced_ai_prompt",
		"system_prompt_length": len(systemPrompt),
		"system_prompt_preview": func() string {
			if len(systemPrompt) > 100 {
				return systemPrompt[:100] + "..."
			}
			return systemPrompt
		}(),
		"instance_from_node": instance,
		"api_provider_from_node": apiProvider,
		"user_input": userInput,
		"node_data_keys": func() []string {
			keys := make([]string, 0, len(node.Data))
			for k := range node.Data {
				keys = append(keys, k)
			}
			return keys
		}(),
	}).Info("ðŸ” ADVANCED_AI_PROMPT_DEBUG: Extracted node configuration")

	// Get device settings for fallback values
	deviceSettings, err := s.deviceSettingsService.GetByIDDevice(execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get device settings for advanced AI prompt")
	}

	// Use device settings as fallback
	if instance == "" && deviceSettings != nil {
		if deviceSettings.Instance.Valid {
			instance = deviceSettings.Instance.String
		}
	}
	if apiProvider == "" && deviceSettings != nil {
		apiProvider = deviceSettings.Provider
	}

	// ðŸ” DEBUG TRACE: Log device settings and final values for advanced AI prompt
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"device_settings_found": deviceSettings != nil,
		"device_instance_valid": func() bool {
			if deviceSettings != nil {
				return deviceSettings.Instance.Valid
			}
			return false
		}(),
		"device_instance_value": func() string {
			if deviceSettings != nil && deviceSettings.Instance.Valid {
				return deviceSettings.Instance.String
			}
			return "null"
		}(),
		"device_provider": func() string {
			if deviceSettings != nil {
				return deviceSettings.Provider
			}
			return "null"
		}(),
		"final_instance": instance,
		"final_api_provider": apiProvider,
		"final_system_prompt_length": len(systemPrompt),
	}).Info("ðŸ” ADVANCED_AI_PROMPT_DEBUG: Device settings and final configuration")

	// Use global settings as fallback
	if apiProvider == "" {
		apiProvider = flow.Niche
	}

	logrus.WithFields(logrus.Fields{
		"system_prompt_length": len(systemPrompt),
		"instance": instance,
		"api_provider": apiProvider,
	}).Info("ðŸ¤– ADVANCED_AI_PROMPT: Configuration loaded")

	// Check if we have complete AI configuration
	if systemPrompt == "" {
		logrus.Error("ðŸ¤– ADVANCED_AI_PROMPT: No system prompt configured")
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}
	if instance == "" {
		logrus.Error("ðŸ¤– ADVANCED_AI_PROMPT: No instance configured")
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}
	if apiProvider == "" {
		logrus.Error("ðŸ¤– ADVANCED_AI_PROMPT: No API provider configured")
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}

	// Get execution variables for prompt replacement
	// GetFlowExecutionVariables removed`nvariables := make(map[string]interface{}))
	}

	// Replace variables in system prompt
	originalSystemPrompt := systemPrompt
	systemPrompt = s.flowService.ReplaceVariables(systemPrompt, variables)

	// ðŸ” DEBUG TRACE: Log variable replacement for advanced AI prompt
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"variables_count": len(variables),
		"original_prompt_length": len(originalSystemPrompt),
		"final_prompt_length": len(systemPrompt),
		"prompt_changed": originalSystemPrompt != systemPrompt,
	}).Info("ðŸ” ADVANCED_AI_PROMPT_DEBUG: Variable replacement completed")

	// ðŸ” DEBUG TRACE: Log final AI service call parameters for advanced AI prompt
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"system_prompt_length": len(systemPrompt),
		"system_prompt_preview": func() string {
			if len(systemPrompt) > 300 {
				return systemPrompt[:300] + "..."
			}
			return systemPrompt
		}(),
		"user_input": userInput,
		"instance": instance,
		"api_provider": apiProvider,
		"device_id": execution.IDDevice,
		"prospect_num": execution.ProspectNum,
	}).Info("ðŸ” ADVANCED_AI_PROMPT_DEBUG: Final parameters for AI service call")

	// Get actual API key from device settings
	var actualAPIKey string
	if deviceSettings != nil && deviceSettings.APIKey.Valid {
		actualAPIKey = deviceSettings.APIKey.String
	}

	// Generate AI response with advanced JSON parsing
	logrus.WithFields(logrus.Fields{
		"id_device": execution.IDDevice,
		"api_provider": apiProvider,
		"api_key_provided": actualAPIKey != "",
		"user_input": userInput,
	}).Info("ðŸ¤– ADVANCED_AI_PROMPT: Generating AI response")

	rawResponse, err := s.aiService.GenerateResponse(systemPrompt, userInput, actualAPIKey, execution.IDDevice, []models.ConversationMessage{})
	if err != nil {
		logrus.WithError(err).Error("Failed to generate advanced AI response")
		return "I'm sorry, I'm having trouble processing your request right now. Please try again later.", nil
	}

	// Parse the AI response JSON to extract media URLs and handle multiple response items
	var response string
	parsedResponse, err := s.aiWhatsappService.ParseAIResponse(rawResponse)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"raw_response": rawResponse,
			"node_id": node.ID,
		}).Warn("ðŸ§  ADVANCED_AI: Failed to parse JSON response, treating as plain text")
		// Fallback to plain text if JSON parsing fails
		response = rawResponse
	} else {
		// Successfully parsed JSON response - handle multiple response items
		logrus.WithFields(logrus.Fields{
			"stage": parsedResponse.Stage,
			"response_count": len(parsedResponse.Response),
			"node_id": node.ID,
		}).Info("ðŸ§  ADVANCED_AI: Successfully parsed JSON response with multiple items")
		
		// Process each response item and send them individually
		for i, item := range parsedResponse.Response {
			logrus.WithFields(logrus.Fields{
				"item_index": i,
				"item_type": item.Type,
				"content_length": len(item.Content),
			}).Info("ðŸ§  ADVANCED_AI: Processing response item")
			
			switch item.Type {
			case "text":
				err := s.SendMessageFromDevice(execution.IDDevice, execution.ProspectNum, item.Content)
				if err != nil {
					logrus.WithError(err).Error("Failed to send text message from advanced AI")
				}
			case "image", "audio", "video":
				err := s.SendMediaMessage(execution.IDDevice, execution.ProspectNum, item.Content)
				if err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"media_type": item.Type,
						"media_url": item.Content,
					}).Error("Failed to send media message from advanced AI")
				}
			default:
				logrus.WithField("type", item.Type).Warn("Unknown response type in advanced AI")
			}
			
			// Add delay between messages for better user experience
			if i < len(parsedResponse.Response)-1 {
				time.Sleep(2 * time.Second)
			}
		}
		
		// Update conversation stage if provided
		if parsedResponse.Stage != "" {
			err = s.aiWhatsappService.UpdateConversationStage(execution.ProspectNum, parsedResponse.Stage)
			if err != nil {
				logrus.WithError(err).Error("Failed to update conversation stage")
			}
		}
		
		// For JSON responses, we've already sent all messages, so return empty string
		// to prevent duplicate sending in the main flow processing logic
		response = ""
	}

	// Check if next node exists and advance to it
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸ§  ADVANCED_AI: Advanced AI response generated, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return response, err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return response, err
			}
			
			return response, nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after advanced AI prompt node")
			return response, err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after advanced AI prompt")
			return response, err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return response + "\n" + nextResponse, nil
		}
	} else {
		// End of flow
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	return response, nil
}

// processManualNode processes a manual node (human intervention required)
func (s *Service) processManualNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// For now, just return a message indicating manual intervention
	// Human mode would be implemented through a separate table or field
	logrus.WithFields(logrus.Fields{
		"prospect_id": execution.IDProspect,
		"node_id":     node.ID,
	}).Info("ðŸ‘¤ MANUAL: Manual intervention node triggered")

	// Get manual response message
	message := "Your message has been forwarded to our support team. We'll get back to you soon."
	if msg, ok := node.Data["message"].(string); ok {
		message = msg
	}

	// Check if next node exists and advance to it
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸ‘¤ MANUAL: Manual response sent, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return message, err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return message, err
			}
			
			return message, nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after manual node")
			return message, err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after manual")
			return message, err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return message + "\n" + nextResponse, nil
		}
	} else {
		// End of flow
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	return message, nil
}

// processMessageNode processes a simple message node
func (s *Service) processMessageNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// Get message from node data
	message := ""
	if msg, ok := node.Data["message"].(string); ok {
		message = msg
	}

	// Replace variables in message
	// GetFlowExecutionVariables removed`nvariables := make(map[string]interface{}))
	}
	message = s.flowService.ReplaceVariables(message, variables)

	// Check if next node exists and advance to it
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸ“¤ MESSAGE: Message sent, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return message, err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return message, err
			}
			
			return message, nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after message node")
			return message, err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after message")
			return message, err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return message + "\n" + nextResponse, nil
		}
	} else {
		// End of flow
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
		}).Info("ðŸ MESSAGE: End of flow reached, completing execution")
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	return message, nil
}

// processImageNode processes an image node
func (s *Service) processImageNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// Get image URL from node data
	imageURL := ""
	if url, ok := node.Data["imageUrl"].(string); ok {
		imageURL = url
	} else if url, ok := node.Data["image"].(string); ok {
		imageURL = url
	}

	// Console log for tracing image URL extraction
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"raw_image_url": imageURL,
		"node_data_keys": func() []string {
			keys := make([]string, 0, len(node.Data))
			for k := range node.Data {
				keys = append(keys, k)
			}
			return keys
		}(),
		"url_source": func() string {
			if _, ok := node.Data["imageUrl"]; ok {
				return "imageUrl"
			} else if _, ok := node.Data["image"]; ok {
				return "image"
			}
			return "none"
		}(),
	}).Info("ðŸ” IMAGE NODE: RAW URL EXTRACTED FOR TRACING")

	// Replace variables in image URL
	// GetFlowExecutionVariables removed`nvariables := make(map[string]interface{}))
	}
	imageURL = s.flowService.ReplaceVariables(imageURL, variables)

	// Console log for tracing processed image URL
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"processed_image_url": imageURL,
		"variables_count": len(variables),
	}).Info("ðŸ” IMAGE NODE: PROCESSED URL EXTRACTED FOR TRACING")

	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"image_url":    imageURL,
	}).Info("ðŸ–¼ï¸ IMAGE: Processing image node")

	// Check if next node exists and advance to it
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"execution_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸ–¼ï¸ IMAGE: Image processed, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return imageURL, err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return imageURL, err
			}
			
			// Return raw image URL for media detection service to process
			return imageURL, nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after image node")
			return imageURL, err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after image")
			return imageURL, err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return fmt.Sprintf("%s\n%s", imageURL, nextResponse), nil
		}
	} else {
		// End of flow
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
		}).Info("ðŸ IMAGE: End of flow reached, completing execution")
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	// Return raw image URL for media detection service to process
	return imageURL, nil
}

// processAudioNode processes an audio node
func (s *Service) processAudioNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// Get audio URL from node data
	audioURL := ""
	if url, ok := node.Data["audioUrl"].(string); ok {
		audioURL = url
	} else if url, ok := node.Data["audio"].(string); ok {
		audioURL = url
	} else if url, ok := node.Data["mediaUrl"].(string); ok {
		audioURL = url
	}

	// Console log for tracing audio URL extraction
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"raw_audio_url": audioURL,
		"node_data_keys": func() []string {
			keys := make([]string, 0, len(node.Data))
			for k := range node.Data {
				keys = append(keys, k)
			}
			return keys
		}(),
		"url_source": func() string {
			if _, ok := node.Data["audioUrl"]; ok {
				return "audioUrl"
			} else if _, ok := node.Data["audio"]; ok {
				return "audio"
			} else if _, ok := node.Data["mediaUrl"]; ok {
				return "mediaUrl"
			}
			return "none"
		}(),
	}).Info("ðŸ” AUDIO NODE: RAW URL EXTRACTED FOR TRACING")

	// Replace variables in audio URL
	// GetFlowExecutionVariables removed`nvariables := make(map[string]interface{}))
	}
	audioURL = s.flowService.ReplaceVariables(audioURL, variables)

	// Console log for tracing processed audio URL
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"processed_audio_url": audioURL,
		"variables_count": len(variables),
	}).Info("ðŸ” AUDIO NODE: PROCESSED URL EXTRACTED FOR TRACING")

	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"audio_url":    audioURL,
	}).Info("ðŸŽµ AUDIO: Processing audio node")

	// Check if next node exists and advance to it
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"execution_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸŽµ AUDIO: Audio processed, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return audioURL, err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return audioURL, err
			}
			
			// Return raw audio URL for media detection service to process
			return audioURL, nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after audio node")
			return audioURL, err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after audio")
			return audioURL, err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return fmt.Sprintf("%s\n%s", audioURL, nextResponse), nil
		}
	} else {
		// End of flow
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
		}).Info("ðŸ AUDIO: End of flow reached, completing execution")
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	// Return raw audio URL for media detection service to process
	return audioURL, nil
}

// processVideoNode processes a video node
func (s *Service) processVideoNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// Get video URL from node data
	videoURL := ""
	if url, ok := node.Data["videoUrl"].(string); ok {
		videoURL = url
	} else if url, ok := node.Data["video"].(string); ok {
		videoURL = url
	} else if url, ok := node.Data["mediaUrl"].(string); ok {
		videoURL = url
	}

	// Console log for tracing video URL extraction
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"raw_video_url": videoURL,
		"node_data_keys": func() []string {
			keys := make([]string, 0, len(node.Data))
			for k := range node.Data {
				keys = append(keys, k)
			}
			return keys
		}(),
		"url_source": func() string {
			if _, ok := node.Data["videoUrl"]; ok {
				return "videoUrl"
			} else if _, ok := node.Data["video"]; ok {
				return "video"
			} else if _, ok := node.Data["mediaUrl"]; ok {
				return "mediaUrl"
			}
			return "none"
		}(),
	}).Info("ðŸ” VIDEO NODE: RAW URL EXTRACTED FOR TRACING")

	// Replace variables in video URL
	// GetFlowExecutionVariables removed`nvariables := make(map[string]interface{}))
	}
	videoURL = s.flowService.ReplaceVariables(videoURL, variables)

	// Console log for tracing processed video URL
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"processed_video_url": videoURL,
		"variables_count": len(variables),
	}).Info("ðŸ” VIDEO NODE: PROCESSED URL EXTRACTED FOR TRACING")

	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"video_url":    videoURL,
	}).Info("ðŸŽ¬ VIDEO: Processing video node")

	// Check if next node exists and advance to it
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"execution_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸŽ¬ VIDEO: Video processed, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return videoURL, err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return videoURL, err
			}
			
			// Return raw video URL for media detection service to process
			return videoURL, nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after video node")
			return videoURL, err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after video")
			return videoURL, err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return fmt.Sprintf("%s\n%s", videoURL, nextResponse), nil
		}
	} else {
		// End of flow
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
		}).Info("ðŸ VIDEO: End of flow reached, completing execution")
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	// Return raw video URL for media detection service to process
	return videoURL, nil
}

// processDelayNode processes a delay node
func (s *Service) processDelayNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"flow_id":      flow.ID,
	}).Info("ðŸ• DELAY: Processing delay node")
	
	// Get delay time from node data (default to 5 seconds if not specified)
	delaySeconds := 5
	if delay, ok := node.Data["delay"].(float64); ok {
		delaySeconds = int(delay)
	} else if delay, ok := node.Data["delaySeconds"].(float64); ok {
		delaySeconds = int(delay)
	}
	
	logrus.WithFields(logrus.Fields{
		"execution_id":   execution.IDProspect,
		"delay_seconds":  delaySeconds,
		"phone_number":   execution.ProspectNum,
		"device_id":      execution.IDDevice,
	}).Info("ðŸ• DELAY: Scheduling delayed message")
	
	// Get next node to process after delay
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err != nil || nextNode == nil {
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
		}).Info("ðŸ• DELAY: No next node found, completing execution")
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
		return "", nil
	}
	
	// DO NOT update execution here - let ProcessFlowContinuation handle the transition
	// This ensures proper sequential flow processing
	logrus.WithFields(logrus.Fields{
		"execution_id":     execution.IDProspect,
		"current_node":     node.ID,
		"next_node":        nextNode.ID,
		"delay_seconds":    delaySeconds,
	}).Info("ðŸ• DELAY: Keeping execution at current node, will advance after delay")
	
	// Create delayed message for queue processing
	delayedMessage := &services.QueueMessage{
		ID:          fmt.Sprintf("delayed_%s_%s_%d", execution.IDProspect, nextNode.ID, time.Now().Unix()),
		DeviceID:    execution.IDDevice,
		PhoneNumber: execution.ProspectNum,
		Content:     userInput, // Pass the original user input
		MessageType: "flow_continuation",
		FlowID:      flow.ID,
		ExecutionID: fmt.Sprintf("%d", execution.IDProspect),
		NodeID:      nextNode.ID, // This is the node to process AFTER the delay
		Delay:       time.Duration(delaySeconds) * time.Second,
		CreatedAt:   time.Now(),
	}
	
	// Queue the delayed message
	err = s.queueService.EnqueueDelayedMessage(delayedMessage)
	if err != nil {
		logrus.WithError(err).Error("ðŸ• DELAY: Failed to queue delayed message")
		return "", fmt.Errorf("failed to queue delayed message: %w", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"execution_id":   execution.IDProspect,
		"message_id":     delayedMessage.ID,
		"delay_seconds":  delaySeconds,
		"next_node_id":   nextNode.ID,
	}).Info("ðŸ• DELAY: Message queued successfully for delayed processing")
	
	// Return empty string as no immediate response is needed
	// The delayed message will be processed later by the queue processor
	return "", nil
}

// processConditionNode processes a condition node
func (s *Service) processConditionNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// Evaluate condition based on user input and move to appropriate next node
	nextNode, err := s.flowService.EvaluateConditionNode(flow, node.ID, userInput)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸ”€ CONDITION: Condition evaluated, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return "", err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return "", err
			}
			
			return "", nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after condition node")
			return "", err
		}
		
		// Recursively process the next node if it's not a delay
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processStageNode processes a stage node
func (s *Service) processStageNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// For now, just log the stage transition
	// Stage tracking would be implemented through a separate field or table
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":     node.ID,
		"stage":       node.Data["stage"],
	}).Info("ðŸŽ¯ STAGE: Stage transition node processed")

	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸŽ¯ STAGE: Stage processed, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return "", err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return "", err
			}
			
			return "", nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after stage node")
			return "", err
		}
		
		// Recursively process the next node if it's not a delay
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processUserReplyNode processes a user reply node
// handleUserReplyResume handles user reply when execution is waiting and resumes flow
func (s *Service) handleUserReplyResume(execution *models.AIWhatsapp, userInput string) error {
	// Get the flow data
	flow, err := s.flowService.GetFlow(execution.FlowID.String)
	if err != nil {
		logrus.WithError(err).Error("âŒ USER_REPLY: Failed to get flow for resume")
		return err
	}
	
	if flow == nil {
		logrus.WithField("flow_id", execution.FlowID.String).Error("âŒ USER_REPLY: Flow not found for resume")
		return fmt.Errorf("flow not found for resume")
	}
	
	// Validate that we have a valid current node ID
	if !execution.CurrentNodeIDID.Valid || execution.CurrentNodeIDID.String == "" {
		logrus.Error("âŒ USER_REPLY: Invalid current node ID for resume")
		return fmt.Errorf("invalid current node ID for resume")
	}
	
	// Save user message to conversation history
	err = s.aiWhatsappService.SaveConversationHistory(execution.ProspectNum, execution.IDDevice, userInput, "", "")
	if err != nil {
		logrus.WithError(err).Error("âŒ USER_REPLY: Failed to save user message to conversation")
		return err
	}
	
	// Get the next node after the user_reply node
	nextNode, err := s.flowService.GetNextNode(flow, execution.CurrentNodeIDID.String)
	if err != nil {
		logrus.WithError(err).Error("âŒ USER_REPLY: Failed to get next node after user reply")
		return err
	}
	
	if nextNode == nil {
		logrus.WithField("current_node_id", execution.CurrentNodeIDID.String).Info("ðŸ USER_REPLY: No next node found, completing flow")
		
		// Clear waiting state and complete flow
		err = s.updateFlowTrackingFields(execution, execution.CurrentNodeIDID.String, execution.FlowID.String, false)
		if err != nil {
			logrus.WithError(err).Error("Failed to clear waiting state")
			return err
		}
		
		// Complete the flow execution
		return s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}
	
	logrus.WithFields(logrus.Fields{
		"execution_id": sql.NullString{String: "", Valid: false}.String,
		"current_node": execution.CurrentNodeIDID.String,
		"next_node":    nextNode.ID,
		"next_type":    nextNode.Type,
		"user_input":   userInput,
	}).Info("ðŸ”„ USER_REPLY: Resuming flow execution from next node")
	
	// Clear waiting state and update to next node
	err = s.updateFlowTrackingFields(execution, nextNode.ID, execution.FlowID.String, false)
	if err != nil {
		logrus.WithError(err).Error("Failed to update flow tracking for resume")
		return err
	}
	
	// Update the current node in execution for processing
	s.updateCurrentNode(execution, nextNode.ID)
	
	// Process the next node
	response, err := s.processFlowMessage(flow, execution, userInput)
	if err != nil {
		logrus.WithError(err).Error("âŒ USER_REPLY: Failed to process next node after user reply")
		return err
	}
	
	// Send response if there is one
	if response != "" {
		logrus.WithFields(logrus.Fields{
			"execution_id":    sql.NullString{String: "", Valid: false}.String,
			"response_length": len(response),
		}).Info("ðŸ“¤ USER_REPLY: Sending response after flow resume")
		
		// Send the response
		err = s.SendMessageFromDevice(execution.IDDevice, execution.ProspectNum, response)
		if err != nil {
			logrus.WithError(err).Error("âŒ USER_REPLY: Failed to send response after resume")
			return err
		}
		
		// Save bot response to conversation history
		err = s.aiWhatsappService.SaveConversationHistory(execution.ProspectNum, execution.IDDevice, "", response, "")
		if err != nil {
			logrus.WithError(err).Error("âŒ USER_REPLY: Failed to save bot response to conversation")
			return err
		}
	}
	
	logrus.WithField("execution_id", sql.NullString{String: "", Valid: false}.String).Info("âœ… USER_REPLY: Flow resumed successfully after user reply")
	return nil
}

// updateCurrentNode updates both new and legacy current node fields
func (s *Service) updateCurrentNode(execution *models.AIWhatsapp, nodeID string) {
	// Update new flow tracking field
	execution.CurrentNodeIDID.String = nodeID
	execution.CurrentNodeIDID.Valid = true
	
	// Update legacy field for backward compatibility
	execution.CurrentNodeID.String = nodeID
	execution.CurrentNodeID.Valid = true
}

// updateFlowTrackingFields updates the flow tracking fields for user reply handling
// Uses repository's UpdateFlowTrackingFields to preserve conversation history
func (s *Service) updateFlowTrackingFields(execution *models.AIWhatsapp, currentNodeID, flowID string, waitingForReply bool) error {
	// Determine last node ID
	lastNodeID := ""
	if execution.CurrentNodeIDID.Valid && execution.CurrentNodeIDID.String != "" {
		lastNodeID = execution.CurrentNodeIDID.String
	}
	
	// Set waiting_for_reply flag
	waitingForReplyValue := 0
	if waitingForReply {
		waitingForReplyValue = 1
	}
	
	// Get execution ID
	executionID := ""
	if sql.NullString{String: "", Valid: false}.Valid {
		executionID = sql.NullString{String: "", Valid: false}.String
	}
	
	// Update flow tracking fields directly in repository to preserve conversation history
	err := s.aiWhatsappService.GetRepository().UpdateFlowTrackingFields(
		execution.ProspectNum, execution.IDDevice,
		flowID, // flowID
		currentNodeID, // currentNodeID
		lastNodeID, // lastNodeID
		waitingForReplyValue, // waitingForReply
		"active", // executionStatus
		executionID, // executionID
	)
	if err != nil {
		return fmt.Errorf("failed to update flow tracking fields: %w", err)
	}
	
	// Update the execution model in memory for consistency
	execution.CurrentNodeIDID.String = currentNodeID
	execution.CurrentNodeIDID.Valid = true
	execution.FlowID.String = flowID
	execution.FlowID.Valid = true
	execution.LastNodeID.String = lastNodeID
	execution.LastNodeID.Valid = (lastNodeID != "")
	execution.WaitingForReply.Int32 = int32(waitingForReplyValue)
	execution.WaitingForReply.Valid = true
	
	logrus.WithFields(logrus.Fields{
		"prospect_id":       execution.IDProspect,
		"current_node_id":   currentNodeID,
		"flow_id":           flowID,
		"waiting_for_reply": waitingForReply,
		"last_node_id":      execution.LastNodeID.String,
	}).Info("âœ… FLOW_TRACKING: Updated flow tracking fields successfully")
	
	return nil
}

// processUserReplyNode processes a user reply node by setting waiting state
func (s *Service) processUserReplyNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"prospect_id": execution.IDProspect,
		"node_id":     node.ID,
		"user_input":  userInput,
	}).Info("ðŸ’¬ USER_REPLY: Processing user reply node - setting waiting state")

	// Set the flow to waiting for user reply state
	// Update the flow tracking fields to indicate we're waiting for user input
	err := s.updateFlowTrackingFields(execution, node.ID, flow.ID, true)
	if err != nil {
		logrus.WithError(err).Error("Failed to update flow tracking fields for waiting state")
		return "", err
	}

	logrus.WithFields(logrus.Fields{
		"prospect_id": execution.IDProspect,
		"node_id":     node.ID,
		"flow_id":     flow.ID,
	}).Info("âœ… USER_REPLY: Flow set to waiting for user reply state")

	// Return empty response as we're now waiting for user input
	return "", nil
}

// processWaitingReplyTimesNode processes a waiting reply times node
func (s *Service) processWaitingReplyTimesNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// Handle reply timing logic and move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("â±ï¸ WAITING_REPLY: Reply timing processed, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return "", err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return "", err
			}
			
			return "", nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after waiting reply times node")
			return "", err
		}
		
		// Recursively process the next node if it's not a delay
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processStartNode processes a start node
func (s *Service) processStartNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// Move to next node from start
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸš€ START: Start node processed, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return "", err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return "", err
			}
			
			return "", nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after start node")
			return "", err
		}
		
		// Recursively process the next node if it's not a delay
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processDefaultNode processes any unrecognized node type
func (s *Service) processDefaultNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// Default behavior - move to next node or end flow
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id": execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("ðŸ”§ DEFAULT: Default node processed, advancing to delay node")
			
			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return "", err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return "", err
			}
			
			return "", nil
		}
		
		// For non-delay nodes, continue processing immediately
		s.updateCurrentNode(execution, nextNode.ID)
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after default node")
			return "", err
		}
		
		// Recursively process the next node if it's not a delay
		return s.processFlowMessage(flow, execution, userInput)
	}
	s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	return "", nil
}

// StartQueueProcessor starts the queue processor for handling queued messages
func (s *Service) StartQueueProcessor() {
	logrus.Info("ðŸš€ QUEUE: Starting queue processor")

	// For now, just log that the queue processor would start
	// Queue processing would be implemented through the queue service
	logrus.Info("ðŸ“‹ QUEUE: Queue processor started (placeholder implementation)")
}

// processQueuedMessage processes a queued message from the queue service
func (s *Service) processQueuedMessage(message *services.QueueMessage) error {
	// For now, just log the queued message processing
	// Queue message processing would be implemented based on the actual QueueMessage structure
	logrus.WithFields(logrus.Fields{
		"message_id": message.ID,
		"content":    message.Content,
	}).Info("ðŸ“‹ QUEUE: Processing queued message (placeholder implementation)")
	return nil
}

// ProcessFlowContinuation processes flow continuation after delay
// This method is called by the queue service when a delayed message is ready
func (s *Service) ProcessFlowContinuation(executionID, flowID, nodeID, phoneNumber, deviceID, userInput string) error {
	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"flow_id":      flowID,
		"node_id":      nodeID,
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("ðŸ”„ FLOW: Processing flow continuation after delay")

	// First try to get active execution, then try any execution (including completed ones)
	// This handles cases where execution was completed but delayed messages are still pending
	execution, err := s.aiWhatsappService.GetActiveFlowExecution(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).Error("âŒ FLOW: Failed to get active execution for continuation")
		return fmt.Errorf("failed to get active execution: %w", err)
	}

	// If no active execution found, try to get any execution (including completed)
	if execution == nil {
		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("ðŸ”„ FLOW: No active execution found, checking for any existing execution")
		
		// Get any execution (regardless of status) to continue delayed processing
		execution, err = s.aiWhatsappService.GetFlowExecutionByProspectAndDevice(phoneNumber, deviceID)
		if err != nil {
			logrus.WithError(err).Error("âŒ FLOW: Failed to get any execution for continuation")
			return fmt.Errorf("failed to get any execution: %w", err)
		}
		
		if execution == nil {
			// Log as debug instead of warn to reduce noise - this is expected for cleaned up executions
			logrus.WithField("execution_id", executionID).Debug("âš ï¸ FLOW: No execution found for continuation (likely cleaned up)")
			return fmt.Errorf("execution not found: %s", executionID)
		}
		
		// Reactivate the execution for delayed processing
		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"previous_status": sql.NullString{String: "active", Valid: true}.String,
		}).Info("ðŸ”„ FLOW: Reactivating execution for delayed message processing")
		
		// Set execution status back to active for processing
		sql.NullString{String: "active", Valid: true}.String = "active"
		sql.NullString{String: "active", Valid: true}.Valid = true
	}

	// Get the flow
	flow, err := s.flowService.GetFlow(flowID)
	if err != nil {
		logrus.WithError(err).Error("âŒ FLOW: Failed to get flow for continuation")
		return fmt.Errorf("failed to get flow: %w", err)
	}

	if flow == nil {
		logrus.WithField("flow_id", flowID).Warn("âš ï¸ FLOW: Flow not found for continuation")
		return fmt.Errorf("flow not found: %s", flowID)
	}

	// Get the target node (the node to process after delay)
	targetNode, err := s.flowService.FindNodeByID(flow, nodeID)
	if err != nil {
		logrus.WithError(err).Error("âŒ FLOW: Failed to get target node for continuation")
		return fmt.Errorf("failed to get target node: %w", err)
	}

	if targetNode == nil {
		logrus.WithField("node_id", nodeID).Warn("âš ï¸ FLOW: Target node not found for continuation")
		return fmt.Errorf("target node not found: %s", nodeID)
	}

	// Update execution to the target node (advance from delay node to next node)
	logrus.WithFields(logrus.Fields{
		"execution_id":   executionID,
		"previous_node":  execution.CurrentNodeID.String,
		"target_node":    nodeID,
	}).Info("ðŸ”„ FLOW: Advancing execution to target node after delay")
	
	s.updateCurrentNode(execution, nodeID)
	err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeIDID.String, make(map[string]interface{}), "active")
	if err != nil {
		logrus.WithError(err).Error("âŒ FLOW: Failed to update execution to target node")
		return fmt.Errorf("failed to update execution: %w", err)
	}

	// Process the target node
	response, err := s.processFlowMessage(flow, execution, userInput)
	if err != nil {
		logrus.WithError(err).Error("âŒ FLOW: Failed to process flow continuation")
		return fmt.Errorf("failed to process flow: %w", err)
	}

	// Send response if available
	if response != "" {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
			"response":     response,
		}).Info("ðŸ“¤ FLOW: Sending delayed response to user")

		// Check if response contains media URLs using the new detection service
		if s.mediaDetectionService.HasMedia(response) {
			mediaInfo := s.mediaDetectionService.ExtractFirstMedia(response)
			if mediaInfo != nil {
				logrus.WithFields(logrus.Fields{
					"media_type": mediaInfo.MediaType,
					"media_url":  mediaInfo.MediaURL,
					"device_id":  deviceID,
				}).Info("ðŸ–¼ï¸ FLOW: Extracted media URL from delayed response, sending as media message")
				
				// Send as media message instead of text
				err = s.SendMediaMessage(deviceID, phoneNumber, mediaInfo.MediaURL)
				if err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"device_id":    deviceID,
						"phone_number": phoneNumber,
						"media_url":    mediaInfo.MediaURL,
						"media_type":   mediaInfo.MediaType,
					}).Error("âŒ FLOW: Failed to send delayed media message")
					return fmt.Errorf("failed to send delayed media message: %w", err)
				}
			} else {
				// Fallback to text if extraction failed
				err = s.SendMessageFromDevice(deviceID, phoneNumber, response)
				if err != nil {
					logrus.WithError(err).Error("âŒ FLOW: Failed to send delayed response as text fallback")
					return fmt.Errorf("failed to send delayed response: %w", err)
				}
			}
		} else {
			// Send as regular text message
			err = s.SendMessageFromDevice(deviceID, phoneNumber, response)
			if err != nil {
				logrus.WithError(err).Error("âŒ FLOW: Failed to send delayed response")
				return fmt.Errorf("failed to send response: %w", err)
			}
		}

		// Add bot response to ai_whatsapp_nodepath conversation
		err = s.aiWhatsappService.SaveConversationHistory(phoneNumber, deviceID, "", response, "")
		if err != nil {
			logrus.WithError(err).Error("âŒ FLOW: Failed to add bot message to ai_whatsapp_nodepath")
		}

		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"response":     response,
		}).Info("âœ… FLOW: Delayed response sent successfully")
	} else {
		logrus.WithField("execution_id", executionID).Info("â„¹ï¸ FLOW: No response generated from delayed flow continuation")
	}

	return nil
}