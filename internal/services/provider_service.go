package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// ProviderService handles message sending through external providers (Wablas, Whacenter)
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
		"device_id":    deviceSettings.IDDevice.String,
		"phone_number": phoneNumber,
	}).Info("📤 MESSAGE: Sending message through provider")

	switch provider {
	case "wablas":
		return ps.sendWablasMessage(deviceSettings, phoneNumber, message)
	case "whacenter":
		return ps.sendWhacenterMessage(deviceSettings, phoneNumber, message)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// SendMediaMessage sends a media message through the appropriate provider
func (ps *ProviderService) SendMediaMessage(deviceSettings *models.DeviceSettings, phoneNumber, caption, mediaURL string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"caption": caption,
		"media_url": mediaURL,
	}).Info("📤 PROVIDER: [TRACE] Starting SendMediaMessage")
	
	if deviceSettings == nil {
		logrus.Error("❌ PROVIDER: [TRACE] Device settings is nil")
		return fmt.Errorf("device settings cannot be nil")
	}

	// Get provider from device settings
	provider := strings.ToLower(deviceSettings.Provider)
	logrus.WithFields(logrus.Fields{
		"provider":     provider,
		"device_id":    deviceSettings.IDDevice.String,
		"phone_number": phoneNumber,
		"media_url":    mediaURL,
		"caption": caption,
		"instance_valid": deviceSettings.Instance.Valid,
		"api_key_valid": deviceSettings.APIKey.Valid,
	}).Info("📤 PROVIDER: [TRACE] Device settings and provider info")

	switch provider {
	case "wablas":
		logrus.WithFields(logrus.Fields{
			"provider": "wablas",
			"media_url": mediaURL,
		}).Info("📤 PROVIDER: [TRACE] Routing to Wablas media service")
		return ps.sendWablasImageMessage(deviceSettings, phoneNumber, caption, mediaURL)
	case "whacenter":
		logrus.WithFields(logrus.Fields{
			"provider": "whacenter",
			"media_url": mediaURL,
		}).Info("📤 PROVIDER: [TRACE] Routing to Whacenter media service")
		return ps.sendWhacenterMediaMessage(deviceSettings, phoneNumber, caption, mediaURL)
	default:
		logrus.WithFields(logrus.Fields{
			"unsupported_provider": provider,
			"media_url": mediaURL,
		}).Error("❌ PROVIDER: [TRACE] Unsupported provider")
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// sendWablasMessage sends a text message via Wablas API
// Updated to match PHP implementation - only use instance for authorization
func (ps *ProviderService) sendWablasMessage(deviceSettings *models.DeviceSettings, phoneNumber, message string) error {
	if !deviceSettings.Instance.Valid {
		return fmt.Errorf("no instance found for Wablas")
	}

	apiURL := "https://my.wablas.com/api/send-message"
	
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"message_len":  len(message),
	}).Debug("[WABLAS-TEXT] Preparing request")

	// Prepare form data to match PHP implementation
	data := url.Values{}
	data.Set("phone", phoneNumber)
	data.Set("message", message)

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers as per PHP implementation
	req.Header.Set("Authorization", deviceSettings.Instance.String)
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
	}).Debug("[WABLAS-TEXT] Response received")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wablas API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
	}).Info("[WABLAS-TEXT] ✅ Message sent successfully")

	return nil
}

// sendWablasImageMessage sends a media message via Wablas API
// Enhanced to support different media types like the handlers implementation
func (ps *ProviderService) sendWablasImageMessage(deviceSettings *models.DeviceSettings, phoneNumber, caption, mediaURL string) error {
	if !deviceSettings.Instance.Valid {
		return fmt.Errorf("no instance found for Wablas media message")
	}

	// Determine file type and corresponding API endpoint
	fileType := ps.getFileTypeFromURL(mediaURL)
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
	
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"media_url":    mediaURL,
		"file_type":    fileType,
		"caption_len":  len(caption),
	}).Debug("[WABLAS-MEDIA] Preparing request")

	// Prepare form data using correct field names
	data := url.Values{}
	data.Set("phone", phoneNumber)
	data.Set(fieldName, mediaURL) // Use dynamic field name based on media type
	if caption != "" {
		data.Set("caption", caption)
	}

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - Wablas uses instance for authorization
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", deviceSettings.Instance.String)

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
		"file_type":   fileType,
	}).Debug("[WABLAS-MEDIA] Response received")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wablas API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
		"file_type":    fileType,
	}).Info("[WABLAS-MEDIA] ✅ Media sent successfully")

	return nil
}

// getFileTypeFromURL determines file type from URL extension
func (ps *ProviderService) getFileTypeFromURL(fileURL string) string {
	fileURL = strings.ToLower(fileURL)
	if strings.Contains(fileURL, ".mp4") || strings.Contains(fileURL, ".avi") || strings.Contains(fileURL, ".mov") {
		return "video"
	}
	if strings.Contains(fileURL, ".mp3") || strings.Contains(fileURL, ".wav") || strings.Contains(fileURL, ".ogg") {
		return "audio"
	}
	return "image" // default to image
}

