package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// AIWhatsappPayload represents the payload structure for AI API calls
type AIWhatsappPayload struct {
	Model             string              `json:"model"`
	Messages          []AIWhatsappMessage `json:"messages"`
	Temperature       float64             `json:"temperature"`
	TopP              float64             `json:"top_p"`
	RepetitionPenalty float64             `json:"repetition_penalty"`
}

// AIWhatsappMessage represents individual messages in the payload
type AIWhatsappMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeviceSetting represents device configuration
type DeviceSetting struct {
	IDDevice     string `gorm:"column:id_device;primaryKey" json:"id_device"`
	Provider     string `gorm:"column:provider" json:"provider"`
	Instance     string `gorm:"column:instance" json:"instance"`
	APIKey       string `gorm:"column:api_key" json:"api_key"`
	APIKeyOption string `gorm:"column:api_key_option" json:"api_key_option"`
}

// AISetting represents AI behavior settings
type AISetting struct {
	IDAI          int    `gorm:"column:id_ai;primaryKey" json:"id_ai"`
	BehavePrompt  string `gorm:"column:behave_prompt" json:"behave_prompt"`
	ClosingPrompt string `gorm:"column:closing_prompt" json:"closing_prompt"`
}

// AIWhatsapp represents the ai_whatsapp_nodepath table structure
type AIWhatsapp struct {
	IDProspect   int       `gorm:"column:id_prospect" json:"id_prospect"`
	IDDevice     string    `gorm:"column:id_device" json:"id_device"`
	ProspectNum  string    `gorm:"column:prospect_num" json:"prospect_num"`
	Stage        string    `gorm:"column:stage" json:"stage"`
	DateOrder    time.Time `gorm:"column:date_order" json:"date_order"`
	ConvLast     string    `gorm:"column:conv_last" json:"conv_last"`
	ConvCurrent  string    `gorm:"column:conv_current" json:"conv_current"`
	Human        int       `gorm:"column:human" json:"human"`
	Niche        string    `gorm:"column:niche" json:"niche"`
	Jam          string    `gorm:"column:jam" json:"jam"`
	Intro        string    `gorm:"column:intro" json:"intro"`
	CatatanStaff string    `gorm:"column:catatan_staff" json:"catatan_staff"`
	Balas        string    `gorm:"column:balas" json:"balas"` // Changed to string to handle empty values
	DataImage    string    `gorm:"column:data_image" json:"data_image"`
	ConvStage    string    `gorm:"column:conv_stage" json:"conv_stage"`
	BotBalas     time.Time `gorm:"column:bot_balas" json:"bot_balas"`
	KeywordIklan string    `gorm:"column:keywordiklan" json:"keywordiklan"`
	Marketer     string    `gorm:"column:marketer" json:"marketer"`
	UpdateToday  time.Time `gorm:"column:update_today" json:"update_today"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// ConversationHistory represents conversation records
type ConversationHistory struct {
	ID            int       `gorm:"column:id;primaryKey" json:"id"`
	IDDevice      string    `gorm:"column:id_device" json:"id_device"`
	ProspectNum   string    `gorm:"column:prospect_num" json:"prospect_num"`
	ConvLast      string    `gorm:"column:conv_last" json:"conv_last"`
	ConvCurrent   string    `gorm:"column:conv_current" json:"conv_current"`
	Stage         string    `gorm:"column:stage" json:"stage"`
	HumanTakeover int       `gorm:"column:human_takeover" json:"human_takeover"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (DeviceSetting) TableName() string {
	return "device_setting_nodepath"
}

func (AISetting) TableName() string {
	return "ai_setting_nodepath"
}

func (ConversationHistory) TableName() string {
	return "conversation_history_nodepath"
}

func (AIWhatsapp) TableName() string {
	return "ai_whatsapp_nodepath"
}

func main() {
	fmt.Println("=== AI PAYLOAD DEBUG FOR REPETITIVE RESPONSE ISSUE ===")
	fmt.Println("Analyzing why AI keeps repeating the same response...")
	fmt.Println()

	// Test data from your conversation (using actual prospect number from database)
	deviceID := "FakhriAidilTLW-001"
	prospectNum := "60179645043" // Actual prospect number found in ai_whatsapp_nodepath table
	currentText := "Betul" // Latest user input

	// Connect to database
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(mysqlURI), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		fmt.Println("\n🔧 SIMULATING WITH MOCK DATA...")
		simulateWithMockData(deviceID, prospectNum, currentText)
		return
	}

	fmt.Println("✅ Database connected successfully")
	fmt.Println()

	// Debug the actual data retrieval process
	debugDataRetrieval(db, deviceID, prospectNum, currentText)
}

