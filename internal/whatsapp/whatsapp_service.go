package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"

	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"
)

// QueuedMessage represents a message in the processing queue
type QueuedMessage struct {
	DeviceID string
	Message  *events.Message
	Retries  int
	Timestamp time.Time
}

// Service handles WhatsApp integration using whatsmeow with multi-device support
type Service struct {
	cfg          *config.Config
	
	// Multi-device support
	clients      map[string]*whatsmeow.Client // deviceID -> client
	containers   map[string]*sqlstore.Container // deviceID -> container
	connections  map[string]bool // deviceID -> connection status
	clientMutex  sync.RWMutex
	
	// Configuration
	maxDevices     int
	storageDir     string
	currentDevices int
	
	// Services
	chatService    *services.ChatService
	queueService   *services.QueueService
	flowService    *services.FlowService
	aiService      *services.AIService
	websocketService *services.WebSocketService
	
	// Performance optimizations
	messageQueue chan *QueuedMessage
	processingWG sync.WaitGroup
	isConnected  bool
}

// NewService creates a new WhatsApp service with multi-device support and performance optimizations
func NewService(cfg *config.Config, chatService *services.ChatService, queueService *services.QueueService, flowService *services.FlowService, aiService *services.AIService, websocketService *services.WebSocketService) (*Service, error) {
	service := &Service{
		cfg:              cfg,
		clients:          make(map[string]*whatsmeow.Client),
		containers:       make(map[string]*sqlstore.Container),
		connections:      make(map[string]bool),
		maxDevices:       cfg.WhatsAppMaxDevices,
		storageDir:       cfg.WhatsAppStoragePath,
		chatService:      chatService,
		queueService:     queueService,
		flowService:      flowService,
		aiService:        aiService,
		websocketService: websocketService,
		messageQueue:     make(chan *QueuedMessage, 1000), // Buffered queue for performance
		isConnected:      false,
	}
	
	// Start message processing workers
	for i := 0; i < 5; i++ { // 5 worker goroutines for message processing
		go service.messageProcessor()
	}
	
	return service, nil
}

// messageProcessor processes queued messages in background
func (s *Service) messageProcessor() {
	for queuedMsg := range s.messageQueue {
		s.processingWG.Add(1)
		go func(msg *QueuedMessage) {
			defer s.processingWG.Done()
			
			// Process the message with retry logic
			for i := 0; i < 3; i++ {
				if err := s.processQueuedMessageInternal(msg); err != nil {
					logrus.WithFields(logrus.Fields{
						"device_id": msg.DeviceID,
						"retry": i + 1,
						"error": err,
					}).Warn("Failed to process queued message, retrying")
					time.Sleep(time.Duration(i+1) * time.Second)
					continue
				}
				break
			}
		}(queuedMsg)
	}
}

// processQueuedMessageInternal processes a single queued message
func (s *Service) processQueuedMessageInternal(msg *QueuedMessage) error {
	if msg.Message == nil {
		return fmt.Errorf("message is nil")
	}

	// Extract phone number and content from the message
	phoneNumber := msg.Message.Info.Sender.User
	content := ""
	
	if msg.Message.Message.GetConversation() != "" {
		content = msg.Message.Message.GetConversation()
	} else if msg.Message.Message.GetExtendedTextMessage() != nil {
		content = msg.Message.Message.GetExtendedTextMessage().GetText()
	}

	// Process the message
	s.processIncomingMessage(phoneNumber, content)
	return nil
}

// SetServices sets additional services after initialization
func (s *Service) SetServices(flowService *services.FlowService, aiService *services.AIService) {
	s.flowService = flowService
	s.aiService = aiService
}

