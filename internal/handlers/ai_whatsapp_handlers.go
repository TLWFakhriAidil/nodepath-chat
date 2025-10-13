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
	protected.Put("/ai-whatsapp/:id/human", h.UpdateHumanStatus)
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
	Phone    string `json:"phone"`
	Message  string `json:"message"`
	Device   string `json:"device"`
	Time     string `json:"time"`
	IsFromMe bool   `json:"isFromMe"`
}

// WhacenterWebhookRequest represents incoming Whacenter webhook data
type WhacenterWebhookRequest struct {
	Number string `json:"number"`
	Text   string `json:"text"`
	Device string `json:"device"`
	Date   string `json:"date"`
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
	ProspectNum  string `json:"prospect_num"`
	IDDevice     string `json:"id_device"`
	ProspectName string `json:"prospect_name"`
	Niche        string `json:"niche"`
	Stage        string `json:"stage"`
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
	go h.processIncomingMessage(req.From, req.Message, deviceID, "whatsapp", req.From)

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
		"device_id":  deviceID,
		"phone":      req.Phone,
		"message":    req.Message,
		"is_from_me": req.IsFromMe,
	}).Info("Received Wablas webhook")

	// WABLAS Command Processing (only when isFromMe=true, matching PHP logic)
	if req.IsFromMe {
		cleanText := strings.TrimSpace(req.Message)

		// Command 1: Text starts with '%' → set text to "Teruskan", name to "Sis"
		if len(cleanText) > 0 && cleanText[0] == '%' {
			logrus.WithFields(logrus.Fields{
				"device_id": deviceID,
				"phone":     req.Phone,
			}).Info("🔧 WABLAS: Processing % command - set text to 'Teruskan'")

			// Update message and continue processing
			req.Message = "Teruskan"
			senderName := "Sis"

			// Process through standardized flow
			go h.processIncomingMessage(req.Phone, req.Message, deviceID, "wablas", senderName)

			return c.JSON(fiber.Map{
				"status": "success",
				"type":   "percent_command",
			})
		}

		// Command 2: Text equals 'cmd' → set human=1, return empty
		if cleanText == "cmd" {
			logrus.WithFields(logrus.Fields{
				"device_id": deviceID,
				"phone":     req.Phone,
			}).Info("🔧 WABLAS: Processing cmd command - set human=1")

			go func() {
				whats, err := h.AIRepo.GetAIWhatsappByProspectAndDevice(req.Phone, deviceID)
				if err == nil && whats != nil {
					whats.Human = 1
					h.AIRepo.UpdateAIWhatsapp(whats)
					logrus.Info("✅ WABLAS: Successfully set human=1 for cmd command")
				}
			}()

			return c.JSON(fiber.Map{
				"status": "success",
				"type":   "cmd_command",
			})
		}

		// Command 3: Text equals 'dmc' → set human=null, return empty
		if cleanText == "dmc" {
			logrus.WithFields(logrus.Fields{
				"device_id": deviceID,
				"phone":     req.Phone,
			}).Info("🔧 WABLAS: Processing dmc command - set human=null")

			go func() {
				whats, err := h.AIRepo.GetAIWhatsappByProspectAndDevice(req.Phone, deviceID)
				if err == nil && whats != nil {
					whats.Human = 0 // Set to 0 (null equivalent in Go)
					h.AIRepo.UpdateAIWhatsapp(whats)
					logrus.Info("✅ WABLAS: Successfully set human=null for dmc command")
				}
			}()

			return c.JSON(fiber.Map{
				"status": "success",
				"type":   "dmc_command",
			})
		}

		// Other isFromMe messages → ignore
		logrus.Info("⏭️ WABLAS: Ignoring other isFromMe message (not %, cmd, or dmc)")
		return c.JSON(fiber.Map{
			"status": "ignored",
			"reason": "isFromMe message (not %, cmd, or dmc)",
		})
	}

	// Validate phone number length (matching PHP: if strlen($wa_no) > 13 return)
	if len(req.Phone) > 13 {
		logrus.WithFields(logrus.Fields{
			"device_id":    deviceID,
			"phone":        req.Phone,
			"phone_length": len(req.Phone),
		}).Warn("⚠️ WABLAS: Phone number length exceeds 13 characters - terminating")
		return c.JSON(fiber.Map{
			"status": "ignored",
			"reason": "phone number length > 13",
		})
	}

	// Process the message asynchronously
	go h.processIncomingMessage(req.Phone, req.Message, deviceID, "wablas", req.Phone)

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

	// WHACENTER Command Processing (matching PHP logic)
	cleanText := strings.TrimSpace(req.Text)
	phoneNumber := req.Number
	message := req.Text
	senderName := "Sis"

	// Command 1: Text starts with '#' → extract phone, set text to "Teruskan"
	if len(cleanText) > 0 && cleanText[0] == '#' {
		phoneNumber = cleanText[1:]
		message = "Teruskan"
		senderName = "Sis"

		logrus.WithFields(logrus.Fields{
			"device_id":       deviceID,
			"extracted_phone": phoneNumber,
		}).Info("🔧 WHACENTER: Processing # command - extract phone and set text to 'Teruskan'")
	}

	// Command 2: Text starts with '/' → extract phone, set human=1, return empty
	if len(cleanText) > 0 && cleanText[0] == '/' {
		extractedPhone := cleanText[1:]
		logrus.WithFields(logrus.Fields{
			"device_id":       deviceID,
			"extracted_phone": extractedPhone,
		}).Info("🔧 WHACENTER: Processing / command - set human=1")

		go func() {
			whats, err := h.AIRepo.GetAIWhatsappByProspectAndDevice(extractedPhone, deviceID)
			if err == nil && whats != nil {
				whats.Human = 1
				h.AIRepo.UpdateAIWhatsapp(whats)
				logrus.Info("✅ WHACENTER: Successfully set human=1 for / command")
			}
		}()

		return c.JSON(fiber.Map{
			"status": "success",
			"type":   "slash_command",
		})
	}

	// Command 3: Text starts with '?' → extract phone, set human=null, return empty
	if len(cleanText) > 0 && cleanText[0] == '?' {
		extractedPhone := cleanText[1:]
		logrus.WithFields(logrus.Fields{
			"device_id":       deviceID,
			"extracted_phone": extractedPhone,
		}).Info("🔧 WHACENTER: Processing ? command - set human=null")

		go func() {
			whats, err := h.AIRepo.GetAIWhatsappByProspectAndDevice(extractedPhone, deviceID)
			if err == nil && whats != nil {
				whats.Human = 0 // Set to 0 (null equivalent in Go)
				h.AIRepo.UpdateAIWhatsapp(whats)
				logrus.Info("✅ WHACENTER: Successfully set human=null for ? command")
			}
		}()

		return c.JSON(fiber.Map{
			"status": "success",
			"type":   "question_command",
		})
	}

	// Validate phone number length (matching PHP: if strlen($wa_no) > 13 return)
	if len(phoneNumber) > 13 {
		logrus.WithFields(logrus.Fields{
			"device_id":    deviceID,
			"number":       phoneNumber,
			"phone_length": len(phoneNumber),
		}).Warn("⚠️ WHACENTER: Phone number length exceeds 13 characters - terminating")
		return c.JSON(fiber.Map{
			"status": "ignored",
			"reason": "phone number length > 13",
		})
	}

	// Process the message asynchronously
	go h.processIncomingMessage(phoneNumber, message, deviceID, "whacenter", senderName)

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
		"device_id":    deviceID,
		"payload_size": len(body),
		"content_type": c.Get("Content-Type"),
		"user_agent":   c.Get("User-Agent"),
		"headers":      headers,
		"raw_payload":  string(body),
		"method":       c.Method(),
		"url":          c.OriginalURL(),
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
			"sender_name":  extractedData.SenderName,
			"device_id":    deviceID,
		}).Info("⏭️ WAHA: Ignoring group message as per requirements")
		return c.JSON(fiber.Map{
			"status":         "ignored",
			"reason":         "group message",
			"extracted_data": extractedData,
		})
	}

	// Validate required fields
	if extractedData.SenderPhone == "" || extractedData.Message == "" {
		logrus.WithFields(logrus.Fields{
			"sender_phone": extractedData.SenderPhone,
			"message":      truncateString(extractedData.Message, 100),
			"sender_name":  extractedData.SenderName,
			"is_from_me":   extractedData.IsFromMe,
			"is_group":     extractedData.IsGroup,
		}).Warn("⚠️ WAHA: Missing required fields in extracted data")
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing required fields",
			"missing": map[string]bool{
				"sender_phone": extractedData.SenderPhone == "",
				"message":      extractedData.Message == "",
			},
			"extracted_data": extractedData,
		})
	}

	// Clean phone number format (remove @c.us suffix if present)
	if strings.HasSuffix(extractedData.SenderPhone, "@c.us") {
		extractedData.SenderPhone = strings.TrimSuffix(extractedData.SenderPhone, "@c.us")
		logrus.WithFields(logrus.Fields{
			"device_id":    deviceID,
			"cleaned_from": extractedData.SenderPhone,
		}).Info("🔧 WAHA: Phone number cleaned - stripped @c.us suffix")
	}

	// SECOND PASS: Validate phone number length and re-extract if needed (matching PHP: if strlen($wa_no) > 13)
	if len(extractedData.SenderPhone) > 13 {
		logrus.WithFields(logrus.Fields{
			"device_id":    deviceID,
			"sender_phone": extractedData.SenderPhone,
			"phone_length": len(extractedData.SenderPhone),
			"raw_phone":    extractedData.RawPhone,
		}).Warn("⚠️ WAHA: Phone number length exceeds 13 - attempting re-extraction from raw")

		// Re-extract phone using suffix logic (matching PHP second pass)
		idNoWaha := extractedData.RawPhone
		var finalPhone string

		if strings.HasSuffix(idNoWaha, "@c.us") {
			// Normal contact - extract number before @
			finalPhone = strings.Split(idNoWaha, "@")[0]
			logrus.WithField("from", finalPhone).Info("🔍 WAHA: Re-extracted from @c.us")
		} else if strings.HasSuffix(idNoWaha, "@g.us") {
			// Group - skip processing
			logrus.Info("🔍 WAHA: Detected @g.us group - terminating")
			return c.JSON(fiber.Map{
				"status": "ignored",
				"reason": "group message (@g.us)",
			})
		} else if strings.HasSuffix(idNoWaha, "@lid") {
			// LID mapping - try SenderAlt or RecipientAlt
			logrus.Info("🔍 WAHA: Detected @lid - attempting LID mapping")

			var payload map[string]interface{}
			var rawPayload map[string]interface{}
			json.Unmarshal(body, &rawPayload)
			if payloadData, ok := rawPayload["payload"].(map[string]interface{}); ok {
				payload = payloadData
			} else {
				payload = rawPayload
			}

			if _dataObj, ok := payload["_data"].(map[string]interface{}); ok {
				if infoObj, ok := _dataObj["Info"].(map[string]interface{}); ok {
					senderAlt := strings.TrimSpace(getStringValue(infoObj["SenderAlt"]))
					recipientAlt := strings.TrimSpace(getStringValue(infoObj["RecipientAlt"]))

					logrus.WithFields(logrus.Fields{
						"senderAlt":    senderAlt,
						"recipientAlt": recipientAlt,
					}).Info("🔍 WAHA: LID mapping alternatives")

					// Try both alternatives
					for _, alt := range []string{senderAlt, recipientAlt} {
						if alt == "" {
							continue
						}
						if strings.HasSuffix(alt, "@c.us") {
							finalPhone = strings.Split(alt, "@")[0]
							logrus.WithField("from", finalPhone).Info("🔍 WAHA: LID mapped via @c.us")
							break
						} else if strings.HasSuffix(alt, "@s.whatsapp.net") {
							finalPhone = strings.Split(alt, "@")[0]
							logrus.WithField("from", finalPhone).Info("🔍 WAHA: LID mapped via @s.whatsapp.net")
							break
						}
					}
				}
			}

			if finalPhone == "" {
				logrus.Warn("🔍 WAHA: LID mapping failed - terminating")
				return c.JSON(fiber.Map{
					"status": "ignored",
					"reason": "LID mapping failed",
				})
			}
		} else {
			// Unknown format - terminate
			logrus.Warn("⚠️ WAHA: Unknown phone format after validation - terminating")
			return c.JSON(fiber.Map{
				"status": "ignored",
				"reason": "unknown phone format",
				"phone":  idNoWaha,
			})
		}

		// Update extracted phone with the re-extracted value
		extractedData.SenderPhone = finalPhone
		logrus.WithField("final_phone", finalPhone).Info("✅ WAHA: Phone re-extracted successfully")
	}

	// WAHA Command Processing (matching PHP logic exactly)
	cleanText := strings.TrimSpace(extractedData.Message)

	// Command 1: Text starts with '0' → extract phone, set text to "Teruskan"
	if len(cleanText) > 0 && cleanText[0] == '0' {
		extractedPhone := cleanText[1:]
		logrus.WithFields(logrus.Fields{
			"device_id":       deviceID,
			"original_phone":  extractedData.SenderPhone,
			"extracted_phone": extractedPhone,
		}).Info("🔧 WAHA: Processing 0 command - extract phone and set text to 'Teruskan'")

		extractedData.SenderPhone = extractedPhone
		extractedData.Message = "Teruskan"
		extractedData.SenderName = "Sis"

		// Process through standardized flow
		webhookData := map[string]interface{}{
			"from":         extractedData.SenderPhone,
			"message":      extractedData.Message,
			"message_type": "text",
			"is_group":     extractedData.IsGroup,
			"sender_name":  extractedData.SenderName,
		}

		go func() {
			if h.mainHandlers != nil {
				h.mainHandlers.processWebhookMessage(webhookData, deviceID, "waha")
			}
		}()

		return c.JSON(fiber.Map{
			"status": "success",
			"type":   "0_command",
		})
	}

	// Command 2: Text starts with '/' → extract phone, set human=1, return empty
	if len(cleanText) > 0 && cleanText[0] == '/' {
		extractedPhone := cleanText[1:]
		logrus.WithFields(logrus.Fields{
			"device_id":       deviceID,
			"extracted_phone": extractedPhone,
		}).Info("🔧 WAHA: Processing / command - set human=1")

		go func() {
			whats, err := h.AIRepo.GetAIWhatsappByProspectAndDevice(extractedPhone, deviceID)
			if err == nil && whats != nil {
				whats.Human = 1
				h.AIRepo.UpdateAIWhatsapp(whats)
				logrus.Info("✅ WAHA: Successfully set human=1 for / command")
			}
		}()

		return c.JSON(fiber.Map{
			"status": "success",
			"type":   "slash_command",
		})
	}

	// Command 3: Text starts with '!' → extract phone, set human=null, return empty
	if len(cleanText) > 0 && cleanText[0] == '!' {
		extractedPhone := cleanText[1:]
		logrus.WithFields(logrus.Fields{
			"device_id":       deviceID,
			"extracted_phone": extractedPhone,
		}).Info("🔧 WAHA: Processing ! command - set human=null")

		go func() {
			whats, err := h.AIRepo.GetAIWhatsappByProspectAndDevice(extractedPhone, deviceID)
			if err == nil && whats != nil {
				whats.Human = 0 // Set to 0 (null equivalent in Go)
				h.AIRepo.UpdateAIWhatsapp(whats)
				logrus.Info("✅ WAHA: Successfully set human=null for ! command")
			}
		}()

		return c.JSON(fiber.Map{
			"status": "success",
			"type":   "exclamation_command",
		})
	}

	// Logic 3: Otherwise → treat as normal customer message
	logrus.WithFields(logrus.Fields{
		"sender_phone": extractedData.SenderPhone,
		"sender_name":  extractedData.SenderName,
		"message":      truncateString(extractedData.Message, 100),
		"device_id":    deviceID,
	}).Info("💬 WAHA: Processing normal customer message through standardized flow routing")

	// STANDARDIZED FLOW ROUTING: Use the same flow processing logic as Whacenter
	// Create webhook data structure compatible with processWebhookMessage
	webhookData := map[string]interface{}{
		"from":         extractedData.SenderPhone,
		"message":      extractedData.Message,
		"message_type": "text",
		"is_group":     extractedData.IsGroup,
		"sender_name":  extractedData.SenderName,
		"is_from_me":   extractedData.IsFromMe,
	}

	// Route through the standardized webhook processing system
	// This ensures WAHA follows the same flow node logic as Whacenter
	go func() {
		if h.mainHandlers != nil {
			err := h.mainHandlers.processWebhookMessage(webhookData, deviceID, "waha")
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"device_id":    deviceID,
					"sender_phone": extractedData.SenderPhone,
				}).Error("❌ WAHA: Failed to process message through standardized flow routing")
			} else {
				logrus.WithFields(logrus.Fields{
					"device_id":    deviceID,
					"sender_phone": extractedData.SenderPhone,
				}).Info("✅ WAHA: Successfully processed message through standardized flow routing")
			}
		} else {
			logrus.Error("❌ WAHA: Main handlers not available, falling back to direct AI processing")
			// Fallback to direct processing if main handlers not available
			h.processIncomingMessage(extractedData.SenderPhone, extractedData.Message, deviceID, "waha", extractedData.SenderName)
		}
	}()

	return c.JSON(fiber.Map{
		"status":         "success",
		"type":           "customer_message",
		"routing":        "standardized_flow",
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
	RawPhone    string `json:"raw_phone"` // Keep original for second pass
	SenderName  string `json:"sender_name"`
	Message     string `json:"message"`
	IsFromMe    bool   `json:"is_from_me"`
	IsGroup     bool   `json:"is_group"`
}

// extractWahaWebhookData extracts WAHA webhook data using the exact format specified by user
// Implements the required extraction logic: $payload = $data['payload'], etc.
// Handles isFromMe messages with specific command processing
func (h *AIWhatsappHandlers) extractWahaWebhookData(webhookPayload map[string]interface{}) WahaWebhookData {
	var result WahaWebhookData

	logrus.WithFields(logrus.Fields{
		"payload_keys": getMapKeys(webhookPayload),
		"has_payload":  webhookPayload["payload"] != nil,
	}).Info("🔍 WAHA: Starting data extraction with user-specified format")

	// Extract using exact user-specified format: $payload = $data['payload']
	var payload map[string]interface{}
	if payloadData, ok := webhookPayload["payload"].(map[string]interface{}); ok {
		payload = payloadData
		logrus.Info("🔍 WAHA: Using nested payload structure")
	} else {
		// Fallback to direct structure if no nested payload
		payload = webhookPayload
		logrus.Info("🔍 WAHA: Using direct payload structure as fallback")
	}

	// Check for _data nested structure (common in WAHA webhooks)
	var dataPayload map[string]interface{}
	if dataObj, ok := payload["_data"].(map[string]interface{}); ok {
		dataPayload = dataObj
		logrus.Info("🔍 WAHA: Found _data nested structure")
	} else {
		dataPayload = payload
		logrus.Info("🔍 WAHA: Using direct payload structure")
	}

	// Extract data using user-specified format:
	// $wa_text = $payload['body'] or $payload['_data']['body']
	if bodyVal, ok := dataPayload["body"].(string); ok {
		result.Message = bodyVal
		logrus.WithField("extraction_method", "data_body").Info("🔍 WAHA: Message extracted from data.body")
	} else if bodyVal, ok := payload["body"].(string); ok {
		result.Message = bodyVal
		logrus.WithField("extraction_method", "payload_body").Info("🔍 WAHA: Message extracted from payload.body")
	}

	// $wa_no_raw = $payload['from'] or $payload['_data']['from'] ?? null
	// FIRST PASS: Extract RAW phone (with suffix intact)
	var idNoWaha string
	if fromVal, ok := dataPayload["from"].(string); ok {
		idNoWaha = fromVal
		logrus.WithField("extraction_method", "data_from").Info("🔍 WAHA: Sender phone extracted from data.from")
	} else if fromVal, ok := payload["from"].(string); ok {
		idNoWaha = fromVal
		logrus.WithField("extraction_method", "payload_from").Info("🔍 WAHA: Sender phone extracted from payload.from")
	}

	// Store raw phone without processing (matching PHP: first get raw value)
	result.SenderPhone = idNoWaha
	result.RawPhone = idNoWaha // Keep original for second pass validation

	// $wa_nama = $payload['_data']['Info']['PushName'] ?? 'Sis'
	if _dataObj, ok := payload["_data"].(map[string]interface{}); ok {
		if infoObj, ok := _dataObj["Info"].(map[string]interface{}); ok {
			if pushNameVal, ok := infoObj["PushName"].(string); ok {
				result.SenderName = pushNameVal
				logrus.WithField("extraction_method", "payload._data.Info.PushName").Info("🔍 WAHA: Sender name extracted from payload._data.Info.PushName")
			}
		}
	}

	// Default to 'Sis' if no sender name found
	if result.SenderName == "" {
		result.SenderName = "Sis"
		logrus.Info("🔍 WAHA: Using default sender name 'Sis'")
	}

	// Extract isFromMe for special handling
	if infoObj, ok := dataPayload["info"].(map[string]interface{}); ok {
		if isFromMeVal, ok := infoObj["fromMe"].(bool); ok {
			result.IsFromMe = isFromMeVal
			logrus.WithField("is_from_me", isFromMeVal).Info("🔍 WAHA: IsFromMe extracted from data.info.fromMe")
		}
	} else if isFromMeVal, ok := payload["isFromMe"].(bool); ok {
		result.IsFromMe = isFromMeVal
		logrus.WithField("is_from_me", isFromMeVal).Info("🔍 WAHA: IsFromMe extracted from payload.isFromMe")
	}

	// Extract additional fields for completeness
	if infoObj, ok := dataPayload["info"].(map[string]interface{}); ok {
		if isGroupVal, ok := infoObj["isGroup"].(bool); ok {
			result.IsGroup = isGroupVal
		}
	} else if mediaObj, ok := payload["media"].(map[string]interface{}); ok {
		if infoObj, ok := mediaObj["Info"].(map[string]interface{}); ok {
			if isGroupVal, ok := infoObj["IsGroup"].(bool); ok {
				result.IsGroup = isGroupVal
			}
		}
	}

	// Log extraction results with production debugging
	logrus.WithFields(logrus.Fields{
		"sender_phone":       result.SenderPhone,
		"sender_name":        result.SenderName,
		"message":            truncateString(result.Message, 100),
		"is_from_me":         result.IsFromMe,
		"is_group":           result.IsGroup,
		"extraction_success": result.SenderPhone != "" && result.Message != "",
	}).Error("🚨 WAHA PRODUCTION: Final extraction results")

	// Log critical error if fields are still missing after all fallbacks
	if result.SenderPhone == "" || result.Message == "" {
		logrus.WithFields(logrus.Fields{
			"missing_sender_phone": result.SenderPhone == "",
			"missing_message":      result.Message == "",
			"all_payload_keys":     getMapKeys(webhookPayload),
			"payload_structure":    analyzePayloadDepth(webhookPayload),
		}).Error("🚨 WAHA PRODUCTION CRITICAL: All extraction methods failed - payload structure unknown")
	}

	// Console debug output for checking extracted data
	logrus.WithFields(logrus.Fields{
		"sender_phone": result.SenderPhone,
		"sender_name":  result.SenderName,
		"message":      result.Message,
		"is_from_me":   result.IsFromMe,
		"is_group":     result.IsGroup,
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
		ProspectNum:  req.ProspectNum,
		IDDevice:     req.IDDevice,
		ProspectName: sql.NullString{String: req.ProspectName, Valid: req.ProspectName != ""},
		Stage:        sql.NullString{String: req.Stage, Valid: req.Stage != ""},
		Human:        0, // AI active by default
		Niche:        req.Niche,
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
	response, err := h.AIWhatsappService.ProcessAIConversation(req.ProspectNum, req.IDDevice, req.Message, req.Stage, "User")
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
			"error":   "Invalid request format",
			"details": err.Error(),
		})
	}

	// Log the raw payload structure for debugging
	logrus.WithFields(logrus.Fields{
		"raw_payload":  rawPayload,
		"payload_keys": getMapKeys(rawPayload),
	}).Info("🧪 WAHA TEST: Raw payload received")

	// Extract standardized webhook data
	extractedData := h.extractWahaWebhookData(rawPayload)

	// Log the test extraction
	logrus.WithFields(logrus.Fields{
		"sender_phone": extractedData.SenderPhone,
		"sender_name":  extractedData.SenderName,
		"message":      truncateString(extractedData.Message, 100),
		"is_from_me":   extractedData.IsFromMe,
		"is_group":     extractedData.IsGroup,
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
// Updated to accept senderName parameter to properly save prospect_name
func (h *AIWhatsappHandlers) processIncomingMessage(prospectNum, message, deviceID, provider, senderName string) {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"device_id":    deviceID,
		"provider":     provider,
		"message":      message,
		"sender_name":  senderName,
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

	// Process AI conversation with actual sender name
	response, err := h.AIWhatsappService.ProcessAIConversation(prospectNum, deviceID, message, stage, senderName)
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
			"idDevice":  req.DeviceID,
			"startDate": req.StartDate,
			"endDate":   req.EndDate,
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
	userID, ok := c.Locals("user_id").(string)
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
		"totalConversations":      summary["total_conversations"],
		"aiActiveConversations":   summary["ai_active"],
		"humanTakeovers":          summary["human_takeover"],
		"uniqueDevices":           summary["unique_devices"],
		"uniqueNiches":            summary["unique_niches"],
		"conversationsWithStages": summary["conversations_with_stage"],
		"aiActivePercentage":      summary["ai_active_percentage"],
		"humanTakeoverPercentage": summary["human_takeover_percentage"],
		"dailyBreakdown":          analyticsData["daily_data"],
		"stageDistribution":       stageDistributionMap,
		"dateRange":               analyticsData["date_range"],
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

	// Support for both parameter names (device_id and user_device_ids)
	if deviceFilter == "" {
		deviceFilter = c.Query("user_device_ids", "")
	}

	// Support for date filtering (startDate/endDate)
	startDateStr := c.Query("startDate", "")
	endDateStr := c.Query("endDate", "")

	// Parse date parameters
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		} else {
			logrus.WithError(err).WithField("startDate", startDateStr).Warn("Invalid startDate format, ignoring")
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day for endDate
			endOfDay := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
			endDate = &endOfDay
		} else {
			logrus.WithError(err).WithField("endDate", endDateStr).Warn("Invalid endDate format, ignoring")
		}
	}

	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Get user ID from authentication context
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		logrus.Error("User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Authentication required",
		})
	}

	logrus.WithFields(logrus.Fields{
		"page":            page,
		"limit":           limit,
		"deviceFilter":    deviceFilter,
		"stageFilter":     stageFilter,
		"search":          search,
		"startDate":       startDateStr,
		"endDate":         endDateStr,
		"startDateParsed": startDate,
		"endDateParsed":   endDate,
		"userID":          userID,
	}).Info("GetAllAIWhatsappData called with parameters")

	// Get data from repository with user-specific filtering including date range
	data, total, err := h.AIRepo.GetAllAIWhatsappData(limit, offset, deviceFilter, stageFilter, search, userID, startDate, endDate)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp data")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve AI WhatsApp data",
			"error":   err.Error(),
		})
	}

	// Handle empty data gracefully
	if data == nil {
		data = []models.AIWhatsapp{}
	}

	// Transform data to handle sql.NullString fields properly
	transformedData := make([]map[string]interface{}, len(data))
	for i, item := range data {
		transformed := map[string]interface{}{
			"id_prospect":  item.IDProspect,
			"id_device":    item.IDDevice,
			"prospect_num": item.ProspectNum,
			"human":        item.Human,
			"niche":        item.Niche,
			"intro":        item.Intro,
			"created_at":   item.CreatedAt,
			"updated_at":   item.UpdatedAt,
		}

		// Handle nullable fields
		if item.ProspectName.Valid {
			transformed["prospect_name"] = item.ProspectName.String
		} else {
			transformed["prospect_name"] = nil
		}

		if item.Stage.Valid {
			transformed["stage"] = item.Stage.String
		} else {
			transformed["stage"] = nil
		}

		if item.Balas.Valid {
			transformed["balas"] = item.Balas.String
		} else {
			transformed["balas"] = nil
		}

		if item.KeywordIklan.Valid {
			transformed["keywordiklan"] = item.KeywordIklan.String
		} else {
			transformed["keywordiklan"] = nil
		}

		if item.Marketer.Valid {
			transformed["marketer"] = item.Marketer.String
		} else {
			transformed["marketer"] = nil
		}

		if item.DateOrder != nil {
			transformed["date_order"] = item.DateOrder
		} else {
			transformed["date_order"] = nil
		}

		if item.ConvCurrent.Valid {
			transformed["conv_current"] = item.ConvCurrent.String
		} else {
			transformed["conv_current"] = nil
		}

		if item.UpdateToday != nil {
			transformed["update_today"] = item.UpdateToday
		} else {
			transformed["update_today"] = nil
		}

		// Handle JSON fields
		if item.ConvLast.Valid && len(item.ConvLast.String) > 0 && item.ConvLast.String != "null" {
			transformed["conv_last"] = item.ConvLast.String
		} else {
			transformed["conv_last"] = nil
		}

		// Flow execution fields
		if item.FlowReference.Valid {
			transformed["flow_reference"] = item.FlowReference.String
		} else {
			transformed["flow_reference"] = nil
		}

		// No current_node field anymore - removed from schema

		if item.ExecutionStatus.Valid {
			transformed["execution_status"] = item.ExecutionStatus.String
		} else {
			transformed["execution_status"] = nil
		}

		if item.ExecutionID.Valid {
			transformed["execution_id"] = item.ExecutionID.String
		} else {
			transformed["execution_id"] = nil
		}

		if item.CurrentNodeID.Valid {
			transformed["current_node_id"] = item.CurrentNodeID.String
		} else {
			transformed["current_node_id"] = nil
		}

		if item.WaitingForReply.Valid {
			transformed["waiting_for_reply"] = item.WaitingForReply.Int32
		} else {
			transformed["waiting_for_reply"] = nil
		}

		if item.FlowID.Valid {
			transformed["flow_id"] = item.FlowID.String
		} else {
			transformed["flow_id"] = nil
		}

		if item.LastNodeID.Valid {
			transformed["last_node_id"] = item.LastNodeID.String
		} else {
			transformed["last_node_id"] = nil
		}

		transformedData[i] = transformed
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit

	// Return paginated response with transformed data
	return c.JSON(fiber.Map{
		"success": true,
		"data":    transformedData,
		"pagination": fiber.Map{
			"current_page":  page,
			"total_pages":   totalPages,
			"total_records": total,
			"limit":         limit,
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
	userIDStr, ok := c.Locals("user_id").(string)
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

	if !deviceSettings.UserID.Valid || deviceSettings.UserID.String != userIDStr {
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
		"user_id":     userIDStr,
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
				"type":      "object",
				"keys":      getMapKeys(v),
				"key_count": len(v),
			}
		case []interface{}:
			analysis[key] = map[string]interface{}{
				"type":   "array",
				"length": len(v),
			}
		case string:
			analysis[key] = map[string]interface{}{
				"type":   "string",
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

// getStringValue safely extracts string value from interface{}
func getStringValue(val interface{}) string {
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
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
		"device_id":    deviceID,
		"payload_size": len(body),
		"content_type": c.Get("Content-Type"),
		"user_agent":   c.Get("User-Agent"),
		"headers":      headers,
		"raw_payload":  string(body),
		"method":       c.Method(),
		"url":          c.OriginalURL(),
		"ip":           c.IP(),
	}).Error("🚨 WAHA DEBUG ENDPOINT: Complete webhook request details")

	// Parse as generic map for structure analysis
	var rawPayload map[string]interface{}
	if err := json.Unmarshal(body, &rawPayload); err != nil {
		logrus.WithError(err).Error("🚨 WAHA DEBUG: Failed to parse JSON")
		return c.Status(400).JSON(fiber.Map{
			"success":  false,
			"error":    "Invalid JSON format",
			"raw_body": string(body),
		})
	}

	// Perform complete payload analysis
	payloadAnalysis := analyzePayloadDepth(rawPayload)
	extractedData := h.extractWahaWebhookData(rawPayload)

	// Log detailed analysis
	logrus.WithFields(logrus.Fields{
		"payload_keys":       getMapKeys(rawPayload),
		"payload_analysis":   payloadAnalysis,
		"extracted_data":     extractedData,
		"extraction_success": extractedData.SenderPhone != "" && extractedData.Message != "",
	}).Error("🚨 WAHA DEBUG: Complete payload analysis")

	// Return comprehensive debug information
	return c.JSON(fiber.Map{
		"success": true,
		"debug_info": fiber.Map{
			"device_id":          deviceID,
			"payload_size":       len(body),
			"headers":            headers,
			"raw_payload":        rawPayload,
			"payload_keys":       getMapKeys(rawPayload),
			"payload_analysis":   payloadAnalysis,
			"extracted_data":     extractedData,
			"extraction_success": extractedData.SenderPhone != "" && extractedData.Message != "",
			"timestamp":          time.Now().Unix(),
		},
		"message": "Debug data logged successfully",
	})
}

// UpdateHumanStatus updates the human status for a conversation
func (h *AIWhatsappHandlers) UpdateHumanStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid prospect ID")
	}

	var req struct {
		Human int `json:"human"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate human value (should be 0 or 1)
	if req.Human != 0 && req.Human != 1 {
		return h.errorResponse(c, fiber.StatusBadRequest, "Human value must be 0 (AI) or 1 (Human)")
	}

	// Update human status in database
	err := h.AIRepo.UpdateHumanStatus(idStr, req.Human)
	if err != nil {
		logrus.WithError(err).Error("Failed to update human status")
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to update human status")
	}

	logrus.WithFields(logrus.Fields{
		"id_prospect": idStr,
		"human":       req.Human,
	}).Info("Human status updated successfully")

	return h.successResponse(c, map[string]interface{}{
		"id":     idStr,
		"human":  req.Human,
		"status": "updated",
	})
}
