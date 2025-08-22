package whatsapp

import (
	"fmt"
	"regexp"
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
func NewService(cfg *config.Config, queueService *services.QueueService, flowService *services.FlowService, aiService *services.AIService, aiWhatsappService services.AIWhatsappService, websocketService *services.WebSocketService, deviceSettingsService *services.DeviceSettingsService, providerService *services.ProviderService) (*Service, error) {
	service := &Service{
		cfg:                   cfg,
		queueService:          queueService,
		flowService:           flowService,
		aiService:             aiService,
		aiWhatsappService:     aiWhatsappService,
		websocketService:      websocketService,
		deviceSettingsService: deviceSettingsService,
		providerService:       providerService,
		messageQueue:          make(chan *WebhookMessage, 1000), // Buffered queue for performance
	}

	// Start message processing workers for high performance
	for i := 0; i < 10; i++ { // 10 worker goroutines for handling 3000+ devices
		go service.messageProcessor()
	}

	logrus.Info("🚀 WHATSAPP: Simplified webhook-based service initialized")
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
	}).Info("📨 WEBHOOK: Processing incoming message")

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
	}).Info("📤 MESSAGE: Sending message (not implemented)")
	return nil
}

// SendMessageFromDevice sends a message from a specific device through the appropriate provider
func (s *Service) SendMessageFromDevice(deviceID, phoneNumber, message string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"message":      message,
	}).Info("📤 MESSAGE: Sending message from device")

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
func (s *Service) SendMediaMessage(deviceID, phoneNumber, caption, mediaURL string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"caption":      caption,
		"media_url":    mediaURL,
	}).Info("📤 MEDIA: Sending media message")

	// Get device settings by device_id
	deviceSettings, err := s.deviceSettingsService.GetByIDDevice(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device settings for %s: %w", deviceID, err)
	}

	// Send media message through provider service
	err = s.providerService.SendMediaMessage(deviceSettings, phoneNumber, caption, mediaURL)
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
	}).Info("🔍 FLOW: Processing incoming message")

	// Check for personal commands (%, #, cmd)
	if strings.HasPrefix(content, "%") || strings.HasPrefix(content, "#") || strings.HasPrefix(content, "cmd") {
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"command":   content,
		}).Info("🔧 COMMAND: Personal command detected")
		return s.handlePersonalCommand(phoneNumber, content, deviceID)
	}

	// First check if there's any existing conversation record (regardless of execution status)
	existingRecord, err := s.aiWhatsappService.GetAIWhatsappByProspectAndDevice(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get existing conversation record")
		return err
	}

	if existingRecord != nil {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
			"current_stage": existingRecord.Stage,
			"flow_reference": existingRecord.FlowReference.String,
		}).Info("🔄 FLOW: Found existing conversation record, continuing conversation")
		
		// Continue with existing conversation - check if it's a flow or AI conversation
		if existingRecord.FlowReference.Valid && existingRecord.FlowReference.String != "" {
			// It's a flow execution - continue with flow processing
			logrus.WithField("flow_reference", existingRecord.FlowReference.String).Info("🔄 FLOW: Continuing flow execution")
			
			// Get the flow data from chatbot_flows_nodepath for existing conversation
			logrus.WithFields(logrus.Fields{
				"execution_id":   existingRecord.ExecutionID.String,
				"flow_reference": existingRecord.FlowReference.String,
			}).Info("📊 FLOW: Retrieving flow data for existing conversation")

			flow, err := s.flowService.GetFlow(existingRecord.FlowReference.String)
			if err != nil {
				logrus.WithError(err).Error("❌ FLOW: Failed to get flow from database for existing conversation")
				return err
			}

			if flow == nil {
				logrus.WithField("flow_reference", existingRecord.FlowReference.String).Error("❌ FLOW: Flow not found in database for existing conversation")
				return fmt.Errorf("flow not found for existing conversation")
			}

			logrus.WithFields(logrus.Fields{
				"flow_id":    flow.ID,
				"flow_name":  flow.Name,
				"current_node": existingRecord.CurrentNode.String,
			}).Info("✅ FLOW: Successfully retrieved flow data for existing conversation")
			
			_, err = s.processFlowMessage(flow, existingRecord, content)
			return err
		} else {
			// It's an AI conversation - continue with AI processing
			logrus.WithField("stage", existingRecord.Stage).Info("🤖 AI: Continuing AI conversation")
			return s.processAIConversation(phoneNumber, content, deviceID)
		}
	}

	// No existing record found - check for default flow to start new conversation
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("🆕 FLOW: No existing conversation found, checking for default flow")

	// Get default flow for device
	defaultFlow, err := s.flowService.GetDefaultFlowForDevice(deviceID)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get default flow for device")
		return err
	}

	if defaultFlow == nil {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("⚠️ FLOW: No default flow found for device, falling back to AI conversation")
		
		// Fallback to AI conversation when no flow is configured
		return s.processAIConversation(phoneNumber, content, deviceID)
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
		"flow_id":      defaultFlow.ID,
		"flow_name":    defaultFlow.Name,
	}).Info("🚀 FLOW: Starting new execution with default flow")

	// Start new execution with default flow
	variables := make(map[string]interface{})
	aiExecution, err := s.aiWhatsappService.StartFlowExecution(phoneNumber, deviceID, defaultFlow.ID, variables)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to start new execution")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": aiExecution.ExecutionID.String,
		"flow_id":      defaultFlow.ID,
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("✅ FLOW: New execution started successfully")

	// Note: Human mode checking would be implemented through a separate table or field
	// For now, we'll process all messages through the flow

	// Get the flow data from chatbot_flows_nodepath
	logrus.WithFields(logrus.Fields{
		"execution_id":   aiExecution.ExecutionID.String,
		"flow_reference": aiExecution.FlowReference.String,
	}).Info("📊 FLOW: Retrieving flow data from chatbot_flows_nodepath")

	flow, err := s.flowService.GetFlow(aiExecution.FlowReference.String)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get flow from database")
		return err
	}

	if flow == nil {
		logrus.WithField("flow_reference", aiExecution.FlowReference.String).Error("❌ FLOW: Flow not found in database")
		return fmt.Errorf("flow not found")
	}

	logrus.WithFields(logrus.Fields{
		"flow_id":    flow.ID,
		"flow_name":  flow.Name,
		"flow_niche": flow.Niche,
		"device_id":  flow.IdDevice,
	}).Info("✅ FLOW: Successfully retrieved flow data from chatbot_flows_nodepath")

	// Store user message for later conversation history saving
	logrus.WithFields(logrus.Fields{
			"execution_id": aiExecution.IDProspect,
			"message_type": "USER",
			"content":      content,
		}).Info("💬 FLOW: Processing user message for conversation history")

	// Process the message through the flow
	logrus.WithFields(logrus.Fields{
			"execution_id": aiExecution.IDProspect,
			"flow_id":      flow.ID,
			"current_node": aiExecution.CurrentNode.String,
			"user_input":   content,
		}).Info("⚙️ FLOW: Processing message through flow engine")

	response, err := s.processFlowMessage(flow, aiExecution, content)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to process flow message")
		return err
	}

	logrus.WithFields(logrus.Fields{
			"execution_id":    aiExecution.IDProspect,
			"response_length": len(response),
			"has_response":    response != "",
		}).Info("🔄 FLOW: Flow processing completed")

	if response != "" {
		logrus.WithFields(logrus.Fields{
			"phone_number":    phoneNumber,
			"device_id":       deviceID,
			"response":        response,
			"response_length": len(response),
		}).Info("📤 FLOW: Sending response back to user")

		// Send response back to user using specific device
		err = s.SendMessageFromDevice(deviceID, phoneNumber, response)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"device_id":    deviceID,
				"phone_number": phoneNumber,
			}).Error("❌ FLOW: Failed to send response message")
			return err
		}

		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"response":     response,
		}).Info("✅ FLOW: Response sent successfully")

		// Save complete conversation pair (user message + bot response)
		logrus.WithFields(logrus.Fields{
				"execution_id": aiExecution.IDProspect,
				"user_message": content,
				"bot_response": response,
			}).Info("💬 FLOW: Saving complete conversation pair to ai_whatsapp_nodepath")

		err = s.aiWhatsappService.SaveConversationHistory(phoneNumber, deviceID, content, response, "")
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to save conversation pair to ai_whatsapp_nodepath")
		} else {
			logrus.WithField("execution_id", aiExecution.IDProspect).Info("✅ FLOW: Complete conversation pair saved to ai_whatsapp_nodepath successfully")
		}
	} else {
		logrus.WithField("execution_id", aiExecution.IDProspect).Info("ℹ️ FLOW: No response generated from flow processing")
		
		// Create AI WhatsApp record as fallback when no flow response is generated
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("🤖 FLOW: Creating AI WhatsApp record for prospect tracking")
		
		// Check if AI WhatsApp record already exists
		existingRecord, err := s.aiWhatsappService.GetAIWhatsappByProspectAndDevice(phoneNumber, deviceID)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to check existing AI WhatsApp record")
		} else if existingRecord == nil {
			// Create new AI WhatsApp record for prospect tracking
			err = s.aiWhatsappService.CreateAIWhatsappRecord(phoneNumber, deviceID, content, flow.Niche)
			if err != nil {
				logrus.WithError(err).Error("❌ FLOW: Failed to create AI WhatsApp record")
			} else {
				logrus.WithFields(logrus.Fields{
					"phone_number": phoneNumber,
					"device_id":    deviceID,
					"niche":        flow.Niche,
				}).Info("✅ FLOW: AI WhatsApp record created successfully")
			}
		} else {
			// Save user message only when no bot response is generated
			logrus.WithFields(logrus.Fields{
				"phone_number": phoneNumber,
				"user_message": content,
				"stage": existingRecord.Stage,
			}).Info("💬 FLOW: Saving user message (no bot response generated)")
			
			err = s.aiWhatsappService.SaveConversationHistory(phoneNumber, deviceID, content, "", existingRecord.Stage)
			if err != nil {
				logrus.WithError(err).Error("❌ FLOW: Failed to save user message to conversation history")
			} else {
				logrus.WithField("phone_number", phoneNumber).Info("✅ FLOW: User message saved to conversation history successfully")
			}
		}
	}

	return nil
}