// initializeClient initializes a whatsmeow client for a specific device
func (s *Service) initializeClient(deviceID string) error {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	// Check if we've reached the maximum number of devices
	if len(s.clients) >= s.maxDevices {
		return fmt.Errorf("maximum number of devices (%d) reached", s.maxDevices)
	}

	// Ensure storage directory exists
	storageDir := s.cfg.WhatsAppStoragePath
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Initialize SQLite store for session data
	dbPath := filepath.Join(storageDir, fmt.Sprintf("whatsapp_%s.db", deviceID))
	container, err := sqlstore.New("sqlite", fmt.Sprintf("%s?_pragma=foreign_keys(1)", dbPath), nil)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	// Get device store
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return fmt.Errorf("failed to get device store: %w", err)
	}

	// Create WhatsApp client
	client := whatsmeow.NewClient(deviceStore, nil)

	// Add event handlers
	client.AddEventHandler(s.handleEvent)

	// Store client and container
	s.clients[deviceID] = client
	s.containers[deviceID] = container
	s.connections[deviceID] = false
	s.currentDevices++

	logrus.WithField("device_id", deviceID).Info("WhatsApp client initialized")
	return nil
}

// Connect connects to WhatsApp for a specific device
func (s *Service) Connect(deviceID string) error {
	s.clientMutex.RLock()
	client, exists := s.clients[deviceID]
	s.clientMutex.RUnlock()

	if !exists {
		// Initialize client if it doesn't exist
		if err := s.initializeClient(deviceID); err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}
		s.clientMutex.RLock()
		client = s.clients[deviceID]
		s.clientMutex.RUnlock()
	}

	if client.Store.ID == nil {
		// Generate QR code for pairing
		qrChan, err := client.GetQRChannel(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get QR channel: %w", err)
		}

		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					logrus.WithFields(logrus.Fields{
						"device_id": deviceID,
						"qr_code": evt.Code,
					}).Info("QR code for WhatsApp pairing")
					// In a real implementation, you'd display this QR code in the web interface
				} else {
					logrus.WithFields(logrus.Fields{
						"device_id": deviceID,
						"event": evt.Event,
					}).Info("QR channel event")
				}
			}
		}()
	}

	// Connect to WhatsApp
	err := client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to WhatsApp: %w", err)
	}

	// Update connection status
	s.clientMutex.Lock()
	s.connections[deviceID] = true
	s.clientMutex.Unlock()

	// Update global connection status if this is the first device
	if !s.isConnected {
		s.isConnected = true
	}

	logrus.WithField("device_id", deviceID).Info("Connected to WhatsApp")
	return nil
}

// Disconnect disconnects from WhatsApp for a specific device or all devices
func (s *Service) Disconnect(deviceID ...string) {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	if len(deviceID) == 0 {
		// Disconnect all devices
		for id, client := range s.clients {
			if client != nil {
				client.Disconnect()
				s.connections[id] = false
				logrus.WithField("device_id", id).Info("Disconnected from WhatsApp")
			}
		}
		s.isConnected = false
	} else {
		// Disconnect specific device
		id := deviceID[0]
		if client, exists := s.clients[id]; exists && client != nil {
			client.Disconnect()
			s.connections[id] = false
			logrus.WithField("device_id", id).Info("Disconnected from WhatsApp")
			
			// Check if any devices are still connected
			anyConnected := false
			for _, connected := range s.connections {
				if connected {
					anyConnected = true
					break
				}
			}
			s.isConnected = anyConnected
		}
	}
}

// IsConnected returns the connection status for a specific device or overall
func (s *Service) IsConnected(deviceID ...string) bool {
	s.clientMutex.RLock()
	defer s.clientMutex.RUnlock()

	if len(deviceID) == 0 {
		// Check overall connection status
		return s.isConnected
	}

	// Check specific device
	id := deviceID[0]
	if client, exists := s.clients[id]; exists {
		return s.connections[id] && client != nil && client.IsConnected()
	}
	return false
}

