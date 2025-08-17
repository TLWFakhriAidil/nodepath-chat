package platform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/antipattern"
)

// PlatformSender handles sending messages via external platforms
type PlatformSender struct {
	client            *http.Client
	messageRandomizer *antipattern.MessageRandomizer
	greetingProcessor *antipattern.GreetingProcessor
}

// NewPlatformSender creates a new platform sender
func NewPlatformSender() *PlatformSender {
	return &PlatformSender{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		messageRandomizer: antipattern.NewMessageRandomizer(),
		greetingProcessor: antipattern.NewGreetingProcessor(),
	}
}

// SendMessage sends a message via external platform
// NOTE: Anti-spam is already applied by BroadcastWorker - we just send raw content
func (ps *PlatformSender) SendMessage(platform, instance, phone, recipientName, message, imageURL, deviceID string) error {
	// NO ANTI-SPAM HERE - Already handled by BroadcastWorker
	
	_ = time.Now() // startTime - for future use
	var err error
	
	switch platform {
	case "Wablas":
		err = ps.sendViaWablas(instance, phone, message, imageURL)
	case "Whacenter":
		err = ps.sendViaWhacenter(instance, phone, message, imageURL)
	default:
		err = fmt.Errorf("unknown platform: %s", platform)
	}
	
	if err != nil {
		return err
	}
	
	return nil
}

// sendViaWablas sends message via Wablas API
func (ps *PlatformSender) sendViaWablas(token, phone, message, imageURL string) error {
	if imageURL != "" {
		// Send image with caption
		return ps.sendWablasImage(token, phone, message, imageURL)
	}
	// Send text only
	return ps.sendWablasText(token, phone, message)
}

// sendWablasText sends text message via Wablas
func (ps *PlatformSender) sendWablasText(token, phone, message string) error {
	apiURL := "https://my.wablas.com/api/send-message"
	
	// Prepare form data
	data := url.Values{}
	data.Set("phone", phone)
	data.Set("message", message)
	
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// Send request
	resp, err := ps.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Check status
	if status, ok := result["status"].(bool); ok && !status {
		if msg, ok := result["message"].(string); ok {
			return fmt.Errorf("wablas error: %s", msg)
		}
		return fmt.Errorf("wablas returned false status: %s", string(body))
	}
	
	return nil
}

// sendWablasImage sends image with caption via Wablas
func (ps *PlatformSender) sendWablasImage(token, phone, caption, imageURL string) error {
	apiURL := "https://my.wablas.com/api/send-image"
	
	// Prepare form data
	data := url.Values{}
	data.Set("phone", phone)
	data.Set("image", imageURL)
	data.Set("caption", caption)
	
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// Send request
	resp, err := ps.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Check status
	if status, ok := result["status"].(bool); ok && !status {
		if msg, ok := result["message"].(string); ok {
			return fmt.Errorf("wablas error: %s", msg)
		}
		return fmt.Errorf("wablas returned false status")
	}
	
	return nil
}

// sendViaWhacenter sends message via Whacenter API
func (ps *PlatformSender) sendViaWhacenter(deviceID, phone, message, imageURL string) error {
	apiURL := "https://app.whacenter.com/api/send"
	
	// Format phone number (remove 62 prefix if exists, add it back)
	phone = strings.TrimPrefix(phone, "62")
	phone = strings.TrimPrefix(phone, "+62")
	phone = "62" + phone
	
	// Create form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add fields
	writer.WriteField("device_id", deviceID)
	writer.WriteField("number", phone)
	writer.WriteField("message", message)
	
	// Add image if provided
	if imageURL != "" {
		writer.WriteField("file", imageURL)
	}
	
	writer.Close()
	
	// Create request
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Send request
	resp, err := ps.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w, body: %s", err, string(respBody))
	}
	
	// Check status
	if status, ok := result["status"].(bool); ok && !status {
		if msg, ok := result["msg"].(string); ok {
			return fmt.Errorf("whacenter error: %s", msg)
		}
		return fmt.Errorf("whacenter returned false status")
	}
	
	return nil
}

// truncateString truncates a string to specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}