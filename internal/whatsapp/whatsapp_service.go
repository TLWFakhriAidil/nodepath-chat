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
// Rebuilt from scratch for proper flow execution
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

	logrus.Info("🚀 WHATSAPP: Rebuilt flow processing service initialized")
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

// SetServices sets the flow and AI services
func (s *Service) SetServices(flowService *services.FlowService, aiService *services.AIService) {
	s.flowService = flowService
	s.aiService = aiService
}

// ProcessIncomingMessageFromWebhook processes incoming messages from webhook
func (s *Service) ProcessIncomingMessageFromWebhook(phoneNumber, content, deviceID, provider string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
		"provider":     provider,
		"content":      content,
	}).Info("📨 WEBHOOK: Received message")

	// Queue message for processing
	webhookMsg := &WebhookMessage{
		PhoneNumber: phoneNumber,
		Content:     content,
		DeviceID:    deviceID,
		Provider:    provider,
		Timestamp:   time.Now(),
		Retries:     0,
	}

	// Non-blocking queue insertion
	select {
	case s.messageQueue <- webhookMsg:
		logrus.Info("✅ WEBHOOK: Message queued for processing")
	default:
		logrus.Error("❌ WEBHOOK: Message queue is full, processing directly")
		return s.processWebhookMessageInternal(webhookMsg)
	}

	return nil
}

// SendMessage sends a message (placeholder for compatibility)
func (s *Service) SendMessage(phoneNumber, message string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"message":      message,
	}).Info("📤 SEND: Sending message")
	return nil
}

// SendMessageFromDevice sends a message from a specific device
func (s *Service) SendMessageFromDevice(deviceID, phoneNumber, message string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"message":      message,
	}).Info("📤 DEVICE: Sending message from device")

	// Get device settings for provider information
	deviceSettings, err := s.deviceSettingsService.GetByIDDevice(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device settings: %w", err)
	}

	// Send message using provider service
	return s.providerService.SendMessage(deviceSettings, phoneNumber, message)
}

// SendMediaMessage sends a media message from a specific device
func (s *Service) SendMediaMessage(deviceID, phoneNumber, caption, mediaURL string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"caption":      caption,
		"media_url":    mediaURL,
	}).Info("📤 MEDIA: Sending media message")

	// Get device settings for provider information
	deviceSettings, err := s.deviceSettingsService.GetByIDDevice(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device settings: %w", err)
	}

	// Send media message using provider service
	return s.providerService.SendMediaMessage(deviceSettings, phoneNumber, caption, mediaURL)
}

// ============================================================================
// MAIN MESSAGE PROCESSING - COMPLETELY REBUILT FROM SCRATCH
// ============================================================================

// processIncomingMessage handles incoming messages with proper flow logic
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

	// Check if there's an active flow for this device
	flow, err := s.flowService.GetDefaultFlowForDevice(deviceID)
	if err != nil {
		logrus.WithError(err).Info("No default flow found, falling back to AI conversation")
		return s.processAIConversation(phoneNumber, content, deviceID)
	}

	if flow == nil {
		// No flow configured, use AI conversation
		logrus.Info("⚠️ FLOW: No default flow found, using AI conversation")
		return s.processAIConversation(phoneNumber, content, deviceID)
	}

	// Check for existing conversation
	existingRecord, err := s.aiWhatsappService.GetAIWhatsappByProspectAndDevice(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get existing conversation record")
		return err
	}

	if existingRecord != nil {
		// Continue existing conversation
		return s.continueExistingConversation(existingRecord, content)
	}

	// Start new conversation with default flow
	return s.startNewConversation(phoneNumber, deviceID, content)
}