// handlePersonalCommand handles personal device commands (%, #, cmd)
func (s *Service) handlePersonalCommand(phoneNumber, command, deviceID string) error {
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"command":   command,
	}).Info("🔧 COMMAND: Processing personal command")

	if command == "cmd" {
		// For now, just send a response indicating command received
		// Human mode toggle would be implemented through a separate service
		return s.SendMessageFromDevice(deviceID, phoneNumber, "Command received. Human mode toggle not yet implemented.")
	}

	// Handle % and # commands for triggering AI based on current stage
	return s.processAIConversation(phoneNumber, command, deviceID)
}

// processAIConversation processes AI conversation when flow is not available
func (s *Service) processAIConversation(phoneNumber, content, deviceID string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
	}).Info("🤖 AI: Processing AI conversation")

	// Get current conversation stage from AI WhatsApp service
	var stage string
	// Note: We pass empty stage initially, the AI service will handle stage determination
	stage = "" // Default stage, AI service will determine appropriate stage

	// Process AI conversation through AI WhatsApp service
	response, err := s.aiWhatsappService.ProcessAIConversation(phoneNumber, deviceID, content, stage)
	if err != nil {
		logrus.WithError(err).Error("Failed to process AI conversation")
		// Send fallback message
		return s.SendMessageFromDevice(deviceID, phoneNumber, "I'm sorry, I'm having trouble processing your message right now. Please try again later.")
	}

	// Send AI response if we have one
	if response != nil && len(response.Response) > 0 {
		return s.sendAIResponse(phoneNumber, deviceID, response)
	}

	return nil
}