// GetQRCode returns the QR code for pairing (if needed) for a specific device
func (s *Service) GetQRCode(deviceID string) (string, error) {
	s.clientMutex.RLock()
	client, exists := s.clients[deviceID]
	s.clientMutex.RUnlock()

	if !exists {
		// Initialize client if it doesn't exist
		if err := s.initializeClient(deviceID); err != nil {
			return "", fmt.Errorf("failed to initialize client: %w", err)
		}
		s.clientMutex.RLock()
		client = s.clients[deviceID]
		s.clientMutex.RUnlock()
	}

	if client.Store.ID != nil {
		return "", fmt.Errorf("device already paired")
	}

	qrChan, err := client.GetQRChannel(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get QR channel: %w", err)
	}

	// Wait for QR code
	select {
	case evt := <-qrChan:
		if evt.Event == "code" {
			return evt.Code, nil
		}
		return "", fmt.Errorf("unexpected QR event: %s", evt.Event)
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("timeout waiting for QR code")
	}
}

// SendMessage sends a text message using the first available device
func (s *Service) SendMessage(phoneNumber, message string) error {
	return s.SendMessageFromDevice("", phoneNumber, message)
}

// SendMessageFromDevice sends a text message from a specific device
func (s *Service) SendMessageFromDevice(deviceID, phoneNumber, message string) error {
	s.clientMutex.RLock()
	defer s.clientMutex.RUnlock()

	// Get client for the specified device or first available
	var client *whatsmeow.Client
	var selectedDeviceID string
	
	if deviceID != "" {
		// Use specific device
		if c, exists := s.clients[deviceID]; exists && s.connections[deviceID] {
			client = c
			selectedDeviceID = deviceID
		} else {
			return fmt.Errorf("device %s not connected", deviceID)
		}
	} else {
		// Use first available connected device
		for id, c := range s.clients {
			if s.connections[id] && c != nil && c.IsConnected() {
				client = c
				selectedDeviceID = id
				break
			}
		}
	}

	if client == nil {
		return fmt.Errorf("no WhatsApp devices connected")
	}

	// Parse phone number to JID
	jid, err := s.parsePhoneNumber(phoneNumber)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Create message
	msg := &waProto.Message{
		Conversation: proto.String(message),
	}

	// Send message
	_, err = client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"device_id":    selectedDeviceID,
		"phone_number": phoneNumber,
		"message_len":  len(message),
	}).Info("Message sent via WhatsApp")

	return nil
}

// SendMediaMessage sends a media message
func (s *Service) SendMediaMessage(phoneNumber, caption, mediaURL, mediaType string) error {
	if !s.IsConnected() {
		return fmt.Errorf("WhatsApp not connected")
	}

	// For now, send as text with media URL
	// In a full implementation, you'd download and upload the media
	message := caption
	if mediaURL != "" {
		if message != "" {
			message += "\n\n"
		}
		message += fmt.Sprintf("Media (%s): %s", mediaType, mediaURL)
	}

	return s.SendMessage(phoneNumber, message)
}

// handleEvent handles incoming WhatsApp events
func (s *Service) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		s.handleIncomingMessage(v)
	case *events.Connected:
		logrus.Info("WhatsApp connected")
		s.isConnected = true
	case *events.Disconnected:
		logrus.Info("WhatsApp disconnected")
		s.isConnected = false
	case *events.LoggedOut:
		logrus.Info("WhatsApp logged out")
		s.isConnected = false
	default:
		logrus.WithField("event_type", fmt.Sprintf("%T", v)).Debug("Unhandled WhatsApp event")
	}
}

// handleIncomingMessage processes incoming WhatsApp messages
func (s *Service) handleIncomingMessage(evt *events.Message) {
	// Skip messages from self
	if evt.Info.IsFromMe {
		return
	}

	// Extract message content
	content := ""
	if evt.Message.GetConversation() != "" {
		content = evt.Message.GetConversation()
	} else if evt.Message.GetExtendedTextMessage() != nil {
		content = evt.Message.GetExtendedTextMessage().GetText()
	}

	if content == "" {
		logrus.Debug("Received empty message, ignoring")
		return
	}

	// Extract phone number
	phoneNumber := evt.Info.Sender.User

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"content":      content,
		"message_id":   evt.Info.ID,
	}).Info("Received WhatsApp message")

	// Process message through flow engine
	go s.processIncomingMessage(phoneNumber, content, "default")
}

