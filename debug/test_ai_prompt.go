package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookData represents the webhook payload structure
type WebhookData struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
	PushName    string `json:"pushName"`
	Source      string `json:"source"`
	Timestamp   string `json:"timestamp"`
	IsGroup     bool   `json:"is_group"`
	IDGroup     string `json:"id_group"`
	Media       string `json:"media"`
	AdReply     map[string]interface{} `json:"ad_reply"`
}

func main() {
	fmt.Println("=== Testing AI Prompt Node Execution ===")

	// Create webhook payload to test AI prompt
	webhookData := WebhookData{
		From:        "60179645043",
		To:          "60179645043",
		Message:     "Test AI prompt execution",
		MessageType: "text",
		PushName:    "Test User",
		Source:      "WHACENTER",
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		IsGroup:     false,
		IDGroup:     "",
		Media:       "",
		AdReply: map[string]interface{}{
			"source_id":   nil,
			"source_type": nil,
			"source_url":  nil,
		},
	}

	// Convert to JSON
	payload, err := json.Marshal(webhookData)
	if err != nil {
		fmt.Printf("❌ Error marshaling webhook data: %v\n", err)
		return
	}

	fmt.Printf("📤 Sending webhook to test AI prompt node...\n")
	fmt.Printf("📱 Phone: %s\n", webhookData.From)
	fmt.Printf("💬 Message: %s\n", webhookData.Message)
	fmt.Printf("🔧 Device: FakhriAidilTLW-001\n")

	// Send webhook request to correct endpoint format
	resp, err := http.Post(
		"http://localhost:8080/api/webhook/FakhriAidilTLW-001/65ec33c0-7574-46db-839f-4eadde18008a",
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		fmt.Printf("❌ Error sending webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Webhook sent successfully! Status: %s\n", resp.Status)
	fmt.Println("📋 Check server logs to see if AI prompt node executes properly.")
	fmt.Println("🔍 Look for logs containing 'AI_PROMPT' to track execution.")
}