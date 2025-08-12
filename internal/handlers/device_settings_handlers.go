package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"nodepath-chat/internal/models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

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
	
	// TODO: Process the webhook data based on provider type
	// This is where you would integrate with your chatbot flow engine
	// For now, we just acknowledge receipt
	
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
	if deviceID == "" {
		return h.errorResponse(c, 400, "Device ID is required")
	}

	// Get device settings
	device, err := h.deviceSettingsService.GetByID(deviceID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings")
		return h.errorResponse(c, 404, "Device not found")
	}

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
		status = h.checkWhacenterStatus(device, status)
	case "wablas":
		status = h.checkWablasStatus(device, status)
	default:
		status["status"] = "unsupported_provider"
		status["details"] = map[string]interface{}{
			"error": "Provider not supported for status checking",
		}
	}

	return h.successResponse(c, status)
}

// checkWhacenterStatus checks the status of a Whacenter device
func (h *Handlers) checkWhacenterStatus(device *models.DeviceSettings, status map[string]interface{}) map[string]interface{} {
	if !device.Instance.Valid || device.Instance.String == "" {
		status["status"] = "not_configured"
		status["details"] = map[string]interface{}{
			"error": "Device instance not configured",
		}
		return status
	}

	// Make API call to check Whacenter device status
	client := &http.Client{Timeout: 10 * time.Second}
	apiKey := "abebe840-156c-441c-8252-da0342c5a07c" // Use the same hardcoded API key

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.whacenter.com/api/device/%s/status", device.Instance.String), nil)
	if err != nil {
		status["status"] = "error"
		status["details"] = map[string]interface{}{
			"error": "Failed to create status request",
		}
		return status
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		status["status"] = "connection_error"
		status["details"] = map[string]interface{}{
			"error": "Failed to connect to Whacenter API",
		}
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var apiResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err == nil {
			if connected, ok := apiResponse["connected"].(bool); ok {
				status["connected"] = connected
				if connected {
					status["status"] = "connected"
				} else {
					status["status"] = "disconnected"
				}
			}
			status["details"] = apiResponse
		}
	} else {
		status["status"] = "api_error"
		status["details"] = map[string]interface{}{
			"http_status": resp.StatusCode,
			"error":       "API returned error status",
		}
	}

	return status
}

// checkWablasStatus checks the status of a Wablas device
func (h *Handlers) checkWablasStatus(device *models.DeviceSettings, status map[string]interface{}) map[string]interface{} {
	if !device.Instance.Valid || device.Instance.String == "" {
		status["status"] = "not_configured"
		status["details"] = map[string]interface{}{
			"error": "Device instance not configured",
		}
		return status
	}

	// Make API call to check Wablas device status
	client := &http.Client{Timeout: 10 * time.Second}
	token := device.Instance.String // The instance contains the auth token for Wablas

	req, err := http.NewRequest("GET", "https://my.wablas.com/api/device/status", nil)
	if err != nil {
		status["status"] = "error"
		status["details"] = map[string]interface{}{
			"error": "Failed to create status request",
		}
		return status
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		status["status"] = "connection_error"
		status["details"] = map[string]interface{}{
			"error": "Failed to connect to Wablas API",
		}
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var apiResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err == nil {
			if statusField, ok := apiResponse["status"].(bool); ok {
				status["connected"] = statusField
				if statusField {
					status["status"] = "connected"
				} else {
					status["status"] = "disconnected"
				}
			}
			status["details"] = apiResponse
		}
	} else {
		status["status"] = "api_error"
		status["details"] = map[string]interface{}{
			"http_status": resp.StatusCode,
			"error":       "API returned error status",
		}
	}

	return status
}