// ProcessIncomingMessageFromWebhook processes an incoming message from webhook through the flow engine
func (s *Service) ProcessIncomingMessageFromWebhook(phoneNumber, content, deviceID, provider string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"content": content,
		"device_id": deviceID,
		"provider": provider,
	}).Info("Processing webhook message through flow engine")

	// Process the message through the flow engine
	s.processIncomingMessage(phoneNumber, content, deviceID)
	return nil
}

// processIncomingMessage processes an incoming message through the flow engine
func (s *Service) processIncomingMessage(phoneNumber, content string, deviceID ...string) {
	// Use provided device ID or default
	idDevice := "default"
	if len(deviceID) > 0 && deviceID[0] != "" {
		idDevice = deviceID[0]
	}

	// Get or create active execution
	execution, err := s.chatService.GetActiveExecution(phoneNumber, idDevice)
	if err != nil {
		logrus.WithError(err).Error("Failed to get active execution")
		return
	}

	if execution == nil {
		// No active execution - you might want to start a default flow here
		logrus.WithField("phone_number", phoneNumber).Info("No active execution found for incoming message")
		return
	}

	// Get the flow
	flow, err := s.flowService.GetFlow(execution.FlowReference)
	if err != nil {
		logrus.WithError(err).Error("Failed to get flow")
		return
	}

	if flow == nil {
		logrus.WithField("flow_reference", execution.FlowReference).Error("Flow not found")
		return
	}

	// Add user message to conversation
	err = s.chatService.AddConversationMessage(execution, "USER", content)
	if err != nil {
		logrus.WithError(err).Error("Failed to add user message")
		return
	}

	// Process the message through the flow
	response, err := s.processFlowMessage(flow, execution, content)
	if err != nil {
		logrus.WithError(err).Error("Failed to process flow message")
		return
	}

	if response != "" {
		// Send response back to user
		err = s.SendMessage(phoneNumber, response)
		if err != nil {
			logrus.WithError(err).Error("Failed to send response message")
			return
		}

		// Add bot response to conversation
		err = s.chatService.AddConversationMessage(execution, "BOT", response)
		if err != nil {
			logrus.WithError(err).Error("Failed to add bot message")
		}
	}
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

// processAdvancedAIPromptNode processes an advanced AI prompt node with structured response
func (s *Service) processAdvancedAIPromptNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
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

	// Replace variables in system prompt and closing prompt
	systemPrompt = s.flowService.ReplaceVariables(systemPrompt, variables)
	if closingPrompt != "" {
		closingPrompt = s.flowService.ReplaceVariables(closingPrompt, variables)
	}

	// Generate advanced AI response
	response, err := s.aiService.GenerateAdvancedResponse(systemPrompt, userInput, apiProvider, history, closingPrompt)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate advanced AI response")
		return "I'm sorry, I'm having trouble processing your request right now. Please try again later.", nil
	}

	// Update execution stage if provided
	if response.Stage != "" && response.Stage != "error" {
		// Store the stage in execution variables for future reference
		err := s.chatService.SetExecutionVariable(execution, "current_stage", response.Stage)
		if err != nil {
			logrus.WithError(err).Warn("Failed to update execution stage variable")
		}
	}

	// Process response parts and build final message
	finalResponse := s.buildResponseFromParts(response.Response)

	// Move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
	} else {
		// End of flow
		s.chatService.CompleteExecution(execution.ID)
	}

	return finalResponse, nil
}

// buildResponseFromParts builds a final response string from AI response parts
func (s *Service) buildResponseFromParts(responseParts []models.AIResponsePart) string {
	var finalResponse string
	var textParts []string

	for _, part := range responseParts {
		switch part.Type {
		case "text":
			if part.Jenis == "onemessage" {
				// Combine with other text parts
				textParts = append(textParts, part.Content)
			} else {
				// Send as separate message (for now, just add to final response)
				if finalResponse != "" {
					finalResponse += "\n\n"
				}
				finalResponse += part.Content
			}
		case "image":
			// For now, just mention the image URL in the response
			// In a full implementation, this would trigger an image send
			if part.URL != "" {
				if finalResponse != "" {
					finalResponse += "\n\n"
				}
				finalResponse += "[Image: " + part.URL + "]"
			}
		}
	}

	// Combine all text parts marked as "onemessage"
	if len(textParts) > 0 {
		combinedText := strings.Join(textParts, " ")
		if finalResponse != "" {
			finalResponse = combinedText + "\n\n" + finalResponse
		} else {
			finalResponse = combinedText
		}
	}

	if finalResponse == "" {
		finalResponse = "I'm sorry, I couldn't generate a proper response. Please try again."
	}

	return finalResponse
}

