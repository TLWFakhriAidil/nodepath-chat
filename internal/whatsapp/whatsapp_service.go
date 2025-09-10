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
	"nodepath-chat/internal/utils"

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
	urlValidator          *utils.URLValidator

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
		urlValidator:          utils.NewURLValidator(),
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
// Now includes URL validation to prevent sending broken links
func (s *Service) SendMediaMessage(deviceID, phoneNumber, mediaURL string) error {
	// Console log for tracing media URL extraction
	logrus.WithFields(logrus.Fields{
		"device_id":        deviceID,
		"phone_number":     phoneNumber,
		"media_url":        mediaURL,
		"media_url_length": len(mediaURL),
		"media_url_preview": func() string {
			if len(mediaURL) > 100 {
				return mediaURL[:100] + "..."
			}
			return mediaURL
		}(),
	}).Info("📤 MEDIA: Sending media message - URL EXTRACTED FOR TRACING")

	// Validate URL before sending to prevent 404 errors
	isValid, mediaType, validationErr := s.urlValidator.ValidateMediaURL(mediaURL)
	if !isValid {
		logrus.WithError(validationErr).WithFields(logrus.Fields{
			"device_id":    deviceID,
			"phone_number": phoneNumber,
			"media_url":    mediaURL,
		}).Warn("❌ MEDIA: URL validation failed, sending fallback message instead")

		// Send fallback text message instead of broken media URL
		fallbackMessage := fmt.Sprintf("Sorry, the media content is currently unavailable. Please try again later.\n\nOriginal URL: %s", mediaURL)
		return s.SendMessageFromDevice(deviceID, phoneNumber, fallbackMessage)
	}

	logrus.WithFields(logrus.Fields{
		"device_id":  deviceID,
		"media_url":  mediaURL,
		"media_type": mediaType,
	}).Info("✅ MEDIA: URL validation successful, proceeding with media send")

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
	}).Info("🔍 FLOW: Checking for active execution in ai_whatsapp_nodepath")

	// Check for personal commands (%, #, cmd)
	if strings.HasPrefix(content, "%") || strings.HasPrefix(content, "#") || strings.HasPrefix(content, "cmd") {
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"command":   content,
		}).Info("🔧 COMMAND: Personal command detected")
		return s.handlePersonalCommand(phoneNumber, content, deviceID)
	}

	// Get or create active execution from ai_whatsapp_nodepath
	aiExecution, err := s.aiWhatsappService.GetActiveFlowExecution(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get active execution from ai_whatsapp_nodepath")
		return err
	}

	if aiExecution == nil {
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
			}).Info("⚠️ FLOW: No default flow found for device, falling back to AI conversation")

			// Fallback to AI conversation when no flow is configured
			return s.processAIConversation(phoneNumber, content, deviceID)
		}

		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
			"flow_id":      defaultFlow.ID,
			"flow_name":    defaultFlow.Name,
		}).Info("🚀 FLOW: Starting new execution with default flow in ai_whatsapp_nodepath")

		// Start new execution with default flow in ai_whatsapp_nodepath
		variables := make(map[string]interface{})
		aiExecution, err = s.aiWhatsappService.StartFlowExecution(phoneNumber, deviceID, defaultFlow.ID, variables)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to start new execution in ai_whatsapp_nodepath")
			return err
		}

		logrus.WithFields(logrus.Fields{
			"execution_id": aiExecution.ExecutionID.String,
			"flow_id":      defaultFlow.ID,
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("✅ FLOW: New execution started successfully in ai_whatsapp_nodepath")

		// Process the new execution through the flow
		return s.processNewFlowExecution(aiExecution, content, phoneNumber, deviceID)
	} else {
		logrus.WithFields(logrus.Fields{
			"execution_id":   aiExecution.ExecutionID.String,
			"flow_reference": aiExecution.FlowReference.String,
			"phone_number":   phoneNumber,
			"device_id":      deviceID,
			"current_node":   aiExecution.CurrentNode.String,
		}).Info("🔄 FLOW: Found existing active execution in ai_whatsapp_nodepath")

		// Check if the execution is waiting for user reply
		if aiExecution.WaitingForReply.Valid && aiExecution.WaitingForReply.Int32 == 1 {
			logrus.WithFields(logrus.Fields{
				"execution_id":    aiExecution.ExecutionID.String,
				"current_node_id": aiExecution.CurrentNodeID.String,
				"flow_id":         aiExecution.FlowID.String,
				"user_input":      content,
			}).Info("💬 USER_REPLY: Processing user reply for waiting execution")

			// Handle the user reply and resume flow from the correct node
			return s.handleUserReplyResume(aiExecution, content)
		} else {
			// Execution exists but not waiting for reply - this means the flow is already completed or in progress
			// We should not restart the flow, instead fall back to AI conversation
			// Note: AI conversation will handle its own conversation saving to prevent duplicates
			logrus.WithFields(logrus.Fields{
				"execution_id":      aiExecution.ExecutionID.String,
				"current_node_id":   aiExecution.CurrentNodeID.String,
				"waiting_for_reply": aiExecution.WaitingForReply.Int32,
				"user_input":        content,
			}).Info("ℹ️ FLOW: Existing execution not waiting for reply, falling back to AI conversation")

			// Fall back to AI conversation instead of restarting flow
			// AI conversation will handle conversation saving internally
			return s.processAIConversation(phoneNumber, content, deviceID)
		}
	}

	return nil
}

