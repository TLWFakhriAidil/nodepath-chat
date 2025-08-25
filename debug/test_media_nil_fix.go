package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Test webhook payload structure - matches webhook handler expectations
type WebhookPayload struct {
	From        string `json:"from"`        // Phone number field expected by webhook
	Message     string `json:"message"`     // Message content
	MessageType string `json:"message_type"` // Message type (text, image, etc.)
	IsGroup     bool   `json:"is_group"`    // Whether it's a group message
	Timestamp   int64  `json:"timestamp"`   // Message timestamp
}

func main() {
	// Test payload with a message that should trigger Advanced AI Prompt node with media response
	// This simulates a user input that would generate a JSON response with media
	payload := WebhookPayload{
		From:        "601137508067", // Use the test phone number from custom instructions
		Message:     "show me an image", // This should trigger AI response with media
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

	// Send webhook request
	url := "http://localhost:8080/api/webhook/FakhriAidilTLW-001/65ec33c0-7574-46db-839f-4eadde18008a"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Webhook sent successfully! Status: %s\n", resp.Status)
	fmt.Printf("Testing media response to verify no <nil> message is sent...\n")
	fmt.Printf("Check server logs to confirm:")
	fmt.Printf("1. Advanced AI Prompt node processes the request\n")
	fmt.Printf("2. JSON response is parsed successfully\n")
	fmt.Printf("3. Media items are sent individually\n")
	fmt.Printf("4. No additional <nil> message is sent\n")
	fmt.Printf("5. Look for log: 'Skipping empty message to prevent <nil> message'\n")

	// Wait a moment for processing
	time.Sleep(2 * time.Second)
	fmt.Printf("\nTest completed. Check server logs for confirmation.\n")
}