// processManualNode processes a manual node
func (s *Service) processManualNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get manual response from node data
	response := "Thank you for your message."
	if msg, ok := node.Data["message"].(string); ok && msg != "" {
		response = msg
	}

	// Get execution variables for response replacement
	variables, err := s.chatService.GetExecutionVariables(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in response
	response = s.flowService.ReplaceVariables(response, variables)

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

// processDefaultNode processes other node types
func (s *Service) processDefaultNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// For other node types, just move to the next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}

	// End of flow
	s.chatService.CompleteExecution(execution.ID)
	return "Thank you for using our service!", nil
}

// processUserReplyNode processes a user reply node that waits indefinitely for user input
func (s *Service) processUserReplyNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// User reply node waits for any user input before proceeding
	// Once user provides input, move to the next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}

	// End of flow
	s.chatService.CompleteExecution(execution.ID)
	return "Thank you for your response!", nil
}

// processWaitingReplyTimesNode processes a waiting reply times node with configurable timeout
func (s *Service) processWaitingReplyTimesNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get wait time from node data (default to 5 seconds if not specified)
	waitTime := 5
	if wt, ok := node.Data["waitTime"].(float64); ok {
		waitTime = int(wt)
	} else if wt, ok := node.Data["waitTimeSeconds"].(float64); ok {
		waitTime = int(wt)
	}

	// For now, we'll treat this as immediate processing since the timeout logic
	// would require more complex scheduling infrastructure
	// In a production system, this would involve setting up a timer
	// and handling timeout scenarios
	// TODO: Implement actual timeout logic using waitTime (%d seconds)
	_ = waitTime // Suppress unused variable warning

	// Move to next node after processing user input
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}

	// End of flow
	s.chatService.CompleteExecution(execution.ID)
	return "Thank you for your response!", nil
}

// processMessageNode processes a message node that sends a text message
func (s *Service) processMessageNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get message from node data
	message := "Hello! This is an automated message."
	if msg, ok := node.Data["message"].(string); ok && msg != "" {
		message = msg
	}

	// Get execution variables for message replacement
	variables, err := s.chatService.GetExecutionVariables(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in message
	message = s.flowService.ReplaceVariables(message, variables)

	// Move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		// Continue processing the next node
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			return message, nil // Return current message even if next fails
		}
		// If next node also has a response, combine them
		if nextResponse != "" {
			return message + "\n" + nextResponse, nil
		}
	} else {
		// End of flow
		s.chatService.CompleteExecution(execution.ID)
	}

	return message, nil
}

// processImageNode processes an image node that sends an image with caption
func (s *Service) processImageNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get image URL and caption from node data
	imageURL := ""
	caption := "Image"
	
	if url, ok := node.Data["imageUrl"].(string); ok && url != "" {
		imageURL = strings.Trim(url, " `")
	}
	if cap, ok := node.Data["caption"].(string); ok && cap != "" {
		caption = cap
	}

	// Get execution variables for caption replacement
	variables, err := s.chatService.GetExecutionVariables(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in caption
	caption = s.flowService.ReplaceVariables(caption, variables)

	// Move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		// Continue processing the next node
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			return fmt.Sprintf("[Image: %s] %s", imageURL, caption), nil
		}
		// If next node also has a response, combine them
		if nextResponse != "" {
			return fmt.Sprintf("[Image: %s] %s\n%s", imageURL, caption, nextResponse), nil
		}
	} else {
		// End of flow
		s.chatService.CompleteExecution(execution.ID)
	}

	return fmt.Sprintf("[Image: %s] %s", imageURL, caption), nil
}

