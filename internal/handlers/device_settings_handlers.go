package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"

	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetDeviceSettings retrieves all device settings
func (h *Handlers) GetDeviceSettings(c *fiber.Ctx) error {
	settings, err := h.deviceSettingsService.GetAll()
	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings")
		return h.errorResponse(c, 500, "Failed to retrieve device settings")
	}

	return h.successResponse(c, settings)
}

// GetDeviceSettingsById retrieves a device setting by ID
func (h *Handlers) GetDeviceSettingsById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.errorResponse(c, 400, "Device setting ID is required")
	}

	setting, err := h.deviceSettingsService.GetByID(id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get device setting")
		if err.Error() == "device setting not found" {
			return h.errorResponse(c, 404, "Device setting not found")
		}
		return h.errorResponse(c, 500, "Failed to retrieve device setting")
	}

	return h.successResponse(c, setting)
}

// CreateDeviceSettings creates a new device setting
func (h *Handlers) CreateDeviceSettings(c *fiber.Ctx) error {
	var req models.CreateDeviceSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	// Validate required fields
	if req.IDDevice == "" {
		return h.errorResponse(c, 400, "ID Device is required")
	}
	if req.IDERP == "" {
		return h.errorResponse(c, 400, "ID ERP is required")
	}
	if req.IDAdmin == "" {
		return h.errorResponse(c, 400, "ID Admin is required")
	}
	// DeviceID is optional - it will be generated later if not provided

	setting, err := h.deviceSettingsService.Create(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create device setting")
		return h.errorResponse(c, 500, "Failed to create device setting")
	}

	return h.successMessageResponse(c, "Device setting created successfully", setting)
}

// UpdateDeviceSettings updates an existing device setting
func (h *Handlers) UpdateDeviceSettings(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.errorResponse(c, 400, "Device setting ID is required")
	}

	var req models.UpdateDeviceSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	setting, err := h.deviceSettingsService.Update(id, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to update device setting")
		if err.Error() == "device setting not found" {
			return h.errorResponse(c, 404, "Device setting not found")
		}
		return h.errorResponse(c, 500, "Failed to update device setting")
	}

	return h.successMessageResponse(c, "Device setting updated successfully", setting)
}

// DeleteDeviceSettings deletes a device setting
func (h *Handlers) DeleteDeviceSettings(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.errorResponse(c, 400, "Device setting ID is required")
	}

	err := h.deviceSettingsService.Delete(id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete device setting")
		if err.Error() == "device setting not found" {
			return h.errorResponse(c, 404, "Device setting not found")
		}
		return h.errorResponse(c, 500, "Failed to delete device setting")
	}

	return h.successMessageResponse(c, "Device setting deleted successfully", nil)
}

// GetDeviceIDs retrieves all device IDs for dropdown selection
func (h *Handlers) GetDeviceIDs(c *fiber.Ctx) error {
	settings, err := h.deviceSettingsService.GetAll()
	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings")
		return h.errorResponse(c, 500, "Failed to retrieve device settings")
	}

	// Extract device IDs and create dropdown options
	type DeviceOption struct {
		Value string `json:"value"`
		Label string `json:"label"`
	}

	var options []DeviceOption
	for _, setting := range settings {
		if setting.IDDevice.Valid && setting.IDDevice.String != "" {
			label := setting.IDDevice.String
			if setting.Provider != "" {
				label += " (" + setting.Provider + ")"
			}
			options = append(options, DeviceOption{
				Value: setting.IDDevice.String,
				Label: label,
			})
		}
	}

	return h.successResponse(c, options)
}

