package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"
	"nodepath-chat/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// AIWhatsappHandlers contains all AI WhatsApp webhook handlers
type AIWhatsappHandlers struct {
	AIWhatsappService services.AIWhatsappService
	AIRepo            repository.AIWhatsappRepository
	DeviceRepo        repository.DeviceSettingsRepository
	mainHandlers      *Handlers // Reference to main handlers for flow routing
}

// NewAIWhatsappHandlers creates a new AI WhatsApp handlers instance
func NewAIWhatsappHandlers(
	aiWhatsappService services.AIWhatsappService,
	aiRepo repository.AIWhatsappRepository,
	deviceRepo repository.DeviceSettingsRepository,
) *AIWhatsappHandlers {
	return &AIWhatsappHandlers{
		AIWhatsappService: aiWhatsappService,
		AIRepo:            aiRepo,
		DeviceRepo:        deviceRepo,
		mainHandlers:      nil, // Will be set after main handlers initialization
	}
}

// SetMainHandlers sets the reference to main handlers for flow routing
func (h *AIWhatsappHandlers) SetMainHandlers(mainHandlers *Handlers) {
	h.mainHandlers = mainHandlers
}

// getAuthMiddleware returns the authentication middleware from main handlers
func (h *AIWhatsappHandlers) getAuthMiddleware() fiber.Handler {
	if h.mainHandlers != nil && h.mainHandlers.authHandlers != nil {
		return h.mainHandlers.authHandlers.AuthMiddleware()
	}
	// Fallback: return a middleware that always denies access
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication middleware not available",
		})
	}
}

// getDeviceRequiredMiddleware returns the device required middleware from main handlers
func (h *AIWhatsappHandlers) getDeviceRequiredMiddleware() fiber.Handler {
	if h.mainHandlers != nil && h.mainHandlers.authHandlers != nil {
		return h.mainHandlers.authHandlers.DeviceRequiredMiddleware()
	}
	// Fallback: return a middleware that always denies access
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Device required middleware not available",
		})
	}
}

// SetupAIWhatsappRoutes sets up AI WhatsApp webhook routes
func (h *AIWhatsappHandlers) SetupAIWhatsappRoutes(api fiber.Router) {
	// Webhook endpoints for receiving WhatsApp messages (no auth required for webhooks)
	api.Post("/webhook/whatsapp/:device_id", h.HandleWhatsappWebhook)
	api.Post("/webhook/wablas/:device_id", h.HandleWablasWebhook)
	api.Post("/webhook/whacenter/:device_id", h.HandleWhacenterWebhook)
	api.Post("/webhook/waha/:device_id", h.HandleWahaWebhook)

	// Test endpoints for webhook data extraction (no auth required for testing)
	api.Post("/test/waha/extraction", h.TestWahaExtraction)
	
	// Production debugging endpoint - logs everything and returns payload structure (no auth required for debugging)
	api.Post("/debug/waha/:device_id", h.DebugWahaWebhook)

	// Device command processing (no auth required for webhook commands)
	api.Post("/ai/device/command", h.ProcessDeviceCommand)

	// Protected routes requiring authentication
	protected := api.Group("/ai")
	protected.Use(h.getAuthMiddleware())
	protected.Use(h.getDeviceRequiredMiddleware())

	// AI conversation management endpoints
	protected.Post("/conversation/start", h.StartAIConversation)
	protected.Post("/conversation/process", h.ProcessAIMessage)
	protected.Post("/conversation/toggle-human", h.ToggleHumanTakeover)
	protected.Get("/conversation/history/:prospect_num", h.GetConversationHistory)
	protected.Get("/conversation/status/:prospect_num", h.GetConversationStatus)

	// AI settings management
	protected.Get("/settings/:staff_id", h.GetAISettings)
	protected.Post("/settings", h.CreateAISettings)
	protected.Put("/settings/:staff_id", h.UpdateAISettings)
	protected.Delete("/settings/:staff_id", h.DeleteAISettings)

	// Analytics endpoints
	protected.Get("/analytics", h.GetAnalytics)
	protected.Post("/analytics", h.GetAnalytics)

	// Data table endpoints
	protected.Get("/ai-whatsapp/data", h.GetAllAIWhatsappData)
	protected.Delete("/ai-whatsapp/data/:id", h.DeleteAIWhatsappData)
}

// WhatsappWebhookRequest represents incoming WhatsApp webhook data
type WhatsappWebhookRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	DeviceID  string `json:"device_id"`
}

// WablasWebhookRequest represents incoming Wablas webhook data
type WablasWebhookRequest struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
	Device  string `json:"device"`
	Time    string `json:"time"`
}

// WhacenterWebhookRequest represents incoming Whacenter webhook data
type WhacenterWebhookRequest struct {
	Number  string `json:"number"`
	Text    string `json:"text"`
	Device  string `json:"device"`
	Date    string `json:"date"`
}

// WahaWebhookRequest represents incoming WAHA webhook data
// WAHA uses nested payload structure with _data containing message info
type WahaWebhookRequest struct {
	Event   string `json:"event"`
	Session string `json:"session"`
	Payload struct {
		Data struct {
			From string `json:"from"`
			Body string `json:"body"`
			Info struct {
				IsGroup bool `json:"IsGroup"`
			} `json:"Info"`
		} `json:"_data"`
	} `json:"payload"`
}

// StartAIConversationRequest represents request to start AI conversation
type StartAIConversationRequest struct {
	ProspectNum string `json:"prospect_num"`
	IDDevice    string `json:"id_device"`
	Niche       string `json:"niche"`
	Stage       string `json:"stage"`
}

// ProcessAIMessageRequest represents request to process AI message
type ProcessAIMessageRequest struct {
	ProspectNum string `json:"prospect_num"`
	IDDevice    string `json:"id_device"`
	Message     string `json:"message"`
	Stage       string `json:"stage"`
}