// processAudioNode processes an audio node that sends an audio file
func (s *Service) processAudioNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get audio URL from node data
	audioURL := ""
	duration := 30
	
	if url, ok := node.Data["audioUrl"].(string); ok && url != "" {
		audioURL = strings.Trim(url, " `")
	}
	if dur, ok := node.Data["duration"].(float64); ok {
		duration = int(dur)
	}

	// Move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		// Continue processing the next node
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			return fmt.Sprintf("[Audio: %s (%ds)]", audioURL, duration), nil
		}
		// If next node also has a response, combine them
		if nextResponse != "" {
			return fmt.Sprintf("[Audio: %s (%ds)]\n%s", audioURL, duration, nextResponse), nil
		}
	} else {
		// End of flow
		s.chatService.CompleteExecution(execution.ID)
	}

	return fmt.Sprintf("[Audio: %s (%ds)]", audioURL, duration), nil
}

// processVideoNode processes a video node that sends a video file with caption
func (s *Service) processVideoNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get video URL and caption from node data
	videoURL := ""
	caption := "Video"
	duration := 60
	
	if url, ok := node.Data["videoUrl"].(string); ok && url != "" {
		videoURL = strings.Trim(url, " `")
	}
	if cap, ok := node.Data["caption"].(string); ok && cap != "" {
		caption = cap
	}
	if dur, ok := node.Data["duration"].(float64); ok {
		duration = int(dur)
	}

	// Get execution variables for caption replacement
	variables, err := s.chatService.GetExecutionVariables(execution)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}

	// Replace variables in caption
	caption = s.flowService.ReplaceVariables(caption, variables)

	// Move to next node
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		// Continue processing the next node
		nextResponse, err := s.processFlowMessage(flow, execution, userInput)
		if err != nil {
			return fmt.Sprintf("[Video: %s (%ds)] %s", videoURL, duration, caption), nil
		}
		// If next node also has a response, combine them
		if nextResponse != "" {
			return fmt.Sprintf("[Video: %s (%ds)] %s\n%s", videoURL, duration, caption, nextResponse), nil
		}
	} else {
		// End of flow
		s.chatService.CompleteExecution(execution.ID)
	}

	return fmt.Sprintf("[Video: %s (%ds)] %s", videoURL, duration, caption), nil
}

// processDelayNode processes a delay node that waits for specified seconds
func (s *Service) processDelayNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get delay from node data (default to 3 seconds if not specified)
	delay := 3
	if d, ok := node.Data["delay"].(float64); ok {
		delay = int(d)
	} else if d, ok := node.Data["delaySeconds"].(float64); ok {
		delay = int(d)
	}

	// For now, we'll process immediately and note the delay
	// In a production system, this would involve actual delay implementation
	// TODO: Implement actual delay logic using delay (%d seconds)
	_ = delay // Suppress unused variable warning

	// Move to next node immediately (delay would be handled by queue system)
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}

	// End of flow
	s.chatService.CompleteExecution(execution.ID)
	return "", nil // Delay nodes don't return messages
}