// GenerateWhacenterDevice generates a device using Whacenter API
func (h *Handlers) GenerateWhacenterDevice(c *fiber.Ctx) error {
	var req struct {
		models.CreateDeviceSettingsRequest
		WebhookURL string `json:"webhook_url"`
		DeviceData struct {
			DeviceName string `json:"device_name"`
			WebhookURL string `json:"webhook_url"`
		} `json:"device_data"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	// Validate required fields
	if req.PhoneNumber == "" {
		return h.errorResponse(c, 400, "Phone number is required")
	}
	if req.IDDevice == "" {
		return h.errorResponse(c, 400, "ID Device is required")
	}

	// Check existing device settings by IDDevice to get instance value
	existingDevice, err := h.deviceSettingsService.GetByIDDevice(req.IDDevice)
	var whacenterAPIKey string
	
	if err != nil {
		// No existing device found, create new with hardcoded API key
		logrus.WithFields(logrus.Fields{
			"id_device": req.IDDevice,
			"action": "create_new",
		}).Info("🆕 WHACENTER: No existing device found, creating new device")
		whacenterAPIKey = "abebe840-156c-441c-8252-da0342c5a07c" // Hardcoded API key for new devices
	} else {
		// Existing device found, check instance column
		if !existingDevice.Instance.Valid || existingDevice.Instance.String == "" {
			// Instance is null, create new device with hardcoded API key
			logrus.WithFields(logrus.Fields{
				"id_device": req.IDDevice,
				"action": "create_new_null_instance",
			}).Info("🆕 WHACENTER: Instance is null, creating new device")
			whacenterAPIKey = "abebe840-156c-441c-8252-da0342c5a07c" // Hardcoded API key for new devices
		} else {
			// Instance is not null, delete existing device data using instance value
			logrus.WithFields(logrus.Fields{
				"id_device": req.IDDevice,
				"instance": existingDevice.Instance.String,
				"action": "delete_existing",
			}).Info("🗑️ WHACENTER: Instance found, deleting existing device data")
			
			// Delete existing device using instance value as device_id
			deleteURL := fmt.Sprintf("https://api.whacenter.com/api/deleteDevice?api_key=%s&device_id=%s", 
				"abebe840-156c-441c-8252-da0342c5a07c", existingDevice.Instance.String)
			
			deleteClient := &http.Client{Timeout: 30 * time.Second}
			deleteReq, err := http.NewRequest("GET", deleteURL, nil)
			if err != nil {
				logrus.WithError(err).Warn("Failed to create delete request")
			} else {
				deleteReq.Header.Set("Accept", "application/json")
				deleteReq.Header.Set("Content-Type", "application/json")
				
				deleteResp, err := deleteClient.Do(deleteReq)
				if err != nil {
					logrus.WithError(err).Warn("Failed to delete existing device")
				} else {
					defer deleteResp.Body.Close()
					logrus.WithFields(logrus.Fields{
						"status": deleteResp.StatusCode,
						"device_id": existingDevice.Instance.String,
					}).Info("📥 WHACENTER: Device deletion attempted")
				}
			}
			
			// Now create new device with hardcoded API key
			whacenterAPIKey = "abebe840-156c-441c-8252-da0342c5a07c"
		}
	}

	// Prepare Whacenter API request with GET parameters (without webhook initially)
	whacenterURL := fmt.Sprintf("https://api.whacenter.com/api/addDevice?api_key=%s&name=%s&number=%s", 
		whacenterAPIKey, req.IDDevice, req.PhoneNumber)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create GET request with proper headers
	request, err := http.NewRequest("GET", whacenterURL, nil)
	if err != nil {
		return h.errorResponse(c, 500, "Failed to create request")
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	// Make request to Whacenter API
	logrus.WithFields(logrus.Fields{
		"provider": "whacenter",
		"url": whacenterURL,
		"device_name": req.IDDevice,
		"phone_number": req.PhoneNumber,
		"webhook_url": req.WebhookURL,
		"api_key_length": len(req.APIKey),
	}).Info("🔵 WHACENTER: Making external API request")
	
	// Log request headers (without sensitive data)
	logrus.WithFields(logrus.Fields{
		"content_type": request.Header.Get("Content-Type"),
		"has_auth_header": request.Header.Get("Authorization") != "",
		"request_method": "GET",
	}).Info("🔵 WHACENTER: Request details")
	
	resp, err := client.Do(request)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider": "whacenter",
			"url": whacenterURL,
			"error": err.Error(),
		}).Error("❌ WHACENTER: Failed to call external API")
		return h.errorResponse(c, 500, fmt.Sprintf("Failed to communicate with Whacenter API: %v", err))
	}
	defer resp.Body.Close()
	
	logrus.WithFields(logrus.Fields{
		"provider": "whacenter",
		"status_code": resp.StatusCode,
		"status": resp.Status,
		"content_type": resp.Header.Get("Content-Type"),
		"content_length": resp.Header.Get("Content-Length"),
	}).Info("📥 WHACENTER: Received response from external API")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider": "whacenter",
			"error": err.Error(),
		}).Error("❌ WHACENTER: Failed to read response body")
		return h.errorResponse(c, 500, "Failed to read API response")
	}

	logrus.WithFields(logrus.Fields{
		"provider": "whacenter",
		"response_body": string(body),
		"response_length": len(body),
	}).Info("📄 WHACENTER: API response body received")

	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		logrus.WithFields(logrus.Fields{
			"provider": "whacenter",
			"error": err.Error(),
			"response_body": string(body),
		}).Error("❌ WHACENTER: Failed to unmarshal response JSON")
		return h.errorResponse(c, 500, "Failed to parse API response")
	}

	// Check if API call was successful
	if success, ok := apiResponse["success"].(bool); !ok || !success {
		message := "Unknown error"
		if msg, exists := apiResponse["message"].(string); exists {
			message = msg
		}
		return h.errorResponse(c, 500, fmt.Sprintf("Whacenter API error: %s", message))
	}

	// Extract device information from response
	data, ok := apiResponse["data"].(map[string]interface{})
	if !ok {
		return h.errorResponse(c, 500, "Invalid API response format")
	}

	// Extract device information from nested device object
	device, ok := data["device"].(map[string]interface{})
	if !ok {
		return h.errorResponse(c, 500, "Invalid device data format")
	}

	deviceID, _ := device["device_id"].(string)
	apiKey, _ := device["device_key"].(string)
	
	// If device_key is empty, use the whacenterAPIKey as fallback
	if apiKey == "" {
		apiKey = whacenterAPIKey
	}

	// Construct production webhook URL using the actual device_id from API response
	productionWebhookURL := fmt.Sprintf("https://nodepath-chat-production.up.railway.app/api/webhook/%s/%s", req.IDDevice, deviceID)
	
	// Set webhook for the created device
	setWebhookURL := fmt.Sprintf("https://api.whacenter.com/api/setWebhook?device_id=%s&webhook=%s", 
		deviceID, url.QueryEscape(productionWebhookURL))
	
	logrus.WithFields(logrus.Fields{
		"provider": "whacenter",
		"device_id": deviceID,
		"webhook_url": productionWebhookURL,
		"set_webhook_url": setWebhookURL,
	}).Info("🔗 WHACENTER: Setting webhook for device")
	
	// Create webhook request
	webhookRequest, err := http.NewRequest("GET", setWebhookURL, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to create webhook request")
	} else {
		webhookRequest.Header.Set("Accept", "application/json")
		
		// Execute webhook request
		webhookResp, err := client.Do(webhookRequest)
		if err != nil {
			logrus.WithError(err).Error("Failed to set webhook")
		} else {
			defer webhookResp.Body.Close()
			webhookBody, _ := io.ReadAll(webhookResp.Body)
			
			logrus.WithFields(logrus.Fields{
				"status_code": webhookResp.StatusCode,
				"response": string(webhookBody),
			}).Info("📥 WHACENTER: Webhook set response")
		}
	}

	// Save device data to database - Whacenter mapping: webhook_id stores webhook_url, instance stores device_id, device_id should be null
	createReq := &models.CreateDeviceSettingsRequest{
		// DeviceID is intentionally left empty (null) for Whacenter devices
		APIKeyOption: req.APIKeyOption,
		WebhookID:    productionWebhookURL, // Store webhook URL
		Provider:     "whacenter",
		PhoneNumber:  req.PhoneNumber,
		APIKey:       req.APIKey, // Preserve the original OpenRouter API key
		IDDevice:     req.IDDevice,
		IDERP:        req.IDERP,
		IDAdmin:      req.IDAdmin,
		Instance:     deviceID, // Store device_id as instance for Whacenter
	}

	// Debug logging for database save
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"webhook_id": productionWebhookURL,
		"instance": deviceID,
		"provider": "whacenter",
		"phone_number": req.PhoneNumber,
	}).Info("💾 WHACENTER: Saving device data to database")

	// Upsert device setting in database (update if exists, create if not)
	deviceSetting, err := h.deviceSettingsService.Upsert(createReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to save device setting to database")
		// Continue with success response even if database save fails
	} else {
		logrus.WithField("device_setting_id", deviceSetting.ID).Info("Device setting saved to database")
	}

	// Log successful device generation
	logrus.WithFields(logrus.Fields{
		"provider": "whacenter",
		"device_id": deviceID,
		"webhook_url": req.WebhookURL,
		"phone_number": req.PhoneNumber,
	}).Info("✅ WHACENTER: Device generated successfully")
	
	// Return success response
	return h.successResponse(c, map[string]interface{}{
		"success": true,
		"message": "Device generated successfully via Whacenter",
		"data": map[string]interface{}{
			"device_id":   deviceID,
			"webhook_url": productionWebhookURL,
			"api_key":     apiKey,
			"provider":    "whacenter",
		},
	})
}

// HandleWebhook processes incoming webhook requests from WhatsApp providers
func (h *Handlers) HandleWebhook(c *fiber.Ctx) error {
	idDevice := c.Params("id_device")
	instance := c.Params("instance")
	
	if idDevice == "" {
		return h.errorResponse(c, 400, "ID Device is required")
	}
	if instance == "" {
		return h.errorResponse(c, 400, "Instance is required")
	}
	
	// Get the raw webhook payload
	body := c.Body()
	
	logrus.WithFields(logrus.Fields{
		"id_device": idDevice,
		"instance": instance,
		"content_type": c.Get("Content-Type"),
		"user_agent": c.Get("User-Agent"),
		"payload_size": len(body),
	}).Info("📨 WEBHOOK: Received webhook request")
	
	// Verify the device exists in our database
	deviceSettings, err := h.deviceSettingsService.GetByIDDevice(idDevice)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id_device": idDevice,
			"error": err.Error(),
		}).Warn("⚠️ WEBHOOK: Device not found in database")
		return h.errorResponse(c, 404, "Device not found")
	}
	
	// Verify the instance matches
	if !deviceSettings.Instance.Valid || deviceSettings.Instance.String != instance {
		logrus.WithFields(logrus.Fields{
			"id_device": idDevice,
			"expected_instance": deviceSettings.Instance.String,
			"received_instance": instance,
		}).Warn("⚠️ WEBHOOK: Instance mismatch")
		return h.errorResponse(c, 401, "Invalid instance")
	}
	
	// Parse the webhook payload based on provider
	var webhookData map[string]interface{}
	if err := json.Unmarshal(body, &webhookData); err != nil {
		logrus.WithFields(logrus.Fields{
			"id_device": idDevice,
			"error": err.Error(),
			"payload": string(body),
		}).Error("❌ WEBHOOK: Failed to parse JSON payload")
		return h.errorResponse(c, 400, "Invalid JSON payload")
	}
	
	logrus.WithFields(logrus.Fields{
		"id_device": idDevice,
		"provider": deviceSettings.Provider,
		"instance": instance,
		"webhook_data": webhookData,
	}).Info("✅ WEBHOOK: Successfully processed webhook")
	
	// Process the webhook data based on provider type and integrate with AI WhatsApp
	go h.processWebhookMessage(webhookData, idDevice, deviceSettings.Provider)
	
	return h.successResponse(c, map[string]interface{}{
		"success": true,
		"message": "Webhook received and processed",
		"id_device": idDevice,
		"provider": deviceSettings.Provider,
	})
}

// GenerateWablasDevice generates a device using Wablas API
func (h *Handlers) GenerateWablasDevice(c *fiber.Ctx) error {
	var req struct {
		models.CreateDeviceSettingsRequest
		WebhookURL string `json:"webhook_url"`
		DeviceData struct {
			DeviceName string `json:"device_name"`
			WebhookURL string `json:"webhook_url"`
		} `json:"device_data"`
	}

	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	// Validate required fields
	if req.PhoneNumber == "" {
		return h.errorResponse(c, 400, "Phone number is required")
	}
	if req.IDDevice == "" {
		return h.errorResponse(c, 400, "ID Device is required")
	}

	// Check existing device settings by IDDevice to get instance value
	existingDevice, err := h.deviceSettingsService.GetByIDDevice(req.IDDevice)
	var wablasToken string
	
	if err != nil {
		// No existing device found, create new with hardcoded token
		logrus.WithFields(logrus.Fields{
			"id_device": req.IDDevice,
			"action": "create_new",
		}).Info("🆕 WABLAS: No existing device found, creating new device")
		wablasToken = "j0oB1aibqYDQlgyk9SIqLyfeGgRJjjmOUFMVqxGd8Irk6JCwl1ZxYtY.7hDkbW0f" // Hardcoded token for new devices
	} else {
		// Existing device found, check instance column
		if !existingDevice.Instance.Valid || existingDevice.Instance.String == "" {
			// Instance is null, create new device with hardcoded token
			logrus.WithFields(logrus.Fields{
				"id_device": req.IDDevice,
				"action": "create_new_null_instance",
			}).Info("🆕 WABLAS: Instance is null, creating new device")
			wablasToken = "j0oB1aibqYDQlgyk9SIqLyfeGgRJjjmOUFMVqxGd8Irk6JCwl1ZxYtY.7hDkbW0f" // Hardcoded token for new devices
		} else {
			// Instance is not null, delete existing device data using instance value
			logrus.WithFields(logrus.Fields{
				"id_device": req.IDDevice,
				"instance": existingDevice.Instance.String,
				"action": "delete_existing",
			}).Info("🗑️ WABLAS: Instance found, deleting existing device data")
			
			// Delete existing device using instance value as authorization
			deleteURL := "https://my.wablas.com/api/device/delete"
			
			// Create HTTP client for delete request
			deleteClient := &http.Client{
				Timeout: 30 * time.Second,
			}
			
			// Create delete request
			deleteRequest, err := http.NewRequest("DELETE", deleteURL, nil)
			if err != nil {
				logrus.WithError(err).Error("Failed to create delete request")
			} else {
				// Set headers for delete request using instance value
				deleteRequest.Header.Set("Authorization", existingDevice.Instance.String)
				deleteRequest.Header.Set("Accept", "application/json")
				
				// Execute delete request
				deleteResp, err := deleteClient.Do(deleteRequest)
				if err != nil {
					logrus.WithError(err).Error("Failed to delete existing Wablas device")
				} else {
					defer deleteResp.Body.Close()
					logrus.WithFields(logrus.Fields{
						"status_code": deleteResp.StatusCode,
						"auth_token": existingDevice.Instance.String,
					}).Info("📥 WABLAS: Device deletion attempted")
				}
			}
			
			// Now create new device with hardcoded token
			wablasToken = "j0oB1aibqYDQlgyk9SIqLyfeGgRJjjmOUFMVqxGd8Irk6JCwl1ZxYtY.7hDkbW0f"
		}
	}

	// Prepare Wablas API request for device creation
	wablasURL := "https://my.wablas.com/api/device/create"
	
	// Prepare form data
	formData := url.Values{}
	formData.Set("name", req.IDDevice)
	formData.Set("phone", req.PhoneNumber)
	formData.Set("bank", "BCA")
	formData.Set("periode", "monthly")
	formData.Set("product", "large")
	
	formDataEncoded := formData.Encode()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request with proper headers
	request, err := http.NewRequest("POST", wablasURL, strings.NewReader(formDataEncoded))
	if err != nil {
		return h.errorResponse(c, 500, "Failed to create request")
	}

	// Use the determined Wablas token for device creation
	authHeader := wablasToken
	
	request.Header.Set("Authorization", authHeader)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make request to Wablas API
	logrus.WithFields(logrus.Fields{
		"provider": "wablas",
		"url": wablasURL,
		"device_name": req.IDDevice,
		"phone_number": req.PhoneNumber,
		"api_key_length": len(req.APIKey),
	}).Info("🟡 WABLAS: Making external API request")
	
	// Log request headers (without sensitive data)
	logrus.WithFields(logrus.Fields{
		"content_type": request.Header.Get("Content-Type"),
		"has_auth_header": request.Header.Get("Authorization") != "",
		"request_body": formDataEncoded,
	}).Info("🟡 WABLAS: Request details")
	
	resp, err := client.Do(request)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider": "wablas",
			"url": wablasURL,
			"error": err.Error(),
		}).Error("❌ WABLAS: Failed to call external API")
		return h.errorResponse(c, 500, fmt.Sprintf("Failed to communicate with Wablas API: %v", err))
	}
	defer resp.Body.Close()
	
	logrus.WithFields(logrus.Fields{
		"provider": "wablas",
		"status_code": resp.StatusCode,
		"status": resp.Status,
		"content_type": resp.Header.Get("Content-Type"),
		"content_length": resp.Header.Get("Content-Length"),
	}).Info("📥 WABLAS: Received response from external API")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider": "wablas",
			"error": err.Error(),
		}).Error("❌ WABLAS: Failed to read response body")
		return h.errorResponse(c, 500, "Failed to read API response")
	}

	logrus.WithFields(logrus.Fields{
		"provider": "wablas",
		"response_body": string(body),
		"response_length": len(body),
	}).Info("📄 WABLAS: API response body received")

	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		logrus.WithFields(logrus.Fields{
			"provider": "wablas",
			"error": err.Error(),
			"response_body": string(body),
		}).Error("❌ WABLAS: Failed to unmarshal response JSON")
		return h.errorResponse(c, 500, "Failed to parse API response")
	}

	// Check if API call was successful
	if status, ok := apiResponse["status"].(bool); !ok || !status {
		message := "Unknown error"
		if msg, exists := apiResponse["message"].(string); exists {
			message = msg
		}
		return h.errorResponse(c, 500, fmt.Sprintf("Wablas API error: %s", message))
	}

	// Extract device information from response
	data, ok := apiResponse["data"].(map[string]interface{})
	if !ok {
		return h.errorResponse(c, 500, "Invalid API response format")
	}

	deviceID, _ := data["device"].(string)
	deviceToken, _ := data["token"].(string)
	deviceSecret, _ := data["secret_key"].(string)

	// Create new auth header with device token and secret
	newAuthHeader := fmt.Sprintf("%s.%s", deviceToken, deviceSecret)

	// Use production webhook URL
	productionWebhookURL := fmt.Sprintf("https://nodepath-chat-production.up.railway.app/api/webhook/%s/%s", req.IDDevice, newAuthHeader)

	// Setup webhook configuration using the correct endpoint
	webhookFormData := url.Values{}
	webhookFormData.Set("webhook_url", productionWebhookURL)
	
	webhookFormEncoded := webhookFormData.Encode()

	// Setup webhook
	webhookRequest, err := http.NewRequest("POST", "https://my.wablas.com/api/device/change-webhook-url", strings.NewReader(webhookFormEncoded))
	if err != nil {
		logrus.WithError(err).Error("Failed to create webhook request")
	} else {
		webhookRequest.Header.Set("Authorization", newAuthHeader)
		webhookRequest.Header.Set("Accept", "application/json")
		webhookRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		webhookResp, err := client.Do(webhookRequest)
		if err != nil {
			logrus.WithError(err).Error("Failed to setup webhook")
		} else {
			defer webhookResp.Body.Close()
			
			// Read webhook response
			webhookBody, err := io.ReadAll(webhookResp.Body)
			if err != nil {
				logrus.WithError(err).Error("Failed to read webhook response")
			} else {
				var webhookResponse map[string]interface{}
				if err := json.Unmarshal(webhookBody, &webhookResponse); err == nil {
					if status, ok := webhookResponse["status"].(bool); ok && status {
						logrus.Info("Webhook configured successfully")
					} else {
						logrus.WithField("response", string(webhookBody)).Warn("Webhook setup may have failed")
					}
				}
			}
		}
	}

	// Save device data to database - Wablas mapping: device_id stores device_id, webhook_id stores webhook_url, instance stores api_key
	createReq := &models.CreateDeviceSettingsRequest{
		DeviceID:     deviceID, // Store device_id
		APIKeyOption: req.APIKeyOption,
		WebhookID:    productionWebhookURL, // Store webhook URL
		Provider:     "wablas",
		PhoneNumber:  req.PhoneNumber,
		APIKey:       req.APIKey, // Preserve the original OpenRouter API key
		IDDevice:     req.IDDevice,
		IDERP:        req.IDERP,
		IDAdmin:      req.IDAdmin,
		Instance:     newAuthHeader, // Store API key as instance for Wablas
	}

	// Debug logging for database save
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"webhook_id": productionWebhookURL,
		"instance": newAuthHeader,
		"provider": "wablas",
		"phone_number": req.PhoneNumber,
	}).Info("💾 WABLAS: Saving device data to database")

	// Upsert device setting in database (update if exists, create if not)
	deviceSetting, err := h.deviceSettingsService.Upsert(createReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to save device setting to database")
		// Continue with success response even if database save fails
	} else {
		logrus.WithField("device_setting_id", deviceSetting.ID).Info("Device setting saved to database")
	}

	// Log successful device generation
	logrus.WithFields(logrus.Fields{
		"provider": "wablas",
		"device_id": deviceID,
		"webhook_url": productionWebhookURL,
		"phone_number": req.PhoneNumber,
	}).Info("✅ WABLAS: Device generated successfully")
	
	// Return success response
	return h.successResponse(c, map[string]interface{}{
		"success": true,
		"message": "Device generated successfully via Wablas",
		"data": map[string]interface{}{
			"device_id":   deviceID,
			"webhook_url": productionWebhookURL,
			"api_key":     newAuthHeader,
			"provider":    "wablas",
		},
	})
}

// GetDeviceStatus checks the connection status of a device
func (h *Handlers) GetDeviceStatus(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	logrus.WithField("device_id", deviceID).Info("[STATUS] Starting device status check")
	
	if deviceID == "" {
		logrus.Error("[STATUS] Device ID is empty")
		return h.errorResponse(c, 400, "Device ID is required")
	}

	// Get device settings
	device, err := h.deviceSettingsService.GetByID(deviceID)
	if err != nil {
		logrus.WithError(err).WithField("device_id", deviceID).Error("[STATUS] Failed to get device settings")
		return h.errorResponse(c, 404, "Device not found")
	}

	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"provider":  device.Provider,
		"instance":  device.Instance.String,
	}).Info("[STATUS] Device found, checking status")

	// Initialize status response
	status := map[string]interface{}{
		"device_id":    deviceID,
		"provider":     device.Provider,
		"connected":    false,
		"status":       "disconnected",
		"last_checked": time.Now(),
		"details":      map[string]interface{}{},
	}

	// Check status based on provider
	switch device.Provider {
	case "whacenter":
		logrus.Info("[STATUS] Checking Whacenter status")
		status = h.checkWhacenterStatus(device, status)
	case "wablas":
		logrus.Info("[STATUS] Checking Wablas status")
		status = h.checkWablasStatus(device, status)
	default:
		logrus.WithField("provider", device.Provider).Warn("[STATUS] Unsupported provider")
		status["status"] = "unsupported_provider"
		status["details"] = map[string]interface{}{
			"error": "Provider not supported for status checking",
		}
	}

	logrus.WithField("final_status", status).Info("[STATUS] Returning final status")
	return h.successResponse(c, status)
}

// checkWhacenterStatus checks the status of a Whacenter device
func (h *Handlers) checkWhacenterStatus(device *models.DeviceSettings, status map[string]interface{}) map[string]interface{} {
	logrus.WithFields(logrus.Fields{
		"device_id": device.ID,
		"instance_valid": device.Instance.Valid,
		"instance_value": device.Instance.String,
	}).Info("[WHACENTER] Starting Whacenter status check")
	
	if !device.Instance.Valid || device.Instance.String == "" {
		logrus.Error("[WHACENTER] Device instance not configured")
		status["status"] = "not_configured"
		status["details"] = map[string]interface{}{
			"error": "Device instance not configured",
		}
		return status
	}

	// Make API call to check Whacenter device status using the correct endpoint
	client := &http.Client{Timeout: 10 * time.Second}
	// Use the hardcoded API key for whacenter requests
	whacenterAPIKey := "abebe840-156c-441c-8252-da0342c5a07c"
	// Use the correct statusDevice API endpoint with device_id and api_key parameters
	apiURL := fmt.Sprintf("https://api.whacenter.com/api/statusDevice?api_key=%s&device_id=%s", 
		whacenterAPIKey, url.QueryEscape(device.Instance.String))
	
	logrus.WithFields(logrus.Fields{
		"api_url": apiURL,
	}).Info("[WHACENTER] Making API request")

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logrus.WithError(err).Error("[WHACENTER] Failed to create HTTP request")
		status["status"] = "error"
		status["details"] = map[string]interface{}{
			"error": "Failed to create status request",
			"details": err.Error(),
		}
		return status
	}

	// No authorization header needed for statusDevice endpoint
	req.Header.Set("Accept", "application/json")

	logrus.WithFields(logrus.Fields{
		"headers": req.Header,
	}).Info("[WHACENTER] Request headers set")

	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("[WHACENTER] HTTP request failed")
		status["status"] = "connection_error"
		status["details"] = map[string]interface{}{
			"error": "Failed to connect to Whacenter API",
			"details": err.Error(),
		}
		return status
	}
	defer resp.Body.Close()

	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"headers": resp.Header,
	}).Info("[WHACENTER] Received API response")

	// Read response body for logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("[WHACENTER] Failed to read response body")
		status["status"] = "error"
		status["details"] = map[string]interface{}{
			"error": "Failed to read API response",
			"details": err.Error(),
		}
		return status
	}
	
	logrus.WithFields(logrus.Fields{
		"response_body": string(bodyBytes),
		"body_length": len(bodyBytes),
	}).Info("[WHACENTER] Response body received")

	if resp.StatusCode == 200 {
		var apiResponse map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &apiResponse); err == nil {
			logrus.WithField("parsed_response", apiResponse).Info("[WHACENTER] Successfully parsed JSON response")
			
			// Parse the response according to statusDevice API format
			if data, ok := apiResponse["data"].(map[string]interface{}); ok {
				if deviceStatus, ok := data["status"].(string); ok {
					logrus.WithField("device_status", deviceStatus).Info("[WHACENTER] Found device status")
					
					if deviceStatus == "NOT CONNECTED" {
						status["connected"] = false
						status["status"] = "disconnected"
						
						// Fetch QR code when device is not connected
						qrCode := h.getWhacenterQRCode(device.Instance.String)
						if qrCode != "" {
							status["qr_code"] = qrCode
						}
					} else {
						status["connected"] = true
						status["status"] = "connected"
					}
					status["device_status"] = deviceStatus
				} else {
					logrus.Warn("[WHACENTER] No 'status' field found in data")
					status["status"] = "unknown"
				}
				status["details"] = data
			} else {
				logrus.Warn("[WHACENTER] No 'data' field found in response")
				status["status"] = "invalid_response"
				status["details"] = apiResponse
			}
		} else {
			logrus.WithError(err).Error("[WHACENTER] Failed to parse JSON response")
			status["status"] = "parse_error"
			status["details"] = map[string]interface{}{
				"error": "Failed to parse API response",
				"raw_response": string(bodyBytes),
				"parse_error": err.Error(),
			}
		}
	} else if resp.StatusCode == 404 {
		// Handle 404 specifically - device not found in Whacenter
		logrus.WithFields(logrus.Fields{
			"device_instance": device.Instance.String,
			"api_url": apiURL,
		}).Warn("[WHACENTER] Device not found in Whacenter system")
		
		status["status"] = "device_not_found"
		status["connected"] = false
		status["details"] = map[string]interface{}{
			"error": "Device not found in Whacenter system",
			"message": "The device may have been deleted from Whacenter or the device ID is incorrect",
			"device_instance": device.Instance.String,
			"http_status": 404,
			"response_body": string(bodyBytes),
			"suggestion": "Please regenerate the device or check if it exists in your Whacenter dashboard",
		}
	} else {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response_body": string(bodyBytes),
		}).Error("[WHACENTER] API returned non-200 status")
		
		status["status"] = "api_error"
		status["details"] = map[string]interface{}{
			"http_status": resp.StatusCode,
			"error":       "API returned error status",
			"response_body": string(bodyBytes),
		}
	}

	logrus.WithField("final_status", status).Info("[WHACENTER] Returning status")
	return status
}

// getWhacenterQRCode fetches QR code for Whacenter device when not connected
func (h *Handlers) getWhacenterQRCode(deviceID string) string {
	logrus.WithField("device_id", deviceID).Info("[WHACENTER] Fetching QR code")
	
	client := &http.Client{Timeout: 10 * time.Second}
	// Use the hardcoded API key for whacenter requests
	whacenterAPIKey := "abebe840-156c-441c-8252-da0342c5a07c"
	qrURL := fmt.Sprintf("https://api.whacenter.com/api/qr?api_key=%s&device_id=%s", 
		whacenterAPIKey, url.QueryEscape(deviceID))
	
	req, err := http.NewRequest("GET", qrURL, nil)
	if err != nil {
		logrus.WithError(err).Error("[WHACENTER] Failed to create QR request")
		return ""
	}
	
	// Accept both JSON and image responses
	req.Header.Set("Accept", "application/json, image/png")
	
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("[WHACENTER] QR request failed")
		return ""
	}
	defer resp.Body.Close()
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("[WHACENTER] Failed to read QR response")
		return ""
	}
	
	if resp.StatusCode == 200 {
		// Check if response is a PNG image (like in the PHP code)
		// PNG signature: \x89PNG\r\n\x1a\n (first 8 bytes)
		logrus.WithFields(logrus.Fields{
			"response_length": len(bodyBytes),
			"first_8_bytes": fmt.Sprintf("%x", bodyBytes[:min(8, len(bodyBytes))]),
			"content_type": resp.Header.Get("Content-Type"),
		}).Info("[WHACENTER] Analyzing QR response format")
		
		if len(bodyBytes) >= 8 {
			// Check for PNG signature: first byte is 0x89, followed by "PNG"
			if bodyBytes[0] == 0x89 && string(bodyBytes[1:4]) == "PNG" {
				// It's a valid PNG image, convert to base64 data URL
				logrus.Info("[WHACENTER] Successfully fetched QR code as PNG image")
				return fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(bodyBytes))
			}
		}
		
		// If not PNG, try to parse as JSON response
		logrus.Info("[WHACENTER] Response is not PNG format, trying JSON parsing")
		var qrResponse map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &qrResponse); err == nil {
			logrus.WithField("json_response", qrResponse).Info("[WHACENTER] Successfully parsed JSON response")
			if data, ok := qrResponse["data"].(map[string]interface{}); ok {
				if qrCode, ok := data["qr"].(string); ok {
					logrus.Info("[WHACENTER] Successfully fetched QR code from JSON")
					return qrCode
				}
				logrus.Warn("[WHACENTER] No 'qr' field found in JSON data")
			} else {
				logrus.Warn("[WHACENTER] No 'data' field found in JSON response")
			}
		} else {
			logrus.WithError(err).Warn("[WHACENTER] Failed to parse response as JSON")
			// Log raw response for debugging
			logrus.WithField("raw_response", string(bodyBytes[:min(200, len(bodyBytes))])).Warn("[WHACENTER] Raw response preview")
		}
	}
	
	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response_length": len(bodyBytes),
		"content_type": resp.Header.Get("Content-Type"),
	}).Warn("[WHACENTER] Failed to fetch QR code")
	
	return ""
}

// checkWablasStatus checks the status of a Wablas device
func (h *Handlers) checkWablasStatus(device *models.DeviceSettings, status map[string]interface{}) map[string]interface{} {
	logrus.WithFields(logrus.Fields{
		"device_id": device.ID,
		"instance_valid": device.Instance.Valid,
		"instance_value": device.Instance.String,
	}).Info("[WABLAS] Starting Wablas status check")
	
	// Check if instance (API key) is configured
	if !device.Instance.Valid || device.Instance.String == "" {
		logrus.Error("[WABLAS] Device instance not configured")
		status["status"] = "NOT CONNECTED"
		status["qr"] = "timeout"
		status["details"] = map[string]interface{}{
			"error": "Device instance not configured",
		}
		return status
	}

	// Extract token from instance - following PHP pattern: $token = explode('.', $auth_header)[0];
	authHeader := device.Instance.String
	var token string
	if strings.Contains(authHeader, ".") {
		parts := strings.Split(authHeader, ".")
		token = parts[0]
	} else {
		token = authHeader // Use full string if no dot found
	}

	// **STEP 1: CHECK DEVICE STATUS** - following PHP pattern
	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := fmt.Sprintf("https://my.wablas.com/api/device/info?token=%s", url.QueryEscape(token))
	
	// Log API request (without sensitive token details)
	logrus.WithFields(logrus.Fields{
		"api_url": "https://my.wablas.com/api/device/info",
		"token_prefix": token[:min(8, len(token))] + "...",
	}).Info("[WABLAS] Making API request")
	
	logrus.WithFields(logrus.Fields{
		"api_url": apiURL,
		"token_prefix": token[:min(8, len(token))] + "...",
	}).Info("[WABLAS] Making API request")

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logrus.WithError(err).Error("[WABLAS] Failed to create HTTP request")
		status["status"] = "NOT CONNECTED"
		status["qr"] = "timeout"
		status["details"] = map[string]interface{}{
			"error": "Failed to create status request",
			"details": err.Error(),
		}
		return status
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("[WABLAS] HTTP request failed")
		status["status"] = "NOT CONNECTED"
		status["qr"] = "timeout"
		status["details"] = map[string]interface{}{
			"error": "Failed to connect to Wablas API",
			"details": err.Error(),
		}
		return status
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("[WABLAS] Failed to read response body")
		status["status"] = "NOT CONNECTED"
		status["qr"] = "timeout"
		status["details"] = map[string]interface{}{
			"error": "Failed to read API response",
			"details": err.Error(),
		}
		return status
	}
	
	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response_body": string(bodyBytes),
	}).Info("[WABLAS] Received API response")

	if resp.StatusCode != 200 {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response_body": string(bodyBytes),
		}).Error("[WABLAS] API returned non-200 status")
		
		status["status"] = "NOT CONNECTED"
		status["qr"] = "timeout"
		status["details"] = map[string]interface{}{
			"http_status": resp.StatusCode,
			"error":       "API returned error status",
			"response_body": string(bodyBytes),
		}
		return status
	}

	// **Decode JSON Response** - following PHP pattern
	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		logrus.WithError(err).Error("[WABLAS] Failed to parse JSON response")
		status["status"] = "NOT CONNECTED"
		status["qr"] = "timeout"
		status["details"] = map[string]interface{}{
			"error": "Failed to parse API response",
			"raw_response": string(bodyBytes),
			"parse_error": err.Error(),
		}
		return status
	}

	// Check API response status - following PHP pattern
	if apiStatus, ok := data["status"].(bool); !ok || !apiStatus {
		logrus.Warn("[WABLAS] API response status is false or missing")
		status["status"] = "NOT CONNECTED"
		status["qr"] = "timeout"
		status["details"] = data
		return status
	}

	// **Extract Device Status** - following PHP pattern
	var deviceStatus string = "UNKNOWN"
	var deviceID string
	var image interface{} = nil

	if dataObj, ok := data["data"].(map[string]interface{}); ok {
		if ds, ok := dataObj["status"].(string); ok {
			deviceStatus = ds
		}
		if serial, ok := dataObj["serial"].(string); ok {
			deviceID = serial
		}
	}

	// **STEP 2: FETCH QR CODE IF NOT CONNECTED** - following PHP pattern
	if deviceStatus == "disconnected" && deviceID != "" {
		qrURL := fmt.Sprintf("https://my.wablas.com/api/device/scan?token=%s", url.QueryEscape(token))
		image = qrURL
	}

	// **Return Final Response** - following PHP pattern
	status["status"] = deviceStatus
	status["provider"] = "wablas"
	if dataObj, ok := data["data"].(map[string]interface{}); ok {
		status["data"] = dataObj
	} else {
		status["data"] = map[string]interface{}{}
	}
	if image != nil {
		status["image"] = image
		status["qr"] = image // Also set qr field for compatibility
	} else {
		status["image"] = nil
		status["qr"] = nil
	}
	if message, ok := data["message"].(string); ok {
		status["message"] = message
	} else {
		status["message"] = "No message returned"
	}

	logrus.WithField("final_status", status).Info("[WABLAS] Returning status")
	return status
}

// getWablasQRCode fetches QR code from Wablas API when device is disconnected
func (h *Handlers) getWablasQRCode(token string) string {
	client := &http.Client{Timeout: 10 * time.Second}
	qrURL := fmt.Sprintf("https://my.wablas.com/api/device/scan?token=%s", url.QueryEscape(token))
	
	logrus.WithField("qr_url", qrURL).Info("[WABLAS] Fetching QR code")
	
	req, err := http.NewRequest("GET", qrURL, nil)
	if err != nil {
		logrus.WithError(err).Error("[WABLAS] Failed to create QR request")
		return ""
	}
	
	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("[WABLAS] Failed to fetch QR code")
		return ""
	}
	defer resp.Body.Close()
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("[WABLAS] Failed to read QR response body")
		return ""
	}
	
	if resp.StatusCode == 200 {
		var qrResponse map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &qrResponse); err == nil {
			if data, ok := qrResponse["data"].(map[string]interface{}); ok {
				if qrCode, ok := data["qr"].(string); ok {
					logrus.WithField("qr_code_length", len(qrCode)).Info("[WABLAS] QR code fetched successfully")
					return qrCode
				}
			}
		}
	}
	
	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response_body": string(bodyBytes),
	}).Warn("[WABLAS] Failed to get QR code")
	
	return ""
}

// DebugDevices returns all device settings for debugging
func (h *Handlers) DebugDevices(c *fiber.Ctx) error {
	devices, err := h.deviceSettingsService.GetAll()
	if err != nil {
		return h.errorResponse(c, 500, "Failed to get device settings")
	}

	// Create a simplified view for debugging
	var debugData []map[string]interface{}
	for _, device := range devices {
		data := map[string]interface{}{
				"id":          device.ID,
				"provider":    device.Provider,
				"id_device":   getStringFromNullString(device.IDDevice),
				"instance":    getStringFromNullString(device.Instance),
				"device_id":   getStringFromNullString(device.DeviceID),
				"phone_number": getStringFromNullString(device.PhoneNumber),
				"created_at":  device.CreatedAt,
			}
		debugData = append(debugData, data)
	}

	return h.successResponse(c, map[string]interface{}{
		"total_devices": len(devices),
		"devices":       debugData,
	})
}

// Helper function to convert sql.NullString to string
// processWebhookMessage processes incoming webhook messages and integrates with AI WhatsApp service
func (h *Handlers) processWebhookMessage(webhookData map[string]interface{}, idDevice, provider string) {
	logrus.WithFields(logrus.Fields{
		"id_device": idDevice,
		"provider": provider,
		"webhook_data": webhookData,
	}).Info("🔄 WEBHOOK: Processing webhook message for AI integration")

	// Extract message data based on provider
	var from, message, messageType string
	var isGroup bool

	switch provider {
	case "whacenter":
		// Extract data for Whacenter provider
		if fromVal, ok := webhookData["from"].(string); ok {
			from = fromVal
		}
		if msgVal, ok := webhookData["message"].(string); ok {
			message = msgVal
		}
		if msgTypeVal, ok := webhookData["message_type"].(string); ok {
			messageType = msgTypeVal
		}
		if isGroupVal, ok := webhookData["is_group"].(bool); ok {
			isGroup = isGroupVal
		}

	case "wablas":
		// Extract data for Wablas provider
		if fromVal, ok := webhookData["phone"].(string); ok {
			from = fromVal
		}
		if msgVal, ok := webhookData["message"].(string); ok {
			message = msgVal
		}
		if msgTypeVal, ok := webhookData["type"].(string); ok {
			messageType = msgTypeVal
		}
		// Wablas doesn't have is_group field, default to false
		isGroup = false

	default:
		// Generic webhook format
		if fromVal, ok := webhookData["from"].(string); ok {
			from = fromVal
		}
		if msgVal, ok := webhookData["message"].(string); ok {
			message = msgVal
		}
		if msgTypeVal, ok := webhookData["message_type"].(string); ok {
			messageType = msgTypeVal
		} else if msgTypeVal, ok := webhookData["type"].(string); ok {
			messageType = msgTypeVal
		}
		if isGroupVal, ok := webhookData["is_group"].(bool); ok {
			isGroup = isGroupVal
		}
	}

	// Validate required fields
	if from == "" || message == "" {
		logrus.WithFields(logrus.Fields{
			"id_device": idDevice,
			"from": from,
			"message": message,
		}).Warn("⚠️ WEBHOOK: Missing required fields (from or message)")
		return
	}

	// Skip group messages if configured to do so
	if isGroup {
		logrus.WithFields(logrus.Fields{
			"id_device": idDevice,
			"from": from,
		}).Info("📱 WEBHOOK: Skipping group message")
		return
	}

	// Check for media URLs in bracket format and extract clean text for processing
	// This allows proper handling of bracket format media URLs as user input
	if h.mediaDetectionService.HasMedia(message) {
		mediaResults := h.mediaDetectionService.DetectMedia(message)
		if len(mediaResults) > 0 {
			// Use the clean text (with bracket format removed) for further processing
			cleanMessage := mediaResults[0].CleanText
			
			logrus.WithFields(logrus.Fields{
				"id_device": idDevice,
				"from": from,
				"original_message": message,
				"clean_message": cleanMessage,
				"detected_media_count": len(mediaResults),
			}).Info("📎 WEBHOOK: Detected bracket format media URLs, using clean text for processing")
			
			// Update message to clean text for further processing
			message = strings.TrimSpace(cleanMessage)
			
			// If clean message is empty after removing media URLs, skip processing
			if message == "" {
				logrus.WithFields(logrus.Fields{
					"id_device": idDevice,
					"from": from,
				}).Info("📎 WEBHOOK: Message contained only media URLs, skipping text processing")
				return
			}
		}
	}

	// Only process text messages for non-media content
	if messageType != "text" && messageType != "" {
		logrus.WithFields(logrus.Fields{
			"id_device": idDevice,
			"from": from,
			"message_type": messageType,
		}).Info("📱 WEBHOOK: Skipping non-text message")
		return
	}

	// Check if this is a device command (%, #, cmd)
	if strings.HasPrefix(message, "%") || strings.HasPrefix(message, "#") || strings.ToLower(strings.TrimSpace(message)) == "cmd" {
		logrus.WithFields(logrus.Fields{
			"id_device": idDevice,
			"from": from,
			"command": message,
		}).Info("⚙️ WEBHOOK: Processing device command")

		// Process device command through AI WhatsApp handlers
		if h.aiWhatsappHandlers != nil && h.aiWhatsappHandlers.AIWhatsappService != nil {
			err := h.aiWhatsappHandlers.AIWhatsappService.ProcessDeviceCommand(from, message, idDevice)
			if err != nil {
				logrus.WithError(err).Error("❌ WEBHOOK: Failed to process device command")
			}
		} else {
			logrus.Error("❌ WEBHOOK: AI WhatsApp service not available")
		}
		return
	}



	// Check if device has a configured flow - prioritize flow engine over AI conversation
	flows, err := h.flowService.GetFlowsByDevice(idDevice)
	if err != nil {
		logrus.WithError(err).Warn("⚠️ WEBHOOK: Failed to check for device flows")
	}

	// If device has configured flows, use the flow engine
	if len(flows) > 0 {
		logrus.WithFields(logrus.Fields{
			"id_device": idDevice,
			"from": from,
			"message": message,
			"provider": provider,
			"flow_count": len(flows),
		}).Info("🔄 WEBHOOK: Processing message through flow engine")

		// Process message through WhatsApp service flow engine
		if h.whatsappService != nil {
			err := h.whatsappService.ProcessIncomingMessageFromWebhook(from, message, idDevice, provider)
			if err != nil {
				logrus.WithError(err).Error("❌ WEBHOOK: Failed to process message through flow engine")
				// Fallback to AI conversation if flow processing fails
				h.processAIConversation(from, message, idDevice, provider)
			}
		} else {
			logrus.Error("❌ WEBHOOK: WhatsApp service not available, falling back to AI conversation")
			h.processAIConversation(from, message, idDevice, provider)
		}
		return
	}

	// No flows configured, use AI conversation system
	logrus.WithFields(logrus.Fields{
		"id_device": idDevice,
		"from": from,
		"message": message,
		"provider": provider,
	}).Info("🤖 WEBHOOK: No flows configured, processing message through AI conversation")

	h.processAIConversation(from, message, idDevice, provider)
}

// processAIConversation handles message processing through the AI conversation system
func (h *Handlers) processAIConversation(from, message, idDevice, provider string) {
	// Get current conversation stage from AI WhatsApp repository
	var stage string
	if h.aiWhatsappHandlers != nil && h.aiWhatsappHandlers.AIRepo != nil {
		aiConv, err := h.aiWhatsappHandlers.AIRepo.GetAIWhatsappByProspectNum(from)
		if err != nil {
			logrus.WithError(err).Warn("⚠️ WEBHOOK: Failed to get AI conversation stage")
		} else if aiConv != nil {
			stage = aiConv.Stage
		}
	}

	// Process AI conversation through AI WhatsApp service
	if h.aiWhatsappHandlers != nil && h.aiWhatsappHandlers.AIWhatsappService != nil {
		response, err := h.aiWhatsappHandlers.AIWhatsappService.ProcessAIConversation(from, idDevice, message, stage)
		if err != nil {
			logrus.WithError(err).Error("❌ WEBHOOK: Failed to process AI conversation")
			return
		}

		// Save conversation history and send response if we have a response
		if response != nil {
			// Extract bot response text from response array
			var botResponseText string
			for _, item := range response.Response {
				if item.Type == "text" {
					if botResponseText != "" {
						botResponseText += " "
					}
					botResponseText += item.Content
				}
			}

			// Save conversation history to conv_last field
			err = h.aiWhatsappHandlers.AIWhatsappService.SaveConversationHistory(from, idDevice, message, botResponseText, response.Stage)
			if err != nil {
				logrus.WithError(err).Error("❌ WEBHOOK: Failed to save conversation history")
			}

			logrus.WithFields(logrus.Fields{
				"id_device": idDevice,
				"to": from,
				"provider": provider,
			}).Info("📤 WEBHOOK: Sending AI response back to WhatsApp")

			// Send response through the appropriate provider
			h.sendWhatsappResponse(from, idDevice, provider, response)
		}
	} else {
		logrus.Error("❌ WEBHOOK: AI WhatsApp service not available")
	}
}

// sendWhatsappResponse sends AI response back to WhatsApp through the appropriate provider
func (h *Handlers) sendWhatsappResponse(to, idDevice, provider string, response interface{}) {
	logrus.WithFields(logrus.Fields{
		"to": to,
		"id_device": idDevice,
		"provider": provider,
	}).Info("📤 WHATSAPP: Sending response")

	// Get device settings to retrieve API credentials
	deviceSettings, err := h.deviceSettingsService.GetByIDDevice(idDevice)
	if err != nil {
		logrus.WithError(err).Error("❌ WHATSAPP: Failed to get device settings")
		return
	}

	// Parse response data - handle AIWhatsappResponse struct
	var aiResponse *services.AIWhatsappResponse
	switch v := response.(type) {
	case *services.AIWhatsappResponse:
		aiResponse = v
	case services.AIWhatsappResponse:
		aiResponse = &v
	default:
		logrus.WithField("response_type", fmt.Sprintf("%T", response)).Error("❌ WHATSAPP: Invalid response format")
		return
	}

	// Validate response structure
	if aiResponse == nil || len(aiResponse.Response) == 0 {
		logrus.Error("❌ WHATSAPP: No response messages found")
		return
	}

	// Send each response message
	for _, respItem := range aiResponse.Response {
		if respItem.Content == "" {
			continue
		}

		switch respItem.Type {
		case "text":
			h.sendTextMessage(to, respItem.Content, deviceSettings, provider)
		case "image":
			h.sendImageMessage(to, respItem.Content, deviceSettings, provider)
		case "audio":
			// Send audio message using sendChatMessage for multimedia support
			h.sendChatMessage(to, "", respItem.Content, deviceSettings, 1*time.Second)
		case "video":
			// Send video message using sendChatMessage for multimedia support
			h.sendChatMessage(to, "", respItem.Content, deviceSettings, 1*time.Second)
		default:
			// Default to text message
			h.sendTextMessage(to, respItem.Content, deviceSettings, provider)
		}

		// Add small delay between messages to avoid rate limiting
		time.Sleep(5000 * time.Millisecond)
	}
}

// sendTextMessage sends a text message through the appropriate provider with delay support
func (h *Handlers) sendTextMessage(to, message string, deviceSettings *models.DeviceSettings, provider string) {
	// Add delay before sending (similar to PHP delax parameter)
	delay := 1 * time.Second
	time.Sleep(delay)

	// Determine provider based on instance length if not specified
	if provider == "" {
		provider = h.determineProviderFromInstance(deviceSettings.Instance.String)
	}

	switch provider {
	case "whacenter":
		h.sendWhacenterTextMessage(to, message, deviceSettings)
	case "wablas":
		h.sendWablasTextMessage(to, message, deviceSettings)

	default:
		logrus.WithField("provider", provider).Warn("⚠️ WHATSAPP: Unsupported provider for text message")
	}
}

// sendImageMessage sends an image message through the appropriate provider with delay support
func (h *Handlers) sendImageMessage(to, imageURL string, deviceSettings *models.DeviceSettings, provider string) {
	// Add delay before sending (similar to PHP delax parameter)
	delay := 1 * time.Second
	time.Sleep(delay)

	// Determine provider based on instance length if not specified
	if provider == "" {
		provider = h.determineProviderFromInstance(deviceSettings.Instance.String)
	}

	switch provider {
	case "whacenter":
		h.sendWhacenterImageMessage(to, imageURL, deviceSettings)
	case "wablas":
		h.sendWablasImageMessage(to, imageURL, deviceSettings)

	default:
		logrus.WithField("provider", provider).Warn("⚠️ WHATSAPP: Unsupported provider for image message")
	}
}

// sendWhacenterTextMessage sends text message via Whacenter API
func (h *Handlers) sendWhacenterTextMessage(to, message string, deviceSettings *models.DeviceSettings) {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WHACENTER: No instance available")
		return
	}

	// Whacenter API endpoint for sending messages
	apiURL := "https://api.whacenter.com/api/send"

	// Prepare request payload - Use instance for device_id as per Whacenter API requirements
	payload := map[string]interface{}{
		"device_id": deviceSettings.Instance.String, // ✅ Use instance
		"number": to,
		"message": message,
		"type": "text",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to marshal payload")
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to create request")
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+deviceSettings.Instance.String)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to send message")
		return
	}
	defer resp.Body.Close()

	// Read response body for error details
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to read response body")
		return
	}

	// Log response details
	logFields := logrus.Fields{
		"to": to,
		"status_code": resp.StatusCode,
		"response_body": string(respBody),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logFields["status"] = "success"
		logrus.WithFields(logFields).Info("📤 WHACENTER: Text message sent successfully")
	} else {
		logFields["status"] = "error"
		logrus.WithFields(logFields).Error("❌ WHACENTER: Text message failed")
	}
}

// sendWablasTextMessage sends text message via Wablas API
func (h *Handlers) sendWablasTextMessage(to, message string, deviceSettings *models.DeviceSettings) {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WABLAS: No instance available")
		return
	}

	// Wablas API endpoint for sending messages
	apiURL := "https://my.wablas.com/api/send-message"

	// Prepare form data
	formData := url.Values{}
	formData.Set("phone", to)
	formData.Set("message", message)
	formData.Set("isGroup", "false")

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		logrus.WithError(err).Error("❌ WABLAS: Failed to create request")
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", deviceSettings.Instance.String)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WABLAS: Failed to send message")
		return
	}
	defer resp.Body.Close()

	// Read response body for error details
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("❌ WABLAS: Failed to read response body")
		return
	}

	// Log response details
	logFields := logrus.Fields{
		"to": to,
		"status_code": resp.StatusCode,
		"response_body": string(respBody),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logFields["status"] = "success"
		logrus.WithFields(logFields).Info("📤 WABLAS: Text message sent successfully")
	} else {
		logFields["status"] = "error"
		logrus.WithFields(logFields).Error("❌ WABLAS: Text message failed")
	}
}

// sendWhacenterImageMessage sends image message via Whacenter API
func (h *Handlers) sendWhacenterImageMessage(to, imageURL string, deviceSettings *models.DeviceSettings) {
	h.sendWhacenterMultimediaMessage(to, "", imageURL, "image", deviceSettings)
}

// sendWhacenterMultimediaMessage sends multimedia messages (video, audio, image) via Whacenter API
// Equivalent to PHP sendChatMessage function for Whacenter provider
func (h *Handlers) sendWhacenterMultimediaMessage(to, caption, fileURL, fileType string, deviceSettings *models.DeviceSettings) {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WHACENTER: No instance available")
		return
	}

	// Whacenter API endpoint for sending media
	apiURL := "https://api.whacenter.com/api/send-media"

	// Prepare request payload - Use instance for device_id as per Whacenter API requirements
	payload := map[string]interface{}{
		"device_id": deviceSettings.Instance.String, // ✅ Use instance
		"number":    to,
		"media_url": fileURL,
		"type":      fileType,
	}

	// Add caption if provided
	if caption != "" {
		payload["caption"] = caption
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to marshal payload")
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to create request")
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+deviceSettings.Instance.String)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to send multimedia message")
		return
	}
	defer resp.Body.Close()

	// Read response body for error details
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: Failed to read response body")
		return
	}

	// Log response details
	logFields := logrus.Fields{
		"to":          to,
		"file_type":   fileType,
		"status_code": resp.StatusCode,
		"response_body": string(respBody),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logFields["status"] = "success"
		logrus.WithFields(logFields).Info("📤 WHACENTER: Multimedia message sent successfully")
	} else {
		logFields["status"] = "error"
		logrus.WithFields(logFields).Error("❌ WHACENTER: Multimedia message failed")
	}
}

// sendWablasImageMessage sends image message via Wablas API
func (h *Handlers) sendWablasImageMessage(to, imageURL string, deviceSettings *models.DeviceSettings) {
	h.sendWablasMultimediaMessage(to, "", imageURL, "image", deviceSettings)
}

// sendWablasMultimediaMessage sends multimedia messages (video, audio, image) via Wablas API
// Equivalent to PHP sendChatMessage function for Wablas provider
func (h *Handlers) sendWablasMultimediaMessage(to, caption, fileURL, fileType string, deviceSettings *models.DeviceSettings) {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WABLAS: No instance available")
		return
	}

	// Determine API endpoint based on file type
	var apiURL string
	var fieldName string

	switch fileType {
	case "video":
		apiURL = "https://my.wablas.com/api/send-video"
		fieldName = "video"
	case "audio":
		apiURL = "https://my.wablas.com/api/send-audio"
		fieldName = "audio"
	default: // image
		apiURL = "https://my.wablas.com/api/send-image"
		fieldName = "image"
	}

	// Prepare form data
	formData := url.Values{}
	formData.Set("phone", to)
	formData.Set(fieldName, fileURL)
	if caption != "" {
		formData.Set("caption", caption)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		logrus.WithError(err).Error("❌ WABLAS: Failed to create request")
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", deviceSettings.Instance.String)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WABLAS: Failed to send multimedia message")
		return
	}
	defer resp.Body.Close()

	logrus.WithFields(logrus.Fields{
		"to":          to,
		"file_type":   fileType,
		"status_code": resp.StatusCode,
	}).Info("📤 WABLAS: Multimedia message sent")
}

// determineProviderFromInstance determines the provider based on instance string length
// Based on PHP logic: if instance length > 20 then Whacenter, else Wablas
func (h *Handlers) determineProviderFromInstance(instance string) string {
	characterCount := len(instance)
	if characterCount <= 20 {
		return "wablas"
	}
	return "whacenter"
}

// sendChatMessage sends multimedia messages (video, audio, image) with caption support
// Equivalent to PHP sendChatMessage function
func (h *Handlers) sendChatMessage(to, reply, fileURL string, deviceSettings *models.DeviceSettings, delay time.Duration) {
	// Console log for tracing media URL in handlers
	logrus.WithFields(logrus.Fields{
		"to": to,
		"file_url": fileURL,
		"device_id": deviceSettings.IDDevice,
		"file_url_length": len(fileURL),
		"delay_ms": delay.Milliseconds(),
	}).Info("🔍 HANDLERS: MEDIA URL RECEIVED FOR TRACING")

	// Add delay before sending
	time.Sleep(delay)

	// Determine provider based on instance length
	provider := h.determineProviderFromInstance(deviceSettings.Instance.String)

	// Determine file type based on extension
	fileType := h.getFileType(fileURL)

	switch provider {
	case "wablas":
		h.sendWablasMultimediaMessage(to, reply, fileURL, fileType, deviceSettings)
	case "whacenter":
		h.sendWhacenterMultimediaMessage(to, reply, fileURL, fileType, deviceSettings)

	default:
		logrus.WithField("provider", provider).Warn("⚠️ WHATSAPP: Unsupported provider for multimedia message")
	}
}

// getFileType determines file type based on file extension
func (h *Handlers) getFileType(fileURL string) string {
	var fileType string
	if strings.Contains(fileURL, ".mp4") {
		fileType = "video"
	} else if strings.Contains(fileURL, ".mp3") {
		fileType = "audio"
	} else {
		fileType = "image"
	}

	// Console log for tracing file type determination
	logrus.WithFields(logrus.Fields{
		"file_url": fileURL,
		"determined_type": fileType,
		"has_mp4": strings.Contains(fileURL, ".mp4"),
		"has_mp3": strings.Contains(fileURL, ".mp3"),
		"default_to_image": !strings.Contains(fileURL, ".mp4") && !strings.Contains(fileURL, ".mp3"),
	}).Info("🔍 HANDLERS: FILE TYPE DETERMINED FOR TRACING")

	return fileType
}



func getStringFromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}