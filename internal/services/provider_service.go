package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// ProviderService handles message sending through external providers (Wablas, Whacenter, WAHA)
type ProviderService struct {
	httpClient *http.Client
}

// NewProviderService creates a new provider service instance
func NewProviderService() *ProviderService {
	return &ProviderService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendMessage sends a message through the appropriate provider based on device settings
func (ps *ProviderService) SendMessage(deviceSettings *models.DeviceSettings, phoneNumber, message string) error {
	if deviceSettings == nil {
		return fmt.Errorf("device settings cannot be nil")
	}

	// Get provider from device settings
	provider := strings.ToLower(deviceSettings.Provider)
	logrus.WithFields(logrus.Fields{
		"provider":     provider,
		"device_id":    deviceSettings.Instance.String,
		"phone_number": phoneNumber,
	}).Info("📤 MESSAGE: Sending message through provider")

	switch provider {
	case "wablas":
		return ps.sendWablasMessage(deviceSettings, phoneNumber, message)
	case "whacenter":
		return ps.sendWhacenterMessage(deviceSettings, phoneNumber, message)
	case "waha":
		return ps.sendWahaMessage(deviceSettings, phoneNumber, message)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// SendMediaMessage sends a media message through the appropriate provider
func (ps *ProviderService) SendMediaMessage(deviceSettings *models.DeviceSettings, phoneNumber, mediaURL string) error {
	if deviceSettings == nil {
		return fmt.Errorf("device settings cannot be nil")
	}

	// Get provider from device settings
	provider := strings.ToLower(deviceSettings.Provider)
	logrus.WithFields(logrus.Fields{
		"provider":     provider,
		"device_id":    deviceSettings.Instance.String,
		"phone_number": phoneNumber,
		"media_url":    mediaURL,
	}).Info("📤 MEDIA: Sending media message through provider")

	switch provider {
	case "wablas":
		return ps.sendWablasImageMessage(deviceSettings, phoneNumber, mediaURL)
	case "whacenter":
		return ps.sendWhacenterMediaMessage(deviceSettings, phoneNumber, mediaURL)
	case "waha":
		return ps.sendWahaMediaMessage(deviceSettings, phoneNumber, mediaURL)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// sendWablasMessage sends a text message via Wablas API
// Uses the exact API format specified by user requirements
func (ps *ProviderService) sendWablasMessage(deviceSettings *models.DeviceSettings, phoneNumber, message string) error {
	// Prevent sending empty or whitespace-only messages to avoid <nil> messages
	if message == "" || strings.TrimSpace(message) == "" {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceSettings.Instance.String,
		}).Warn("[WABLAS-TEXT] Skipping empty message to prevent <nil> message")
		return nil
	}

	apiURL := "https://my.wablas.com/api/send-message"

	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"message_len":  len(message),
		"device_id":    deviceSettings.Instance.String,
	}).Debug("[WABLAS-TEXT] Preparing request")

	// Get instance for authorization (as per user requirements)
	instance := ""
	if deviceSettings.Instance.Valid {
		instance = deviceSettings.Instance.String
	} else {
		return fmt.Errorf("no instance found for Wablas device %s", deviceSettings.Instance.String)
	}

	// Prepare form data exactly as specified by user
	data := url.Values{}
	data.Set("phone", phoneNumber) // Recipient phone number
	data.Set("message", message)   // Message content

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers exactly as specified by user
	req.Header.Set("Authorization", instance) // Set the Authorization header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	startTime := time.Now()
	resp, err := ps.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	duration := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response":    string(body),
		"duration":    duration,
		"instance":    instance,
	}).Debug("[WABLAS-TEXT] Response received")

	// Check for success (200-299 status codes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wablas API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
		"device_id":    deviceSettings.Instance.String,
	}).Info("[WABLAS-TEXT] ✅ Message sent successfully")

	return nil
}