// ToggleHumanTakeoverRequest represents request to toggle human takeover
type ToggleHumanTakeoverRequest struct {
	ProspectNum string `json:"prospect_num"`
	Human       bool   `json:"human"`
}

// ProcessDeviceCommandRequest represents request to process device command
type ProcessDeviceCommandRequest struct {
	ProspectNum string `json:"prospect_num"`
	Command     string `json:"command"`
	IDDevice    string `json:"id_device"`
}

// AnalyticsRequest represents the request structure for analytics endpoint
type AnalyticsRequest struct {
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
	DeviceID  string `json:"device_id" form:"device_id"`
}

// AnalyticsResponse represents the response structure for analytics endpoint
type AnalyticsResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// HandleWhatsappWebhook handles generic WhatsApp webhook messages
func (h *AIWhatsappHandlers) HandleWhatsappWebhook(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	if deviceID == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Device ID is required")
	}

	var req WhatsappWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse WhatsApp webhook request")
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	req.DeviceID = deviceID

	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"from":      req.From,
		"message":   req.Message,
	}).Info("Received WhatsApp webhook")

	// Process the message asynchronously
	go h.processIncomingMessage(req.From, req.Message, deviceID, "whatsapp")

	return h.successResponse(c, map[string]string{"status": "received"})
}

// HandleWablasWebhook handles Wablas provider webhook messages
func (h *AIWhatsappHandlers) HandleWablasWebhook(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	if deviceID == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Device ID is required")
	}

	var req WablasWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse Wablas webhook request")
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"phone":     req.Phone,
		"message":   req.Message,
	}).Info("Received Wablas webhook")

	// Process the message asynchronously
	go h.processIncomingMessage(req.Phone, req.Message, deviceID, "wablas")

	return h.successResponse(c, map[string]string{"status": "received"})
}

// HandleWhacenterWebhook handles Whacenter provider webhook messages
func (h *AIWhatsappHandlers) HandleWhacenterWebhook(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	if deviceID == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Device ID is required")
	}

	var req WhacenterWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse Whacenter webhook request")
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"number":    req.Number,
		"text":      req.Text,
	}).Info("Received Whacenter webhook")

	// Process the message asynchronously
	go h.processIncomingMessage(req.Number, req.Text, deviceID, "whacenter")

	return h.successResponse(c, map[string]string{"status": "received"})
}

