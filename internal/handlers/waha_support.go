package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"nodepath-chat/internal/models"
	"github.com/sirupsen/logrus"
)

// sendWahaTextMessage sends text message via WAHA API with retry logic
func (h *Handlers) sendWahaTextMessage(to, message string, deviceSettings *models.DeviceSettings) {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WAHA: No instance available")
		return
	}

	// WAHA API configuration
	apiBase := "https://waha-plus-production-705f.up.railway.app"
	apiKey := "dckr_pat_vxeqEu_CqRi5O3CBHnD7FxhnBz0"
	sessionName := fmt.Sprintf("user_%s", deviceSettings.IDDevice)

	// WAHA API endpoint for sending text messages
	apiURL := fmt.Sprintf("%s/api/sendText", apiBase)

	// Prepare request payload
	payload := map[string]interface{}{
		"session": sessionName,
		"chatId":  to + "@c.us",
		"text":    message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to marshal text payload")
		return
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	baseDelay := 2 * time.Second
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Create fresh request for each attempt
		req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payloadBytes))
		if err != nil {
			logrus.WithError(err).Error("❌ WAHA: Failed to create text request")
			return
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", apiKey)

		// Send request
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		
		if err != nil {
			if attempt < maxRetries {
				delay := time.Duration(attempt) * baseDelay
				logrus.WithFields(logrus.Fields{
					"attempt":    attempt,
					"next_retry": delay,
					"error":      err.Error(),
				}).Warn("⚠️ WAHA: Network error, retrying...")
				time.Sleep(delay)
				continue
			}
			logrus.WithError(err).Error("❌ WAHA: Network error after all retries")
			return
		}
		
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		
		// Check if successful
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			logrus.WithFields(logrus.Fields{
				"to":      to,
				"session": sessionName,
				"attempt": attempt,
			}).Info("✅ WAHA: Text message sent successfully")
			return
		}
		
		// Handle specific websocket errors
		if resp.StatusCode == http.StatusInternalServerError {
			bodyStr := string(body)
			if strings.Contains(bodyStr, "websocket not connected") ||
			   strings.Contains(bodyStr, "failed to get device list") ||
			   strings.Contains(bodyStr, "failed to send usync query") {
				
				if attempt < maxRetries {
					// Longer delay for websocket issues
					delay := time.Duration(attempt) * 5 * time.Second
					logrus.WithFields(logrus.Fields{
						"attempt":    attempt,
						"next_retry": delay,
						"session":    sessionName,
					}).Warn("⚠️ WAHA: Websocket issue, waiting longer before retry...")
					time.Sleep(delay)
					continue
				}
			}
		}
		
		// Log error and retry if attempts remaining
		if attempt < maxRetries {
			delay := time.Duration(attempt) * 3 * time.Second
			logrus.WithFields(logrus.Fields{
				"attempt":    attempt,
				"status":     resp.StatusCode,
				"next_retry": delay,
			}).Warn("⚠️ WAHA: Server error, retrying...")
			time.Sleep(delay)
			continue
		}
		
		// Final failure
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
			"to":     to,
		}).Error("❌ WAHA: Failed to send text message after all retries")
	}
}

// sendWahaImageMessage sends image message via WAHA API with retry logic
func (h *Handlers) sendWahaImageMessage(to, imageURL string, deviceSettings *models.DeviceSettings) {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WAHA: No instance available")
		return
	}

	// WAHA API configuration
	apiBase := "https://waha-plus-production-705f.up.railway.app"
	apiKey := "dckr_pat_vxeqEu_CqRi5O3CBHnD7FxhnBz0"
	sessionName := fmt.Sprintf("user_%s", deviceSettings.IDDevice)

	// WAHA API endpoint for sending image messages
	apiURL := fmt.Sprintf("%s/api/sendImage", apiBase)

	// Prepare request payload
	payload := map[string]interface{}{
		"session": sessionName,
		"chatId":  to + "@c.us",
		"file": map[string]interface{}{
			"url": imageURL,
		},
		"caption": "",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to marshal image payload")
		return
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	baseDelay := 2 * time.Second
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Create fresh request for each attempt
		req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payloadBytes))
		if err != nil {
			logrus.WithError(err).Error("❌ WAHA: Failed to create image request")
			return
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", apiKey)

		// Send request
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		
		if err != nil {
			if attempt < maxRetries {
				delay := time.Duration(attempt) * baseDelay
				logrus.WithFields(logrus.Fields{
					"attempt":    attempt,
					"next_retry": delay,
					"error":      err.Error(),
				}).Warn("⚠️ WAHA: Network error on image send, retrying...")
				time.Sleep(delay)
				continue
			}
			logrus.WithError(err).Error("❌ WAHA: Network error on image after all retries")
			return
		}
		
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		
		// Check if successful
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			logrus.WithFields(logrus.Fields{
				"to":      to,
				"session": sessionName,
				"image":   imageURL,
				"attempt": attempt,
			}).Info("✅ WAHA: Image message sent successfully")
			return
		}
		
		// Handle specific websocket errors
		if resp.StatusCode == http.StatusInternalServerError {
			bodyStr := string(body)
			if strings.Contains(bodyStr, "websocket not connected") ||
			   strings.Contains(bodyStr, "failed to get device list") ||
			   strings.Contains(bodyStr, "failed to send usync query") {
				
				if attempt < maxRetries {
					// Longer delay for websocket issues
					delay := time.Duration(attempt) * 5 * time.Second
					logrus.WithFields(logrus.Fields{
						"attempt":    attempt,
						"next_retry": delay,
						"session":    sessionName,
					}).Warn("⚠️ WAHA: Websocket issue on image, waiting longer...")
					time.Sleep(delay)
					continue
				}
			}
		}
		
		// Log error and retry if attempts remaining
		if attempt < maxRetries {
			delay := time.Duration(attempt) * 3 * time.Second
			logrus.WithFields(logrus.Fields{
				"attempt":    attempt,
				"status":     resp.StatusCode,
				"next_retry": delay,
			}).Warn("⚠️ WAHA: Server error on image, retrying...")
			time.Sleep(delay)
			continue
		}
		
		// Final failure
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
			"to":     to,
			"image":  imageURL,
		}).Error("❌ WAHA: Failed to send image after all retries")
	}
}

// Helper function already exists in ai_whatsapp_handlers.go
// truncateString truncates a string to specified length