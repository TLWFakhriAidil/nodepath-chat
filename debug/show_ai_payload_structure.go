package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// AIWhatsappPayload represents the payload sent to AI API
type AIWhatsappPayload struct {
	Model             string                `json:"model"`
	Messages          []AIWhatsappMessage   `json:"messages"`
	Temperature       float64               `json:"temperature"`
	TopP              float64               `json:"top_p"`
	RepetitionPenalty float64               `json:"repetition_penalty"`
}

// AIWhatsappMessage represents a message in the AI conversation
type AIWhatsappMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	fmt.Println("=== AI PAYLOAD GENERATION AND DEBUG CONSOLE OUTPUT ===")
	fmt.Println("This shows how the AI generates replies and the debug information before generation.")
	fmt.Println()

	// Simulate the AI payload generation process
	deviceID := "FakhriAidilTLW-001"
	prospectNum := "601137508067"
	currentText := "Hello, I need help with my order"

	fmt.Printf("🔍 DEBUG: Processing message for device: %s\n", deviceID)
	fmt.Printf("🔍 DEBUG: Prospect number: %s\n", prospectNum)
	fmt.Printf("🔍 DEBUG: User input: %s\n", currentText)
	fmt.Println()

	// Step 1: Device Configuration (simulated)
	fmt.Println("📋 STEP 1: Device Configuration")
	fmt.Println("==============================")
	provider := "wablas"
	instance := "instance-001"
	apiKeyOption := "anthropic/claude-3.5-sonnet"
	apiKey := "sk-ant-api03-xxx...xxx"

	fmt.Printf("✅ Device settings loaded:\n")
	fmt.Printf("   - Provider: %s\n", provider)
	fmt.Printf("   - Instance: %s\n", instance)
	fmt.Printf("   - API Key Option (Model): %s\n", apiKeyOption)
	fmt.Printf("   - API Key: %s...\n", maskAPIKey(apiKey))
	fmt.Println()

	// Step 2: AI Settings (simulated)
	fmt.Println("🤖 STEP 2: AI Settings")
	fmt.Println("======================")
	behavePrompt := `You are Layla, a friendly and knowledgeable customer support assistant for an e-commerce platform. You help customers with their orders, product inquiries, and general support needs. Always be polite, helpful, and professional.`
	closingPrompt := `Always end your responses with helpful next steps and ask if there's anything else you can help with.`

	fmt.Printf("✅ AI settings loaded:\n")
	fmt.Printf("   - Behave Prompt Length: %d characters\n", len(behavePrompt))
	fmt.Printf("   - Closing Prompt Length: %d characters\n", len(closingPrompt))
	fmt.Printf("   - Behave Prompt Preview: %s...\n", getPreview(behavePrompt, 100))
	fmt.Println()

	// Step 3: Conversation Context (simulated)
	fmt.Println("💬 STEP 3: Conversation Context")
	fmt.Println("===============================")
	currentStage := "Problem Identification"
	humanTakeover := 0
	niche := "customer_support"
	lastAIResponse := "Hello! How can I help you today?"

	fmt.Printf("✅ Conversation context loaded:\n")
	fmt.Printf("   - Current Stage: %s\n", currentStage)
	fmt.Printf("   - Human Takeover: %d (0=AI active, 1=Human active)\n", humanTakeover)
	fmt.Printf("   - Niche: %s\n", niche)
	fmt.Printf("   - Last AI Response: %s\n", getPreview(lastAIResponse, 50))
	fmt.Println()

	// Step 4: Build AI Prompt Content
	fmt.Println("📝 STEP 4: Building AI Prompt Content")
	fmt.Println("=====================================")
	promptContent := buildAIPromptContent(behavePrompt, closingPrompt, currentStage)
	fmt.Printf("✅ AI prompt content built (Length: %d characters)\n", len(promptContent))
	fmt.Printf("   - Prompt Preview: %s...\n", getPreview(promptContent, 200))
	fmt.Println()

	// Step 5: API Configuration
	fmt.Println("🔧 STEP 5: API Configuration")
	fmt.Println("============================")
	apiURL := getAPIURL(deviceID)
	model := getAIModel(deviceID, apiKeyOption)
	finalAPIKey := getAPIKey(deviceID, apiKey)

	fmt.Printf("✅ API configuration determined:\n")
	fmt.Printf("   - API URL: %s\n", apiURL)
	fmt.Printf("   - Model: %s\n", model)
	fmt.Printf("   - API Key: %s...\n", maskAPIKey(finalAPIKey))
	fmt.Println()

	// Step 6: Create AI Payload
	fmt.Println("🚀 STEP 6: Creating AI Payload")
	fmt.Println("==============================")
	payload := AIWhatsappPayload{
		Model: model,
		Messages: []AIWhatsappMessage{
			{Role: "system", Content: promptContent},
			{Role: "assistant", Content: lastAIResponse},
			{Role: "user", Content: currentText},
		},
		Temperature:       0.67,
		TopP:              1.0,
		RepetitionPenalty: 1.0,
	}

	fmt.Println("✅ AI payload created successfully!")
	fmt.Println()

	// Step 7: Display Complete Payload Structure
	fmt.Println("📦 STEP 7: COMPLETE AI PAYLOAD STRUCTURE")
	fmt.Println("=========================================")

	// Convert payload to JSON for display
	payloadJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("❌ ERROR: Failed to marshal payload: %v\n", err)
		return
	}

	fmt.Println(string(payloadJSON))
	fmt.Println()

	// Step 8: Console Debug Output Before API Call
	fmt.Println("🔍 STEP 8: CONSOLE DEBUG OUTPUT BEFORE AI API CALL")
	fmt.Println("===================================================")
	fmt.Printf("[DEBUG] Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("[DEBUG] Device ID: %s\n", deviceID)
	fmt.Printf("[DEBUG] Prospect Number: %s\n", prospectNum)
	fmt.Printf("[DEBUG] Current Stage: %s\n", currentStage)
	fmt.Printf("[DEBUG] Human Takeover: %d\n", humanTakeover)
	fmt.Printf("[DEBUG] API Provider: %s\n", getProviderName(deviceID))
	fmt.Printf("[DEBUG] API URL: %s\n", apiURL)
	fmt.Printf("[DEBUG] Model: %s\n", model)
	fmt.Printf("[DEBUG] Temperature: %.2f\n", payload.Temperature)
	fmt.Printf("[DEBUG] Top P: %.1f\n", payload.TopP)
	fmt.Printf("[DEBUG] Repetition Penalty: %.1f\n", payload.RepetitionPenalty)
	fmt.Printf("[DEBUG] Message Count: %d\n", len(payload.Messages))
	fmt.Printf("[DEBUG] System Prompt Length: %d chars\n", len(payload.Messages[0].Content))
	fmt.Printf("[DEBUG] Assistant Context Length: %d chars\n", len(payload.Messages[1].Content))
	fmt.Printf("[DEBUG] User Input Length: %d chars\n", len(payload.Messages[2].Content))
	fmt.Printf("[DEBUG] Total Payload Size: %d bytes\n", len(payloadJSON))
	fmt.Printf("[DEBUG] API Key Masked: %s\n", maskAPIKey(finalAPIKey))
	fmt.Println()

	// Step 9: Message Breakdown
	fmt.Println("📋 STEP 9: MESSAGE BREAKDOWN")
	fmt.Println("============================")
	for i, msg := range payload.Messages {
		fmt.Printf("Message %d [%s]:\n", i+1, msg.Role)
		fmt.Printf("  Length: %d characters\n", len(msg.Content))
		fmt.Printf("  Preview: %s...\n", getPreview(msg.Content, 150))
		fmt.Println()
	}

	// Step 10: Simulated API Call Debug
	fmt.Println("🌐 STEP 10: SIMULATED API CALL DEBUG")
	fmt.Println("====================================")
	fmt.Printf("[API_CALL] Making POST request to: %s\n", apiURL)
	fmt.Printf("[API_CALL] Headers: Content-Type: application/json\n")
	fmt.Printf("[API_CALL] Headers: Authorization: Bearer %s\n", maskAPIKey(finalAPIKey))
	fmt.Printf("[API_CALL] Payload size: %d bytes\n", len(payloadJSON))
	fmt.Printf("[API_CALL] Request timeout: 30 seconds\n")
	fmt.Printf("[API_CALL] Retry attempts: 3\n")
	fmt.Printf("[API_CALL] Circuit breaker: CLOSED\n")
	fmt.Printf("[API_CALL] Rate limit status: OK\n")
	fmt.Println()

	// Step 11: Expected Response Format
	fmt.Println("📤 STEP 11: EXPECTED AI RESPONSE FORMAT")
	fmt.Println("=======================================")
	expectedResponse := `{
  "Stage": "Problem Identification",
  "Response": [
    {
      "type": "text",
      "content": "I understand you need help with your order. Let me assist you with that."
    },
    {
      "type": "text", 
      "content": "Could you please provide your order number so I can look up the details?"
    }
  ]
}`
	fmt.Println(expectedResponse)
	fmt.Println()

	fmt.Println("🎯 SUMMARY")
	fmt.Println("==========")
	fmt.Println("This demonstrates the complete AI payload generation process:")
	fmt.Println("")
	fmt.Println("1. 📋 Device configuration loading (provider, instance, API keys)")
	fmt.Println("2. 🤖 AI settings retrieval (behavior prompts, closing prompts)")
	fmt.Println("3. 💬 Conversation context gathering (stage, history, user input)")
	fmt.Println("4. 📝 System prompt construction with instructions")
	fmt.Println("5. 🔧 API configuration determination (URL, model, authentication)")
	fmt.Println("6. 🚀 Payload creation with OpenRouter/OpenAI format")
	fmt.Println("7. 📦 JSON serialization for API transmission")
	fmt.Println("8. 🔍 Comprehensive debug logging before API call")
	fmt.Println("9. 📋 Message structure breakdown and analysis")
	fmt.Println("10. 🌐 API call execution with retry logic")
	fmt.Println("11. 📤 Response parsing and stage management")
	fmt.Println("")
	fmt.Println("The system follows the exact payload structure defined in your custom instructions:")
	fmt.Println("- Model selection based on device ID")
	fmt.Println("- Three-message format: system, assistant, user")
	fmt.Println("- Temperature: 0.67, Top P: 1, Repetition Penalty: 1")
	fmt.Println("- Special handling for SCHQ-S94 and SCHQ-S12 devices (OpenAI)")
	fmt.Println("- All other devices use OpenRouter with custom models")
	fmt.Println("")
	fmt.Println("All debug information is logged to console before the AI generates its reply.")
}