// sendAIResponse sends AI response with multiple message types (text and images)
// extractMediaURL extracts URL from text content based on file extensions
func (s *Service) extractMediaURL(content string) (string, string) {
	// First check for bracketed media format: [IMAGE: url], [AUDIO: url], [VIDEO: url]
	bracketedPatterns := map[string]string{
		`\[IMAGE:\s*([^\]]+)\]`: "image",
		`\[AUDIO:\s*([^\]]+)\]`: "audio",
		`\[VIDEO:\s*([^\]]+)\]`: "video",
	}
	
	for pattern, mediaType := range bracketedPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			// Extract URL from the bracketed format and trim whitespace
			mediaURL := strings.TrimSpace(matches[1])
			logrus.WithFields(logrus.Fields{
				"media_url": mediaURL,
				"media_type": mediaType,
				"pattern": pattern,
			}).Info("📤 MEDIA: Extracted media URL from bracketed format")
			return mediaURL, mediaType
		}
	}
	
	// Fallback to direct URL pattern with media file extensions
	mediaPattern := `https?://[^\s\[\]]+\.(png|jpg|jpeg|gif|mp3|wav|mp4|avi|mov|webm)`
	re := regexp.MustCompile(mediaPattern)
	match := re.FindString(content)
	
	if match != "" {
		// Determine media type based on file extension
		if regexp.MustCompile(`\.(png|jpg|jpeg|gif)$`).MatchString(match) {
			return match, "image"
		} else if regexp.MustCompile(`\.(mp3|wav)$`).MatchString(match) {
			return match, "audio"
		} else if regexp.MustCompile(`\.(mp4|avi|mov|webm)$`).MatchString(match) {
			return match, "video"
		}
	}
	
	return "", ""
}

