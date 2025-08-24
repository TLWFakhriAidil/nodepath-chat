package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookPayload represents the webhook payload structure for WhatsApp providers
type WebhookPayload struct {
	Message     string `json:"message"`
	From        string `json:"from"`
	To          string `json:"to"`
	MessageType string `json:"message_type"`
	IsGroup     bool   `json:"is_group"`
}

func main() {
	// Test webhook payload
	payload := WebhookPayload{
		Message:     "Hello, I need help with my health concerns",
		From:        "60179645043",
		To:          "device",
		MessageType: "text",
		IsGroup:     false,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ Failed to marshal JSON: %v\n", err)
		return
	}

	fmt.Printf("📤 Sending webhook test message...\n")
	fmt.Printf("📱 Device: FakhriAidilTLW-001\n")
	fmt.Printf("📞 From: %s\n", payload.From)
	fmt.Printf("💬 Message: %s\n", payload.Message)
	fmt.Println()

	// Send HTTP POST request to webhook endpoint
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Use the correct webhook endpoint format: /api/webhook/{id_device}/{instance}
	// For testing, we'll use a dummy instance value
	webhookURL := "http://localhost:8080/api/webhook/FakhriAidilTLW-001/test-instance"
	
	resp, err := client.Post(
		webhookURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("❌ Failed to send request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("📥 Response Status: %s\n", resp.Status)
	fmt.Printf("📄 Response Body: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✅ Webhook test completed successfully!")
		fmt.Println("🔍 Check the server logs to see if flow-based AI prompt was triggered.")
	} else {
		fmt.Printf("⚠️ Webhook returned status: %d\n", resp.StatusCode)
	}
}