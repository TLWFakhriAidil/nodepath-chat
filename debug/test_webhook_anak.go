package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// WebhookPayload represents the structure of the webhook request
type WebhookPayload struct {
	From    string `json:"from"`
	Message string `json:"message"`
	Device  string `json:"device"`
}

func main() {
	fmt.Println("=== Testing Webhook with 'anak' ===")
	fmt.Println("Sending webhook to test condition evaluation...")
	fmt.Println()

	// Create webhook payload with "anak" message
	payload := WebhookPayload{
		From:    "60179645043",
		Message: "anak",
		Device:  "FakhriAidilTLW-001",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Printf("Payload: %s\n", string(jsonData))
	fmt.Println()

	// Send POST request to correct webhook endpoint with device ID and instance
	webhookURL := "http://localhost:8080/api/webhook/FakhriAidilTLW-001/65ec33c0-7574-46db-839f-4eadde18008a"
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(body))
	fmt.Println()
	fmt.Println("✅ Webhook sent successfully!")
	fmt.Println("Check server logs to see if 'anak' triggers the correct response.")
}