// HandleWahaWebhook handles incoming WAHA webhook requests
// Processes WhatsApp messages and triggers AI responses based on device settings
// Implements standardized WAHA webhook data extraction and processing logic
func (h *AIWhatsappHandlers) HandleWahaWebhook(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	body := c.Body()
	
	// Enhanced logging for production debugging - log ALL headers and payload details
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"payload_size": len(body),
		"content_type": c.Get("Content-Type"),
		"user_agent": c.Get("User-Agent"),
		"headers": headers,
		"raw_payload": string(body),
		"method": c.Method(),
		"url": c.OriginalURL(),
	}).Error("🚨 WAHA PRODUCTION DEBUG: Complete webhook request details")

	// Parse as generic map first for flexible extraction
	var rawPayload map[string]interface{}
	if err := json.Unmarshal(body, &rawPayload); err != nil {
		logrus.WithError(err).Error("Failed to parse WAHA webhook JSON")
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid JSON format",
		})
	}

	// Extract standardized webhook data according to requirements
	extractedData := h.extractWahaWebhookData(rawPayload)

	// Logic 1: If is_group = true → ignore (do not execute)
	if extractedData.IsGroup {
		logrus.WithFields(logrus.Fields{
			"sender_phone": extractedData.SenderPhone,
			"sender_name": extractedData.SenderName,
			"device_id": deviceID,
		}).Info("⏭️ WAHA: Ignoring group message as per requirements")
		return c.JSON(fiber.Map{
			"status": "ignored", 
			"reason": "group message",
			"extracted_data": extractedData,
		})
	}

	// Validate required fields
	if extractedData.SenderPhone == "" || extractedData.Message == "" {
		logrus.WithFields(logrus.Fields{
			"sender_phone": extractedData.SenderPhone,
			"message": truncateString(extractedData.Message, 100),
			"sender_name": extractedData.SenderName,
			"is_from_me": extractedData.IsFromMe,
			"is_group": extractedData.IsGroup,
		}).Warn("⚠️ WAHA: Missing required fields in extracted data")
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing required fields",
			"missing": map[string]bool{
				"sender_phone": extractedData.SenderPhone == "",
				"message": extractedData.Message == "",
			},
			"extracted_data": extractedData,
		})
	}

	// Clean phone number format (remove @c.us suffix if present)
	if strings.HasSuffix(extractedData.SenderPhone, "@c.us") {
		extractedData.SenderPhone = strings.TrimSuffix(extractedData.SenderPhone, "@c.us")
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"cleaned_from": extractedData.SenderPhone,
		}).Info("🔧 WAHA: Phone number cleaned - stripped @c.us suffix")
	}

	// Logic 2: If is_from_me = true → check message starting with %, #, cmd, or & (system/device commands)
	if extractedData.IsFromMe {
		message := strings.TrimSpace(extractedData.Message)
		isSystemCommand := strings.HasPrefix(message, "%") || 
						 strings.HasPrefix(message, "#") || 
						 strings.HasPrefix(message, "cmd") || 
						 strings.HasPrefix(message, "&")
		
		if isSystemCommand {
			logrus.WithFields(logrus.Fields{
				"sender_phone": extractedData.SenderPhone,
				"command": truncateString(message, 50),
				"device_id": deviceID,
			}).Info("🔧 WAHA: Processing system/device command through standardized flow routing")
			
			// STANDARDIZED FLOW ROUTING: Use the same flow processing logic as Whacenter for commands
			// Create webhook data structure compatible with processWebhookMessage
			webhookData := map[string]interface{}{
				"from": extractedData.SenderPhone,
				"message": message,
				"message_type": "text",
				"is_group": extractedData.IsGroup,
				"sender_name": extractedData.SenderName,
				"is_from_me": extractedData.IsFromMe,
			}

			// Route through the standardized webhook processing system
			go func() {
				if h.mainHandlers != nil {
					err := h.mainHandlers.processWebhookMessage(webhookData, deviceID, "waha")
					if err != nil {
						logrus.WithError(err).WithFields(logrus.Fields{
							"device_id": deviceID,
							"sender_phone": extractedData.SenderPhone,
							"command": message,
						}).Error("❌ WAHA: Failed to process system command through standardized flow routing")
					} else {
						logrus.WithFields(logrus.Fields{
							"device_id": deviceID,
							"sender_phone": extractedData.SenderPhone,
							"command": message,
						}).Info("✅ WAHA: Successfully processed system command through standardized flow routing")
					}
				} else {
					logrus.Error("❌ WAHA: Main handlers not available, falling back to direct AI processing for command")
					// Fallback to direct processing if main handlers not available
					h.processIncomingMessage(extractedData.SenderPhone, message, deviceID, "waha")
				}
			}()
			
			return c.JSON(fiber.Map{
				"status": "success",
				"type": "system_command",
				"routing": "standardized_flow",
				"extracted_data": extractedData,
			})
		} else {
			logrus.WithFields(logrus.Fields{
				"sender_phone": extractedData.SenderPhone,
				"message": truncateString(message, 100),
				"device_id": deviceID,
			}).Info("⏭️ WAHA: Ignoring non-system message from sender (is_from_me=true)")
			
			return c.JSON(fiber.Map{
				"status": "ignored", 
				"reason": "non-system message from sender",
				"extracted_data": extractedData,
			})
		}
	}

	// Logic 3: Otherwise → treat as normal customer message
	logrus.WithFields(logrus.Fields{
		"sender_phone": extractedData.SenderPhone,
		"sender_name": extractedData.SenderName,
		"message": truncateString(extractedData.Message, 100),
		"device_id": deviceID,
	}).Info("💬 WAHA: Processing normal customer message through standardized flow routing")

	// STANDARDIZED FLOW ROUTING: Use the same flow processing logic as Whacenter
	// Create webhook data structure compatible with processWebhookMessage
	webhookData := map[string]interface{}{
		"from": extractedData.SenderPhone,
		"message": extractedData.Message,
		"message_type": "text",
		"is_group": extractedData.IsGroup,
		"sender_name": extractedData.SenderName,
		"is_from_me": extractedData.IsFromMe,
	}

	// Route through the standardized webhook processing system
	// This ensures WAHA follows the same flow node logic as Whacenter
	go func() {
		if h.mainHandlers != nil {
			err := h.mainHandlers.processWebhookMessage(webhookData, deviceID, "waha")
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"device_id": deviceID,
					"sender_phone": extractedData.SenderPhone,
				}).Error("❌ WAHA: Failed to process message through standardized flow routing")
			} else {
				logrus.WithFields(logrus.Fields{
					"device_id": deviceID,
					"sender_phone": extractedData.SenderPhone,
				}).Info("✅ WAHA: Successfully processed message through standardized flow routing")
			}
		} else {
			logrus.Error("❌ WAHA: Main handlers not available, falling back to direct AI processing")
			// Fallback to direct processing if main handlers not available
			h.processIncomingMessage(extractedData.SenderPhone, extractedData.Message, deviceID, "waha")
		}
	}()

	return c.JSON(fiber.Map{
		"status": "success",
		"type": "customer_message",
		"routing": "standardized_flow",
		"extracted_data": extractedData,
	})
}

// extractWahaFields extracts fields from WAHA webhook payload using multiple fallback methods
// Handles different WAHA payload structures that may vary in production
func (h *AIWhatsappHandlers) extractWahaFields(payload map[string]interface{}) (from, message, event, session string, isGroup bool) {
	// Use the new standardized extraction function
	extractedData := h.extractWahaWebhookData(payload)
	
	// Map to old function signature for backward compatibility
	from = extractedData.SenderPhone
	message = extractedData.Message
	isGroup = extractedData.IsGroup
	
	// Extract event and session from top level for backward compatibility
	if eventVal, ok := payload["event"].(string); ok {
		event = eventVal
	}
	if sessionVal, ok := payload["session"].(string); ok {
		session = sessionVal
	}
	
	return from, message, event, session, isGroup
}

// WahaWebhookData represents the standardized extracted data from WAHA webhook
type WahaWebhookData struct {
	SenderPhone string `json:"sender_phone"`
	SenderName  string `json:"sender_name"`
	Message     string `json:"message"`
	IsFromMe    bool   `json:"is_from_me"`
	IsGroup     bool   `json:"is_group"`
}