// sendWablasImageMessage sends a media message via Wablas API with type detection
// Handles video, audio, and image files with appropriate API endpoints
func (ps *ProviderService) sendWablasImageMessage(deviceSettings *models.DeviceSettings, phoneNumber, mediaURL string) error {
	// Detect media type based on file extension
	mediaType := ""
	var apiURL string
	var fieldName string

	if strings.Contains(mediaURL, ".mp4") {
		mediaType = "video"
		apiURL = "https://my.wablas.com/api/send-video"
		fieldName = "video"
	} else if strings.Contains(mediaURL, ".mp3") {
		mediaType = "audio"
		apiURL = "https://my.wablas.com/api/send-audio"
		fieldName = "audio"
	} else {
		// Default to image for all other file types
		mediaType = "image"
		apiURL = "https://my.wablas.com/api/send-image"
		fieldName = "image"
	}

	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"media_url":    mediaURL,
		"media_type":   mediaType,
		"device_id":    deviceSettings.Instance.String,
	}).Debug("[WABLAS-MEDIA] Preparing request")

	// Get instance for authorization (as per user requirements)
	instance := ""
	if deviceSettings.Instance.Valid {
		instance = deviceSettings.Instance.String
	} else {
		return fmt.Errorf("no instance found for Wablas device %s", deviceSettings.Instance.String)
	}

	// Prepare form data with appropriate field name
	data := url.Values{}
	data.Set("phone", phoneNumber) // Recipient phone number
	data.Set(fieldName, mediaURL)  // Media file URL with correct field name

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (using instance for authorization as per user requirements)
	req.Header.Set("Authorization", instance)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	startTime := time.Now()
	resp, err := ps.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	duration := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response":    string(body),
		"duration":    duration,
		"instance":    instance,
		"media_type":  mediaType,
	}).Debug("[WABLAS-MEDIA] Response received")

	// Check for success (200-299 status codes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wablas API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
		"device_id":    deviceSettings.Instance.String,
		"media_type":   mediaType,
	}).Info("[WABLAS-MEDIA] ✅ Media sent successfully")

	return nil
}

// sendWhacenterMessage sends a text message via Whacenter API
// Uses the exact API format specified by user requirements
func (ps *ProviderService) sendWhacenterMessage(deviceSettings *models.DeviceSettings, phoneNumber, message string) error {
	// Prevent sending empty or whitespace-only messages to avoid <nil> messages
	if message == "" || strings.TrimSpace(message) == "" {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceSettings.Instance.String,
		}).Warn("[WHACENTER] Skipping empty message to prevent <nil> message")
		return nil
	}

	apiURL := "https://api.whacenter.com/api/send"

	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"message_len":  len(message),
		"device_id":    deviceSettings.Instance.String,
	}).Debug("[WHACENTER] Preparing request")

	// Get instance for device_id (as per user requirements)
	instance := ""
	if deviceSettings.Instance.Valid {
		instance = deviceSettings.Instance.String
	} else {
		return fmt.Errorf("no instance found for Whacenter device %s", deviceSettings.Instance.String)
	}

	// Prepare form data exactly as specified by user
	data := url.Values{}
	data.Set("device_id", instance) // device_id from instance
	data.Set("number", phoneNumber) // recipient number
	data.Set("message", message)    // message content

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (form data, no authorization header as per user example)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	startTime := time.Now()
	resp, err := ps.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	duration := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response":    string(body),
		"duration":    duration,
		"instance":    instance,
	}).Debug("[WHACENTER] Response received")

	// Check for success (200-299 status codes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("whacenter API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
		"device_id":    deviceSettings.Instance.String,
	}).Info("[WHACENTER] ✅ Message sent successfully")

	return nil
}

// sendWhacenterMediaMessage sends a media message via Whacenter API
// Uses the exact API format specified by user requirements with type detection
func (ps *ProviderService) sendWhacenterMediaMessage(deviceSettings *models.DeviceSettings, phoneNumber, mediaURL string) error {
	apiURL := "https://api.whacenter.com/api/send"

	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"media_url":    mediaURL,
		"device_id":    deviceSettings.Instance.String,
	}).Debug("[WHACENTER] Preparing media request")

	// Get instance for device_id (as per user requirements)
	instance := ""
	if deviceSettings.Instance.Valid {
		instance = deviceSettings.Instance.String
	} else {
		return fmt.Errorf("no instance found for Whacenter device %s", deviceSettings.Instance.String)
	}

	// Detect media type based on file extension (as per PHP code)
	mediaType := ""
	if strings.Contains(mediaURL, ".mp4") {
		mediaType = "video"
	} else if strings.Contains(mediaURL, ".mp3") {
		mediaType = "audio"
	} else {
		mediaType = "image"
	}

	// Prepare form data exactly as specified by user PHP code
	data := url.Values{}
	data.Set("device_id", instance) // device_id from instance
	data.Set("number", phoneNumber) // recipient number
	data.Set("file", mediaURL)      // media file URL

	// Add type parameter for video and audio only (as per PHP code)
	if mediaType != "" && mediaType != "image" {
		data.Set("type", mediaType)
	}

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (form data, no authorization header as per user example)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	startTime := time.Now()
	resp, err := ps.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	duration := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response":    string(body),
		"duration":    duration,
		"instance":    instance,
		"media_type":  mediaType,
	}).Debug("[WHACENTER] Media response received")

	// Check for success (200-299 status codes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("whacenter API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
		"device_id":    deviceSettings.Instance.String,
		"media_type":   mediaType,
	}).Info("[WHACENTER] ✅ Media sent successfully")

	return nil
}

