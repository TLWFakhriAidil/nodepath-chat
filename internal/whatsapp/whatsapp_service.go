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
	chatService      *services.ChatService
	queueService     *services.QueueService
	flowService      *services.FlowService
	aiService        *services.AIService
	websocketService *services.WebSocketService

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
func NewService(cfg *config.Config, chatService *services.ChatService, queueService *services.QueueService, flowService *services.FlowService, aiService *services.AIService, websocketService *services.WebSocketService) (*Service, error) {
	service := &Service{
		cfg:              cfg,
		chatService:      chatService,
		queueService:     queueService,
		flowService:      flowService,
		aiService:        aiService,
		websocketService: websocketService,
		messageQueue:     make(chan *WebhookMessage, 1000), // Buffered queue for performance
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
	// For webhook-based system, we need to use external provider APIs
	// This will be handled by the queue service or external API calls
	return fmt.Errorf("SendMessage not implemented for webhook-based system - use provider-specific APIs")
}

// SendMessageFromDevice sends a message from a specific device using provider APIs
func (s *Service) SendMessageFromDevice(deviceID, phoneNumber, message string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"message":      message,
	}).Info("📤 SENDING: Message via provider API")

	// For webhook-based system, delegate to queue service for provider API calls
	if s.queueService != nil {
		return s.queueService.SendMessage(deviceID, phoneNumber, message)
	}

	return fmt.Errorf("queue service not available for sending messages")
}

// SendMediaMessage sends a media message using provider APIs
func (s *Service) SendMediaMessage(phoneNumber, caption, mediaURL, mediaType string) error {
	// For webhook-based system, delegate to queue service for provider API calls
	if s.queueService != nil {
		return s.queueService.SendMediaMessage(phoneNumber, caption, mediaURL, mediaType)
	}

	return fmt.Errorf("queue service not available for sending media messages")
}

// processIncomingMessage processes incoming messages and handles flow/AI logic
func (s *Service) processIncomingMessage(phoneNumber, content string, deviceID string) error {
	logrus.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"content":      content,
	}).Info("🔍 FLOW: Checking for active execution")

	// Check for personal commands (%, #, cmd)
	if strings.HasPrefix(content, "%") || strings.HasPrefix(content, "#") || strings.HasPrefix(content, "cmd") {
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"command":   content,
		}).Info("🔧 COMMAND: Personal command detected")
		return s.handlePersonalCommand(phoneNumber, content, deviceID)
	}

	// Get or create active execution
	execution, err := s.chatService.GetActiveExecution(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get active execution")
		return err
	}

	if execution == nil {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("🆕 FLOW: No active execution found, checking for default flow")

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
			}).Info("⚠️ FLOW: No default flow found for device")
			return nil
		}

		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
			"flow_id":      defaultFlow.ID,
			"flow_name":    defaultFlow.Name,
		}).Info("🚀 FLOW: Starting new execution with default flow")

		// Start new execution with default flow
		execution, err = s.chatService.StartExecution(defaultFlow.ID, phoneNumber, deviceID)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to start new execution")
			return err
		}

		logrus.WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"flow_id":      defaultFlow.ID,
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("✅ FLOW: New execution started successfully")
	} else {
		logrus.WithFields(logrus.Fields{
			"execution_id":   execution.ID,
			"flow_reference": execution.FlowReference,
			"phone_number":   phoneNumber,
			"device_id":      deviceID,
		}).Info("🔄 FLOW: Found existing active execution")
	}

	// Check if human mode is active (no AI reply)
	if execution.HumanMode == 1 {
		logrus.WithFields(logrus.Fields{
			"device_id":    deviceID,
			"phone_number": phoneNumber,
		}).Info("👤 HUMAN: Human mode active, skipping AI processing")
		return nil
	}

	// Get the flow data from chatbot_flows_nodepath
	logrus.WithFields(logrus.Fields{
		"execution_id":   execution.ID,
		"flow_reference": execution.FlowReference,
	}).Info("📊 FLOW: Retrieving flow data from chatbot_flows_nodepath")

	flow, err := s.flowService.GetFlow(execution.FlowReference)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get flow from database")
		return err
	}

	if flow == nil {
		logrus.WithField("flow_reference", execution.FlowReference).Error("❌ FLOW: Flow not found in database")
		return fmt.Errorf("flow not found")
	}

	logrus.WithFields(logrus.Fields{
		"flow_id":    flow.ID,
		"flow_name":  flow.Name,
		"flow_niche": flow.Niche,
		"device_id":  flow.IdDevice,
	}).Info("✅ FLOW: Successfully retrieved flow data from chatbot_flows_nodepath")

	// Add user message to conversation
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"message_type": "USER",
		"content":      content,
	}).Info("💬 FLOW: Adding user message to conversation")

	err = s.chatService.AddConversationMessage(execution, "USER", content)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to add user message to conversation")
		return err
	}

	logrus.WithField("execution_id", execution.ID).Info("✅ FLOW: User message added to conversation successfully")

	// Process the message through the flow
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"flow_id":      flow.ID,
		"current_node": execution.CurrentNode,
		"user_input":   content,
	}).Info("⚙️ FLOW: Processing message through flow engine")

	response, err := s.processFlowMessage(flow, execution, content)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to process flow message")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"execution_id":    execution.ID,
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

		// Add bot response to conversation
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"message_type": "BOT",
			"response":     response,
		}).Info("💬 FLOW: Adding bot response to conversation")

		err = s.chatService.AddConversationMessage(execution, "BOT", response)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to add bot message to conversation")
		} else {
			logrus.WithField("execution_id", execution.ID).Info("✅ FLOW: Bot response added to conversation successfully")
		}
	} else {
		logrus.WithField("execution_id", execution.ID).Info("ℹ️ FLOW: No response generated from flow processing")
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
		// Toggle human mode
		execution, err := s.chatService.GetOrCreateExecution(phoneNumber, deviceID)
		if err != nil {
			return err
		}

		newHumanMode := 1
		if execution.HumanMode == 1 {
			newHumanMode = 0
		}

		err = s.chatService.UpdateHumanMode(execution.ID, newHumanMode)
		if err != nil {
			return err
		}

		response := "AI mode activated"
		if newHumanMode == 1 {
			response = "Human mode activated"
		}

		return s.SendMessageFromDevice(deviceID, phoneNumber, response)
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

	if s.aiService == nil {
		return fmt.Errorf("AI service not available")
	}

	// Get execution for context
	execution, err := s.chatService.GetOrCreateExecution(phoneNumber, deviceID)
	if err != nil {
		return err
	}

	// Process AI conversation
	response, err := s.aiService.ProcessConversation(execution, content)
	if err != nil {
		return err
	}

	// Send AI response
	if response != "" {
		return s.SendMessageFromDevice(deviceID, phoneNumber, response)
	}

	return nil
}