// continueExistingConversation continues an existing conversation
func (s *Service) continueExistingConversation(execution *models.AIWhatsapp, userInput string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": execution.ProspectNum,
		"device_id":    execution.IDDevice,
		"current_node": execution.CurrentNode.String,
		"flow_ref":     execution.FlowReference.String,
	}).Info("🔄 FLOW: Continuing existing conversation")

	// Check if it's a flow or AI conversation
	if execution.FlowReference.Valid && execution.FlowReference.String != "" {
		// Continue flow execution
		return s.executeFlow(execution, userInput)
	} else {
		// Check if device has flows configured - if so, start flow execution
		defaultFlow, err := s.flowService.GetDefaultFlowForDevice(execution.IDDevice)
		if err != nil {
			logrus.WithError(err).Warn("⚠️ FLOW: Failed to check for default flow, falling back to AI conversation")
			return s.processAIConversation(execution.ProspectNum, userInput, execution.IDDevice)
		}

		if defaultFlow != nil {
			// Device has flows configured, convert existing record to flow execution
			logrus.WithFields(logrus.Fields{
				"phone_number": execution.ProspectNum,
				"device_id":    execution.IDDevice,
				"flow_id":      defaultFlow.ID,
			}).Info("🔄 FLOW: Converting existing AI conversation to flow execution")

			// Start flow execution on existing record
			variables := make(map[string]interface{})
			updatedExecution, err := s.aiWhatsappService.StartFlowExecution(execution.ProspectNum, execution.IDDevice, defaultFlow.ID, variables)
			if err != nil {
				logrus.WithError(err).Error("❌ FLOW: Failed to convert to flow execution, falling back to AI conversation")
				return s.processAIConversation(execution.ProspectNum, userInput, execution.IDDevice)
			}

			// Execute the flow with updated execution record
			return s.executeFlow(updatedExecution, userInput)
		} else {
			// No flows configured, continue AI conversation
			logrus.Info("🤖 FLOW: No flows configured for device, continuing AI conversation")
			return s.processAIConversation(execution.ProspectNum, userInput, execution.IDDevice)
		}
	}
}

// startNewConversation starts a new conversation with default flow
func (s *Service) startNewConversation(phoneNumber, deviceID, userInput string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("🆕 FLOW: Starting new conversation")

	// Get default flow for device
	defaultFlow, err := s.flowService.GetDefaultFlowForDevice(deviceID)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get default flow")
		return err
	}

	if defaultFlow == nil {
		// No flow configured, use AI conversation
		logrus.Info("⚠️ FLOW: No default flow found, using AI conversation")
		return s.processAIConversation(phoneNumber, userInput, deviceID)
	}

	// Start new flow execution
	variables := make(map[string]interface{})
	execution, err := s.aiWhatsappService.StartFlowExecution(phoneNumber, deviceID, defaultFlow.ID, variables)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to start flow execution")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": execution.ExecutionID.String,
		"flow_id":      defaultFlow.ID,
	}).Info("✅ FLOW: New flow execution started")

	// For first-time users, execute the flow without user input first to send initial greeting
	// This will process the start node and advance to the first message node
	logrus.Info("🚀 FLOW: Executing initial flow to send greeting message")
	err = s.executeFlow(execution, "")
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to execute initial flow")
		// Don't return error, continue with user input processing
	}

	// Now process the user's actual input against the current flow state
	if userInput != "" {
		logrus.Info("📝 FLOW: Processing user input after initial greeting")
		// Get updated execution state
		updatedExecution, err := s.aiWhatsappService.GetAIWhatsappByProspectAndDevice(phoneNumber, deviceID)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to get updated execution state")
			return err
		}
		
		// Execute flow with user input
		return s.executeFlow(updatedExecution, userInput)
	}

	return nil
}

// executeFlow executes the flow logic using the new FlowEngine
func (s *Service) executeFlow(execution *models.AIWhatsapp, userInput string) error {
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.ExecutionID.String,
		"flow_ref":     execution.FlowReference.String,
		"current_node": execution.CurrentNode.String,
	}).Info("⚙️ FLOW: Executing flow with FlowEngine")

	// Create FlowEngine instance
	flowEngine := NewFlowEngine(
		s.flowService,
		s.aiWhatsappService,
		s.aiService,
		s.providerService,
		s.deviceSettingsService,
	)

	// Execute flow using the FlowEngine
	err := flowEngine.ExecuteFlow(execution, userInput)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: FlowEngine execution failed")
		return err
	}

	logrus.Info("✅ FLOW: FlowEngine execution completed successfully")
	return nil
}

// processFlowExecution is deprecated - use FlowEngine.ExecuteFlow instead
// This function is kept for backward compatibility but should not be used
func (s *Service) processFlowExecution(flow *models.ChatbotFlow, execution *models.AIWhatsapp, userInput string) (string, error) {
	logrus.Warn("⚠️ DEPRECATED: processFlowExecution called - use FlowEngine.ExecuteFlow instead")
	return "", fmt.Errorf("deprecated function called - use FlowEngine.ExecuteFlow instead")
}