// processNewFlowExecution handles flow processing for new executions only
// This function contains the logic that was previously running for both new and existing executions
// Fixed: Consolidated conversation saving to prevent duplicate entries
func (s *Service) processNewFlowExecution(aiExecution *models.AIWhatsapp, content, phoneNumber, deviceID string) error {
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

	// Save user message to conversation history (single save point for user input)
	logrus.WithFields(logrus.Fields{
		"execution_id": aiExecution.IDProspect,
		"message_type": "USER",
		"content":      content,
	}).Info("💬 FLOW: Adding user message to ai_whatsapp_nodepath")

	err = s.aiWhatsappService.SaveConversationHistory(phoneNumber, deviceID, content, "", "")
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to add user message to ai_whatsapp_nodepath")
		return err
	}

	logrus.WithField("execution_id", aiExecution.IDProspect).Info("✅ FLOW: User message added to conversation successfully")

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

	// Only send response if it's not empty and not just whitespace
	// This prevents sending <nil> messages when Advanced AI Prompt nodes
	// have already sent their individual response items
	if response != "" && strings.TrimSpace(response) != "" {
		logrus.WithFields(logrus.Fields{
			"phone_number":    phoneNumber,
			"device_id":       deviceID,
			"response":        response,
			"response_length": len(response),
		}).Info("📤 FLOW: Sending response back to user")

	// Skip sending if response is empty (already handled by advanced AI nodes)
	if response == "" {
		logrus.WithFields(logrus.Fields{
			"device_id":    deviceID,
			"phone_number": phoneNumber,
		}).Info("🔇 FLOW: Skipping empty response to prevent <nil> message")
	} else {
		// Process AI response using PHP-compatible logic
		stage, messages, err := services.ProcessAIResponsePHP(response, 2000) // 2 second delay
		if err != nil {
			logrus.WithError(err).Error("Failed to process AI response")
			// Fallback to sending as plain text
			err = s.SendMessageFromDevice(deviceID, phoneNumber, response)
			if err != nil {
				logrus.WithError(err).Error("❌ FLOW: Failed to send response message")
				return err
			}
			// Save the fallback response
			err = s.aiWhatsappService.SaveConversationHistory(phoneNumber, deviceID, "", response, "")
			if err != nil {
				logrus.WithError(err).Error("❌ FLOW: Failed to save fallback response to conversation")
			}
		} else {
			// Save the stage if we got one
			if stage != "" {
				logrus.WithFields(logrus.Fields{
					"phone_number": phoneNumber,
					"device_id":    deviceID,
					"stage":        stage,
				}).Info("📋 FLOW: Saving AI stage to database")
				
				// Update the stage in ai_whatsapp_nodepath
				err = s.aiWhatsappService.UpdateStage(phoneNumber, deviceID, stage)
				if err != nil {
					logrus.WithError(err).WithField("stage", stage).Error("❌ FLOW: Failed to update stage")
				}
			}
			// Send each processed message and save EACH ONE separately
			for i, msg := range messages {
				logrus.WithFields(logrus.Fields{
					"index":          i,
					"type":           msg.Type,
					"content_length": len(msg.Content),
				}).Info("📤 FLOW: Sending processed message")

				// Send the message
				sendSuccess := false
				if msg.Type == "text" {
					err = s.SendMessageFromDevice(deviceID, phoneNumber, msg.Content)
					if err != nil {
						logrus.WithError(err).Error("❌ FLOW: Failed to send text message")
					} else {
						sendSuccess = true
					}
				} else if msg.Type == "image" || msg.Type == "audio" || msg.Type == "video" {
					err = s.SendMediaMessage(deviceID, phoneNumber, msg.Content)
					if err != nil {
						logrus.WithError(err).WithFields(logrus.Fields{
							"media_url":  msg.Content,
							"media_type": msg.Type,
						}).Error("❌ FLOW: Failed to send media message")
					} else {
						sendSuccess = true
					}
				}
				
				// Save EACH message to conversation history separately
				// Format the save based on message type to match PHP behavior
				if sendSuccess {
					var saveContent string
					
					// Format based on type (matching PHP format)
					if msg.Type == "text" {
						// For text, save as-is
						saveContent = msg.Content
					} else {
						// For media (image/video/audio), save just the URL
						saveContent = msg.Content
					}
					
					err = s.aiWhatsappService.SaveConversationHistory(phoneNumber, deviceID, "", saveContent, stage)
					if err != nil {
						logrus.WithError(err).WithFields(logrus.Fields{
							"type": msg.Type,
							"content": saveContent,
						}).Error("❌ FLOW: Failed to save message to conversation")
					} else {
						logrus.WithFields(logrus.Fields{
							"type": msg.Type,
							"saved": saveContent,
						}).Debug("✅ FLOW: Saved message to conversation")
					}
				}

				// Add delay between messages
				if i < len(messages)-1 && msg.Delay > 0 {
					time.Sleep(msg.Delay)
				}
			}
		}
	}
	
	// Continue with execution tracking
	if response == "" {
		logrus.WithField("execution_id", aiExecution.IDProspect).Info("ℹ️ FLOW: No response generated from flow processing (Advanced AI nodes handle their own message sending)")
	}

		// Create AI WhatsApp record as fallback when no flow response is generated
		// Note: Conversation history was already saved above for user input
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
			// Note: User message was already saved above, no need to save again
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
			// Existing record found - conversation history was already saved above
			// No need to save again to prevent duplicates
			logrus.WithFields(logrus.Fields{
				"phone_number": phoneNumber,
				"device_id":    deviceID,
				"stage":        existingRecord.Stage,
			}).Info("✅ FLOW: Using existing AI WhatsApp record, conversation already saved")
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

// sendAIResponse sends AI response with multiple message types (text, images, audio, and video)
// Implements PHP onemessage combining logic for text parts with Jenis="onemessage"
func (s *Service) sendAIResponse(phoneNumber, deviceID string, response *services.AIWhatsappResponse) error {
	logrus.WithFields(logrus.Fields{
		"device_id":      deviceID,
		"phone_number":   phoneNumber,
		"stage":          response.Stage,
		"response_count": len(response.Response),
	}).Info("📤 AI: Sending AI response with onemessage combining logic")

	// Variables for onemessage combining logic (from PHP implementation)
	textParts := []string{}
	isOnemessageActive := false
	delayMs := 5000 // 5 second delay between messages

	// Process each response part with PHP-equivalent logic
	for index, part := range response.Response {
		// Validate response part structure
		if part.Type == "" || part.Content == "" {
			logrus.WithFields(logrus.Fields{
				"index": index,
				"part":  part,
			}).Warn("Invalid response part structure, skipping")
			continue
		}

		// Handle text type with "Jenis"="onemessage" combining logic
		if part.Type == "text" && part.Jenis == "onemessage" {
			// Start collecting text parts
			textParts = append(textParts, part.Content)
			isOnemessageActive = true

			// Check if next part isn't also onemessage, then send combined
			nextIsOnemessage := false
			if index+1 < len(response.Response) {
				nextPart := response.Response[index+1]
				if nextPart.Jenis == "onemessage" {
					nextIsOnemessage = true
				}
			}

			if !nextIsOnemessage {
				// Send combined message
				combinedMessage := strings.Join(textParts, "\n")
				err := s.SendMessageFromDevice(deviceID, phoneNumber, combinedMessage)
				if err != nil {
					logrus.WithError(err).Error("Failed to send combined onemessage")
					return err
				}

				// Log conversation with BOT_COMBINED format
				err = s.logConversationMessage(phoneNumber, deviceID, "BOT_COMBINED", combinedMessage)
				if err != nil {
					logrus.WithError(err).Error("Failed to log combined conversation")
				}

				// Reset temporary variables
				textParts = []string{}
				isOnemessageActive = false

				// Add delay
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
			}
		} else {
			// If we just finished onemessage sequence, send combined first
			if isOnemessageActive {
				combinedMessage := strings.Join(textParts, "\n")
				err := s.SendMessageFromDevice(deviceID, phoneNumber, combinedMessage)
				if err != nil {
					logrus.WithError(err).Error("Failed to send combined onemessage before other type")
					return err
				}

				// Log conversation with BOT_COMBINED format
				err = s.logConversationMessage(phoneNumber, deviceID, "BOT_COMBINED", combinedMessage)
				if err != nil {
					logrus.WithError(err).Error("Failed to log combined conversation")
				}

				// Reset variables
				textParts = []string{}
				isOnemessageActive = false

				// Add delay
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
			}

			// Now handle normal text or media
			switch part.Type {
			case "text":
				// Send regular text message
				err := s.SendMessageFromDevice(deviceID, phoneNumber, part.Content)
				if err != nil {
					logrus.WithError(err).WithField("index", index).Error("Failed to send text message")
					return err
				}

				// Log conversation with BOT format
				err = s.logConversationMessage(phoneNumber, deviceID, "BOT", part.Content)
				if err != nil {
					logrus.WithError(err).Error("Failed to log text conversation")
				}

				// Add delay
				time.Sleep(time.Duration(delayMs) * time.Millisecond)

			case "image":
				// Send image message - part.Content contains the image URL
				currentImageURL := strings.TrimSpace(part.Content)
				err := s.SendMediaMessage(deviceID, phoneNumber, currentImageURL)
				if err != nil {
					logrus.WithError(err).WithField("index", index).Error("Failed to send image message")
					return err
				}

				// Log conversation with BOT format for image
				err = s.logConversationMessage(phoneNumber, deviceID, "BOT", currentImageURL)
				if err != nil {
					logrus.WithError(err).Error("Failed to log image conversation")
				}

				// Add delay
				time.Sleep(time.Duration(delayMs) * time.Millisecond)

			case "audio":
				// Send audio message - part.Content contains the audio URL
				err := s.SendMediaMessage(deviceID, phoneNumber, part.Content)
				if err != nil {
					logrus.WithError(err).WithField("index", index).Error("Failed to send audio message")
					return err
				}

				// Log conversation with BOT format for audio
				err = s.logConversationMessage(phoneNumber, deviceID, "BOT", part.Content)
				if err != nil {
					logrus.WithError(err).Error("Failed to log audio conversation")
				}

				// Add delay
				time.Sleep(time.Duration(delayMs) * time.Millisecond)

			case "video":
				// Send video message - part.Content contains the video URL
				err := s.SendMediaMessage(deviceID, phoneNumber, part.Content)
				if err != nil {
					logrus.WithError(err).WithField("index", index).Error("Failed to send video message")
					return err
				}

				// Log conversation with BOT format for video
				err = s.logConversationMessage(phoneNumber, deviceID, "BOT", part.Content)
				if err != nil {
					logrus.WithError(err).Error("Failed to log video conversation")
				}

				// Add delay
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
			}
		}
	}

	return nil
}

// logConversationMessage logs conversation messages with proper format (BOT, BOT_COMBINED)
// This function handles conversation logging similar to PHP implementation
// Updates conv_last field in database and clears conv_current
func (s *Service) logConversationMessage(phoneNumber, deviceID, messageType, content string) error {
	// Create log entry with proper format matching PHP implementation
	var logEntry string
	if messageType == "BOT_COMBINED" {
		// For combined messages, use JSON encoding like PHP
		contentJSON, err := json.Marshal(content)
		if err != nil {
			logrus.WithError(err).Error("Failed to marshal content for BOT_COMBINED")
			logEntry = fmt.Sprintf("%s: %s", messageType, content)
		} else {
			logEntry = fmt.Sprintf("%s: %s", messageType, string(contentJSON))
		}
	} else {
		// For regular BOT messages, use JSON encoding like PHP
		contentJSON, err := json.Marshal(content)
		if err != nil {
			logrus.WithError(err).Error("Failed to marshal content for BOT")
			logEntry = fmt.Sprintf("%s: %s", messageType, content)
		} else {
			logEntry = fmt.Sprintf("%s: %s", messageType, string(contentJSON))
		}
	}

	logrus.WithFields(logrus.Fields{
		"device_id":      deviceID,
		"phone_number":   phoneNumber,
		"message_type":   messageType,
		"content_length": len(content),
		"log_entry":      logEntry,
	}).Info("💬 CONVERSATION: Logging message to database")

	// Get existing conversation to append to conv_last
	existingConv, err := s.aiWhatsappService.GetAIWhatsappByProspectAndDevice(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get existing conversation for logging")
		return err
	}

	if existingConv != nil {
		// Update conv_last by appending new log entry (similar to PHP: $whats->conv_last .= "\n" . $newBotEntry)
		var updatedConvLast string
		if existingConv.ConvLast != nil {
			// Get existing conv_last as string
			existingConvLastStr := string(existingConv.ConvLast)
			// Remove JSON quotes if present
			if len(existingConvLastStr) >= 2 && existingConvLastStr[0] == '"' && existingConvLastStr[len(existingConvLastStr)-1] == '"' {
				existingConvLastStr = existingConvLastStr[1 : len(existingConvLastStr)-1]
				// Unescape JSON escape sequences
				existingConvLastStr = strings.ReplaceAll(existingConvLastStr, "\\n", "\n")
				existingConvLastStr = strings.ReplaceAll(existingConvLastStr, "\\\\", "\\")
				existingConvLastStr = strings.ReplaceAll(existingConvLastStr, "\\\"", "\"")
			}
			updatedConvLast = existingConvLastStr + "\n" + logEntry
		} else {
			updatedConvLast = logEntry
		}

		// Update the conversation record with new conv_last and clear conv_current
		existingConv.ConvLast = json.RawMessage(fmt.Sprintf("\"%s\"", strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(updatedConvLast, "\\", "\\\\"), "\"", "\\\""), "\n", "\\n")))
		existingConv.ConvCurrent = sql.NullString{Valid: false} // Clear conv_current (similar to PHP: $whats->conv_current = null)

		// Save the updated conversation
		err = s.aiWhatsappService.UpdateAIWhatsapp(existingConv)
		if err != nil {
			logrus.WithError(err).Error("Failed to update conversation with new log entry")
			return err
		}

		logrus.WithFields(logrus.Fields{
			"device_id":    deviceID,
			"phone_number": phoneNumber,
			"message_type": messageType,
		}).Info("✅ CONVERSATION: Successfully logged message to database")
	} else {
		logrus.WithFields(logrus.Fields{
			"device_id":    deviceID,
			"phone_number": phoneNumber,
		}).Warn("No existing conversation found for logging")
	}

	return nil
}

