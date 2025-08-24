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
	queueService     *services.QueueService
	flowService      *services.FlowService
	aiService        *services.AIService
	aiWhatsappService services.AIWhatsappService
	websocketService *services.WebSocketService
	
	// Performance optimizations
	messageQueue chan *QueuedMessage
	processingWG sync.WaitGroup
	isConnected  bool
}

// NewService creates a new WhatsApp service with multi-device support and performance optimizations
func NewService(cfg *config.Config, queueService *services.QueueService, flowService *services.FlowService, aiService *services.AIService, aiWhatsappService services.AIWhatsappService, websocketService *services.WebSocketService) (*Service, error) {
	service := &Service{
		cfg:              cfg,
		clients:          make(map[string]*whatsmeow.Client),
		containers:       make(map[string]*sqlstore.Container),
		connections:      make(map[string]bool),
		maxDevices:       cfg.WhatsAppMaxDevices,
		storageDir:       cfg.WhatsAppStoragePath,
		queueService:      queueService,
		flowService:       flowService,
		aiService:         aiService,
		aiWhatsappService: aiWhatsappService,
		websocketService:  websocketService,
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
	go s.processIncomingMessage(phoneNumber, content)
}

// processIncomingMessage processes an incoming message through the flow engine
func (s *Service) processIncomingMessage(phoneNumber, content string) {
	// For now, use a default device ID - in a real implementation, you'd determine this based on routing logic
	idDevice := "default"

	// Process message directly through AI WhatsApp service
	response, err := s.aiWhatsappService.ProcessMessage(phoneNumber, idDevice, content)
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
	}
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