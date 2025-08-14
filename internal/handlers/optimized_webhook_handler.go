package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"nodepath-chat/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// OptimizedWebhookHandler handles incoming WhatsApp webhooks with high performance
type OptimizedWebhookHandler struct {
	messageProcessor *services.MessageProcessor
	logger           *logrus.Entry
}

// WebhookPayload represents the incoming webhook payload
type WebhookPayload struct {
	DeviceID    string                 `json:"device_id"`
	InstanceID  string                 `json:"instance_id"`
	PhoneNumber string                 `json:"phone_number"`
	Message     WebhookMessage         `json:"message"`
	Timestamp   int64                  `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WebhookMessage represents a message in the webhook
type WebhookMessage struct {
	ID        string `json:"id"`
	Type      string `json:"type"`      // text, image, audio, video, document
	Content   string `json:"content"`   // text content or caption
	MediaURL  string `json:"media_url,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	FromMe    bool   `json:"from_me"`
}

// NewOptimizedWebhookHandler creates a new optimized webhook handler
func NewOptimizedWebhookHandler(messageProcessor *services.MessageProcessor) *OptimizedWebhookHandler {
	return &OptimizedWebhookHandler{
		messageProcessor: messageProcessor,
		logger: logrus.WithFields(logrus.Fields{
			"component": "optimized_webhook_handler",
		}),
	}
}