// removeMediaBrackets removes bracketed media text from content
// This function removes patterns like [IMAGE: url], [AUDIO: url], [VIDEO: url] from text
func (s *Service) removeMediaBrackets(content string) string {
	// Define patterns for bracketed media formats
	bracketedPatterns := []string{
		`\[IMAGE:\s*[^\]]+\]`,
		`\[AUDIO:\s*[^\]]+\]`,
		`\[VIDEO:\s*[^\]]+\]`,
	}
	
	cleanedContent := content
	for _, pattern := range bracketedPatterns {
		re := regexp.MustCompile(pattern)
		cleanedContent = re.ReplaceAllString(cleanedContent, "")
	}
	
	// Clean up extra whitespace
	cleanedContent = regexp.MustCompile(`\s+`).ReplaceAllString(cleanedContent, " ")
	cleanedContent = strings.TrimSpace(cleanedContent)
	
	logrus.WithFields(logrus.Fields{
		"original_content": content,
		"cleaned_content": cleanedContent,
	}).Info("📤 MEDIA: Removed bracketed media text from content")
	
	return cleanedContent
}

func (s *Service) sendAIResponse(phoneNumber, deviceID string, response *services.AIWhatsappResponse) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"stage":        response.Stage,
		"response_count": len(response.Response),
	}).Info("📤 AI: Sending AI response")

	// Send each response item in sequence
	for i, item := range response.Response {
		switch item.Type {
		case "text":
			// Check if text content contains media URLs
			mediaURL, mediaType := s.extractMediaURL(item.Content)
			if mediaURL != "" {
				// Send as media message
				logrus.WithFields(logrus.Fields{
					"media_url": mediaURL,
					"media_type": mediaType,
				}).Info("📤 MEDIA: Extracted media URL from text content")
				
				err := s.SendMediaMessage(deviceID, phoneNumber, "", mediaURL)
				if err != nil {
					logrus.WithError(err).WithField("item_index", i).Error("Failed to send extracted media message")
					return err
				}
				
				// Remove the bracketed media text from content and send remaining text if any
				cleanedContent := s.removeMediaBrackets(item.Content)
				if strings.TrimSpace(cleanedContent) != "" {
					err := s.SendMessageFromDevice(deviceID, phoneNumber, cleanedContent)
					if err != nil {
						logrus.WithError(err).WithField("item_index", i).Error("Failed to send cleaned text message")
						return err
					}
				}
			} else {
				// Send as regular text message
				err := s.SendMessageFromDevice(deviceID, phoneNumber, item.Content)
				if err != nil {
					logrus.WithError(err).WithField("item_index", i).Error("Failed to send text message")
					return err
				}
			}
			// Add small delay between messages for better user experience
			time.Sleep(500 * time.Millisecond)

		case "image":
			// Send image message
			err := s.SendMediaMessage(deviceID, phoneNumber, "", item.Content)
			if err != nil {
				logrus.WithError(err).WithField("item_index", i).Error("Failed to send image message")
				return err
			}
			// Add small delay between messages
			time.Sleep(500 * time.Millisecond)

		default:
			logrus.WithField("type", item.Type).Warn("Unknown response type, skipping")
		}
	}

	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"stage":        response.Stage,
	}).Info("✅ AI: Successfully sent AI response")

	return nil
}

// processFlowMessage processes a message through the flow logic
func (s *Service) processFlowMessage(flow *models.ChatbotFlow, aiExecution *models.AIWhatsapp, userInput string) (string, error) {
	// Get current node
	currentNode, err := s.flowService.FindNodeByID(flow, aiExecution.CurrentNode.String)
	if err != nil {
		// If no current node, start from the beginning
		currentNode, err = s.flowService.GetStartNode(flow)
		if err != nil {
			return "", fmt.Errorf("failed to get start node: %w", err)
		}
		aiExecution.CurrentNode.String = currentNode.ID
	}

	// Process based on node type
	switch currentNode.Type {
	case models.NodeTypeStart:
		return s.processStartNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeAIPrompt:
		return s.processAIPromptNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeAdvancedAIPrompt:
		return s.processAdvancedAIPromptNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeManual:
		return s.processManualNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeMessage:
		return s.processMessageNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeImage:
		return s.processImageNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeAudio:
		return s.processAudioNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeVideo:
		return s.processVideoNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeDelay:
		return s.processDelayNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeCondition:
		return s.processConditionNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeStage:
		return s.processStageNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeUserReply:
		return s.processUserReplyNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeWaitingReplyTimes:
		return s.processWaitingReplyTimesNode(flow, aiExecution, currentNode, userInput)
	default:
		return s.processDefaultNode(flow, aiExecution, currentNode, userInput)
	}
}

