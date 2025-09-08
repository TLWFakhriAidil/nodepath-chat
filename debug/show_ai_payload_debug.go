package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	gorm "gorm.io/gorm"
	mysql "gorm.io/driver/mysql"
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

// DeviceSettings represents device configuration
type DeviceSettings struct {
	IDDevice      string `gorm:"column:id_device;primaryKey"`
	APIKey        string `gorm:"column:api_key"`
	APIKeyOption  string `gorm:"column:api_key_option"`
	Provider      string `gorm:"column:provider"`
	Instance      string `gorm:"column:instance"`
}

// TableName returns the table name for DeviceSettings
func (DeviceSettings) TableName() string {
	return "device_setting_nodepath"
}

// AISettings represents AI configuration
type AISettings struct {
	IDDevice     string `gorm:"column:id_device;primaryKey"`
	BehavePrompt string `gorm:"column:behave_prompt"`
	ClosingPrompt string `gorm:"column:closing_prompt"`
}

// TableName returns the table name for AISettings
func (AISettings) TableName() string {
	return "ai_setting_nodepath"
}

// AIWhatsapp represents a conversation record
type AIWhatsapp struct {
	ProspectNum string    `gorm:"column:prospect_num;primaryKey"`
	IDDevice    string    `gorm:"column:id_device"`
	Stage       string    `gorm:"column:stage"`
	Human       int       `gorm:"column:human"`
	Niche       string    `gorm:"column:niche"`
	ConvLast    string    `gorm:"column:conv_last"`
	ConvCurrent string    `gorm:"column:conv_current"`
	DateOrder   time.Time `gorm:"column:date_order"`
}

