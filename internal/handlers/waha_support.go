package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"nodepath-chat/internal/models"
	"github.com/sirupsen/logrus"
)

// sendWahaTextMessage sends text message via WAHA API
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

	// Create HTTP request
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
		logrus.WithError(err).Error("❌ WAHA: Failed to send text message")
		return
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
			"to":     to,
		}).Error("❌ WAHA: Failed to send text message")
		return
	}

	logrus.WithFields(logrus.Fields{
		"to":      to,
		"session": sessionName,
		"message": truncateString(message, 50),
	}).Info("✅ WAHA: Text message sent successfully")
}

// sendWahaImageMessage sends image message via WAHA API
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

	// Create HTTP request
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
		logrus.WithError(err).Error("❌ WAHA: Failed to send image message")
		return
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
			"to":     to,
			"image":  imageURL,
		}).Error("❌ WAHA: Failed to send image message")
		return
	}

	logrus.WithFields(logrus.Fields{
		"to":      to,
		"session": sessionName,
		"image":   imageURL,
	}).Info("✅ WAHA: Image message sent successfully")
}