// getCurrentOrStartNode gets the current node or start node if none set
func (s *Service) getCurrentOrStartNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp) (*models.FlowNode, error) {
	if execution.CurrentNode.String != "" {
		// Get current node
		node, err := s.flowService.FindNodeByID(flow, execution.CurrentNode.String)
		if err == nil && node != nil {
			return node, nil
		}
	}

	// Get start node
	startNode, err := s.flowService.GetStartNode(flow)
	if err != nil {
		return nil, fmt.Errorf("failed to get start node: %w", err)
	}

	// Update execution with start node
	execution.CurrentNode.String = startNode.ID
	err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, startNode.ID, make(map[string]interface{}), "active")
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to update to start node")
	}

	return startNode, nil
}

// processNode processes a single node based on its type
// processNode is deprecated - use FlowEngine node processors instead
// This function is kept for backward compatibility but should not be used
func (s *Service) processNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.Warn("⚠️ DEPRECATED: processNode called - use FlowEngine node processors instead")
	return "", false, fmt.Errorf("deprecated function called - use FlowEngine node processors instead")
}

// ============================================================================
// NODE PROCESSORS - SIMPLE AND CLEAN IMPLEMENTATIONS
// ============================================================================

// processStartNode processes start node
func (s *Service) processStartNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.WithField("node_id", node.ID).Info("🚀 FLOW: Processing start node")
	
	// Start node just initiates the flow, no response needed
	return "", false, nil
}

// processMessageNode processes message node
func (s *Service) processMessageNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.WithField("node_id", node.ID).Info("💬 FLOW: Processing message node")
	
	// Return the message content
	if node.Data != nil {
		if message, ok := node.Data["message"].(string); ok && message != "" {
			return message, false, nil
		}
	}
	
	return "", false, nil
}

// processAIPromptNode processes AI prompt node
func (s *Service) processAIPromptNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.WithField("node_id", node.ID).Info("🤖 FLOW: Processing AI prompt node")
	
	// Get AI prompt from node data
	if node.Data == nil {
		return "", false, nil
	}
	
	prompt, ok := node.Data["prompt"].(string)
	if !ok || prompt == "" {
		return "", false, nil
	}
	
	// Get device settings for AI configuration
	deviceSettings, err := s.deviceSettingsService.GetByIDDevice(execution.IDDevice)
	if err != nil {
		return "", false, fmt.Errorf("failed to get device settings: %w", err)
	}
	
	// Get API key from device settings
	apiKey := ""
	if deviceSettings.APIKey.Valid {
		apiKey = deviceSettings.APIKey.String
	}
	
	// Call AI service with correct parameters
	response, err := s.aiService.GenerateResponse(prompt, userInput, apiKey, nil)
	if err != nil {
		return "", false, fmt.Errorf("AI generation failed: %w", err)
	}
	
	return response, false, nil
}

// processUserReplyNode processes user reply node - STOPS EXECUTION
func (s *Service) processUserReplyNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.WithField("node_id", node.ID).Info("👤 FLOW: Processing user reply node - STOPPING")
	
	// User Reply node stops execution and waits for user input
	// Return empty response and signal to stop
	return "", true, nil
}

// processDelayNode processes delay node
func (s *Service) processDelayNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.WithField("node_id", node.ID).Info("⏰ FLOW: Processing delay node")
	
	// Get delay duration from node data
	if node.Data != nil {
		if delayStr, ok := node.Data["delay"].(string); ok {
			// Parse delay and schedule continuation
			if duration, err := time.ParseDuration(delayStr); err == nil {
				// Schedule flow continuation after delay
				go func() {
					time.Sleep(duration)
					// Continue flow after delay
					s.executeFlow(execution, "")
				}()
				
				// Stop current execution, will continue after delay
				return "", true, nil
			}
		}
	}
	
	// No delay specified, continue normally
	return "", false, nil
}

// processConditionNode processes condition node
func (s *Service) processConditionNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.WithField("node_id", node.ID).Info("🔀 FLOW: Processing condition node")
	
	// Simple condition processing - can be enhanced later
	// For now, just continue to next node
	return "", false, nil
}