// processFlowMessage processes a message through the flow logic
func (s *Service) processFlowMessage(flow *models.ChatbotFlow, execution *models.ChatbotExecution, userInput string) (string, error) {
	// Get current node
	currentNode, err := s.flowService.FindNodeByID(flow, execution.CurrentNode)
	if err != nil {
		// If no current node, start from the beginning
		currentNode, err = s.flowService.GetStartNode(flow)
		if err != nil {
			return "", fmt.Errorf("failed to get start node: %w", err)
		}
		execution.CurrentNode = currentNode.ID
	}

	// Process based on node type
	switch currentNode.Type {
	case models.NodeTypeStart:
		return s.processStartNode(flow, execution, currentNode, userInput)
	case models.NodeTypeAIPrompt:
		return s.processAIPromptNode(flow, execution, currentNode, userInput)
	case models.NodeTypeAdvancedAIPrompt:
		return s.processAdvancedAIPromptNode(flow, execution, currentNode, userInput)
	case models.NodeTypeManual:
		return s.processManualNode(flow, execution, currentNode, userInput)
	case models.NodeTypeMessage:
		return s.processMessageNode(flow, execution, currentNode, userInput)
	case models.NodeTypeImage:
		return s.processImageNode(flow, execution, currentNode, userInput)
	case models.NodeTypeAudio:
		return s.processAudioNode(flow, execution, currentNode, userInput)
	case models.NodeTypeVideo:
		return s.processVideoNode(flow, execution, currentNode, userInput)
	case models.NodeTypeDelay:
		return s.processDelayNode(flow, execution, currentNode, userInput)
	case models.NodeTypeCondition:
		return s.processConditionNode(flow, execution, currentNode, userInput)
	case models.NodeTypeStage:
		return s.processStageNode(flow, execution, currentNode, userInput)
	case models.NodeTypeUserReply:
		return s.processUserReplyNode(flow, execution, currentNode, userInput)
	case models.NodeTypeWaitingReplyTimes:
		return s.processWaitingReplyTimesNode(flow, execution, currentNode, userInput)
	default:
		return s.processDefaultNode(flow, execution, currentNode, userInput)
	}
}