// extractWahaWebhookData extracts WAHA webhook data according to standardized requirements
// Handles both nested payload structure and direct structure
// Returns extracted fields in standardized JSON format
func (h *AIWhatsappHandlers) extractWahaWebhookData(webhookPayload map[string]interface{}) WahaWebhookData {
	var result WahaWebhookData
	
	// Enhanced payload structure analysis for production debugging
	payloadAnalysis := analyzePayloadDepth(webhookPayload)
	logrus.WithFields(logrus.Fields{
		"payload_keys": getMapKeys(webhookPayload),
		"has_payload": webhookPayload["payload"] != nil,
		"payload_analysis": payloadAnalysis,
		"full_payload": webhookPayload,
	}).Error("🚨 WAHA PRODUCTION DEBUG: Complete payload structure analysis")
	
	// Determine the payload object to work with
	var payloadObj map[string]interface{}
	if nestedPayload, ok := webhookPayload["payload"].(map[string]interface{}); ok {
		// Use nested payload structure (real webhook)
		payloadObj = nestedPayload
		logrus.Info("🔍 WAHA: Using nested payload structure")
	} else {
		// Use direct structure (test or alternative format)
		payloadObj = webhookPayload
		logrus.Info("🔍 WAHA: Using direct payload structure")
	}
	
	// Debug log the payload object structure
	logrus.WithFields(logrus.Fields{
		"payload_obj_keys": getMapKeys(payloadObj),
		"payload_obj": payloadObj,
	}).Info("🔍 WAHA DEBUG: Payload object structure")
	
	// FIXED: Extract directly from payload level first, then try _data as fallback
	// Based on production logs, WAHA sends data at payload level: payload.body, payload.from, etc.
	
	// Extract sender_phone = payload.from (primary) or _data.from (fallback)
	if fromVal, ok := payloadObj["from"].(string); ok {
		result.SenderPhone = fromVal
		logrus.WithField("extraction_method", "payload_from").Info("🔍 WAHA: Sender phone extracted from payload.from")
	} else if _dataObj, ok := payloadObj["_data"].(map[string]interface{}); ok {
		if fromVal, ok := _dataObj["from"].(string); ok {
			result.SenderPhone = fromVal
			logrus.WithField("extraction_method", "_data_from").Info("🔍 WAHA: Sender phone extracted from _data.from")
		}
	}
	
	// Extract message = payload.body (primary) or _data.body (fallback)
	if bodyVal, ok := payloadObj["body"].(string); ok {
		result.Message = bodyVal
		logrus.WithField("extraction_method", "payload_body").Info("🔍 WAHA: Message extracted from payload.body")
	} else if _dataObj, ok := payloadObj["_data"].(map[string]interface{}); ok {
		if bodyVal, ok := _dataObj["body"].(string); ok {
			result.Message = bodyVal
			logrus.WithField("extraction_method", "_data_body").Info("🔍 WAHA: Message extracted from _data.body")
		}
	}
	
	// Extract is_from_me = payload.fromMe (primary) or _data.fromMe (fallback)
	if fromMeVal, ok := payloadObj["fromMe"].(bool); ok {
		result.IsFromMe = fromMeVal
		logrus.WithField("extraction_method", "payload_fromMe").Info("🔍 WAHA: IsFromMe extracted from payload.fromMe")
	} else if _dataObj, ok := payloadObj["_data"].(map[string]interface{}); ok {
		if fromMeVal, ok := _dataObj["fromMe"].(bool); ok {
			result.IsFromMe = fromMeVal
			logrus.WithField("extraction_method", "_data_fromMe").Info("🔍 WAHA: IsFromMe extracted from _data.fromMe")
		}
	}
	
	// Extract sender_name from payload.info.pushName (primary) or _data.info.pushName (fallback)
	if infoObj, ok := payloadObj["info"].(map[string]interface{}); ok {
		if pushNameVal, ok := infoObj["pushName"].(string); ok {
			result.SenderName = pushNameVal
			logrus.WithField("extraction_method", "payload_info_pushname").Info("🔍 WAHA: Sender name extracted from payload.info.pushName")
		}
		// Extract is_group from payload.info.IsGroup
		if isGroupVal, ok := infoObj["IsGroup"].(bool); ok {
			result.IsGroup = isGroupVal
			logrus.WithField("extraction_method", "payload_info_isgroup").Info("🔍 WAHA: IsGroup extracted from payload.info.IsGroup")
		}
	} else if _dataObj, ok := payloadObj["_data"].(map[string]interface{}); ok {
		if infoObj, ok := _dataObj["info"].(map[string]interface{}); ok {
			if pushNameVal, ok := infoObj["pushName"].(string); ok {
				result.SenderName = pushNameVal
				logrus.WithField("extraction_method", "_data_info_pushname").Info("🔍 WAHA: Sender name extracted from _data.info.pushName")
			}
			// Extract is_group from _data.info.IsGroup
			if isGroupVal, ok := infoObj["IsGroup"].(bool); ok {
				result.IsGroup = isGroupVal
				logrus.WithField("extraction_method", "_data_info_isgroup").Info("🔍 WAHA: IsGroup extracted from _data.info.IsGroup")
			}
		}
	}
	
	// Additional fallback: try 'me' field for sender information (based on production logs)
	if result.SenderName == "" {
		if meObj, ok := webhookPayload["me"].(map[string]interface{}); ok {
			if pushNameVal, ok := meObj["pushName"].(string); ok {
				result.SenderName = pushNameVal
				logrus.WithField("extraction_method", "me_pushname").Info("🔍 WAHA: Sender name extracted from me.pushName")
			}
		}
	}
	
	// Fallback: try alternative media structure for sender_name and is_group
	if result.SenderName == "" {
		if mediaObj, ok := payloadObj["media"].(map[string]interface{}); ok {
			if infoObj, ok := mediaObj["Info"].(map[string]interface{}); ok {
				if pushNameVal, ok := infoObj["PushName"].(string); ok {
					result.SenderName = pushNameVal
					logrus.WithField("extraction_method", "payload_media_info_pushname").Info("🔍 WAHA: Sender name extracted from payload.media.Info.PushName")
				}
				if isGroupVal, ok := infoObj["IsGroup"].(bool); ok {
					result.IsGroup = isGroupVal
					logrus.WithField("extraction_method", "payload_media_info_isgroup").Info("🔍 WAHA: IsGroup extracted from payload.media.Info.IsGroup")
				}
			}
		}
	}
	
	// PRODUCTION FALLBACK: Try alternative extraction methods if primary extraction failed
	if result.SenderPhone == "" || result.Message == "" {
		logrus.Error("🚨 WAHA PRODUCTION: Primary extraction failed, trying fallback methods")
		
		// Fallback 1: Try direct top-level fields
		if result.SenderPhone == "" {
			if fromVal, ok := webhookPayload["from"].(string); ok {
				result.SenderPhone = fromVal
				logrus.Error("🚨 FALLBACK 1: Extracted sender_phone from top-level 'from'")
			}
		}
		if result.Message == "" {
			if bodyVal, ok := webhookPayload["body"].(string); ok {
				result.Message = bodyVal
				logrus.Error("🚨 FALLBACK 1: Extracted message from top-level 'body'")
			} else if msgVal, ok := webhookPayload["message"].(string); ok {
				result.Message = msgVal
				logrus.Error("🚨 FALLBACK 1: Extracted message from top-level 'message'")
			} else if textVal, ok := webhookPayload["text"].(string); ok {
				result.Message = textVal
				logrus.Error("🚨 FALLBACK 1: Extracted message from top-level 'text'")
			}
		}
		
		// Fallback 2: Try data field without _data prefix
		if dataObj, ok := webhookPayload["data"].(map[string]interface{}); ok {
			logrus.Error("🚨 FALLBACK 2: Found 'data' field, trying extraction")
			if result.SenderPhone == "" {
				if fromVal, ok := dataObj["from"].(string); ok {
					result.SenderPhone = fromVal
					logrus.Error("🚨 FALLBACK 2: Extracted sender_phone from data.from")
				}
			}
			if result.Message == "" {
				if bodyVal, ok := dataObj["body"].(string); ok {
					result.Message = bodyVal
					logrus.Error("🚨 FALLBACK 2: Extracted message from data.body")
				}
			}
		}
		
		// Fallback 3: Try message field variations
		if result.Message == "" {
			for _, key := range []string{"content", "msg", "messageContent", "textContent"} {
				if msgVal, ok := webhookPayload[key].(string); ok {
					result.Message = msgVal
					logrus.WithField("fallback_key", key).Error("🚨 FALLBACK 3: Extracted message from alternative key")
					break
				}
			}
		}
		
		// Fallback 4: Try phone number variations
		if result.SenderPhone == "" {
			for _, key := range []string{"phone", "number", "phoneNumber", "sender", "contact"} {
				if phoneVal, ok := webhookPayload[key].(string); ok {
					result.SenderPhone = phoneVal
					logrus.WithField("fallback_key", key).Error("🚨 FALLBACK 4: Extracted sender_phone from alternative key")
					break
				}
			}
		}
	}
	
	// Log extraction results with production debugging
	logrus.WithFields(logrus.Fields{
		"sender_phone": result.SenderPhone,
		"sender_name": result.SenderName,
		"message": truncateString(result.Message, 100),
		"is_from_me": result.IsFromMe,
		"is_group": result.IsGroup,
		"extraction_success": result.SenderPhone != "" && result.Message != "",
	}).Error("🚨 WAHA PRODUCTION: Final extraction results")
	
	// Log critical error if fields are still missing after all fallbacks
	if result.SenderPhone == "" || result.Message == "" {
		logrus.WithFields(logrus.Fields{
			"missing_sender_phone": result.SenderPhone == "",
			"missing_message": result.Message == "",
			"all_payload_keys": getMapKeys(webhookPayload),
			"payload_structure": analyzePayloadDepth(webhookPayload),
		}).Error("🚨 WAHA PRODUCTION CRITICAL: All extraction methods failed - payload structure unknown")
	}
	
	// Console debug output for checking extracted data
	logrus.WithFields(logrus.Fields{
		"sender_phone": result.SenderPhone,
		"sender_name": result.SenderName,
		"message": result.Message,
		"is_from_me": result.IsFromMe,
		"is_group": result.IsGroup,
	}).Info("🧪 WAHA EXTRACTION DEBUG: Final extracted data")
	
	return result
}