// processImageNode processes image node
func (s *Service) processImageNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.WithField("node_id", node.ID).Info("🖼️ FLOW: Processing image node")
	
	// Get image URL from node data
	if node.Data != nil {
		if imageURL, ok := node.Data["image_url"].(string); ok && imageURL != "" {
			caption := ""
			if cap, ok := node.Data["caption"].(string); ok {
				caption = cap
			}
			
			// Send image
			err := s.SendMediaMessage(execution.IDDevice, execution.ProspectNum, caption, imageURL)
			if err != nil {
				return "", false, fmt.Errorf("failed to send image: %w", err)
			}
			
			// Caption is already sent with the media, no need to return as separate text response
			return "", false, nil
		}
	}
	
	return "", false, nil
}

// processAudioNode processes audio node
func (s *Service) processAudioNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.WithField("node_id", node.ID).Info("🎵 FLOW: Processing audio node")
	
	// Get audio URL from node data
	if node.Data != nil {
		if audioURL, ok := node.Data["audio_url"].(string); ok && audioURL != "" {
			// Send audio
			err := s.SendMediaMessage(execution.IDDevice, execution.ProspectNum, "", audioURL)
			if err != nil {
				return "", false, fmt.Errorf("failed to send audio: %w", err)
			}
			
			return "", false, nil
		}
	}
	
	return "", false, nil
}

// processVideoNode processes video node
func (s *Service) processVideoNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, bool, error) {
	logrus.WithField("node_id", node.ID).Info("🎥 FLOW: Processing video node")
	
	// Get video URL from node data
	if node.Data != nil {
		if videoURL, ok := node.Data["video_url"].(string); ok && videoURL != "" {
			caption := ""
			if cap, ok := node.Data["caption"].(string); ok {
				caption = cap
			}
			
			// Send video
			err := s.SendMediaMessage(execution.IDDevice, execution.ProspectNum, caption, videoURL)
			if err != nil {
				return "", false, fmt.Errorf("failed to send video: %w", err)
			}
			
			// Caption is already sent with the media, no need to return as separate text response
			return "", false, nil
		}
	}
	
	return "", false, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// sendFlowResponse is deprecated - FlowEngine handles response sending internally
// This function is kept for backward compatibility but should not be used
func (s *Service) sendFlowResponse(phoneNumber, deviceID, response string) error {
	logrus.Warn("⚠️ DEPRECATED: sendFlowResponse called - FlowEngine handles responses internally")
	
	// Filter out empty responses to prevent <nil> messages
	if response == "" || response == "<nil>" || response == "nil" {
		logrus.WithField("response", response).Warn("🚫 DEPRECATED: Filtered out invalid response")
		return nil
	}
	
	// Extract media URLs if present
	mediaURL, cleanResponse := s.extractMediaURL(response)
	
	if mediaURL != "" {
		// Send media message
		return s.SendMediaMessage(deviceID, phoneNumber, cleanResponse, mediaURL)
	} else {
		// Send text message
		return s.SendMessageFromDevice(deviceID, phoneNumber, response)
	}
}

// handlePersonalCommand handles personal commands
func (s *Service) handlePersonalCommand(phoneNumber, command, deviceID string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"command":      command,
		"device_id":    deviceID,
	}).Info("🔧 COMMAND: Processing personal command")
	
	// Handle different command types
	if strings.HasPrefix(command, "cmd") {
		// Toggle human mode
		return s.toggleHumanMode(phoneNumber, deviceID)
	}
	
	// Handle provider-specific commands (% for wablas, # for whacenter)
	return s.handleProviderCommand(phoneNumber, command, deviceID)
}

// toggleHumanMode toggles human mode for a conversation
func (s *Service) toggleHumanMode(phoneNumber, deviceID string) error {
	// Implementation for toggling human mode
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("👤 COMMAND: Toggling human mode")
	
	// This would update the human status in the database
	// For now, just log the action
	return nil
}

// handleProviderCommand handles provider-specific commands
func (s *Service) handleProviderCommand(phoneNumber, command, deviceID string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"command":      command,
		"device_id":    deviceID,
	}).Info("🔧 COMMAND: Processing provider command")
	
	// Implementation for provider-specific commands
	return nil
}