// sendWahaMessage sends a text message via WAHA API
// Uses the WAHA HTTP API format as per documentation
func (ps *ProviderService) sendWahaMessage(deviceSettings *models.DeviceSettings, phoneNumber, message string) error {
	// Prevent sending empty or whitespace-only messages to avoid <nil> messages
	if message == "" || strings.TrimSpace(message) == "" {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceSettings.Instance.String,
		}).Warn("[WAHA-TEXT] Skipping empty message to prevent <nil> message")
		return nil
	}

	// Hardcoded API key for WAHA provider
	apiKey := "dckr_pat_vxeqEu_CqRi5O3CBHnD7FxhnBz0"

	// Get instance for session (as per user requirements)
	instance := ""
	if deviceSettings.Instance.Valid {
		instance = deviceSettings.Instance.String
	} else {
		return fmt.Errorf("no instance found for WAHA device %s", deviceSettings.Instance.String)
	}

	// WAHA API endpoint for sending text messages
	apiURL := "https://waha-plus-production-705f.up.railway.app/api/sendText"

	// 🚨 DEBUG: Log API key details (masked for security)
	maskedAPIKey := "<empty>"
	if len(apiKey) > 8 {
		maskedAPIKey = apiKey[:4] + "******" + apiKey[len(apiKey)-4:]
	} else if len(apiKey) > 0 {
		maskedAPIKey = "****" + apiKey[len(apiKey)-2:]
	}

	logrus.WithFields(logrus.Fields{
		"api_url":        apiURL,
		"phone_number":   phoneNumber,
		"message_len":    len(message),
		"device_id":      deviceSettings.Instance.String,
		"instance":       instance,
		"api_key_masked": maskedAPIKey,
		"api_key_length": len(apiKey),
	}).Error("🚨 WAHA DEBUG: Preparing request with API key details")

	// Format phone number for WAHA (international format without + and add @c.us)
	chatId := phoneNumber
	if !strings.HasSuffix(chatId, "@c.us") {
		// Remove + if present and add @c.us
		chatId = strings.TrimPrefix(chatId, "+") + "@c.us"
	}

	// Prepare JSON payload as per WAHA API documentation
	payload := map[string]interface{}{
		"session": instance, // Session name from instance
		"chatId":  chatId,   // Phone number in WAHA format
		"text":    message,  // Message content
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// 🚨 DEBUG: Log complete payload details
	logrus.WithFields(logrus.Fields{
		"payload":   payload,
		"json_data": string(jsonData),
		"chat_id":   chatId,
		"session":   instance,
		"message":   message,
	}).Error("🚨 WAHA DEBUG: Complete payload prepared")

	// Create request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers as per WAHA API documentation
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Api-Key", apiKey) // API key for authentication

	// 🚨 DEBUG: Log request headers (with masked API key)
	headersCopy := make(map[string]string)
	for key, values := range req.Header {
		if key == "X-Api-Key" {
			headersCopy[key] = maskedAPIKey
		} else {
			headersCopy[key] = strings.Join(values, ", ")
		}
	}
	logrus.WithFields(logrus.Fields{
		"method":         "POST",
		"url":            apiURL,
		"headers":        headersCopy,
		"content_length": len(jsonData),
	}).Error("🚨 WAHA DEBUG: Request headers and details")

	// Send request
	startTime := time.Now()
	resp, err := ps.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	duration := time.Since(startTime)

	// 🚨 DEBUG: Log complete response details
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		responseHeaders[key] = strings.Join(values, ", ")
	}

	logrus.WithFields(logrus.Fields{
		"status_code":      resp.StatusCode,
		"response_body":    string(body),
		"response_headers": responseHeaders,
		"duration":         duration,
		"instance":         instance,
		"success":          resp.StatusCode >= 200 && resp.StatusCode < 300,
	}).Error("🚨 WAHA DEBUG: Complete response received")

	// Check for success (200-299 status codes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 🚨 DEBUG: Log error details for 401 Unauthorized
		if resp.StatusCode == 401 {
			logrus.WithFields(logrus.Fields{
				"error_type":       "UNAUTHORIZED",
				"api_key_provided": len(apiKey) > 0,
				"api_key_length":   len(apiKey),
				"api_key_masked":   maskedAPIKey,
				"instance":         instance,
				"endpoint":         apiURL,
				"response_body":    string(body),
			}).Error("🚨 WAHA DEBUG: 401 UNAUTHORIZED ERROR - API Key Issue")
		}
		return fmt.Errorf("WAHA API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// 🚨 DEBUG: Log successful send
	logrus.WithFields(logrus.Fields{
		"phone_number":  phoneNumber,
		"duration":      duration,
		"device_id":     deviceSettings.Instance.String,
		"status_code":   resp.StatusCode,
		"response_body": string(body),
	}).Error("🚨 WAHA DEBUG: ✅ Message sent successfully")

	return nil
}