// StartAIConversation starts a new AI conversation
func (h *AIWhatsappHandlers) StartAIConversation(c *fiber.Ctx) error {
	var req StartAIConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.ProspectNum == "" || req.IDDevice == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Missing required fields")
	}

	// Create AI WhatsApp conversation record
	aiWhatsapp := &models.AIWhatsapp{
		ProspectNum: req.ProspectNum,
		IDDevice:    req.IDDevice,
		Stage:       sql.NullString{String: req.Stage, Valid: req.Stage != ""},
		Human:       0, // AI active by default
		Niche:       req.Niche,
	}

	err := h.AIRepo.CreateAIWhatsapp(aiWhatsapp)
	if err != nil {
		logrus.WithError(err).Error("Failed to create AI conversation")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to start AI conversation")
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num": req.ProspectNum,
		"id_device":    req.IDDevice,
	}).Info("AI conversation started")

	return h.successResponse(c, aiWhatsapp)
}

// ProcessAIMessage processes an AI message manually
func (h *AIWhatsappHandlers) ProcessAIMessage(c *fiber.Ctx) error {
	var req ProcessAIMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.ProspectNum == "" || req.IDDevice == "" || req.Message == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Missing required fields")
	}

	// Process AI conversation
	response, err := h.AIWhatsappService.ProcessAIConversation(req.ProspectNum, req.IDDevice, req.Message, req.Stage)
	if err != nil {
		logrus.WithError(err).Error("Failed to process AI conversation")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to process AI message")
	}

	return h.successResponse(c, response)
}