// processAIConversation processes AI conversation (fallback when no flow)
func (s *Service) processAIConversation(phoneNumber, content, deviceID string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("🤖 AI: Processing AI conversation")
	
	// Get or create AI WhatsApp record
	existingRecord, err := s.aiWhatsappService.GetAIWhatsappByProspectAndDevice(phoneNumber, deviceID)
	if err != nil {
		return err
	}
	
	if existingRecord == nil {
		// Create new AI conversation
		err = s.aiWhatsappService.CreateAIWhatsappRecord(phoneNumber, deviceID, content, "general")
		if err != nil {
			return err
		}
		
		// Get the newly created record
		existingRecord, err = s.aiWhatsappService.GetAIWhatsappByProspectAndDevice(phoneNumber, deviceID)
		if err != nil {
			return err
		}
	}
	
	// Use AI WhatsApp service to process conversation
	response, err := s.aiWhatsappService.ProcessAIConversation(phoneNumber, deviceID, content, existingRecord.Stage)
	if err != nil {
		return err
	}
	
	// Send response
	err = s.sendAIResponse(phoneNumber, deviceID, response)
	if err != nil {
		return err
	}
	
	// Save conversation
	responseText := ""
	if len(response.Response) > 0 {
		responseText = response.Response[0].Content
	}
	
	return s.aiWhatsappService.SaveConversationHistory(phoneNumber, deviceID, content, responseText, response.Stage)
}

// sendAIResponse sends AI response with media support
func (s *Service) sendAIResponse(phoneNumber, deviceID string, response *services.AIWhatsappResponse) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
		"stage":        response.Stage,
	}).Info("📤 AI: Sending AI response")
	
	// Send each response item
	for _, item := range response.Response {
		switch item.Type {
		case "text":
			if item.Content != "" {
				err := s.SendMessageFromDevice(deviceID, phoneNumber, item.Content)
				if err != nil {
					return err
				}
			}
		case "image":
			err := s.SendMediaMessage(deviceID, phoneNumber, "", item.Content)
			if err != nil {
				return err
			}
		case "audio":
			err := s.SendMediaMessage(deviceID, phoneNumber, "", item.Content)
			if err != nil {
				return err
			}
		case "video":
			err := s.SendMediaMessage(deviceID, phoneNumber, "", item.Content)
			if err != nil {
				return err
			}
		}
		
		// Small delay between messages
		time.Sleep(100 * time.Millisecond)
	}
	
	return nil
}

// extractMediaURL extracts media URL from message content
func (s *Service) extractMediaURL(content string) (string, string) {
	// Look for media URLs in brackets like [image:url] or [video:url]
	imageRegex := regexp.MustCompile(`\[image:([^\]]+)\]`)
	videoRegex := regexp.MustCompile(`\[video:([^\]]+)\]`)
	audioRegex := regexp.MustCompile(`\[audio:([^\]]+)\]`)
	
	if matches := imageRegex.FindStringSubmatch(content); len(matches) > 1 {
		cleanContent := imageRegex.ReplaceAllString(content, "")
		return matches[1], strings.TrimSpace(cleanContent)
	}
	
	if matches := videoRegex.FindStringSubmatch(content); len(matches) > 1 {
		cleanContent := videoRegex.ReplaceAllString(content, "")
		return matches[1], strings.TrimSpace(cleanContent)
	}
	
	if matches := audioRegex.FindStringSubmatch(content); len(matches) > 1 {
		cleanContent := audioRegex.ReplaceAllString(content, "")
		return matches[1], strings.TrimSpace(cleanContent)
	}
	
	return "", content
}

// ============================================================================
// QUEUE AND CONTINUATION FUNCTIONS
// ============================================================================

// StartQueueProcessor starts the queue processor
func (s *Service) StartQueueProcessor() {
	logrus.Info("🚀 QUEUE: Starting queue processor")
	// Queue processor is already started in NewService
}

// processQueuedMessage processes a queued message
func (s *Service) processQueuedMessage(message *services.QueueMessage) error {
	logrus.WithField("message_id", message.ID).Info("📋 QUEUE: Processing queued message")
	
	// Implementation for processing queued messages
	return nil
}

// ProcessFlowContinuation processes flow continuation after delay
func (s *Service) ProcessFlowContinuation(executionID, flowID, nodeID, phoneNumber, deviceID, userInput string) error {
	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"flow_id":      flowID,
		"node_id":      nodeID,
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("🔄 FLOW: Processing flow continuation")
	
	// Get execution record
	execution, err := s.aiWhatsappService.GetAIWhatsappByProspectAndDevice(phoneNumber, deviceID)
	if err != nil {
		return err
	}
	
	if execution == nil {
		return fmt.Errorf("execution not found")
	}
	
	// Continue flow execution
	return s.executeFlow(execution, userInput)
}