package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookPayload represents the webhook message structure for Whacenter provider
type WebhookPayload struct {
	From        string `json:"from"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
	IsGroup     bool   `json:"is_group"`
	Timestamp   int64  `json:"timestamp"`
}

func main() {
	// Test webhook payload - simulating user replying "anak" from Whacenter provider
	payload := WebhookPayload{
		From:        "601137508067",
		Message:     "anak",
		MessageType: "text",
		IsGroup:     false,
		Timestamp:   time.Now().Unix(),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	// Send webhook request to correct endpoint with device ID and instance
	webhookURL := "http://localhost:8080/api/webhook/FakhriAidilTLW-001/65ec33c0-7574-46db-839f-4eadde18008a"
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Webhook sent successfully. Status: %s\n", resp.Status)

	// Read response body
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	fmt.Printf("Response: %s\n", buf.String())
}