func debugDataRetrieval(db *gorm.DB, deviceID, prospectNum, currentText string) {
	fmt.Println("🔍 STEP 1: DEVICE CONFIGURATION DEBUG")
	fmt.Println("=====================================\n")

	// 1. Check device settings
	var deviceSetting DeviceSetting
	err := db.Where("id_device = ?", deviceID).First(&deviceSetting).Error
	if err != nil {
		fmt.Printf("❌ ERROR: Device settings not found for %s: %v\n", deviceID, err)
		return
	}

	fmt.Printf("✅ Device Settings Found:\n")
	fmt.Printf("   - Device ID: %s\n", deviceSetting.IDDevice)
	fmt.Printf("   - Provider: %s\n", deviceSetting.Provider)
	fmt.Printf("   - Instance: %s\n", deviceSetting.Instance)
	fmt.Printf("   - API Key: %s...\n", maskAPIKey(deviceSetting.APIKey))
	fmt.Printf("   - Model (API Key Option): %s\n", deviceSetting.APIKeyOption)
	fmt.Println()

	fmt.Println("🤖 STEP 2: AI WHATSAPP CONVERSATION DATA DEBUG")
	fmt.Println("===============================================\n")

	// 2. Get Conversation Data for Test Device
	var conversation AIWhatsapp
	err = db.Where("id_device = ? AND prospect_num = ?", deviceID, prospectNum).First(&conversation).Error
	if err != nil {
		fmt.Printf("❌ Error retrieving conversation data: %v\n", err)
		return
	}
	fmt.Printf("✅ Conversation data retrieved for %s\n", conversation.ProspectNum)
	fmt.Printf("   - Stage: %s\n", conversation.Stage)
	fmt.Printf("   - Human takeover: %d\n", conversation.Human)
	fmt.Printf("   - Conv Last length: %d\n", len(conversation.ConvLast))
	fmt.Printf("   - Conv Current length: %d\n", len(conversation.ConvCurrent))
	fmt.Printf("   - Conv Last Preview: %s\n", getPreview(conversation.ConvLast, 150))
	fmt.Printf("   - Conv Current Preview: %s\n", getPreview(conversation.ConvCurrent, 150))
	fmt.Println()

	fmt.Println("🎯 STEP 3: AI SETTINGS DEBUG")
	fmt.Println("============================\n")

	// 3. Check AI settings (use mock if table doesn't exist)
	var aiSetting AISetting
	err = db.First(&aiSetting).Error
	if err != nil {
		fmt.Printf("⚠️  WARNING: AI settings table not found: %v\n", err)
		fmt.Println("   Using mock AI settings for payload construction...\n")
		
		// Use mock AI settings
		aiSetting.BehavePrompt = "You are Layla, a friendly customer support assistant. Help customers with their health product inquiries."
		aiSetting.ClosingPrompt = "Thank you for choosing our service."
	} else {
		fmt.Printf("✅ AI Settings Found:\n")
		fmt.Printf("   - Behave Prompt Length: %d characters\n", len(aiSetting.BehavePrompt))
		fmt.Printf("   - Closing Prompt Length: %d characters\n", len(aiSetting.ClosingPrompt))
		fmt.Printf("   - Behave Prompt Preview: %s...\n", getPreview(aiSetting.BehavePrompt, 100))
	}
	fmt.Println()

	fmt.Println("🚀 STEP 4: AI PAYLOAD CONSTRUCTION")
	fmt.Println("==================================\n")

	// Build the AI prompt content
	promptContent := buildAIPromptContent(aiSetting.BehavePrompt, aiSetting.ClosingPrompt, conversation.Stage)

	// Determine API configuration
	apiURL, model, apiKey := getAPIConfiguration(deviceID, deviceSetting.APIKeyOption, deviceSetting.APIKey)

	// Create the payload using actual conversation data
	payload := AIWhatsappPayload{
		Model: model,
		Messages: []AIWhatsappMessage{
			{Role: "system", Content: promptContent},
			{Role: "assistant", Content: conversation.ConvLast}, // This is the critical conv_last (lastText)
			{Role: "user", Content: currentText}, // Current user input
		},
		Temperature:       0.67,
		TopP:              1.0,
		RepetitionPenalty: 1.0,
	}

	fmt.Printf("✅ Payload constructed with model: %s\n", payload.Model)
	fmt.Printf("   - API URL: %s\n", apiURL)
	fmt.Printf("   - API Key: %s...\n", maskAPIKey(apiKey))
	fmt.Printf("   - System prompt length: %d\n", len(payload.Messages[0].Content))
	fmt.Printf("   - Assistant message (conv_last/lastText): %d chars\n", len(payload.Messages[1].Content))
	fmt.Printf("   - User message: %s\n", payload.Messages[2].Content)
	fmt.Println()

	// Critical Analysis: Check if conv_last is empty
	if len(conversation.ConvLast) == 0 {
		fmt.Println("🚨 CRITICAL ISSUE DETECTED:")
		fmt.Println("   conv_last (lastText) is EMPTY!")
		fmt.Println("   This causes AI to repeat responses because it has no conversation history.")
		fmt.Println("   The AI cannot maintain context without previous conversation data.")
	} else {
		fmt.Println("✅ conv_last (lastText) contains data - this is good!")
		fmt.Printf("   Preview: %s\n", getPreview(conversation.ConvLast, 200))
	}
	fmt.Println()

	// Display complete payload as JSON
	payloadJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("❌ ERROR: Failed to marshal payload: %v\n", err)
		return
	}

	fmt.Println("📦 COMPLETE AI PAYLOAD:")
	fmt.Println(string(payloadJSON))
	fmt.Println()

	fmt.Println("💬 STEP 5: CONVERSATION HISTORY DEBUG (ADDITIONAL)")
	fmt.Println("==================================================\n")

	// Check conversation history for additional context
	var conversations []ConversationHistory
	err = db.Where("id_device = ? AND prospect_num = ?", deviceID, prospectNum).
		Order("created_at DESC").
		Limit(10).
		Find(&conversations).Error

	if err != nil {
		fmt.Printf("❌ ERROR: Failed to retrieve conversation history: %v\n", err)
		return
	}

	if len(conversations) == 0 {
		fmt.Printf("⚠️  WARNING: No conversation history found for device %s and prospect %s\n", deviceID, prospectNum)
		fmt.Println("   This could indicate the conversation_history_nodepath table is not being populated.")
		fmt.Println()
	} else {
		fmt.Printf("✅ Found %d conversation history records:\n\n", len(conversations))
		for i, conv := range conversations {
			fmt.Printf("   Record %d (ID: %d):\n", i+1, conv.ID)
			fmt.Printf("     - Created: %s\n", conv.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("     - Stage: %s\n", conv.Stage)
			fmt.Printf("     - Human Takeover: %d\n", conv.HumanTakeover)
			fmt.Printf("     - Conv Last: %s\n", getPreview(conv.ConvLast, 150))
			fmt.Printf("     - Conv Current: %s\n", getPreview(conv.ConvCurrent, 150))
			fmt.Println()
		}
	}

	fmt.Println("🔍 STEP 6: FINAL ANALYSIS")
	fmt.Println("=========================\n")

	if len(conversation.ConvLast) == 0 {
		fmt.Println("🚨 PROBLEM IDENTIFIED:")
		fmt.Println("======================\n")
		fmt.Println("❌ The conv_last field in ai_whatsapp_nodepath is EMPTY!")
		fmt.Println("❌ This causes the AI to lose conversation context")
		fmt.Println("❌ Without context, AI repeats the same responses")
		fmt.Println()
		fmt.Println("🔧 SOLUTION:")
		fmt.Println("1. Ensure conv_last is properly updated after each AI response")
		fmt.Println("2. Verify the conversation update logic in your main application")
		fmt.Println("3. Check if the AI response is being saved back to conv_last")
		fmt.Println("4. Ensure lastText is not empty before creating payload")
	} else {
		fmt.Println("✅ Assistant context is present - payload structure looks good!")
		fmt.Println("   If AI is still repeating, check:")
		fmt.Println("   1. API response handling")
		fmt.Println("   2. Temperature and repetition penalty settings")
		fmt.Println("   3. Model-specific behavior")
	}
}