// ToggleHumanTakeover toggles human takeover for a conversation
func (h *AIWhatsappHandlers) ToggleHumanTakeover(c *fiber.Ctx) error {
	var req ToggleHumanTakeoverRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	if req.ProspectNum == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Prospect number is required")
	}

	err := h.AIWhatsappService.ToggleHumanTakeover(req.ProspectNum, req.Human)
	if err != nil {
		logrus.WithError(err).Error("Failed to toggle human takeover")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to toggle human takeover")
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num": req.ProspectNum,
		"human":        req.Human,
	}).Info("Human takeover toggled")

	return h.successResponse(c, map[string]interface{}{
		"prospect_num": req.ProspectNum,
		"human":        req.Human,
		"status":       "updated",
	})
}

// GetConversationHistory retrieves conversation history for a prospect
func (h *AIWhatsappHandlers) GetConversationHistory(c *fiber.Ctx) error {
	prospectNum := c.Params("prospect_num")
	if prospectNum == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Prospect number is required")
	}

	// Get limit from query parameter
	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	history, err := h.AIRepo.GetConversationHistory(prospectNum, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation history")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get conversation history")
	}

	return h.successResponse(c, history)
}

// GetConversationStatus retrieves conversation status for a prospect
func (h *AIWhatsappHandlers) GetConversationStatus(c *fiber.Ctx) error {
	prospectNum := c.Params("prospect_num")
	if prospectNum == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Prospect number is required")
	}

	aiConv, err := h.AIRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation status")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get conversation status")
	}

	if aiConv == nil {
		return h.errorResponse(c, fiber.StatusNotFound, "Conversation not found")
	}

	return h.successResponse(c, aiConv)
}

// GetAISettings retrieves AI settings for a staff member
func (h *AIWhatsappHandlers) GetAISettings(c *fiber.Ctx) error {
	staffID := c.Params("staff_id")
	if staffID == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid staff ID")
	}

	settings, err := h.AIWhatsappService.GetAISettings(staffID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI settings")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get AI settings")
	}

	if settings == nil {
		return h.errorResponse(c, fiber.StatusNotFound, "AI settings not found")
	}

	return h.successResponse(c, settings)
}

// CreateAISettings creates new AI settings
func (h *AIWhatsappHandlers) CreateAISettings(c *fiber.Ctx) error {
	var settings models.AISettings
	if err := c.BodyParser(&settings); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if settings.IDDevice == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Device ID is required")
	}

	// TODO: Implement CreateAISettings method in repository
	logrus.Info("AI settings creation requested but not implemented yet")
	return h.errorResponse(c, fiber.StatusNotImplemented, "AI settings creation not implemented yet")
}

// UpdateAISettings updates existing AI settings
func (h *AIWhatsappHandlers) UpdateAISettings(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid settings ID")
	}

	var settings models.AISettings
	if err := c.BodyParser(&settings); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	settings.ID = idStr
	// TODO: Implement UpdateAISettings method in repository
	logrus.Info("AI settings update requested but not implemented yet")
	return h.errorResponse(c, fiber.StatusNotImplemented, "AI settings update not implemented yet")
}

// TestWahaExtraction tests the WAHA webhook data extraction
// Returns extracted fields in the standardized JSON format for testing
func (h *AIWhatsappHandlers) TestWahaExtraction(c *fiber.Ctx) error {
	// Parse the incoming webhook request as raw map
	var rawPayload map[string]interface{}
	if err := c.BodyParser(&rawPayload); err != nil {
		logrus.WithError(err).Error("❌ WAHA TEST: Failed to parse webhook request")
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request format",
			"details": err.Error(),
		})
	}

	// Log the raw payload structure for debugging
	logrus.WithFields(logrus.Fields{
		"raw_payload": rawPayload,
		"payload_keys": getMapKeys(rawPayload),
	}).Info("🧪 WAHA TEST: Raw payload received")
	
	// Extract standardized webhook data
	extractedData := h.extractWahaWebhookData(rawPayload)

	// Log the test extraction
	logrus.WithFields(logrus.Fields{
		"sender_phone": extractedData.SenderPhone,
		"sender_name": extractedData.SenderName,
		"message": truncateString(extractedData.Message, 100),
		"is_from_me": extractedData.IsFromMe,
		"is_group": extractedData.IsGroup,
	}).Info("🧪 WAHA TEST: Extraction completed")

	// Return extracted fields in standardized JSON format as specified
	return c.JSON(extractedData)
}

// DeleteAISettings deletes AI settings
func (h *AIWhatsappHandlers) DeleteAISettings(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid settings ID")
	}

	// TODO: Implement DeleteAISettings method in repository
	logrus.Info("AI settings deletion requested but not implemented yet")
	return h.errorResponse(c, fiber.StatusNotImplemented, "AI settings deletion not implemented yet")
}

// ProcessDeviceCommand processes device-specific commands
func (h *AIWhatsappHandlers) ProcessDeviceCommand(c *fiber.Ctx) error {
	var req ProcessDeviceCommandRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	if req.ProspectNum == "" || req.Command == "" || req.IDDevice == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Missing required fields")
	}

	err := h.AIWhatsappService.ProcessDeviceCommand(req.ProspectNum, req.Command, req.IDDevice)
	if err != nil {
		logrus.WithError(err).Error("Failed to process device command")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to process device command")
	}

	return h.successResponse(c, map[string]string{"status": "processed"})
}

// processIncomingMessage processes incoming WhatsApp messages asynchronously
func (h *AIWhatsappHandlers) processIncomingMessage(prospectNum, message, deviceID, provider string) {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"device_id":    deviceID,
		"provider":     provider,
		"message":      message,
	}).Info("Processing incoming message")

	// Check if this is a device command
	if strings.HasPrefix(message, "%") || strings.HasPrefix(message, "#") || strings.ToLower(message) == "cmd" {
		err := h.AIWhatsappService.ProcessDeviceCommand(prospectNum, message, deviceID)
		if err != nil {
			logrus.WithError(err).Error("Failed to process device command")
		}
		return
	}

	// Get current conversation stage
	aiConv, err := h.AIRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI conversation")
		return
	}

	var stage string
	if aiConv != nil && aiConv.Stage.Valid {
		stage = aiConv.Stage.String
	}

	// Process AI conversation
	response, err := h.AIWhatsappService.ProcessAIConversation(prospectNum, deviceID, message, stage)
	if err != nil {
		logrus.WithError(err).Error("Failed to process AI conversation")
		return
	}

	// Send response if we have a response
	// Note: ProcessAIConversation already handles conversation logging internally via LogConversation
	// Removed duplicate SaveConversationHistory call to prevent duplicate saves
	if response != nil {
		// Send response back to WhatsApp
		h.sendWhatsappResponse(prospectNum, deviceID, provider, response)
	}
}

