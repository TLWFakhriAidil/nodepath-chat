package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// Service handles WhatsApp integration using whatsmeow
type Service struct {
	cfg          *config.Config
	client       *whatsmeow.Client
	chatService  *services.ChatService
	queueService *services.QueueService
	flowService  *services.FlowService
	aiService    *services.AIService
	isConnected  bool
}

// NewService creates a new WhatsApp service
func NewService(cfg *config.Config, chatService *services.ChatService, queueService *services.QueueService) (*Service, error) {
	s := &Service{
		cfg:          cfg,
		chatService:  chatService,
		queueService: queueService,
	}

	// Initialize WhatsApp client
	if err := s.initializeClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize WhatsApp client: %w", err)
	}

	return s, nil
}

// SetServices sets additional services after initialization
func (s *Service) SetServices(flowService *services.FlowService, aiService *services.AIService) {
	s.flowService = flowService
	s.aiService = aiService
}

// initializeClient initializes the whatsmeow client
func (s *Service) initializeClient() error {
	// Ensure storage directory exists
	storageDir := s.cfg.WhatsAppStoragePath
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Initialize SQLite store for session data
	dbPath := filepath.Join(storageDir, "whatsapp.db")
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
	s.client = whatsmeow.NewClient(deviceStore, nil)

	// Add event handlers
	s.client.AddEventHandler(s.handleEvent)

	logrus.Info("WhatsApp client initialized")
	return nil
}

// Connect connects to WhatsApp
func (s *Service) Connect() error {
	if s.client.Store.ID == nil {
		// Generate QR code for pairing
		qrChan, err := s.client.GetQRChannel(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get QR channel: %w", err)
		}

		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					logrus.WithField("qr_code", evt.Code).Info("QR code for WhatsApp pairing")
					// In a real implementation, you'd display this QR code in the web interface
				} else {
					logrus.WithField("event", evt.Event).Info("QR channel event")
				}
			}
		}()
	}

	// Connect to WhatsApp
	err := s.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to WhatsApp: %w", err)
	}

	s.isConnected = true
	logrus.Info("Connected to WhatsApp")
	return nil
}

// Disconnect disconnects from WhatsApp
func (s *Service) Disconnect() {
	if s.client != nil {
		s.client.Disconnect()
		s.isConnected = false
		logrus.Info("Disconnected from WhatsApp")
	}
}

// IsConnected returns the connection status
func (s *Service) IsConnected() bool {
	return s.isConnected && s.client != nil && s.client.IsConnected()
}

// GetQRCode returns the QR code for pairing (if needed)
func (s *Service) GetQRCode() (string, error) {
	if s.client.Store.ID != nil {
		return "", fmt.Errorf("device already paired")
	}

	qrChan, err := s.client.GetQRChannel(context.Background())
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

// SendMessage sends a message to a phone number
func (s *Service) SendMessage(phoneNumber, message string) error {
	if !s.IsConnected() {
		return fmt.Errorf("WhatsApp not connected")
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
	_, err = s.client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	logrus.WithFields(logrus.Fields{
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
	go s.processIncomingMessage(phoneNumber, content)
}

// processIncomingMessage processes an incoming message through the flow engine
func (s *Service) processIncomingMessage(phoneNumber, content string) {
	// For now, use a default staff ID - in a real implementation, you'd determine this based on routing logic
	staffID := "default"

	// Get or create active execution
	execution, err := s.chatService.GetActiveExecution(phoneNumber, staffID)
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
	case models.NodeTypeAIPrompt:
		return s.processAIPromptNode(flow, execution, currentNode, userInput)
	case models.NodeTypeManual:
		return s.processManualNode(flow, execution, currentNode, userInput)
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
	if instance == "" {
		instance = flow.GlobalInstance
	}
	if apiProvider == "" {
		apiProvider = flow.GlobalOpenRouterKey
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