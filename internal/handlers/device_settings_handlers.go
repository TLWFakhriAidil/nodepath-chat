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

	// Delete existing device first (cleanup)
	if req.DeviceID != "" {
		logrus.Info("Cleaning up existing Whacenter device")
		
		// Call Whacenter delete API
		deleteURL := fmt.Sprintf("https://api.whacenter.com/api/deleteDevice?api_key=%s&device_id=%s", 
			req.APIKey, req.DeviceID)
		
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
				logrus.WithField("status", deleteResp.StatusCode).Info("Device deletion attempted")
			}
		}
	}

	// Check if user has existing device_id in database for Whacenter
	var whacenterAPIKey string
	if req.DeviceID == "" {
		// Use the provided default Whacenter API key when device_id is empty
		whacenterAPIKey = "abebe840-156c-441c-8252-da0342c5a07c"
	} else {
		// Use user's device_id as API key when device_id is not empty
		whacenterAPIKey = req.DeviceID
	}

	// Prepare Whacenter API request with GET parameters
	whacenterURL := fmt.Sprintf("https://api.whacenter.com/api/addDevice?api_key=%s&name=%s&number=%s", 
		whacenterAPIKey, req.IDDevice, req.WebhookURL)

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
	if status, ok := apiResponse["status"].(bool); !ok || !status {
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

	deviceID, _ := data["device_id"].(string)
	apiKey, _ := data["api_key"].(string)

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
			"webhook_url": req.WebhookURL,
			"api_key":     apiKey,
			"provider":    "whacenter",
		},
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

	// Delete existing device first (cleanup)
	if req.DeviceID != "" {
		logrus.Info("Cleaning up existing Wablas device")
		
		// Call Wablas delete API
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
			// Set headers for delete request
			deleteRequest.Header.Set("Authorization", req.DeviceID)
			deleteRequest.Header.Set("Accept", "application/json")
			
			// Execute delete request
			deleteResp, err := deleteClient.Do(deleteRequest)
			if err != nil {
				logrus.WithError(err).Error("Failed to delete existing Wablas device")
			} else {
				defer deleteResp.Body.Close()
				logrus.WithField("status_code", deleteResp.StatusCode).Info("Wablas device deletion attempted")
			}
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

	// Use the provided default Wablas credentials for device creation
	wablasToken := "j0oB1aibqYDQlgyk9SIqLyfeGgRJjjmOUFMVqxGd8Irk6JCwl1ZxYtY.7hDkbW0f"
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

	// Configure webhook URL with auth header
	webhookURL := fmt.Sprintf("https://chatbot.growrvsb.com/chatgpt/%s/%s", req.IDDevice, newAuthHeader)

	// Setup webhook configuration using the correct endpoint
	webhookFormData := url.Values{}
	webhookFormData.Set("webhook_url", webhookURL)
	
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

	// Log successful device generation
	logrus.WithFields(logrus.Fields{
		"provider": "wablas",
		"device_id": deviceID,
		"webhook_url": webhookURL,
		"phone_number": req.PhoneNumber,
	}).Info("✅ WABLAS: Device generated successfully")
	
	// Return success response
	return h.successResponse(c, map[string]interface{}{
		"success": true,
		"message": "Device generated successfully via Wablas",
		"data": map[string]interface{}{
			"device_id":   deviceID,
			"webhook_url": webhookURL,
			"api_key":     newAuthHeader,
			"provider":    "wablas",
		},
	})
}