// sendWhatsappResponse sends AI response back to WhatsApp
func (h *AIWhatsappHandlers) sendWhatsappResponse(prospectNum, deviceID, provider string, response *services.AIWhatsappResponse) {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"device_id":    deviceID,
		"provider":     provider,
		"stage":        response.Stage,
	}).Info("Sending AI response to WhatsApp")

	// TODO: Implement actual WhatsApp sending logic based on provider
	// This would integrate with your existing WhatsApp service
	for _, item := range response.Response {
		if item.Type == "text" {
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"content":      item.Content,
			}).Info("Sending text message")
			// Send text message
		} else if item.Type == "image" {
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"image_url":    item.Content,
			}).Info("Sending image message")
			// Send image message
		}
	}
}

// GetAnalytics retrieves analytics data from ai_whatsapp_nodepath with date filtering
func (h *AIWhatsappHandlers) GetAnalytics(c *fiber.Ctx) error {
	var req AnalyticsRequest

	// Parse query parameters with frontend parameter names or JSON body
	if c.Method() == "GET" {
		// Handle GET request with query parameters
		req.StartDate = c.Query("startDate", "")
		req.EndDate = c.Query("endDate", "")
		req.DeviceID = c.Query("idDevice", "")
		
		// Handle deviceIds parameter (comma-separated list from frontend)
		deviceIds := c.Query("deviceIds", "")
		logrus.WithFields(logrus.Fields{
			"deviceIds": deviceIds,
			"idDevice": req.DeviceID,
			"startDate": req.StartDate,
			"endDate": req.EndDate,
		}).Info("Analytics request received")
		
		if deviceIds != "" && req.DeviceID == "" {
			// Use the first device ID from the list for now
			// TODO: Enhance repository to handle multiple device IDs
			req.DeviceID = "all" // Set to "all" to include all user devices
			logrus.Info("Using all devices for analytics since deviceIds parameter was provided")
		}
	} else {
		// Handle POST request with JSON body
		if err := c.BodyParser(&req); err != nil {
			logrus.WithError(err).Error("Failed to parse analytics request")
			return c.Status(fiber.StatusBadRequest).JSON(AnalyticsResponse{
				Success: false,
				Message: "Invalid request format",
			})
		}
	}

	// Set default date range if not provided (current month start to today)
	now := time.Now()
	var startDate, endDate time.Time
	var err error

	if req.StartDate == "" {
		// Default to first day of current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	} else {
		startDate, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			logrus.WithError(err).Error("Invalid start date format")
			return c.Status(fiber.StatusBadRequest).JSON(AnalyticsResponse{
				Success: false,
				Message: "Invalid start date format. Use YYYY-MM-DD",
			})
		}
	}

	if req.EndDate == "" {
		// Default to today
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	} else {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			logrus.WithError(err).Error("Invalid end date format")
			return c.Status(fiber.StatusBadRequest).JSON(AnalyticsResponse{
				Success: false,
				Message: "Invalid end date format. Use YYYY-MM-DD",
			})
		}
		// Set end time to end of day
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
	}

	// Validate date range
	if startDate.After(endDate) {
		return c.Status(fiber.StatusBadRequest).JSON(AnalyticsResponse{
			Success: false,
			Message: "Start date cannot be after end date",
		})
	}

	// Set default device filter to "all" if not provided
	if req.DeviceID == "" {
		req.DeviceID = "all"
	}

	// Get user ID from authentication context
	userID, ok := c.Locals("userID").(int)
	if !ok {
		logrus.Error("User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(AnalyticsResponse{
			Success: false,
			Message: "Authentication required",
		})
	}

	// Get analytics data from repository with user-specific filtering
	analyticsData, err := h.AIRepo.GetAnalyticsData(startDate, endDate, req.DeviceID, userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get analytics data")
		return c.Status(fiber.StatusInternalServerError).JSON(AnalyticsResponse{
			Success: false,
			Message: "Failed to retrieve analytics data",
		})
	}

	// Log successful analytics request
	logrus.WithFields(logrus.Fields{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"device_id":  req.DeviceID,
	}).Info("Analytics data retrieved successfully")

	// Transform data to match frontend expectations
	summary := analyticsData["summary"].(map[string]interface{})
	
	// Transform stage distribution from array to object format
	stageDistributionArray := analyticsData["stage_distribution"].([]map[string]interface{})
	stageDistributionMap := make(map[string]interface{})
	for _, item := range stageDistributionArray {
		stage := item["stage"].(string)
		count := item["count"]
		stageDistributionMap[stage] = count
	}
	
	responseData := map[string]interface{}{
		"totalConversations":       summary["total_conversations"],
		"aiActiveConversations":    summary["ai_active"],
		"humanTakeovers":           summary["human_takeover"],
		"uniqueDevices":            summary["unique_devices"],
		"uniqueNiches":             summary["unique_niches"],
		"conversationsWithStages":  summary["conversations_with_stage"],
		"aiActivePercentage":       summary["ai_active_percentage"],
		"humanTakeoverPercentage":  summary["human_takeover_percentage"],
		"dailyBreakdown":           analyticsData["daily_data"],
		"stageDistribution":        stageDistributionMap,
		"dateRange":                analyticsData["date_range"],
	}

	return c.JSON(responseData)
}