// processFlowMessage processes a message through the flow logic
func (s *Service) processFlowMessage(flow *models.ChatbotFlow, aiExecution *models.AIWhatsapp, userInput string) (string, error) {
	// Get current node using new flow tracking field
	var currentNodeID string
	if aiExecution.CurrentNodeID.Valid && aiExecution.CurrentNodeID.String != "" {
		currentNodeID = aiExecution.CurrentNodeID.String
	} else {
		// Fallback to legacy field for backward compatibility
		if aiExecution.CurrentNode.Valid && aiExecution.CurrentNode.String != "" {
			currentNodeID = aiExecution.CurrentNode.String
		}
	}

	currentNode, err := s.flowService.FindNodeByID(flow, currentNodeID)
	if err != nil {
		// If no current node, start from the beginning
		currentNode, err = s.flowService.GetStartNode(flow)
		if err != nil {
			return "", fmt.Errorf("failed to get start node: %w", err)
		}
		// Update both new and legacy fields
		s.updateCurrentNode(aiExecution, currentNode.ID)
	}

	// Process based on node type
	switch currentNode.Type {
	case models.NodeTypeStart:
		return s.processStartNode(flow, aiExecution, currentNode, userInput)
	case models.NodeTypeAIPrompt, models.NodeTypeAdvancedAIPrompt, "prompt": // Handle all AI prompt types with one function
		return s.processAIPromptNode(flow, aiExecution, currentNode, userInput)

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

	default:
		return s.processDefaultNode(flow, aiExecution, currentNode, userInput)
	}
}