func simulateWithMockData(deviceID, prospectNum, currentText string) {
	fmt.Println("🔧 SIMULATING AI PAYLOAD WITH MOCK DATA")
	fmt.Println("=======================================\n")

	// Mock data based on your conversation
	mockLastText := "Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni. Anak kerap demam ya? Anak kerap sakit memang merisaukan. Layla faham sangat perasaan akak. Bila kerap demam, imun tubuh anak makin teruk. Ini boleh menyebabkan jangkitan lain seperti jangkitan kuman paru-paru atau dalam darah. Untuk jangka panjang, pelajaran anak akan terganggu kalau selalu terpaksa cuti sekolah. Layla share sikit tentang kita punya rawatan untuk selesaikan masalah ni boleh kak?"
	mockBehavePrompt := "You are Layla, a friendly and knowledgeable customer support assistant for an e-commerce platform. You help customers with their orders, product inquiries, and general support needs. Always be polite, professional, and helpful."
	mockClosingPrompt := "Thank you for choosing our service. Is there anything else I can help you with today?"
	currentStage := "Problem Identification"

	fmt.Printf("Device ID: %s\n", deviceID)
	fmt.Printf("Prospect Number: %s\n", prospectNum)
	fmt.Printf("Current User Input: %s\n", currentText)
	fmt.Printf("Mock Last Text Length: %d characters\n", len(mockLastText))
	fmt.Printf("Current Stage: %s\n", currentStage)
	fmt.Println()

	// Build prompt content
	promptContent := buildAIPromptContent(mockBehavePrompt, mockClosingPrompt, currentStage)

	// Get API configuration
	apiURL, model, apiKey := getAPIConfiguration(deviceID, "anthropic/claude-3.5-sonnet", "sk-ant-api03-mock-key")

	// Create payload with mock data
	payload := AIWhatsappPayload{
		Model: model,
		Messages: []AIWhatsappMessage{
			{Role: "system", Content: promptContent},
			{Role: "assistant", Content: mockLastText},
			{Role: "user", Content: currentText},
		},
		Temperature:       0.67,
		TopP:              1.0,
		RepetitionPenalty: 1.0,
	}

	// Display complete payload structure
	payloadJSON, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Println("📦 COMPLETE MOCK AI PAYLOAD:")
	fmt.Println(string(payloadJSON))
	fmt.Println()

	fmt.Println("🔍 PAYLOAD ANALYSIS:")
	fmt.Println("===================\n")
	fmt.Printf("API URL: %s\n", apiURL)
	fmt.Printf("Model: %s\n", model)
	fmt.Printf("API Key: %s...\n", maskAPIKey(apiKey))
	fmt.Printf("✅ System prompt: %d chars\n", len(payload.Messages[0].Content))
	fmt.Printf("✅ Assistant context: %d chars\n", len(payload.Messages[1].Content))
	fmt.Printf("✅ User input: %d chars\n", len(payload.Messages[2].Content))
	fmt.Println()
	fmt.Println("This payload should work correctly because:")
	fmt.Println("1. Assistant message has proper conversation context")
	fmt.Println("2. System prompt provides clear instructions")
	fmt.Println("3. User input is clear and specific")
	fmt.Println("4. Temperature and repetition penalty are set appropriately")
	fmt.Println()
	fmt.Println("If your AI is still repeating responses, check:")
	fmt.Println("1. Is conv_last being properly saved to ai_whatsapp_nodepath table?")
	fmt.Println("2. Is the conversation retrieval query working correctly?")
	fmt.Println("3. Is lastText being passed correctly to payload construction?")
	fmt.Println("4. Are you using the correct API endpoint and model?")
}

func getAPIConfiguration(deviceID, apiKeyOption, apiKey string) (string, string, string) {
	// Special handling for specific devices
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		return "https://api.openai.com/v1/chat/completions", "gpt-4.1", "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"
	}

	// Default to OpenRouter
	return "https://openrouter.ai/api/v1/chat/completions", apiKeyOption, apiKey
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
		"    {\"type\": \"text\", \"content\": \"Provide the response message here.\"}\n" +
		"  ]\n" +
		"}\n\n" +
		"### Important Rules:\n" +
		"1. **Include the `Stage` field in every response**\n" +
		"2. **Use the Correct Response Format**\n" +
		"3. **Do not repeat previous responses**\n"

	return content
}

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
	return text[:maxLen] + "..."
}