// HandleWebhook processes incoming WhatsApp webhooks with high performance
func (h *OptimizedWebhookHandler) HandleWebhook(c *gin.Context) {
	start := time.Now()
	
	// Parse the webhook payload
	var payload WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.logger.WithError(err).Error("Failed to parse webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}
	
	// Validate required fields
	if err := h.validatePayload(&payload); err != nil {
		h.logger.WithError(err).Error("Invalid webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Skip messages from self
	if payload.Message.FromMe {
		h.logger.Debug("Skipping message from self")
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}
	
	// Skip empty messages
	if strings.TrimSpace(payload.Message.Content) == "" && payload.Message.MediaURL == "" {
		h.logger.Debug("Skipping empty message")
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}
	
	// Create incoming message for processing
	incomingMsg := &services.IncomingMessage{
		ID:          generateMessageID(&payload),
		Type:        h.determineMessageType(&payload),
		DeviceID:    payload.DeviceID,
		PhoneNumber: payload.PhoneNumber,
		Content:     payload.Message.Content,
		MediaURL:    payload.Message.MediaURL,
		MediaType:   payload.Message.MediaType,
		Timestamp:   time.Unix(payload.Timestamp, 0),
		Metadata:    payload.Metadata,
		Priority:    h.calculatePriority(&payload),
		MaxRetries:  3,
	}
	
	// Add request context metadata
	if incomingMsg.Metadata == nil {
		incomingMsg.Metadata = make(map[string]interface{})
	}
	incomingMsg.Metadata["request_id"] = c.GetHeader("X-Request-ID")
	incomingMsg.Metadata["user_agent"] = c.GetHeader("User-Agent")
	incomingMsg.Metadata["client_ip"] = c.ClientIP()
	
	// Process message asynchronously
	if err := h.messageProcessor.ProcessMessage(incomingMsg); err != nil {
		h.logger.WithFields(logrus.Fields{
			"message_id":   incomingMsg.ID,
			"phone_number": payload.PhoneNumber,
			"error":        err.Error(),
		}).Error("Failed to queue message for processing")
		
		// Return appropriate error based on the failure type
		if strings.Contains(err.Error(), "rate limit") {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		} else if strings.Contains(err.Error(), "queue is full") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service temporarily unavailable"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}
	
	// Log successful webhook processing
	processingTime := time.Since(start)
	h.logger.WithFields(logrus.Fields{
		"message_id":      incomingMsg.ID,
		"phone_number":    payload.PhoneNumber,
		"message_type":    incomingMsg.Type,
		"processing_time": processingTime,
	}).Info("Webhook processed successfully")
	
	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":     "accepted",
		"message_id": incomingMsg.ID,
		"queued_at":  time.Now().Unix(),
	})
}

// HandleBulkWebhook processes multiple webhooks in a single request
func (h *OptimizedWebhookHandler) HandleBulkWebhook(c *gin.Context) {
	start := time.Now()
	
	// Parse bulk webhook payload
	var payloads []WebhookPayload
	if err := c.ShouldBindJSON(&payloads); err != nil {
		h.logger.WithError(err).Error("Failed to parse bulk webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}
	
	// Limit bulk size to prevent abuse
	if len(payloads) > 100 {
		h.logger.WithField("payload_count", len(payloads)).Warn("Bulk webhook payload too large")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many messages in bulk request"})
		return
	}
	
	results := make([]map[string]interface{}, 0, len(payloads))
	successCount := 0
	failureCount := 0
	
	// Process each payload
	for i, payload := range payloads {
		result := map[string]interface{}{
			"index": i,
		}
		
		// Validate payload
		if err := h.validatePayload(&payload); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
			failureCount++
			results = append(results, result)
			continue
		}
		
		// Skip messages from self or empty messages
		if payload.Message.FromMe || (strings.TrimSpace(payload.Message.Content) == "" && payload.Message.MediaURL == "") {
			result["status"] = "ignored"
			results = append(results, result)
			continue
		}
		
		// Create incoming message
		incomingMsg := &services.IncomingMessage{
			ID:          generateMessageID(&payload),
			Type:        h.determineMessageType(&payload),
			DeviceID:    payload.DeviceID,
			PhoneNumber: payload.PhoneNumber,
			Content:     payload.Message.Content,
			MediaURL:    payload.Message.MediaURL,
			MediaType:   payload.Message.MediaType,
			Timestamp:   time.Unix(payload.Timestamp, 0),
			Metadata:    payload.Metadata,
			Priority:    h.calculatePriority(&payload),
			MaxRetries:  3,
		}
		
		// Process message
		if err := h.messageProcessor.ProcessMessage(incomingMsg); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
			failureCount++
		} else {
			result["status"] = "accepted"
			result["message_id"] = incomingMsg.ID
			successCount++
		}
		
		results = append(results, result)
	}
	
	// Log bulk processing results
	processingTime := time.Since(start)
	h.logger.WithFields(logrus.Fields{
		"total_messages":   len(payloads),
		"success_count":    successCount,
		"failure_count":    failureCount,
		"processing_time":  processingTime,
	}).Info("Bulk webhook processed")
	
	// Return results
	c.JSON(http.StatusOK, gin.H{
		"total":         len(payloads),
		"success_count": successCount,
		"failure_count": failureCount,
		"results":       results,
		"processed_at":  time.Now().Unix(),
	})
}

// GetProcessorStatus returns the current status of the message processor
func (h *OptimizedWebhookHandler) GetProcessorStatus(c *gin.Context) {
	status := h.messageProcessor.GetStatus()
	metrics := h.messageProcessor.GetMetrics()
	
	c.JSON(http.StatusOK, gin.H{
		"processor_status": status,
		"metrics":          metrics,
		"timestamp":        time.Now().Unix(),
	})
}

// validatePayload validates the webhook payload
func (h *OptimizedWebhookHandler) validatePayload(payload *WebhookPayload) error {
	if payload.DeviceID == "" {
		return fmt.Errorf("device_id is required")
	}
	
	if payload.PhoneNumber == "" {
		return fmt.Errorf("phone_number is required")
	}
	
	if payload.Message.ID == "" {
		return fmt.Errorf("message.id is required")
	}
	
	// Validate phone number format (basic validation)
	if !strings.HasPrefix(payload.PhoneNumber, "+") && !strings.Contains(payload.PhoneNumber, "@") {
		return fmt.Errorf("invalid phone_number format")
	}
	
	return nil
}

// determineMessageType determines the type of message based on the payload
func (h *OptimizedWebhookHandler) determineMessageType(payload *WebhookPayload) services.MessageType {
	// Check if it's a customer service context
	if payload.Metadata != nil {
		if context, exists := payload.Metadata["context"]; exists {
			if contextStr, ok := context.(string); ok {
				switch contextStr {
				case "customer_service":
					return services.MessageTypeCustomerReply
				case "ai_chat":
					return services.MessageTypeAIReply
				case "system":
					return services.MessageTypeSystem
				}
			}
		}
	}
	
	// Default to user reply for regular messages
	return services.MessageTypeUserReply
}

// calculatePriority calculates message priority based on various factors
func (h *OptimizedWebhookHandler) calculatePriority(payload *WebhookPayload) int {
	// Default priority
	priority := 3
	
	// Higher priority for customer service messages
	if payload.Metadata != nil {
		if context, exists := payload.Metadata["context"]; exists {
			if contextStr, ok := context.(string); ok && contextStr == "customer_service" {
				priority = 1 // Highest priority
			}
		}
		
		// Higher priority for VIP customers
		if vip, exists := payload.Metadata["vip"]; exists {
			if vipBool, ok := vip.(bool); ok && vipBool {
				priority = 2
			}
		}
	}
	
	// Higher priority for media messages (they might be more urgent)
	if payload.Message.MediaURL != "" {
		if priority > 2 {
			priority = 2
		}
	}
	
	return priority
}

// generateMessageID generates a unique message ID
func generateMessageID(payload *WebhookPayload) string {
	// Use original message ID if available, otherwise generate UUID
	if payload.Message.ID != "" {
		return fmt.Sprintf("%s_%s", payload.DeviceID, payload.Message.ID)
	}
	
	return fmt.Sprintf("%s_%s", payload.DeviceID, uuid.New().String())
}

// RegisterOptimizedWebhookRoutes registers the optimized webhook routes
func RegisterOptimizedWebhookRoutes(router *gin.Engine, messageProcessor *services.MessageProcessor) {
	handler := NewOptimizedWebhookHandler(messageProcessor)
	
	// Webhook routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/webhook/whatsapp", handler.HandleWebhook)
		v1.POST("/webhook/whatsapp/bulk", handler.HandleBulkWebhook)
		v1.GET("/webhook/status", handler.GetProcessorStatus)
	}
	
	// Health check route
	router.GET("/health", func(c *gin.Context) {
		status := messageProcessor.GetStatus()
		if running, ok := status["running"].(bool); ok && running {
			c.JSON(http.StatusOK, gin.H{"status": "healthy", "timestamp": time.Now().Unix()})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "timestamp": time.Now().Unix()})
		}
	})
}