// processAIPromptNode processes all types of AI prompt nodes (ai_prompt, advanced_ai_prompt, prompt)
// This is the SINGLE standardized function for ALL AI processing nodes
func (s *Service) processAIPromptNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"node_id":      node.ID,
		"node_type":    node.Type,
		"user_input":   userInput,
		"prospect_num": execution.ProspectNum,
		"id_device":    execution.IDDevice,
	}).Info("🤖 AI_PROMPT: Processing AI prompt node (standardized)")

	// Get AI configuration from node data
	var systemPrompt, instance, apiProvider string

	// Check node data for configuration - handle both camelCase and snake_case
	if sp, ok := node.Data["system_prompt"].(string); ok {
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

	// Use global settings as final fallback
	if apiProvider == "" {
		apiProvider = flow.Niche
	}

	logrus.WithFields(logrus.Fields{
		"system_prompt_length": len(systemPrompt),
		"instance":             instance,
		"api_provider":         apiProvider,
	}).Info("🤖 AI_PROMPT: Configuration loaded")

	// Check if we have complete AI configuration
	if systemPrompt == "" {
		logrus.Error("🤖 AI_PROMPT: No system prompt configured")
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}
	if instance == "" {
		logrus.Error("🤖 AI_PROMPT: No instance configured")
		return "I'm sorry, I'm not configured to handle this request. Please contact support.", nil
	}
	if apiProvider == "" {
		logrus.Error("🤖 AI_PROMPT: No API provider configured")
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

	// STANDARDIZED: Add the standardized format instructions for ALL AI nodes
	// This ensures consistent response format across all AI prompt types
	systemPrompt = systemPrompt + "\n\n" +
		"### Instructions:\n" +
		"1. If the current stage is null or undefined, default to the first stage.\n" +
		"2. Always analyze the user's input to determine the appropriate stage. If the input context is unclear, guide the user within the default stage context.\n" +
		"3. Follow all rules and steps strictly. Do not skip or ignore any rules or instructions.\n\n" +
		"4. **Do not repeat the same sentences or phrases that have been used in the recent conversation history.**\n" +
		"5. If the input contains the phrase \"I want this section in add response format [onemessage]\":\n" +
		"   - Add the `Jenis` field with the value `onemessage` at the item level for each text response.\n" +
		"   - The `Jenis` field is only added to `text` types within the `Response` array.\n" +
		"   - If the directive is not present, omit the `Jenis` field entirely.\n\n" +
		"### Response Format:\n" +
		"{\n" +
		"  \"Stage\": \"[Stage]\",  // Specify the current stage explicitly.\n" +
		"  \"Response\": [\n" +
		"    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Provide the first response message here.\"},\n" +
		"    {\"type\": \"image\", \"content\": \"https://example.com/image1.jpg\"},\n" +
		"    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Provide the second response message here.\"}\n" +
		"  ]\n" +
		"}\n\n" +
		"### Example Response:\n" +
		"// If the directive is present\n" +
		"{\n" +
		"  \"Stage\": \"Problem Identification\",\n" +
		"  \"Response\": [\n" +
		"    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni.\"},\n" +
		"    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Kurang selera makan, sembelit, atau kerap demam?\"}\n" +
		"  ]\n" +
		"}\n\n" +
		"// If the directive is NOT present\n" +
		"{\n" +
		"  \"Stage\": \"Problem Identification\",\n" +
		"  \"Response\": [\n" +
		"    {\"type\": \"text\", \"content\": \"Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni.\"},\n" +
		"    {\"type\": \"text\", \"content\": \"Kurang selera makan, sembelit, atau kerap demam?\"}\n" +
		"  ]\n" +
		"}\n\n" +
		"### Important Rules:\n" +
		"1. **Include the `Stage` field in every response**:\n" +
		"   - The `Stage` field must explicitly specify the current stage.\n" +
		"   - If the stage is unclear or missing, default to first stage.\n\n" +
		"2. **Use the Correct Response Format**:\n" +
		"   - Divide long responses into multiple short \"text\" segments for better readability.\n" +
		"   - Include all relevant images provided in the input, interspersed naturally with text responses.\n" +
		"   - If multiple images are provided, create separate `image` entries for each.\n\n" +
		"3. **Dynamic Field for [onemessage]**:\n" +
		"   - If the input specifies \"I want this section in add response format [onemessage]\":\n" +
		"      - Add `\"Jenis\": \"onemessage\"` to each `text` type in the `Response` array.\n" +
		"   - If the directive is not present, omit the `Jenis` field entirely.\n" +
		"   - Non-text types like `image` never include the `Jenis` field.\n\n"

	// Get actual API key from device settings
	var actualAPIKey string
	if deviceSettings != nil && deviceSettings.APIKey.Valid {
		actualAPIKey = deviceSettings.APIKey.String
	}

	// Get conversation history
	var conversationHistory []models.ConversationMessage
	if len(execution.ConvLast) > 0 {
		var convLastStr string
		if err := json.Unmarshal(execution.ConvLast, &convLastStr); err == nil {
			// Successfully unmarshaled
		} else {
			// If unmarshal fails, use it as string
			convLastStr = string(execution.ConvLast)
		}
		// Remove quotes if present
		convLastStr = strings.Trim(convLastStr, "\"")
		
		if convLastStr != "" && convLastStr != "null" {
			conversationHistory = append(conversationHistory, models.ConversationMessage{
				Role:    "assistant",
				Content: convLastStr,
			})
		}
	}

	// Call AI service with configuration
	response, err := s.aiService.GenerateResponse(
		systemPrompt,
		userInput,
		actualAPIKey,
		execution.IDDevice,
		conversationHistory,
	)
	if err != nil {
		logrus.WithError(err).Error("🤖 AI_PROMPT: Failed to generate AI response")
		return "I'm sorry, I couldn't process your request. Please try again later.", nil
	}

	logrus.WithFields(logrus.Fields{
		"response_length": len(response),
		"node_type":       node.Type,
	}).Info("🤖 AI_PROMPT: AI response generated successfully")

	// For advanced_ai_prompt nodes, parse the response and handle it
	if node.Type == models.NodeTypeAdvancedAIPrompt || node.Type == "advanced_ai_prompt" {
		// Parse the AI response JSON for advanced nodes
		parsedResponse, err := s.aiWhatsappService.ParseAIResponse(response)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"raw_response": response,
				"node_id":      node.ID,
			}).Warn("Failed to parse JSON response, treating as plain text")
			// Fallback to plain text if JSON parsing fails
		} else if parsedResponse != nil {
			// Successfully parsed JSON response - handle multiple response items
			logrus.WithFields(logrus.Fields{
				"stage":          parsedResponse.Stage,
				"response_count": len(parsedResponse.Response),
				"node_id":        node.ID,
			}).Info("Successfully parsed JSON response with multiple items")

			// Update stage if provided
			if parsedResponse.Stage != "" {
				execution.Stage.String = parsedResponse.Stage
				execution.Stage.Valid = true
				if err := s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active"); err != nil {
					logrus.WithError(err).Warn("Failed to update execution stage")
				}
			}

			// Send individual messages from parsed response
			if len(parsedResponse.Response) > 0 {
				for i, item := range parsedResponse.Response {
					if i > 0 {
						time.Sleep(2 * time.Second) // Add delay between messages
					}

					switch item.Type {
					case "text":
						err := s.SendMessageFromDevice(execution.ProspectNum, item.Content, execution.IDDevice)
						if err != nil {
							logrus.WithError(err).Error("Failed to send text message")
						}
					case "image", "audio", "video":
						err := s.SendMediaMessage(execution.ProspectNum, item.Content, execution.IDDevice)
						if err != nil {
							logrus.WithError(err).WithFields(logrus.Fields{
								"media_type": item.Type,
								"media_url":  item.Content,
							}).Error("Failed to send media message")
						}
					default:
						logrus.WithField("type", item.Type).Warn("Unknown response type")
					}
				}
				// Return empty string since we've already sent the messages
				return "", nil
			}
		}
	}

	// Handle the next node advancement
	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err != nil || nextNode == nil {
		logrus.WithError(err).Warn("No next node found after AI prompt")
		// Mark execution as completed
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "completed")
		if err != nil {
			logrus.WithError(err).Error("Failed to complete flow execution")
		}
		return response, nil
	}

	// Check if the next node is a delay node
	if nextNode.Type == models.NodeTypeDelay {
		logrus.WithFields(logrus.Fields{
			"prospect_id":  execution.IDProspect,
			"current_node": node.ID,
			"next_node":    nextNode.ID,
			"next_type":    nextNode.Type,
		}).Info("🔄 AI_PROMPT: Response sent, advancing to delay node")

		// Update execution to delay node
		execution.CurrentNode.String = nextNode.ID
		execution.CurrentNode.Valid = true
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution to delay node")
		}

		// Process delay node to schedule next message
		_, err = s.processDelayNode(flow, execution, nextNode, userInput)
		if err != nil {
			logrus.WithError(err).Error("Failed to process delay node after AI prompt")
		}

		return response, nil
	}

	// Update execution to the next node
	execution.CurrentNode.String = nextNode.ID
	execution.CurrentNode.Valid = true
	s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, nextNode.ID, make(map[string]interface{}), "active")

	return response, nil
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
				"prospect_id":  execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("📤 MESSAGE: Message sent, advancing to delay node")

			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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

	// Console log for tracing image URL extraction
	logrus.WithFields(logrus.Fields{
		"node_id":       node.ID,
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
	}).Info("🔍 IMAGE NODE: RAW URL EXTRACTED FOR TRACING")

	// Replace variables in image URL
	variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}
	imageURL = s.flowService.ReplaceVariables(imageURL, variables)

	// Console log for tracing processed image URL
	logrus.WithFields(logrus.Fields{
		"node_id":             node.ID,
		"processed_image_url": imageURL,
		"variables_count":     len(variables),
	}).Info("🔍 IMAGE NODE: PROCESSED URL EXTRACTED FOR TRACING")

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
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		}).Info("🏁 IMAGE: End of flow reached, completing execution")
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
		"node_id":       node.ID,
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
	}).Info("🔍 AUDIO NODE: RAW URL EXTRACTED FOR TRACING")

	// Replace variables in audio URL
	variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}
	audioURL = s.flowService.ReplaceVariables(audioURL, variables)

	// Console log for tracing processed audio URL
	logrus.WithFields(logrus.Fields{
		"node_id":             node.ID,
		"processed_audio_url": audioURL,
		"variables_count":     len(variables),
	}).Info("🔍 AUDIO NODE: PROCESSED URL EXTRACTED FOR TRACING")

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
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		}).Info("🏁 AUDIO: End of flow reached, completing execution")
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
		"node_id":       node.ID,
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
	}).Info("🔍 VIDEO NODE: RAW URL EXTRACTED FOR TRACING")

	// Replace variables in video URL
	variables, err := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get execution variables")
		variables = make(map[string]interface{})
	}
	videoURL = s.flowService.ReplaceVariables(videoURL, variables)

	// Console log for tracing processed video URL
	logrus.WithFields(logrus.Fields{
		"node_id":             node.ID,
		"processed_video_url": videoURL,
		"variables_count":     len(variables),
	}).Info("🔍 VIDEO NODE: PROCESSED URL EXTRACTED FOR TRACING")

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
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		}).Info("🏁 VIDEO: End of flow reached, completing execution")
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
	}).Info("🕐 DELAY: Processing delay node")

	// Get delay time from node data (default to 5 seconds if not specified)
	delaySeconds := 5
	if delay, ok := node.Data["delay"].(float64); ok {
		delaySeconds = int(delay)
	} else if delay, ok := node.Data["delaySeconds"].(float64); ok {
		delaySeconds = int(delay)
	}

	logrus.WithFields(logrus.Fields{
		"execution_id":  execution.IDProspect,
		"delay_seconds": delaySeconds,
		"phone_number":  execution.ProspectNum,
		"device_id":     execution.IDDevice,
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
		"execution_id":  execution.IDProspect,
		"current_node":  node.ID,
		"next_node":     nextNode.ID,
		"delay_seconds": delaySeconds,
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
		"execution_id":  execution.IDProspect,
		"message_id":    delayedMessage.ID,
		"delay_seconds": delaySeconds,
		"next_node_id":  nextNode.ID,
	}).Info("🕐 DELAY: Message queued successfully for delayed processing")

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
				"prospect_id":  execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("🔀 CONDITION: Condition evaluated, advancing to delay node")

			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		"node_id":      node.ID,
		"stage":        node.Data["stage"],
	}).Info("🎯 STAGE: Stage transition node processed")

	nextNode, err := s.flowService.GetNextNode(flow, node.ID)
	if err == nil && nextNode != nil {
		if nextNode.Type == models.NodeTypeDelay {
			// Advance to delay node and process it immediately
			// This ensures the delay is scheduled properly
			logrus.WithFields(logrus.Fields{
				"prospect_id":  execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("🎯 STAGE: Stage processed, advancing to delay node")

			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution after stage node")
			return "", err
		}

		// Recursively process the next node if it's not a delay
		return s.processFlowMessage(flow, execution, userInput)
	}
	return "", nil
}

// handleUserReplyResume handles user reply when execution is waiting and resumes flow
func (s *Service) handleUserReplyResume(execution *models.AIWhatsapp, userInput string) error {
	// Get the flow data
	flow, err := s.flowService.GetFlow(execution.FlowID.String)
	if err != nil {
		logrus.WithError(err).Error("❌ USER_REPLY: Failed to get flow for resume")
		return err
	}

	if flow == nil {
		logrus.WithField("flow_id", execution.FlowID.String).Error("❌ USER_REPLY: Flow not found for resume")
		return fmt.Errorf("flow not found for resume")
	}

	// Validate that we have a valid current node ID
	if !execution.CurrentNodeID.Valid || execution.CurrentNodeID.String == "" {
		logrus.Error("❌ USER_REPLY: Invalid current node ID for resume")
		return fmt.Errorf("invalid current node ID for resume")
	}

	// Save user message to conversation history
	err = s.aiWhatsappService.SaveConversationHistory(execution.ProspectNum, execution.IDDevice, userInput, "", "")
	if err != nil {
		logrus.WithError(err).Error("❌ USER_REPLY: Failed to save user message to conversation")
		return err
	}

	// Get the next node after the user_reply node
	nextNode, err := s.flowService.GetNextNode(flow, execution.CurrentNodeID.String)
	if err != nil {
		logrus.WithError(err).Error("❌ USER_REPLY: Failed to get next node after user reply")
		return err
	}

	if nextNode == nil {
		logrus.WithField("current_node_id", execution.CurrentNodeID.String).Info("🏁 USER_REPLY: No next node found, completing flow")

		// Clear waiting state and complete flow
		err = s.updateFlowTrackingFields(execution, execution.CurrentNodeID.String, execution.FlowID.String, false)
		if err != nil {
			logrus.WithError(err).Error("Failed to clear waiting state")
			return err
		}

		// Complete the flow execution
		return s.aiWhatsappService.CompleteFlowExecution(execution.ProspectNum, execution.IDDevice)
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": execution.ExecutionID.String,
		"current_node": execution.CurrentNodeID.String,
		"next_node":    nextNode.ID,
		"next_type":    nextNode.Type,
		"user_input":   userInput,
	}).Info("🔄 USER_REPLY: Resuming flow execution from next node")

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
		logrus.WithError(err).Error("❌ USER_REPLY: Failed to process next node after user reply")
		return err
	}

	// Send response if there is one
	if response != "" {
		logrus.WithFields(logrus.Fields{
			"execution_id":    execution.ExecutionID.String,
			"response_length": len(response),
		}).Info("📤 USER_REPLY: Sending response after flow resume")

		// Process AI response using PHP-compatible logic
		stage, messages, err := services.ProcessAIResponsePHP(response, 2000) // 2 second delay
		if err != nil {
			logrus.WithError(err).Error("Failed to process AI response")
			// Fallback to sending as plain text
			err = s.SendMessageFromDevice(execution.IDDevice, execution.ProspectNum, response)
			if err != nil {
				logrus.WithError(err).Error("❌ USER_REPLY: Failed to send response after resume")
				return err
			}
			// Save fallback response to conversation
			err = s.aiWhatsappService.SaveConversationHistory(execution.ProspectNum, execution.IDDevice, "", response, "")
			if err != nil {
				logrus.WithError(err).Error("❌ USER_REPLY: Failed to save bot response to conversation")
			}
		} else {
			// Save the stage if we got one
			if stage != "" {
				logrus.WithFields(logrus.Fields{
					"execution_id": execution.ExecutionID.String,
					"stage":        stage,
				}).Info("📋 USER_REPLY: Saving AI stage to database")
				
				// Update the stage in ai_whatsapp_nodepath
				err = s.aiWhatsappService.UpdateStage(execution.ProspectNum, execution.IDDevice, stage)
				if err != nil {
					logrus.WithError(err).WithField("stage", stage).Error("❌ USER_REPLY: Failed to update stage")
				}
			}
			// Send each processed message and save EACH ONE to conversation history
			for i, msg := range messages {
				logrus.WithFields(logrus.Fields{
					"index":          i,
					"type":           msg.Type,
					"content_length": len(msg.Content),
				}).Info("📤 USER_REPLY: Sending processed message")

				// Send the message
				sendSuccess := false
				if msg.Type == "text" {
					err = s.SendMessageFromDevice(execution.IDDevice, execution.ProspectNum, msg.Content)
					if err != nil {
						logrus.WithError(err).Error("❌ USER_REPLY: Failed to send text message")
					} else {
						sendSuccess = true
					}
				} else if msg.Type == "image" || msg.Type == "audio" || msg.Type == "video" {
					err = s.SendMediaMessage(execution.IDDevice, execution.ProspectNum, msg.Content)
					if err != nil {
						logrus.WithError(err).WithFields(logrus.Fields{
							"media_url":  msg.Content,
							"media_type": msg.Type,
						}).Error("❌ USER_REPLY: Failed to send media message")
					} else {
						sendSuccess = true
					}
				}
				
				// Save EACH message to conversation history separately
				// Format the save based on message type to match PHP behavior
				if sendSuccess {
					var saveContent string
					
					// Format based on type (matching PHP format)
					if msg.Type == "text" {
						// For text, save as-is
						saveContent = msg.Content
					} else {
						// For media (image/video/audio), save just the URL
						saveContent = msg.Content
					}
					
					err = s.aiWhatsappService.SaveConversationHistory(execution.ProspectNum, execution.IDDevice, "", saveContent, stage)
					if err != nil {
						logrus.WithError(err).WithFields(logrus.Fields{
							"type": msg.Type,
							"content": saveContent,
						}).Error("❌ USER_REPLY: Failed to save message to conversation")
					} else {
						logrus.WithFields(logrus.Fields{
							"type": msg.Type,
							"saved": saveContent,
						}).Debug("✅ USER_REPLY: Saved message to conversation")
					}
				}

				// Add delay between messages
				if i < len(messages)-1 && msg.Delay > 0 {
					time.Sleep(msg.Delay)
				}
			}
		}
	}

	logrus.WithField("execution_id", execution.ExecutionID.String).Info("✅ USER_REPLY: Flow resumed successfully after user reply")
	return nil
}