// sendWhacenterMessage sends a text message via Whacenter API
// Updated to match PHP implementation - use form data instead of JSON
func (ps *ProviderService) sendWhacenterMessage(deviceSettings *models.DeviceSettings, phoneNumber, message string) error {
	if !deviceSettings.Instance.Valid {
		return fmt.Errorf("no instance found for Whacenter")
	}

	apiURL := "https://api.whacenter.com/api/send"
	
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"message_len":  len(message),
	}).Debug("[WHACENTER] Preparing request")

	// Prepare form data to match PHP implementation
	data := url.Values{}
	data.Set("device_id", deviceSettings.Instance.String)
	data.Set("number", phoneNumber)
	data.Set("message", message)

	// Create request with form data
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers as per PHP implementation (no Authorization header)
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
	}).Debug("[WHACENTER] Response received")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("whacenter API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
	}).Info("[WHACENTER] ✅ Message sent successfully")

	return nil
}

// sendWhacenterMediaMessage sends a media message via Whacenter API
func (ps *ProviderService) sendWhacenterMediaMessage(deviceSettings *models.DeviceSettings, phoneNumber, caption, mediaURL string) error {
	// Use the correct Whacenter media API endpoint - using /api/send as per PHP implementation
	apiURL := "https://api.whacenter.com/api/send"
	
	logrus.WithFields(logrus.Fields{
		"api_url": apiURL,
		"phone_number": phoneNumber,
		"media_url": mediaURL,
		"caption": caption,
	}).Info("📤 WHACENTER: [TRACE] Starting sendWhacenterMediaMessage")
	
	// Determine file type from URL for proper media type detection
	fileType := ps.getFileTypeFromURL(mediaURL)
	
	logrus.WithFields(logrus.Fields{
		"media_url": mediaURL,
		"detected_file_type": fileType,
	}).Info("📤 WHACENTER: [TRACE] File type detection result")

	// Get device ID from instance (Whacenter uses instance as device_id)
	if !deviceSettings.Instance.Valid {
		logrus.WithField("device_id", deviceSettings.IDDevice.String).Error("❌ WHACENTER: [TRACE] No valid instance found")
		return fmt.Errorf("no instance found for Whacenter media message")
	}

	logrus.WithFields(logrus.Fields{
		"device_id": deviceSettings.IDDevice.String,
		"instance": deviceSettings.Instance.String,
		"instance_valid": deviceSettings.Instance.Valid,
	}).Info("📤 WHACENTER: [TRACE] Using instance as device_id")

	// Prepare form data payload to match PHP implementation
	formData := url.Values{}
	formData.Set("device_id", deviceSettings.Instance.String) // Use instance as device_id
	formData.Set("number", phoneNumber)
	formData.Set("file", mediaURL) // Use 'file' parameter as per PHP code
	formData.Set("type", fileType)  // Use detected file type

	// Add message/caption if provided
	if caption != "" {
		formData.Set("message", caption)
		logrus.WithField("caption", caption).Info("📤 WHACENTER: [TRACE] Added caption to payload")
	}

	logrus.WithFields(logrus.Fields{
		"form_data": formData,
	}).Info("📤 WHACENTER: [TRACE] Prepared form data payload")

	// Create request with form data
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: [TRACE] Failed to create HTTP request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for form data - no Authorization header as per PHP implementation
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	logrus.WithFields(logrus.Fields{
		"content_type": "application/x-www-form-urlencoded",
		"method": "POST",
		"url": apiURL,
	}).Info("📤 WHACENTER: [TRACE] HTTP request prepared with headers")

	// Send request
	logrus.Info("📤 WHACENTER: [TRACE] Sending HTTP request to Whacenter API")
	startTime := time.Now()
	resp, err := ps.httpClient.Do(req)
	if err != nil {
		logrus.WithError(err).WithField("api_url", apiURL).Error("❌ WHACENTER: [TRACE] HTTP request failed")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"request_duration": time.Since(startTime),
	}).Info("📤 WHACENTER: [TRACE] HTTP response received")

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("❌ WHACENTER: [TRACE] Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	duration := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response_body": string(body),
		"response_size": len(body),
		"total_duration": duration,
	}).Info("📤 WHACENTER: [TRACE] Complete API response details")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response_body": string(body),
			"media_url": mediaURL,
			"file_type": fileType,
		}).Error("❌ WHACENTER: [TRACE] API returned error status")
		return fmt.Errorf("whacenter API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"media_url": mediaURL,
		"file_type": fileType,
		"duration": duration,
		"status_code": resp.StatusCode,
	}).Info("✅ WHACENTER: [TRACE] Media message sent successfully")

	return nil
}