// processConditionNode processes a condition node that branches based on user input
func (s *Service) processConditionNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get conditions from node data
	conditions, ok := node.Data["conditions"].([]interface{})
	if !ok {
		// No conditions defined, move to default next node
		nextNode, err := s.flowService.GetNextNode(flow, node.ID)
		if err == nil && nextNode != nil {
			execution.CurrentNode = nextNode.ID
			s.chatService.UpdateExecution(execution)
			return s.processFlowMessage(flow, execution, userInput)
		}
		s.chatService.CompleteExecution(execution.ID)
		return "Thank you for your response!", nil
	}

	// Process conditions to find matching one
	userInputLower := strings.ToLower(strings.TrimSpace(userInput))
	var nextNodeID string
	
	for _, condInterface := range conditions {
		cond, ok := condInterface.(map[string]interface{})
		if !ok {
			continue
		}
		
		condType, _ := cond["type"].(string)
		condValue, _ := cond["value"].(string)
		condNextNodeID, _ := cond["nextNodeId"].(string)
		
		switch condType {
		case "contains":
			if strings.Contains(userInputLower, strings.ToLower(condValue)) {
				nextNodeID = condNextNodeID
				break
			}
		case "equals":
			if userInputLower == strings.ToLower(condValue) {
				nextNodeID = condNextNodeID
				break
			}
		case "default":
			if nextNodeID == "" { // Use as fallback
				nextNodeID = condNextNodeID
			}
		}
	}

	// If we found a matching condition with nextNodeID, use it
	if nextNodeID != "" {
		nextNode, err := s.flowService.FindNodeByID(flow, nextNodeID)
		if err == nil && nextNode != nil {
			execution.CurrentNode = nextNode.ID
			s.chatService.UpdateExecution(execution)
			return s.processFlowMessage(flow, execution, userInput)
		}
	}

	// No matching condition, use default flow
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}

	// End of flow
	s.chatService.CompleteExecution(execution.ID)
	return "Thank you for your response!", nil
}

// processStageNode processes a stage node that sets the current stage
func (s *Service) processStageNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Get stage name from node data
	stageName := "default"
	if stage, ok := node.Data["stageName"].(string); ok && stage != "" {
		stageName = stage
	}

	// Set the stage in execution variables
	err := s.chatService.SetExecutionVariable(execution, "current_stage", stageName)
	if err != nil {
		logrus.WithError(err).Warn("Failed to set stage variable")
	}

	// Log stage transition
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"stage_name":   stageName,
		"node_id":      node.ID,
	}).Info("Stage transition")

	// Move to next node immediately (stage nodes don't send messages)
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}

	// End of flow
	s.chatService.CompleteExecution(execution.ID)
	return "", nil // Stage nodes don't return messages
}

// processStartNode processes a start node that initiates the flow
func (s *Service) processStartNode(flow *models.ChatbotFlow, execution *models.ChatbotExecution, node *models.FlowNode, userInput string) (string, error) {
	// Log flow start
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"flow_id":      flow.ID,
		"node_id":      node.ID,
	}).Info("Flow started")

	// Initialize execution variables if needed
	err := s.chatService.SetExecutionVariable(execution, "flow_started", true)
	if err != nil {
		logrus.WithError(err).Warn("Failed to set flow start variable")
	}

	// Move to next node immediately (start nodes don't send messages)
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		execution.CurrentNode = nextNode.ID
		s.chatService.UpdateExecution(execution)
		return s.processFlowMessage(flow, execution, userInput)
	}

	// If no next node, end flow
	s.chatService.CompleteExecution(execution.ID)
	return "", nil // Start nodes don't return messages
}

// parsePhoneNumber converts a phone number string to WhatsApp JID
func (s *Service) parsePhoneNumber(phoneNumber string) (types.JID, error) {
	// Remove any non-digit characters
	cleanNumber := strings.ReplaceAll(phoneNumber, "+", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, " ", "")

	// Create JID
	return types.NewJID(cleanNumber, types.DefaultUserServer), nil
}

// StartQueueProcessor starts the queue processor for outbound messages
func (s *Service) StartQueueProcessor() {
	go func() {
		for {
			// Process delayed messages
			if s.queueService != nil {
				s.queueService.ProcessDelayedMessages()
			}

			// Process outbound messages
			if s.queueService != nil {
				message, err := s.queueService.DequeueOutboundMessage()
				if err != nil {
					logrus.WithError(err).Error("Failed to dequeue message")
					time.Sleep(5 * time.Second)
					continue
				}

				if message != nil {
					err = s.processQueuedMessage(message)
					if err != nil {
						s.queueService.RequeueFailedMessage(message, err)
					}
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

// processQueuedMessage processes a queued outbound message
func (s *Service) processQueuedMessage(message *services.QueueMessage) error {
	if message.MediaURL != "" && message.MediaType != "" {
		return s.SendMediaMessage(message.PhoneNumber, message.Content, message.MediaURL, message.MediaType)
	}
	return s.SendMessage(message.PhoneNumber, message.Content)
}