// updateCurrentNode updates both new and legacy current node fields
func (s *Service) updateCurrentNode(execution *models.AIWhatsapp, nodeID string) {
	// Update new flow tracking field
	execution.CurrentNodeID.String = nodeID
	execution.CurrentNodeID.Valid = true

	// Update legacy field for backward compatibility
	execution.CurrentNode.String = nodeID
	execution.CurrentNode.Valid = true
}

// updateFlowTrackingFields updates the flow tracking fields for user reply handling
// Uses repository's UpdateFlowTrackingFields to preserve conversation history
func (s *Service) updateFlowTrackingFields(execution *models.AIWhatsapp, currentNodeID, flowID string, waitingForReply bool) error {
	// Determine last node ID
	lastNodeID := ""
	if execution.CurrentNodeID.Valid && execution.CurrentNodeID.String != "" {
		lastNodeID = execution.CurrentNodeID.String
	}

	// Set waiting_for_reply flag
	waitingForReplyValue := 0
	if waitingForReply {
		waitingForReplyValue = 1
	}

	// Get execution ID
	executionID := ""
	if execution.ExecutionID.Valid {
		executionID = execution.ExecutionID.String
	}

	// Update flow tracking fields directly in repository to preserve conversation history
	err := s.aiWhatsappService.GetRepository().UpdateFlowTrackingFields(
		execution.ProspectNum, execution.IDDevice,
		flowID,               // flowID
		currentNodeID,        // currentNodeID
		lastNodeID,           // lastNodeID
		waitingForReplyValue, // waitingForReply
		"active",             // executionStatus
		executionID,          // executionID
	)
	if err != nil {
		return fmt.Errorf("failed to update flow tracking fields: %w", err)
	}

	// Update the execution model in memory for consistency
	execution.CurrentNodeID.String = currentNodeID
	execution.CurrentNodeID.Valid = true
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
	}).Info("✅ FLOW_TRACKING: Updated flow tracking fields successfully")

	return nil
}