// processAIPromptNode processes an AI prompt node
func (s *Service) processAIPromptNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
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

	// Get execution variables for prompt replacement
	variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in system prompt
	systemPrompt = s.flowService.ReplaceVariables(systemPrompt, variables)

	// Generate AI response
	response, err := s.aiService.GenerateResponse(systemPrompt, userInput, apiProvider, []models.ConversationMessage{})
	if err != nil {
		logrus.WithError(err).Error("Failed to generate AI response")
		return "I'm sorry, I'm having trouble processing your request right now. Please try again later.", nil
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
			}).Info("🤖 AI_PROMPT: AI response generated, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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

	// Use global settings as fallback
	if apiProvider == "" {
		apiProvider = flow.Niche
	}

	// Check if we have complete AI configuration
	if systemPrompt == "" || instance == "" || apiProvider == "" {
		// Fallback to manual response
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}

	// Get execution variables for prompt replacement
	variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in system prompt
	systemPrompt = s.flowService.ReplaceVariables(systemPrompt, variables)

	// Generate AI response with advanced JSON parsing
	response, err := s.aiService.GenerateResponse(systemPrompt, userInput, apiProvider, []models.ConversationMessage{})
	if err != nil {
		logrus.WithError(err).Error("Failed to generate advanced AI response")
		return "I'm sorry, I'm having trouble processing your request right now. Please try again later.", nil
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
			}).Info("🧠 ADVANCED_AI: Advanced AI response generated, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
	}).Info("👤 MANUAL: Manual intervention node triggered")

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
			}).Info("👤 MANUAL: Manual response sent, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
	variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
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
			}).Info("📤 MESSAGE: Message sent, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		}).Info("🏁 MESSAGE: End of flow reached, completing execution")
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

	// Replace variables in image URL
	variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}
	imageURL = s.flowService.ReplaceVariables(imageURL, variables)

	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"image_url":    imageURL,
	}).Info("🖼️ IMAGE: Processing image node")

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
			}).Info("🖼️ IMAGE: Image processed, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return fmt.Sprintf("[IMAGE: %s]", imageURL), err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return fmt.Sprintf("[IMAGE: %s]", imageURL), err
			}
			
			// Return image URL for webhook-based system
			return fmt.Sprintf("[IMAGE: %s]", imageURL), nil
		}
		
		// For non-delay nodes, continue processing immediately
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after image node")
			return fmt.Sprintf("[IMAGE: %s]", imageURL), err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after image")
			return fmt.Sprintf("[IMAGE: %s]", imageURL), err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return fmt.Sprintf("[IMAGE: %s]\n%s", imageURL, nextResponse), nil
		}
	} else {
		// End of flow
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
		}).Info("🏁 IMAGE: End of flow reached, completing execution")
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	// Return image URL for webhook-based system
	return fmt.Sprintf("[IMAGE: %s]", imageURL), nil
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

	// Replace variables in audio URL
	variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}
	audioURL = s.flowService.ReplaceVariables(audioURL, variables)

	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"audio_url":    audioURL,
	}).Info("🎵 AUDIO: Processing audio node")

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
			}).Info("🎵 AUDIO: Audio processed, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return fmt.Sprintf("[AUDIO: %s]", audioURL), err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return fmt.Sprintf("[AUDIO: %s]", audioURL), err
			}
			
			// Return audio URL for webhook-based system
			return fmt.Sprintf("[AUDIO: %s]", audioURL), nil
		}
		
		// For non-delay nodes, continue processing immediately
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after audio node")
			return fmt.Sprintf("[AUDIO: %s]", audioURL), err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after audio")
			return fmt.Sprintf("[AUDIO: %s]", audioURL), err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return fmt.Sprintf("[AUDIO: %s]\n%s", audioURL, nextResponse), nil
		}
	} else {
		// End of flow
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
		}).Info("🏁 AUDIO: End of flow reached, completing execution")
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	// Return audio URL for webhook-based system
	return fmt.Sprintf("[AUDIO: %s]", audioURL), nil
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

	// Replace variables in video URL
	variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}
	videoURL = s.flowService.ReplaceVariables(videoURL, variables)

	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"video_url":    videoURL,
	}).Info("🎬 VIDEO: Processing video node")

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
			}).Info("🎬 VIDEO: Video processed, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
			if err != nil {
				logrus.WithError(err).Error("Failed to update execution to delay node")
				return fmt.Sprintf("[VIDEO: %s]", videoURL), err
			}
			
			// Process the delay node immediately to schedule the next message
			_, err = s.processDelayNode(flow, execution, nextNode, userInput)
			if err != nil {
				logrus.WithError(err).Error("Failed to process delay node")
				return fmt.Sprintf("[VIDEO: %s]", videoURL), err
			}
			
			// Return video URL for webhook-based system
			return fmt.Sprintf("[VIDEO: %s]", videoURL), nil
		}
		
		// For non-delay nodes, continue processing immediately
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after video node")
			return fmt.Sprintf("[VIDEO: %s]", videoURL), err
		}
		
		// Recursively process the next node if it's not a delay
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process next node after video")
			return fmt.Sprintf("[VIDEO: %s]", videoURL), err
		}
		
		// Combine responses if next node generated content
		if nextResponse != "" {
			return fmt.Sprintf("[VIDEO: %s]\n%s", videoURL, nextResponse), nil
		}
	} else {
		// End of flow
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
		}).Info("🏁 VIDEO: End of flow reached, completing execution")
		s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	// Return video URL for webhook-based system
	return fmt.Sprintf("[VIDEO: %s]", videoURL), nil
}