// TableName returns the table name for AIWhatsapp
func (AIWhatsapp) TableName() string {
	return "ai_whatsapp_nodepath"
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database connection string
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"
	}
	
	// Fix MySQL URI format if needed
	if mysqlURI == "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// Connect to database
	db, err := gorm.Open(mysql.Open(mysqlURI), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("=== AI PAYLOAD GENERATION AND DEBUG CONSOLE OUTPUT ===")
	fmt.Println("This shows how the AI generates replies and the debug information before generation.")
	fmt.Println()

	// Test device ID
	deviceID := "FakhriAidilTLW-001"
	prospectNum := "601137508067"
	currentText := "Hello, I need help with my order"

	fmt.Printf("🔍 DEBUG: Processing message for device: %s\n", deviceID)
	fmt.Printf("🔍 DEBUG: Prospect number: %s\n", prospectNum)
	fmt.Printf("🔍 DEBUG: User input: %s\n", currentText)
	fmt.Println()

	// Step 1: Get device settings
	fmt.Println("📋 STEP 1: Getting device settings...")
	var deviceSettings DeviceSettings
	err = db.Where("id_device = ?", deviceID).First(&deviceSettings).Error
	if err != nil {
		fmt.Printf("❌ ERROR: Failed to get device settings: %v\n", err)
		return
	}

	fmt.Printf("✅ Device settings loaded:\n")
	fmt.Printf("   - Provider: %s\n", deviceSettings.Provider)
	fmt.Printf("   - Instance: %s\n", deviceSettings.Instance)
	fmt.Printf("   - API Key Option: %s\n", deviceSettings.APIKeyOption)
	fmt.Printf("   - API Key: %s...\n", maskAPIKey(deviceSettings.APIKey))
	fmt.Println()

	// Step 2: Get AI settings
	fmt.Println("🤖 STEP 2: Getting AI settings...")
	var aiSettings AISettings
	err = db.Where("id_device = ?", deviceID).First(&aiSettings).Error
	if err != nil {
		fmt.Printf("⚠️  WARNING: AI settings table not found: %v\n", err)
		fmt.Println("📝 Using mock AI settings for demonstration")
		// Create mock AI settings
		aiSettings = AISettings{
			IDDevice:     deviceID,
			BehavePrompt: "You are a helpful customer support assistant. Always be polite and professional. Respond in a friendly manner and help customers with their inquiries.",
			ClosingPrompt: "Thank you for contacting us. Is there anything else I can help you with today?",
		}
	} else {
		fmt.Println("✅ AI settings loaded from database")
	}

	fmt.Printf("✅ AI settings configured:\n")
	fmt.Printf("   - Behave Prompt Length: %d characters\n", len(aiSettings.BehavePrompt))
	fmt.Printf("   - Closing Prompt Length: %d characters\n", len(aiSettings.ClosingPrompt))
	fmt.Printf("   - Behave Prompt Preview: %s...\n", getPreview(aiSettings.BehavePrompt, 100))
	fmt.Println()

	// Step 3: Get conversation context
	fmt.Println("💬 STEP 3: Getting conversation context...")
	var aiConv AIWhatsapp
	err = db.Where("prospect_num = ? AND id_device = ?", prospectNum, deviceID).First(&aiConv).Error
	if err != nil {
		fmt.Printf("⚠️  WARNING: No existing conversation found: %v\n", err)
		// Create mock conversation for demo
		aiConv = AIWhatsapp{
			ProspectNum: prospectNum,
			IDDevice:    deviceID,
			Stage:       "Problem Identification",
			Human:       0,
			Niche:       "customer_support",
			ConvLast:    "Hello! How can I help you today?",
			ConvCurrent: currentText,
			DateOrder:   time.Now(),
		}
		fmt.Println("📝 Created mock conversation for demonstration")
	} else {
		fmt.Println("✅ Existing conversation found")
	}

	fmt.Printf("   - Current Stage: %s\n", aiConv.Stage)
	fmt.Printf("   - Human Takeover: %d (0=AI active, 1=Human active)\n", aiConv.Human)
	fmt.Printf("   - Niche: %s\n", aiConv.Niche)
	fmt.Printf("   - Last AI Response: %s\n", getPreview(aiConv.ConvLast, 50))
	fmt.Println()

	// Step 4: Build AI prompt content
	fmt.Println("📝 STEP 4: Building AI prompt content...")
	promptContent := buildAIPromptContent(aiSettings, aiConv.Stage)
	fmt.Printf("✅ AI prompt content built (Length: %d characters)\n", len(promptContent))
	fmt.Printf("   - Prompt Preview: %s...\n", getPreview(promptContent, 200))
	fmt.Println()

	// Step 5: Determine API configuration
	fmt.Println("🔧 STEP 5: Determining API configuration...")
	apiURL := getAPIURL(deviceID)
	model := getAIModel(deviceID, deviceSettings.APIKeyOption)
	apiKey := getAPIKey(deviceID, deviceSettings.APIKey)

	fmt.Printf("✅ API configuration determined:\n")
	fmt.Printf("   - API URL: %s\n", apiURL)
	fmt.Printf("   - Model: %s\n", model)
	fmt.Printf("   - API Key: %s...\n", maskAPIKey(apiKey))
	fmt.Println()

	// Step 6: Create AI payload
	fmt.Println("🚀 STEP 6: Creating AI payload...")
	payload := AIWhatsappPayload{
		Model: model,
		Messages: []AIWhatsappMessage{
			{Role: "system", Content: promptContent},
			{Role: "assistant", Content: aiConv.ConvLast},
			{Role: "user", Content: currentText},
		},
		Temperature:       0.67,
		TopP:              1.0,
		RepetitionPenalty: 1.0,
	}

	fmt.Println("✅ AI payload created successfully!")
	fmt.Println()

	// Step 7: Display complete payload structure
	fmt.Println("📦 STEP 7: COMPLETE AI PAYLOAD STRUCTURE:")
	fmt.Println("===========================================")

	// Convert payload to JSON for display
	payloadJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("❌ ERROR: Failed to marshal payload: %v\n", err)
		return
	}

	fmt.Println(string(payloadJSON))
	fmt.Println()

	// Step 8: Console debug output before API call
	fmt.Println("🔍 STEP 8: CONSOLE DEBUG OUTPUT BEFORE AI API CALL:")
	fmt.Println("====================================================")
	fmt.Printf("[DEBUG] Device ID: %s\n", deviceID)
	fmt.Printf("[DEBUG] Prospect Number: %s\n", prospectNum)
	fmt.Printf("[DEBUG] Current Stage: %s\n", aiConv.Stage)
	fmt.Printf("[DEBUG] Human Takeover: %d\n", aiConv.Human)
	fmt.Printf("[DEBUG] API Provider: %s\n", getProviderName(deviceID))
	fmt.Printf("[DEBUG] Model: %s\n", model)
	fmt.Printf("[DEBUG] Temperature: %.2f\n", payload.Temperature)
	fmt.Printf("[DEBUG] Top P: %.1f\n", payload.TopP)
	fmt.Printf("[DEBUG] Repetition Penalty: %.1f\n", payload.RepetitionPenalty)
	fmt.Printf("[DEBUG] Message Count: %d\n", len(payload.Messages))
	fmt.Printf("[DEBUG] System Prompt Length: %d chars\n", len(payload.Messages[0].Content))
	fmt.Printf("[DEBUG] Assistant Context Length: %d chars\n", len(payload.Messages[1].Content))
	fmt.Printf("[DEBUG] User Input Length: %d chars\n", len(payload.Messages[2].Content))
	fmt.Printf("[DEBUG] Total Payload Size: %d bytes\n", len(payloadJSON))
	fmt.Printf("[DEBUG] Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Step 9: Show message breakdown
	fmt.Println("📋 STEP 9: MESSAGE BREAKDOWN:")
	fmt.Println("=============================")
	for i, msg := range payload.Messages {
		fmt.Printf("Message %d [%s]:\n", i+1, msg.Role)
		fmt.Printf("  Length: %d characters\n", len(msg.Content))
		fmt.Printf("  Preview: %s...\n", getPreview(msg.Content, 100))
		fmt.Println()
	}

	fmt.Println("🎯 SUMMARY:")
	fmt.Println("===========")
	fmt.Println("This shows the complete AI payload structure and debug information")
	fmt.Println("that the system generates before making the API call to get AI responses.")
	fmt.Println("")
	fmt.Println("The payload follows the OpenRouter/OpenAI format with:")
	fmt.Println("- System prompt (AI behavior instructions)")
	fmt.Println("- Assistant context (last AI response)")
	fmt.Println("- User input (current message)")
	fmt.Println("- Temperature, Top P, and Repetition Penalty settings")
	fmt.Println("")
	fmt.Println("All debug logs are captured in the console before the AI generates its reply.")
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

func buildAIPromptContent(aiSettings AISettings, stage string) string {
	content := aiSettings.BehavePrompt + "\n\n" +
		"### Instructions:\n" +
		"1. If the current stage is null or undefined, default to the first stage.\n" +
		"2. Always analyze the user's input to determine the appropriate stage. If the input context is unclear, guide the user within the default stage context.\n" +
		"3. Follow all rules and steps strictly. Do not skip or ignore any rules or instructions.\n\n" +
		"4. **Do not repeat the same sentences or phrases that have been used in the recent conversation history.**\n" +
		"5. If the input contains the phrase \"I want this section in add response format [onemessage]\":\n" +
		"   - Add the `Jenis` field with the value `onemessage` at the item level for each text response.\n" +
		"   - The `Jenis` field is only added to `text` types within the `Response` array.\n" +
		"   - If the directive is not present, omit the `Jenis` field entirely.\n\n" +
		aiSettings.ClosingPrompt

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