// sendWahaMediaMessage sends media message (image/video/audio) via WAHA API
// Handles video, audio, and image files with appropriate API endpoints matching PHP logic exactly
func (ps *ProviderService) sendWahaMediaMessage(deviceSettings *models.DeviceSettings, phoneNumber, mediaURL string) error {
	// Hardcoded API key for WAHA provider (must match WHATSAPP_API_KEY in container)
	apiKey := "dckr_pat_vxeqEu_CqRi5O3CBHnD7FxhnBz0"

	// Get instance for session
	instance := ""
	if deviceSettings.Instance.Valid {
		instance = deviceSettings.Instance.String
	} else {
		return fmt.Errorf("no instance found for WAHA device %s", deviceSettings.Instance.String)
	}

	// Format phone number - remove all non-numeric characters (matching PHP preg_replace)
	number := strings.NewReplacer(
		"+", "",
		"-", "",
		" ", "",
		"(", "",
		")", "",
	).Replace(phoneNumber)

	// Format chatId for WAHA
	chatId := number + "@c.us"

	// Initialize variables for API endpoint and payload
	var apiURL string
	var payload map[string]interface{}

	// Check file type based on extension (matching PHP logic exactly)
	if strings.Contains(mediaURL, ".mp4") {
		// VIDEO - use sendVideo endpoint
		apiURL = "https://waha-plus-production-705f.up.railway.app/api/sendVideo"
		payload = map[string]interface{}{
			"session": instance,
			"chatId":  chatId,
			"file": map[string]interface{}{
				"mimetype": "video/mp4",
				"url":      mediaURL,
				"filename": "Video",
			},
			"caption": nil, // Can add caption if needed
		}
	} else if strings.Contains(mediaURL, ".mp3") {
		// AUDIO - use sendFile endpoint (matching PHP)
		apiURL = "https://waha-plus-production-705f.up.railway.app/api/sendFile"
		payload = map[string]interface{}{
			"session": instance,
			"chatId":  chatId,
			"file": map[string]interface{}{
				"mimetype": "audio/mp3",
				"url":      mediaURL,
				"filename": "Audio",
			},
			"caption": nil,
		}
	} else {
		// IMAGE or other - determine mimetype from extension
		// Parse URL to get extension
		parsedURL, _ := url.Parse(mediaURL)
		path := parsedURL.Path
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))

		// Mimetype map (matching PHP exactly)
		mimeMap := map[string]string{
			"jpg":  "image/jpeg",
			"jpeg": "image/jpeg",
			"png":  "image/png",
			"gif":  "image/gif",
			"webp": "image/webp",
			"bmp":  "image/bmp",
			"svg":  "image/svg+xml",
		}

		// Step 1: Try using extension
		mimetype := ""
		if ext != "" {
			if mime, ok := mimeMap[ext]; ok {
				mimetype = mime
			}
		}

		// Step 2: If no extension match, try detecting from HTTP headers
		if mimetype == "" {
			// Try HEAD request to get Content-Type
			headReq, _ := http.NewRequest("HEAD", mediaURL, nil)
			headResp, err := ps.httpClient.Do(headReq)
			if err == nil && headResp != nil {
				defer headResp.Body.Close()
				contentType := headResp.Header.Get("Content-Type")
				if contentType != "" {
					// Extract mime type (remove charset etc)
					if idx := strings.Index(contentType, ";"); idx > 0 {
						contentType = contentType[:idx]
					}
					mimetype = strings.TrimSpace(contentType)
				}
			}
		}

		// Step 3: Fallback to default
		if mimetype == "" {
			mimetype = "image/jpeg"
		}

		// Use sendImage endpoint for images
		apiURL = "https://waha-plus-production-705f.up.railway.app/api/sendImage"
		payload = map[string]interface{}{
			"session": instance,
			"chatId":  chatId,
			"file": map[string]interface{}{
				"mimetype": mimetype,
				"url":      mediaURL,
				"filename": "Image",
			},
			"caption": nil,
		}
	}

	// Log the request details
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"session":      instance,
		"chatId":       chatId,
		"media_url":    mediaURL,
		"phone_number": phoneNumber,
		"payload":      payload,
	}).Info("📤 WAHA MEDIA: Sending media message")

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (matching PHP exactly)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	// Send request
	resp, err := ps.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logrus.WithFields(logrus.Fields{
			"status_code":   resp.StatusCode,
			"response_body": string(body),
			"api_url":       apiURL,
			"media_url":     mediaURL,
		}).Error("❌ WAHA MEDIA: Failed to send media")
		return fmt.Errorf("WAHA API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"status_code":   resp.StatusCode,
		"response_body": string(body),
		"media_url":     mediaURL,
	}).Info("✅ WAHA MEDIA: Media sent successfully")

	return nil
}
