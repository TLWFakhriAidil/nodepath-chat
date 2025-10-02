package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"nodepath-chat/internal/models"
)

// sendWahaTextMessage sends text message via WAHA API - NO RETRY
func (h *Handlers) sendWahaTextMessage(to, message string, deviceSettings *models.DeviceSettings) {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WAHA: No instance available")
		return
	}

	apiBase := "https://waha-plus-production-705f.up.railway.app"
	apiKey := "dckr_pat_vxeqEu_CqRi5O3CBHnD7FxhnBz0"
	sessionName := fmt.Sprintf("user_%s", deviceSettings.IDDevice.String)
	apiURL := fmt.Sprintf("%s/api/sendText", apiBase)

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

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to create text request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	// Single attempt only - no retry
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Network error sending text")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Read body for error logging
		body, _ := io.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"error":  string(body),
			"to":     to,
		}).Error("❌ WAHA: Failed to send text")
		return
	}

	logrus.WithFields(logrus.Fields{
		"to":      to,
		"session": sessionName,
	}).Info("✅ WAHA: Text sent")
}

// sendWahaImageMessage sends image message via WAHA API - NO RETRY
func (h *Handlers) sendWahaImageMessage(to, imageURL string, deviceSettings *models.DeviceSettings) {
	if !deviceSettings.Instance.Valid {
		logrus.Error("❌ WAHA: No instance available")
		return
	}

	apiBase := "https://waha-plus-production-705f.up.railway.app"
	apiKey := "dckr_pat_vxeqEu_CqRi5O3CBHnD7FxhnBz0"
	sessionName := fmt.Sprintf("user_%s", deviceSettings.IDDevice.String)
	apiURL := fmt.Sprintf("%s/api/sendImage", apiBase)

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

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Failed to create image request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	// Single attempt only - no retry
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("❌ WAHA: Network error sending image")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Read body for error logging
		body, _ := io.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"error":  string(body),
			"to":     to,
		}).Error("❌ WAHA: Failed to send image")
		return
	}

	logrus.WithFields(logrus.Fields{
		"to":      to,
		"session": sessionName,
	}).Info("✅ WAHA: Image sent")
}