// Helper functions
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
}

func getPreview(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen]
}

func buildAIPromptContent(behavePrompt, closingPrompt, stage string) string {
	content := behavePrompt + "\n\n" +
		"### Instructions:\n" +
		"1. If the current stage is null or undefined, default to the first stage.\n" +
		"2. Always analyze the user's input to determine the appropriate stage. If the input context is unclear, guide the user within the default stage context.\n" +
		"3. Follow all rules and steps strictly. Do not skip or ignore any rules or instructions.\n\n" +
		"4. **Do not repeat the same sentences or phrases that have been used in the recent conversation history.**\n" +
		"5. If the input contains the phrase \"I want this section in add response format [onemessage]\":\n" +
		"   - Add the `Jenis` field with the value `onemessage` at the item level for each text response.\n" +
		"   - The `Jenis` field is only added to `text` types within the `Response` array.\n" +
		"   - If the directive is not present, omit the `Jenis` field entirely.\n\n" +
		"### Response Format:\n" +
		"{\n" +
		"  \"Stage\": \"[Stage]\",  // Specify the current stage explicitly.\n" +
		"  \"Response\": [\n" +
		"    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Provide the first response message here.\"},\n" +
		"    {\"type\": \"image\", \"content\": \"https://example.com/image1.jpg\"},\n" +
		"    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Provide the second response message here.\"}\n" +
		"  ]\n" +
		"}\n\n" +
		closingPrompt

	return content
}

func getAPIURL(deviceID string) string {
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		return "https://api.openai.com/v1/chat/completions"
	}
	return "https://openrouter.ai/api/v1/chat/completions"
}

func getAIModel(deviceID, apiKeyOption string) string {
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		return "gpt-4"
	}
	return apiKeyOption
}

func getAPIKey(deviceID, deviceAPIKey string) string {
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		return "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"
	}
	return deviceAPIKey
}

func getProviderName(deviceID string) string {
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		return "OpenAI"
	}
	return "OpenRouter"
}