// GetAllAIWhatsappData retrieves all AI WhatsApp conversation records for data table display
func (h *AIWhatsappHandlers) GetAllAIWhatsappData(c *fiber.Ctx) error {
	// Parse query parameters for pagination and filtering
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	deviceFilter := c.Query("device_id", "")
	stageFilter := c.Query("stage", "")
	search := c.Query("search", "")

	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Get user ID from authentication context
	userID, ok := c.Locals("userID").(int)
	if !ok {
		logrus.Error("User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Authentication required",
		})
	}

	// Get data from repository with user-specific filtering
	data, total, err := h.AIRepo.GetAllAIWhatsappData(limit, offset, deviceFilter, stageFilter, search, userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp data")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve AI WhatsApp data",
		})
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit

	// Return paginated response
	return c.JSON(fiber.Map{
		"success": true,
		"data": data,
		"pagination": fiber.Map{
			"current_page": page,
			"total_pages":  totalPages,
			"total_records": total,
			"limit":        limit,
		},
	})
}

// DeleteAIWhatsappData deletes an AI WhatsApp conversation record by ID
func (h *AIWhatsappHandlers) DeleteAIWhatsappData(c *fiber.Ctx) error {
	// Get ID from URL parameter
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		logrus.WithError(err).Error("Invalid ID parameter")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid ID parameter",
		})
	}

	// Get user ID from authentication context
	userID, ok := c.Locals("userID").(int)
	if !ok {
		logrus.Error("User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Authentication required",
		})
	}

	// First, verify the record exists and belongs to the user's devices
	record, err := h.AIRepo.GetAIWhatsappByID(id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp record")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve record",
		})
	}

	if record == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Record not found",
		})
	}

	// Verify the record belongs to a device owned by the user
	deviceSettings, err := h.DeviceRepo.GetDeviceSettingsByDevice(record.IDDevice)
	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to verify device ownership",
		})
	}

	if !deviceSettings.UserID.Valid || int(deviceSettings.UserID.Int32) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Access denied: record belongs to different user",
		})
	}

	// Delete the record
	err = h.AIRepo.DeleteAIWhatsapp(id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete AI WhatsApp record")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete record",
		})
	}

	logrus.WithFields(logrus.Fields{
		"id_prospect": id,
		"user_id":     userID,
		"id_device":   record.IDDevice,
	}).Info("AI WhatsApp record deleted successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Record deleted successfully",
	})
}

// Helper methods for consistent response formatting
func (h *AIWhatsappHandlers) successResponse(c *fiber.Ctx, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

func (h *AIWhatsappHandlers) errorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}

// Helper functions for comprehensive WAHA webhook debugging

// getMapKeys returns all keys from a map for debugging payload structure
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// analyzePayloadDepth analyzes the depth and structure of nested payload
func analyzePayloadDepth(payload map[string]interface{}) map[string]interface{} {
	analysis := make(map[string]interface{})
	
	for key, value := range payload {
		switch v := value.(type) {
		case map[string]interface{}:
			analysis[key] = map[string]interface{}{
				"type": "object",
				"keys": getMapKeys(v),
				"key_count": len(v),
			}
		case []interface{}:
			analysis[key] = map[string]interface{}{
				"type": "array",
				"length": len(v),
			}
		case string:
			analysis[key] = map[string]interface{}{
				"type": "string",
				"length": len(v),
			}
		default:
			analysis[key] = map[string]interface{}{
				"type": fmt.Sprintf("%T", v),
			}
		}
	}
	
	return analysis
}

// truncateString truncates a string to specified length for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// DebugWahaWebhook is a special debug endpoint for production WAHA webhook debugging
// Logs complete payload structure and returns detailed analysis without processing
func (h *AIWhatsappHandlers) DebugWahaWebhook(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	body := c.Body()
	
	// Log ALL request details for production debugging
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"payload_size": len(body),
		"content_type": c.Get("Content-Type"),
		"user_agent": c.Get("User-Agent"),
		"headers": headers,
		"raw_payload": string(body),
		"method": c.Method(),
		"url": c.OriginalURL(),
		"ip": c.IP(),
	}).Error("🚨 WAHA DEBUG ENDPOINT: Complete webhook request details")
	
	// Parse as generic map for structure analysis
	var rawPayload map[string]interface{}
	if err := json.Unmarshal(body, &rawPayload); err != nil {
		logrus.WithError(err).Error("🚨 WAHA DEBUG: Failed to parse JSON")
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error": "Invalid JSON format",
			"raw_body": string(body),
		})
	}
	
	// Perform complete payload analysis
	payloadAnalysis := analyzePayloadDepth(rawPayload)
	extractedData := h.extractWahaWebhookData(rawPayload)
	
	// Log detailed analysis
	logrus.WithFields(logrus.Fields{
		"payload_keys": getMapKeys(rawPayload),
		"payload_analysis": payloadAnalysis,
		"extracted_data": extractedData,
		"extraction_success": extractedData.SenderPhone != "" && extractedData.Message != "",
	}).Error("🚨 WAHA DEBUG: Complete payload analysis")
	
	// Return comprehensive debug information
	return c.JSON(fiber.Map{
		"success": true,
		"debug_info": fiber.Map{
			"device_id": deviceID,
			"payload_size": len(body),
			"headers": headers,
			"raw_payload": rawPayload,
			"payload_keys": getMapKeys(rawPayload),
			"payload_analysis": payloadAnalysis,
			"extracted_data": extractedData,
			"extraction_success": extractedData.SenderPhone != "" && extractedData.Message != "",
			"timestamp": time.Now().Unix(),
		},
		"message": "Debug data logged successfully",
	})
}