// processDelayNode processes a delay node
func (s *Service) processDelayNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"flow_id":      flow.ID,
	}).Info("🕐 DELAY: Processing delay node")
	
	// Hardcode delay to 5 seconds for all delay nodes
	delaySeconds := 5
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"hardcoded_delay": delaySeconds,
	}).Info("🕐 DELAY: Using hardcoded 5-second delay")
	
	logrus.WithFields(logrus.Fields{
		"execution_id":   execution.IDProspect,
		"delay_seconds":  delaySeconds,
		"phone_number":   execution.ProspectNum,
		"device_id":      execution.IDDevice,
	}).Info("🕐 DELAY: Scheduling delayed message")
	
	// Get next node to process after delay
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err != nil || nextNode == nil {
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
		}).Info("🕐 DELAY: No next node found, completing execution")
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
	}).Info("🕐 DELAY: Keeping execution at current node, will advance after delay")
	
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
		logrus.WithError(err).Error("🕐 DELAY: Failed to queue delayed message")
		return "", fmt.Errorf("failed to queue delayed message: %w", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"execution_id":   execution.IDProspect,
		"message_id":     delayedMessage.ID,
		"delay_seconds":  delaySeconds,
		"next_node_id":   nextNode.ID,
	}).Info("🕐 DELAY: Message queued successfully for delayed processing")
	
	// Return empty string as no immediate response is needed
	// The delayed message will be processed later by the queue processor
	return "", nil
}

// processConditionNode processes a condition node with proper condition evaluation
// Uses a better approach to distinguish between first-time processing and user replies
// by checking if the current node in execution matches this condition node
func (s *Service) processConditionNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"user_input":   userInput,
		"current_node": execution.CurrentNode.String,
	}).Info("🔀 CONDITION: Processing condition node")

	// Check if this is the first time processing this condition node
	// If the current node in execution is NOT this condition node, it means we're arriving from another node
	if execution.CurrentNode.String != node.ID {
		// Phase 1: Send the condition question/message to user
		message := ""
		if msg, ok := node.Data["message"].(string); ok {
			message = msg
		} else if question, ok := node.Data["question"].(string); ok {
			message = question
		} else {
			// Default message if none specified
			message = "Please choose an option:"
		}

		// Replace variables in message
		variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
		if err != nil {
			logrus.WithError(err).Warn("Failed to get execution variables")
			variables = make(map[string]interface{})
		}
		message = s.flowService.ReplaceVariables(message, variables)

		// Update execution to stay on this condition node and wait for user reply
		execution.CurrentNode.String = node.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution to condition node")
			return message, err
		}

		logrus.WithFields(logrus.Fields{
			"execution_id": execution.IDProspect,
			"node_id":      node.ID,
			"message":      message,
		}).Info("🔀 CONDITION: Sent question, waiting for user reply")

		return message, nil
	}

	// Phase 2: User has replied, now evaluate the conditions
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":      node.ID,
		"user_reply":   userInput,
	}).Info("🔀 CONDITION: Evaluating user reply against conditions")

	// Extract conditions from node data
	conditions, ok := node.Data["conditions"].([]interface{})
	if !ok {
		logrus.WithFields(logrus.Fields{
			"node_id": node.ID,
			"execution_id": execution.IDProspect,
		}).Warn("🔀 CONDITION: No conditions found in node data, using default routing")
		
		// Fallback to default routing if no conditions
		nextNode, err := s.flowService.GetNextNode(flow, node.ID)
		if err != nil || nextNode == nil {
			return "", err
		}
		return s.processNextNodeAfterCondition(flow, execution, node, nextNode, userInput, "default")
	}

	// Normalize user input for comparison
	userInputLower := strings.ToLower(strings.TrimSpace(userInput))
	
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"execution_id": execution.IDProspect,
		"user_input": userInput,
		"conditions_count": len(conditions),
	}).Info("🔀 CONDITION: Evaluating user input against conditions")

	// Evaluate each condition
	var defaultConditionID string
	for _, conditionInterface := range conditions {
		condition, ok := conditionInterface.(map[string]interface{})
		if !ok {
			continue
		}
		
		conditionID, _ := condition["id"].(string)
		conditionType, _ := condition["type"].(string)
		conditionValue, _ := condition["value"].(string)
		conditionLabel, _ := condition["label"].(string)
		
		// Store default condition for fallback
		if conditionType == "default" {
			defaultConditionID = conditionID
			continue
		}
		
		// Evaluate condition based on type
		var conditionMet bool
		switch conditionType {
		case "equals":
			conditionMet = strings.ToLower(strings.TrimSpace(conditionValue)) == userInputLower
		case "contains":
			conditionMet = strings.Contains(userInputLower, strings.ToLower(strings.TrimSpace(conditionValue)))
		}
		
		logrus.WithFields(logrus.Fields{
			"condition_id": conditionID,
			"condition_type": conditionType,
			"condition_value": conditionValue,
			"condition_label": conditionLabel,
			"condition_met": conditionMet,
			"user_reply": userInput,
		}).Info("🔀 CONDITION: Evaluated condition")
		
		if conditionMet {
			// Find next node based on this condition
			nextNode, err := s.flowService.GetNextNodeByCondition(flow, node.ID, conditionID)
			if err != nil {
				logrus.WithError(err).Error("Failed to get next node by condition")
				return "", err
			}
			
			if nextNode != nil {
				logrus.WithFields(logrus.Fields{
					"condition_id": conditionID,
					"condition_label": conditionLabel,
					"next_node": nextNode.ID,
				}).Info("🔀 CONDITION: Condition matched, routing to next node")
				
				return s.processNextNodeAfterCondition(flow, execution, node, nextNode, userInput, conditionLabel)
			}
		}
	}
	
	// No condition matched, use default condition if available
	if defaultConditionID != "" {
		nextNode, err := s.flowService.GetNextNodeByCondition(flow, node.ID, defaultConditionID)
		if err != nil {
			logrus.WithError(err).Error("Failed to get next node by default condition")
			return "", err
		}
		
		if nextNode != nil {
			logrus.WithFields(logrus.Fields{
				"condition_id": defaultConditionID,
				"next_node": nextNode.ID,
			}).Info("🔀 CONDITION: Using default condition path")
			
			return s.processNextNodeAfterCondition(flow, execution, node, nextNode, userInput, "default")
		}
	}
	
	// No matching condition and no default - end flow
	logrus.WithFields(logrus.Fields{
		"node_id": node.ID,
		"execution_id": execution.IDProspect,
		"user_input": userInput,
	}).Warn("🔀 CONDITION: No matching condition found and no default path")
	
	return "", nil
}

