package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	data.Set("phone", phoneNumber)    // Recipient phone number
	data.Set("message", message)      // Message content

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers exactly as specified by user
	req.Header.Set("Authorization", instance)  // Set the Authorization header
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
	data.Set("phone", phoneNumber)        // Recipient phone number
	data.Set(fieldName, mediaURL)         // Media file URL with correct field name

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
	data.Set("device_id", instance)    // device_id from instance
	data.Set("number", phoneNumber)    // recipient number
	data.Set("message", message)       // message content

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
	data.Set("device_id", instance)    // device_id from instance
	data.Set("number", phoneNumber)    // recipient number
	data.Set("file", mediaURL)         // media file URL
	
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

	// Get API key from device settings
	apiKey := ""
	if deviceSettings.APIKey.Valid {
		apiKey = deviceSettings.APIKey.String
	} else {
		return fmt.Errorf("no API key found for WAHA device %s", deviceSettings.Instance.String)
	}

	// Get instance for session (as per user requirements)
	instance := ""
	if deviceSettings.Instance.Valid {
		instance = deviceSettings.Instance.String
	} else {
		return fmt.Errorf("no instance found for WAHA device %s", deviceSettings.Instance.String)
	}

	// WAHA API endpoint for sending text messages
	apiURL := "http://localhost:3000/api/sendText"
	
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"message_len":  len(message),
		"device_id":    deviceSettings.Instance.String,
	}).Debug("[WAHA-TEXT] Preparing request")

	// Format phone number for WAHA (international format without + and add @c.us)
	chatId := phoneNumber
	if !strings.HasSuffix(chatId, "@c.us") {
		// Remove + if present and add @c.us
		chatId = strings.TrimPrefix(chatId, "+") + "@c.us"
	}

	// Prepare JSON payload as per WAHA API documentation
	payload := map[string]interface{}{
		"session": instance,    // Session name from instance
		"chatId":  chatId,      // Phone number in WAHA format
		"text":    message,     // Message content
	}

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

	// Set headers as per WAHA API documentation
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Api-Key", apiKey)  // API key for authentication

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
	}).Debug("[WAHA-TEXT] Response received")

	// Check for success (200-299 status codes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WAHA API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
		"device_id":    deviceSettings.Instance.String,
	}).Info("[WAHA-TEXT] ✅ Message sent successfully")

	return nil
}

// sendWahaMediaMessage sends a media message via WAHA API
// Handles video, audio, and image files with appropriate API endpoints
func (ps *ProviderService) sendWahaMediaMessage(deviceSettings *models.DeviceSettings, phoneNumber, mediaURL string) error {
	// Get API key from device settings
	apiKey := ""
	if deviceSettings.APIKey.Valid {
		apiKey = deviceSettings.APIKey.String
	} else {
		return fmt.Errorf("no API key found for WAHA device %s", deviceSettings.Instance.String)
	}

	// Get instance for session (as per user requirements)
	instance := ""
	if deviceSettings.Instance.Valid {
		instance = deviceSettings.Instance.String
	} else {
		return fmt.Errorf("no instance found for WAHA device %s", deviceSettings.Instance.String)
	}

	// Detect media type and set appropriate API endpoint
	mediaType := ""
	var apiURL string
	
	if strings.Contains(mediaURL, ".mp4") {
		mediaType = "video"
		apiURL = "http://localhost:3000/api/sendVideo"
	} else if strings.Contains(mediaURL, ".mp3") {
		mediaType = "audio"
		apiURL = "http://localhost:3000/api/sendAudio"
	} else {
		// Default to image for all other file types
		mediaType = "image"
		apiURL = "http://localhost:3000/api/sendImage"
	}
	
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"media_url":    mediaURL,
		"media_type":   mediaType,
		"device_id":    deviceSettings.Instance.String,
	}).Debug("[WAHA-MEDIA] Preparing request")

	// Format phone number for WAHA (international format without + and add @c.us)
	chatId := phoneNumber
	if !strings.HasSuffix(chatId, "@c.us") {
		// Remove + if present and add @c.us
		chatId = strings.TrimPrefix(chatId, "+") + "@c.us"
	}

	// Prepare JSON payload as per WAHA API documentation
	payload := map[string]interface{}{
		"session": instance,    // Session name from instance
		"chatId":  chatId,      // Phone number in WAHA format
		"file": map[string]interface{}{
			"url": mediaURL,    // Media file URL
		},
	}

	// Add caption for videos if needed
	if mediaType == "video" {
		payload["caption"] = ""
		payload["asNote"] = false
		payload["convert"] = false
	}

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

	// Set headers as per WAHA API documentation
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Api-Key", apiKey)  // API key for authentication

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
	}).Debug("[WAHA-MEDIA] Response received")

	// Check for success (200-299 status codes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WAHA API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
		"device_id":    deviceSettings.Instance.String,
		"media_type":   mediaType,
	}).Info("[WAHA-MEDIA] ✅ Media sent successfully")

	return nil
}