// processUserReplyNode processes a user reply node by setting waiting state
func (s *Service) processUserReplyNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"prospect_id": execution.IDProspect,
		"node_id":     node.ID,
		"user_input":  userInput,
	}).Info("💬 USER_REPLY: Processing user reply node - setting waiting state")

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
	}).Info("✅ USER_REPLY: Flow set to waiting for user reply state")

	// Return empty response as we're now waiting for user input
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
				"prospect_id":  execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("🚀 START: Start node processed, advancing to delay node")

			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
				"prospect_id":  execution.IDProspect,
				"current_node": node.ID,
				"next_node":    nextNode.ID,
				"next_type":    nextNode.Type,
			}).Info("🔧 DEFAULT: Default node processed, advancing to delay node")

			// Update execution to delay node
			s.updateCurrentNode(execution, nextNode.ID)
			err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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
		err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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

	// First try to get active execution, then try any execution (including completed ones)
	// This handles cases where execution was completed but delayed messages are still pending
	execution, err := s.aiWhatsappService.GetActiveFlowExecution(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get active execution for continuation")
		return fmt.Errorf("failed to get active execution: %w", err)
	}

	// If no active execution found, try to get any execution (including completed)
	if execution == nil {
		logrus.WithFields(logrus.Fields{
			"execution_id": executionID,
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Info("🔄 FLOW: No active execution found, checking for any existing execution")

		// Get any execution (regardless of status) to continue delayed processing
		execution, err = s.aiWhatsappService.GetFlowExecutionByProspectAndDevice(phoneNumber, deviceID)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to get any execution for continuation")
			return fmt.Errorf("failed to get any execution: %w", err)
		}

		if execution == nil {
			// Log as debug instead of warn to reduce noise - this is expected for cleaned up executions
			logrus.WithField("execution_id", executionID).Debug("⚠️ FLOW: No execution found for continuation (likely cleaned up)")
			return fmt.Errorf("execution not found: %s", executionID)
		}

		// Reactivate the execution for delayed processing
		logrus.WithFields(logrus.Fields{
			"execution_id":    executionID,
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
		"execution_id":  executionID,
		"previous_node": execution.CurrentNode.String,
		"target_node":   nodeID,
	}).Info("🔄 FLOW: Advancing execution to target node after delay")

	s.updateCurrentNode(execution, nodeID)
	err = s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, execution.CurrentNodeID.String, make(map[string]interface{}), "active")
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

		// Check if response contains media URLs using the new detection service
		if s.mediaDetectionService.HasMedia(response) {
			mediaInfo := s.mediaDetectionService.ExtractFirstMedia(response)
			if mediaInfo != nil {
				logrus.WithFields(logrus.Fields{
					"media_type": mediaInfo.MediaType,
					"media_url":  mediaInfo.MediaURL,
					"device_id":  deviceID,
				}).Info("🖼️ FLOW: Extracted media URL from delayed response, sending as media message")

				// Send as media message instead of text
				err = s.SendMediaMessage(deviceID, phoneNumber, mediaInfo.MediaURL)
				if err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"device_id":    deviceID,
						"phone_number": phoneNumber,
						"media_url":    mediaInfo.MediaURL,
						"media_type":   mediaInfo.MediaType,
					}).Error("❌ FLOW: Failed to send delayed media message")
					return fmt.Errorf("failed to send delayed media message: %w", err)
				}
			} else {
				// Fallback to text if extraction failed
				err = s.SendMessageFromDevice(deviceID, phoneNumber, response)
				if err != nil {
					logrus.WithError(err).Error("❌ FLOW: Failed to send delayed response as text fallback")
					return fmt.Errorf("failed to send delayed response: %w", err)
				}
			}
		} else {
			// Send as regular text message
			err = s.SendMessageFromDevice(deviceID, phoneNumber, response)
			if err != nil {
				logrus.WithError(err).Error("❌ FLOW: Failed to send delayed response")
				return fmt.Errorf("failed to send response: %w", err)
			}
		}

		// Add bot response to ai_whatsapp_nodepath conversation
		err = s.aiWhatsappService.SaveConversationHistory(phoneNumber, deviceID, "", response, "")
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to add bot message to ai_whatsapp_nodepath")
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