// processNextNodeAfterCondition handles the next node processing after condition evaluation
func (s *Service) processNextNodeAfterCondition(flow *models.ChatbotFlow, execution *models.AIWhatsapp, currentNode *models.FlowNode, nextNode *models.FlowNode, userInput string, conditionLabel string) (string, error) {
	if nextNode.Type == models.NodeTypeDelay {
		// Advance to delay node and process it immediately
		// This ensures the delay is scheduled properly
		logrus.WithFields(logrus.Fields{
			"prospect_id": execution.IDProspect,
			"current_node": currentNode.ID,
			"next_node":    nextNode.ID,
			"next_type":    nextNode.Type,
			"condition": conditionLabel,
		}).Info("🔀 CONDITION: Condition evaluated, advancing to delay node")
		
		// Update execution to delay node
		execution.CurrentNode.String = nextNode.ID
		err := s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
	logrus.WithFields(logrus.Fields{
		"prospect_id": execution.IDProspect,
		"current_node": currentNode.ID,
		"next_node": nextNode.ID,
		"next_type": nextNode.Type,
		"condition": conditionLabel,
	}).Info("🔀 CONDITION: Condition evaluated, continuing to next node")
	
	// Update execution to next node
	execution.CurrentNode.String = nextNode.ID
	err := s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
	if err != nil {
		logrus.WithError(err).Error("Failed to update execution to next node")
		return "", err
	}
	
	// Recursively process the next node
	return s.processFlowMessage(flow, execution, userInput)
}

// processStageNode processes a stage node
func (s *Service) processStageNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// For now, just log the stage transition
	// Stage tracking would be implemented through a separate field or table
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.IDProspect,
		"node_id":     node.ID,
		"stage":       node.Data["stage"],
	}).Info("🎯 STAGE: Stage transition node processed")

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
			}).Info("🎯 STAGE: Stage processed, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
func (s *Service) processUserReplyNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	// Store user input and move to next node
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
			}).Info("💬 USER_REPLY: User reply processed, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after user reply node")
			return "", err
		}
		
		// Recursively process the next node if it's not a delay
		return s.processFlowMessage(flow, execution, userInput)
	}
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
			}).Info("⏱️ WAITING_REPLY: Reply timing processed, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
			}).Info("🚀 START: Start node processed, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
			}).Info("🔧 DEFAULT: Default node processed, advancing to delay node")
			
			// Update execution to delay node
			execution.CurrentNode.String = nextNode.ID
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
		execution.CurrentNode.String = nextNode.ID
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
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
	logrus.Info("🚀 QUEUE: Starting queue processor")

	// For now, just log that the queue processor would start
	// Queue processing would be implemented through the queue service
	logrus.Info("📋 QUEUE: Queue processor started (placeholder implementation)")
}