// processAIPromptNode processes an AI prompt node
func (s *Service) processAIPromptNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
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
	history, err := s.chatService.GetConversationHistory(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get conversation history")
		history = []models.ConversationMessage{}
	}

	// Get execution variables for prompt replacement
	variables, err := s.chatService.GetExecutionVariables(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in system prompt
	systemPrompt = s.flowService.ReplaceVariables(systemPrompt, variables)

	// Generate AI response
	response, err := s.aiService.GenerateResponse(systemPrompt, userInput, apiProvider, history)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate AI response")
		return "I'm sorry, I'm having trouble processing your request right now. Please try again later.", nil
	}

	// Move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
	} else {
		// End of flow
		s.chatService.CompleteExecution(execution.ID)
	}

	return response, nil
}

// processAdvancedAIPromptNode processes an advanced AI prompt node with JSON response parsing
func (s *Service) processAdvancedAIPromptNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Similar to processAIPromptNode but with advanced JSON parsing
	// Implementation would include JSON response parsing and multi-part responses
	return s.processAIPromptNode(flow, execution, node, userInput)
}

// processManualNode processes a manual node (human intervention required)
func (s *Service) processManualNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Set execution to manual mode
	execution.HumanMode = 1
	s.chatService.UpdateExecution(execution)

	// Return manual response message
	if message, ok := node.Data["message"].(string); ok {
		return message, nil
	}
	return "Your message has been forwarded to our support team. We'll get back to you soon.", nil
}

// processMessageNode processes a simple message node
func (s *Service) processMessageNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get message from node data
	message := ""
	if msg, ok := node.Data["message"].(string); ok {
		message = msg
	}

	// Replace variables in message
	variables, err := s.chatService.GetExecutionVariables(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}
	message = s.flowService.ReplaceVariables(message, variables)

	// Move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
	} else {
		// End of flow
		s.chatService.CompleteExecution(execution.ID)
	}

	return message, nil
}

// processImageNode processes an image node
func (s *Service) processImageNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// For webhook-based system, we'll return a text message indicating image would be sent
	// In a full implementation, this would trigger image sending via the provider API
	return s.processMessageNode(flow, execution, node, userInput)
}

// processAudioNode processes an audio node
func (s *Service) processAudioNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// For webhook-based system, we'll return a text message indicating audio would be sent
	return s.processMessageNode(flow, execution, node, userInput)
}

// processVideoNode processes a video node
func (s *Service) processVideoNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// For webhook-based system, we'll return a text message indicating video would be sent
	return s.processMessageNode(flow, execution, node, userInput)
}

// processDelayNode processes a delay node
func (s *Service) processDelayNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// For webhook-based system, we'll skip delays and move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processConditionNode processes a condition node
func (s *Service) processConditionNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Evaluate condition and move to appropriate next node
	// This would include condition evaluation logic
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processStageNode processes a stage node
func (s *Service) processStageNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Update execution stage and move to next node
	if stage, ok := node.Data["stage"].(string); ok {
		execution.CurrentStage = stage
		s.chatService.UpdateExecution(execution)
	}

	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processUserReplyNode processes a user reply node
func (s *Service) processUserReplyNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Store user input and move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processWaitingReplyTimesNode processes a waiting reply times node
func (s *Service) processWaitingReplyTimesNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Handle reply timing logic and move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processStartNode processes a start node
func (s *Service) processStartNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Move to next node from start
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// processDefaultNode processes any unrecognized node type
func (s *Service) processDefaultNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Default behavior - move to next node or end flow
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}
	s.chatService.CompleteExecution(execution.ID)
	return "", nil
}

// StartQueueProcessor starts the queue processor for handling queued messages
func (s *Service) StartQueueProcessor() {
	logrus.Info("🚀 Starting queue processor for WhatsApp service")

	// Start processing queued messages from queue service
	go func() {
		for {
			// Get next message from queue
			message, err := s.queueService.GetNextMessage()
			if err != nil {
				logrus.WithError(err).Error("Failed to get next message from queue")
				time.Sleep(5 * time.Second)
				continue
			}

			if message != nil {
				err = s.processQueuedMessage(message)
				if err != nil {
					logrus.WithError(err).Error("Failed to process queued message")
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

// processQueuedMessage processes a message from the queue
func (s *Service) processQueuedMessage(message *services.QueueMessage) error {
	return s.ProcessIncomingMessageFromWebhook(message.PhoneNumber, message.Content, message.DeviceID, "queue")
}