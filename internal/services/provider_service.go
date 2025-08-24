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
	if deviceSettings == nil {
		return fmt.Errorf("device settings cannot be nil")
	}

	// Get provider from device settings
	provider := strings.ToLower(deviceSettings.Provider)
	logrus.WithFields(logrus.Fields{
		"provider":     provider,
		"device_id":    deviceSettings.IDDevice.String,
		"phone_number": phoneNumber,
		"media_url":    mediaURL,
	}).Info("📤 MEDIA: Sending media message through provider")

	switch provider {
	case "wablas":
		return ps.sendWablasImageMessage(deviceSettings, phoneNumber, caption, mediaURL)
	case "whacenter":
		return ps.sendWhacenterMediaMessage(deviceSettings, phoneNumber, caption, mediaURL)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// sendWablasMessage sends a text message via Wablas API
func (ps *ProviderService) sendWablasMessage(deviceSettings *models.DeviceSettings, phoneNumber, message string) error {
	apiURL := "https://my.wablas.com/api/send-message"
	
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"message_len":  len(message),
	}).Debug("[WABLAS-TEXT] Preparing request")

	// Prepare form data
	data := url.Values{}
	data.Set("phone", phoneNumber)
	data.Set("message", message)

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if deviceSettings.APIKey.Valid {
		req.Header.Set("Authorization", deviceSettings.APIKey.String)
	} else if deviceSettings.Instance.Valid {
		req.Header.Set("Authorization", deviceSettings.Instance.String)
	} else {
		return fmt.Errorf("no API key or instance found for Wablas")
	}
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

// sendWablasImageMessage sends an image message via Wablas API
func (ps *ProviderService) sendWablasImageMessage(deviceSettings *models.DeviceSettings, phoneNumber, caption, imageURL string) error {
	apiURL := "https://my.wablas.com/api/send-image"
	
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"image_url":    imageURL,
		"caption_len":  len(caption),
	}).Debug("[WABLAS-IMAGE] Preparing request")

	// Prepare form data
	data := url.Values{}
	data.Set("phone", phoneNumber)
	data.Set("image", imageURL)
	data.Set("caption", caption)

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if deviceSettings.APIKey.Valid {
		req.Header.Set("Authorization", deviceSettings.APIKey.String)
	} else if deviceSettings.Instance.Valid {
		req.Header.Set("Authorization", deviceSettings.Instance.String)
	} else {
		return fmt.Errorf("no API key or instance found for Wablas")
	}
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
	}).Debug("[WABLAS-IMAGE] Response received")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wablas API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
	}).Info("[WABLAS-IMAGE] ✅ Image sent successfully")

	return nil
}

// sendWhacenterMessage sends a text message via Whacenter API
func (ps *ProviderService) sendWhacenterMessage(deviceSettings *models.DeviceSettings, phoneNumber, message string) error {
	apiURL := "https://api.whacenter.com/api/send"
	
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"message_len":  len(message),
	}).Debug("[WHACENTER] Preparing request")

	// Get device ID from instance or device_id
	deviceID := ""
	if deviceSettings.Instance.Valid {
		deviceID = deviceSettings.Instance.String
	} else if deviceSettings.IDDevice.Valid {
		deviceID = deviceSettings.IDDevice.String
	} else {
		return fmt.Errorf("no device ID found for Whacenter")
	}

	// Prepare JSON payload
	payload := map[string]interface{}{
		"device_id": deviceID,
		"number":    phoneNumber,
		"message":   message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if deviceSettings.APIKey.Valid {
		req.Header.Set("Authorization", "Bearer "+deviceSettings.APIKey.String)
	}

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
	apiURL := "https://api.whacenter.com/api/send"
	
	logrus.WithFields(logrus.Fields{
		"api_url":      apiURL,
		"phone_number": phoneNumber,
		"media_url":    mediaURL,
		"caption_len":  len(caption),
	}).Debug("[WHACENTER] Preparing media request")

	// Get device ID from instance or device_id
	deviceID := ""
	if deviceSettings.Instance.Valid {
		deviceID = deviceSettings.Instance.String
	} else if deviceSettings.IDDevice.Valid {
		deviceID = deviceSettings.IDDevice.String
	} else {
		return fmt.Errorf("no device ID found for Whacenter")
	}

	// Prepare JSON payload for media message
	payload := map[string]interface{}{
		"device_id": deviceID,
		"number":    phoneNumber,
		"message":   caption,
		"file":      mediaURL,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if deviceSettings.APIKey.Valid {
		req.Header.Set("Authorization", "Bearer "+deviceSettings.APIKey.String)
	}

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
	}).Debug("[WHACENTER] Media response received")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("whacenter API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"duration":     duration,
	}).Info("[WHACENTER] ✅ Media sent successfully")

	return nil
}