// processQueuedMessage processes a queued message from the queue service
func (s *Service) processQueuedMessage(message *services.QueueMessage) error {
	// For now, just log the queued message processing
	// Queue message processing would be implemented based on the actual QueueMessage structure
	logrus.WithFields(logrus.Fields{
		"message_id": message.ID,
		"content":    message.Content,
	}).Info("📋 QUEUE: Processing queued message (placeholder implementation)")
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
	}).Info("🔄 FLOW: Processing flow continuation after delay")

	// First try to get execution by specific execution ID
	execution, err := s.aiWhatsappService.GetFlowExecutionByID(executionID)
	if err != nil {
		logrus.WithError(err).WithField("execution_id", executionID).Error("❌ FLOW: Failed to get execution by ID")
		// Fallback to phone number and device ID lookup
		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("🔄 FLOW: Falling back to phone number and device ID lookup")
		
		execution, err = s.aiWhatsappService.GetActiveFlowExecution(phoneNumber, deviceID)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to get active execution for continuation")
			return fmt.Errorf("failed to get active execution: %w", err)
		}
	}

	// If no execution found by ID, try to get any execution (including completed)
	if execution == nil {
		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("🔄 FLOW: No execution found by ID, checking for any existing execution")
		
		// Get any execution (regardless of status) to continue delayed processing
		execution, err = s.aiWhatsappService.GetFlowExecutionByProspectAndDevice(phoneNumber, deviceID)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to get any execution for continuation")
			return fmt.Errorf("failed to get any execution: %w", err)
		}
		
		if execution == nil {
			logrus.WithField("execution_id", executionID).Warn("⚠️ FLOW: No execution found for continuation")
			return fmt.Errorf("execution not found: %s", executionID)
		}
		
		// Reactivate the execution for delayed processing
		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"previous_status": execution.ExecutionStatus.String,
		}).Info("🔄 FLOW: Reactivating execution for delayed message processing")
		
		// Set execution status back to active for processing
		execution.ExecutionStatus.String = "active"
		execution.ExecutionStatus.Valid = true
	}

	// Get the flow
	flow, err := s.flowService.GetFlow(flowID)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get flow for continuation")
		return fmt.Errorf("failed to get flow: %w", err)
	}

	if flow == nil {
		logrus.WithField("flow_id", flowID).Warn("⚠️ FLOW: Flow not found for continuation")
		return fmt.Errorf("flow not found: %s", flowID)
	}

	// Get the target node (the node to process after delay)
	targetNode, err := s.flowService.FindNodeByID(flow, nodeID)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get target node for continuation")
		return fmt.Errorf("failed to get target node: %w", err)
	}

	if targetNode == nil {
		logrus.WithField("node_id", nodeID).Warn("⚠️ FLOW: Target node not found for continuation")
		return fmt.Errorf("target node not found: %s", nodeID)
	}

	// Update execution to the target node (advance from delay node to next node)
	logrus.WithFields(logrus.Fields{
		"execution_id":   executionID,
		"previous_node":  execution.CurrentNode.String,
		"target_node":    nodeID,
	}).Info("🔄 FLOW: Advancing execution to target node after delay")
	
	execution.CurrentNode.String = nodeID
	err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNode.String, make(map[string]interface{}), "active")
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to update execution to target node")
		return fmt.Errorf("failed to update execution: %w", err)
	}

	// Process the target node
	response, err := s.processFlowMessage(flow, execution, userInput)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to process flow continuation")
		return fmt.Errorf("failed to process flow: %w", err)
	}

	// Send response if available
	if response != "" {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
			"response":     response,
		}).Info("📤 FLOW: Sending delayed response to user")

		err = s.SendMessageFromDevice(deviceID, phoneNumber, response)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to send delayed response")
			return fmt.Errorf("failed to send response: %w", err)
		}

		// Save delayed bot response to conversation history (no user message for delayed responses)
		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"bot_response": response,
			"response_type": "delayed",
		}).Info("💬 FLOW: Saving delayed bot response to ai_whatsapp_nodepath")
		
		err = s.aiWhatsappService.SaveConversationHistory(phoneNumber, deviceID, "", response, "")
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to save delayed bot response to ai_whatsapp_nodepath")
		} else {
			logrus.WithField("execution_id", executionID).Info("✅ FLOW: Delayed bot response saved to ai_whatsapp_nodepath successfully")
		}

		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"response":     response,
		}).Info("✅ FLOW: Delayed response sent successfully")
	} else {
		logrus.WithField("execution_id", executionID).Info("ℹ️ FLOW: No response generated from delayed flow continuation")
	}

	return nil
}