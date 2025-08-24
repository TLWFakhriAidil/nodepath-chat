package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"

	"github.com/sirupsen/logrus"
)

// AIWhatsappService interface defines methods for AI WhatsApp conversation management
type AIWhatsappService interface {
	// Process AI conversation
	ProcessAIConversation(prospectNum, idDevice, currentText, stage string) (*AIWhatsappResponse, error)
	
	// Get AI settings
	GetAISettings(idDevice string) (*models.AISettings, error)
	
	// Update conversation stage
	UpdateConversationStage(prospectNum, stage string) error
	
	// Log conversation
	LogConversation(prospectNum string, idDevice string, message, sender, stage string) error
	
	// Save conversation history to conv_last field
	SaveConversationHistory(prospectNum, idDevice, userMessage, botResponse, stage string) error
	
	// Check if human takeover is active
	IsHumanTakeoverActive(prospectNum string) (bool, error)
	
	// Toggle human takeover
	ToggleHumanTakeover(prospectNum string, human bool) error
	
	// Process device commands (%, #, cmd)
	ProcessDeviceCommand(prospectNum, command, idDevice string) error
}

// AIWhatsappResponse represents the response from AI WhatsApp service
type AIWhatsappResponse struct {
	Stage    string                    `json:"Stage"`
	Response []AIWhatsappResponseItem `json:"Response"`
}

// AIWhatsappResponseItem represents individual response items
type AIWhatsappResponseItem struct {
	Type    string `json:"type"`
	Jenis   string `json:"Jenis,omitempty"`
	Content string `json:"content"`
}

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

// AIWhatsappAPIResponse represents the response from AI API
type AIWhatsappAPIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// aiWhatsappService implements AIWhatsappService interface
type aiWhatsappService struct {
	aiRepo       repository.AIWhatsappRepository
	deviceRepo   repository.DeviceSettingsRepository
	flowService  *FlowService
	httpClient   *http.Client
}

// NewAIWhatsappService creates a new instance of AIWhatsappService
func NewAIWhatsappService(aiRepo repository.AIWhatsappRepository, deviceRepo repository.DeviceSettingsRepository, flowService *FlowService) AIWhatsappService {
	return &aiWhatsappService{
		aiRepo:      aiRepo,
		deviceRepo:  deviceRepo,
		flowService: flowService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessAIConversation processes AI conversation and returns response
func (s *aiWhatsappService) ProcessAIConversation(prospectNum, idDevice, currentText, stage string) (*AIWhatsappResponse, error) {
	// Check for device commands first
	if strings.HasPrefix(currentText, "%") || strings.HasPrefix(currentText, "#") || strings.ToLower(currentText) == "cmd" {
		err := s.ProcessDeviceCommand(prospectNum, currentText, idDevice)
		if err != nil {
			logrus.WithError(err).Error("Failed to process device command")
		}
		// Don't return AI response for device commands
		return nil, fmt.Errorf("device command processed")
	}

	// Check if human takeover is active
	humanActive, err := s.IsHumanTakeoverActive(prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to check human takeover status")
		return nil, fmt.Errorf("failed to check human takeover: %w", err)
	}

	if humanActive {
		logrus.WithField("prospect_num", prospectNum).Info("Human takeover is active, skipping AI response")
		return nil, fmt.Errorf("human takeover is active")
	}

	// Get device settings
	deviceSettings, err := s.deviceRepo.GetDeviceSettingsByDevice(idDevice)
	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings")
		return nil, fmt.Errorf("failed to get device settings: %w", err)
	}

	if deviceSettings == nil {
		return nil, fmt.Errorf("device settings not found for device: %s", idDevice)
	}

	// Get AI conversation data
	aiConv, err := s.aiRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI conversation")
		return nil, fmt.Errorf("failed to get AI conversation: %w", err)
	}

	// If prospect doesn't exist, create a new one with proper flow-based stage
	if aiConv == nil {
		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"id_device": idDevice,
		}).Info("Creating new prospect record")
		
		// Get default flow for the device to determine initial stage
		defaultFlow, err := s.flowService.GetDefaultFlowForDevice(idDevice)
		if err != nil {
			logrus.WithError(err).Warn("Failed to get default flow for device, using default stage")
		}
		
		// Determine initial stage and niche from flow
		initialStage := "welcome" // Default stage
		niche := ""
		
		if defaultFlow != nil {
			// Get the start node from the flow to determine initial stage
			startNode, err := s.flowService.GetStartNode(defaultFlow)
			if err == nil && startNode != nil {
				// Use the node ID as the initial stage
				if startNode.ID != "" {
					initialStage = startNode.ID
				}
			}
			niche = defaultFlow.Niche
			logrus.WithFields(logrus.Fields{
				"flow_id": defaultFlow.ID,
				"flow_name": defaultFlow.Name,
				"initial_stage": initialStage,
				"niche": niche,
			}).Info("Using flow-based configuration for new prospect")
		} else {
			logrus.WithField("id_device", idDevice).Warn("No flow found for device, using default configuration")
		}
		
		// Create new AI WhatsApp conversation record
		now := time.Now()
		newAIConv := &models.AIWhatsapp{
			IDDevice:    idDevice, // Use idDevice for device identification
			ProspectNum: prospectNum,
			Stage:       initialStage,
			Human:       0,         // AI is active by default
			Niche:       niche,
			DateOrder:   &now,
		}
		
		err = s.aiRepo.CreateAIWhatsapp(newAIConv)
		if err != nil {
			logrus.WithError(err).Error("Failed to create new prospect record")
			return nil, fmt.Errorf("failed to create new prospect record: %w", err)
		}
		
		// Use the newly created conversation
		aiConv = newAIConv
		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"stage": initialStage,
			"niche": niche,
		}).Info("New prospect record created successfully")
	}

	// Get AI settings
	var aiSettings *models.AISettings
	if aiConv != nil {
		aiSettings, err = s.GetAISettings(aiConv.IDDevice)
		if err != nil {
			logrus.WithError(err).Error("Failed to get AI settings")
			return nil, fmt.Errorf("failed to get AI settings: %w", err)
		}
	}

	// Get conversation history
	convHistory, err := s.aiRepo.GetConversationHistory(prospectNum, 10)
	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation history")
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Build AI prompt content
	promptContent := s.buildAIPromptContent(aiSettings, stage)

	// Get last AI response
	lastText := s.getLastAIResponse(convHistory)

	// Determine API URL and model based on device
	apiURL := s.getAPIURL(idDevice)
	model := s.getAIModel(idDevice, deviceSettings.APIKeyOption)

	// Create AI payload
	payload := AIWhatsappPayload{
		Model: model,
		Messages: []AIWhatsappMessage{
			{Role: "system", Content: promptContent},
			{Role: "assistant", Content: lastText},
			{Role: "user", Content: currentText},
		},
		Temperature:       0.67,
		TopP:              1.0,
		RepetitionPenalty: 1.0,
	}

	// Call AI API
	apiKey := ""
	if deviceSettings.APIKey.Valid {
		apiKey = deviceSettings.APIKey.String
	}
	aiResponse, err := s.callAIAPI(apiURL, apiKey, payload)
	if err != nil {
		logrus.WithError(err).Error("Failed to call AI API")
		return nil, fmt.Errorf("failed to call AI API: %w", err)
	}

	// Parse AI response
	parsedResponse, err := s.parseAIResponse(aiResponse)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Update conversation stage if changed
	if parsedResponse.Stage != "" && parsedResponse.Stage != stage {
		err = s.UpdateConversationStage(prospectNum, parsedResponse.Stage)
		if err != nil {
			logrus.WithError(err).Error("Failed to update conversation stage")
		}
	}

	// Log user message
	var staffID string
	if aiConv != nil {
		staffID = aiConv.IDDevice
	}
	err = s.LogConversation(prospectNum, staffID, currentText, "user", stage)
	if err != nil {
		logrus.WithError(err).Error("Failed to log user message")
	}

	// Log AI response
	aiResponseText := s.formatResponseForLogging(parsedResponse.Response)
	err = s.LogConversation(prospectNum, staffID, aiResponseText, "bot", parsedResponse.Stage)
	if err != nil {
		logrus.WithError(err).Error("Failed to log AI response")
	}

	return parsedResponse, nil
}

// GetAISettings retrieves AI settings for a staff member
func (s *aiWhatsappService) GetAISettings(idDevice string) (*models.AISettings, error) {
	// For now, return a default AI settings since the method doesn't exist
	// TODO: Implement GetAISettingsByStaff method in repository
	return &models.AISettings{
		ID:             "default",
		IDDevice:       idDevice,
		SystemPrompt:   "You are a helpful AI assistant.",
		ClosingPrompt:  "Thank you for using our service.",
		InstancePrompt: "Please provide more details.",
	}, nil
}

// UpdateConversationStage updates the conversation stage
func (s *aiWhatsappService) UpdateConversationStage(prospectNum, stage string) error {
	// For now, we'll use UpdateAIWhatsapp to update the stage
	// TODO: Implement UpdateStage method in repository
	aiConv, err := s.aiRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		return err
	}
	if aiConv == nil {
		return fmt.Errorf("conversation not found for prospect: %s", prospectNum)
	}
	
	aiConv.Stage = stage
	return s.aiRepo.UpdateAIWhatsapp(aiConv)
}

// LogConversation logs a conversation message
func (s *aiWhatsappService) LogConversation(prospectNum string, idDevice string, message, sender, stage string) error {
	convLog := &models.ConversationLog{
		ProspectNum: prospectNum,
		IDDevice:    idDevice,
		Message:     message,
		Sender:      sender,
		Stage:       stage,
	}

	return s.aiRepo.CreateConversationLog(convLog)
}

// IsHumanTakeoverActive checks if human takeover is active
func (s *aiWhatsappService) IsHumanTakeoverActive(prospectNum string) (bool, error) {
	aiConv, err := s.aiRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		return false, err
	}

	if aiConv == nil {
		return false, nil
	}

	return aiConv.Human == 1, nil
}

// ToggleHumanTakeover toggles human takeover status
func (s *aiWhatsappService) ToggleHumanTakeover(prospectNum string, human bool) error {
	humanValue := 0
	if human {
		humanValue = 1
	}

	// For now, we'll use UpdateAIWhatsapp to update the human field
	// TODO: Implement UpdateHuman method in repository
	aiConv, err := s.aiRepo.GetAIWhatsappByProspectNum(prospectNum)
	if err != nil {
		return err
	}
	if aiConv == nil {
		return fmt.Errorf("conversation not found for prospect: %s", prospectNum)
	}
	
	aiConv.Human = humanValue
	return s.aiRepo.UpdateAIWhatsapp(aiConv)
}

// SaveConversationHistory saves conversation history to conv_last field
// Creates new record if phone number and id_device combination doesn't exist
// Updates existing record if combination already exists
func (s *aiWhatsappService) SaveConversationHistory(prospectNum, idDevice, userMessage, botResponse, stage string) error {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"device_id":    idDevice,
		"stage":        stage,
	}).Info("Saving conversation history")

	// Use repository method to handle create or update logic
	return s.aiRepo.SaveConversationHistory(prospectNum, idDevice, userMessage, botResponse, stage)
}

// ProcessDeviceCommand processes device-specific commands
func (s *aiWhatsappService) ProcessDeviceCommand(prospectNum, command, idDevice string) error {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"command":      command,
		"device_id":    idDevice,
	}).Info("Processing device command")

	// Handle different command types
	switch {
	case strings.HasPrefix(command, "%"):
		// Wablas provider command
		logrus.Info("Processing Wablas provider command")
		// TODO: Implement Wablas-specific logic
		return nil
		
	case strings.HasPrefix(command, "#"):
		// Whacenter provider command
		logrus.Info("Processing Whacenter provider command")
		// TODO: Implement Whacenter-specific logic
		return nil
		
	case strings.ToLower(command) == "cmd":
		// Toggle human takeover
		logrus.Info("Toggling human takeover")
		return s.ToggleHumanTakeover(prospectNum, true)
		
	default:
		return fmt.Errorf("unknown device command: %s", command)
	}
}

// buildAIPromptContent builds the AI prompt content according to custom instructions
func (s *aiWhatsappService) buildAIPromptContent(aiSettings *models.AISettings, stage string) string {
	var systemPrompt string
	if aiSettings != nil {
		systemPrompt = aiSettings.SystemPrompt
	}

	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant for WhatsApp conversations."
	}

	// Build the complete prompt content according to the custom instructions
	content := systemPrompt + "\n\n" +
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
		"### Example Response:\n" +
		"// If the directive is present\n" +
		"{\n" +
		"  \"Stage\": \"Problem Identification\",\n" +
		"  \"Response\": [\n" +
		"    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni.\"},\n" +
		"    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Kurang selera makan, sembelit, atau kerap demam?\"}\n" +
		"  ]\n" +
		"}\n\n" +
		"// If the directive is NOT present\n" +
		"{\n" +
		"  \"Stage\": \"Problem Identification\",\n" +
		"  \"Response\": [\n" +
		"    {\"type\": \"text\", \"content\": \"Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni.\"},\n" +
		"    {\"type\": \"text\", \"content\": \"Kurang selera makan, sembelit, atau kerap demam?\"}\n" +
		"  ]\n" +
		"}\n\n" +
		"### Important Rules:\n" +
		"1. **Include the `Stage` field in every response**:\n" +
		"   - The `Stage` field must explicitly specify the current stage.\n" +
		"   - If the stage is unclear or missing, default to first stage.\n\n" +
		"2. **Use the Correct Response Format**:\n" +
		"   - Divide long responses into multiple short \"text\" segments for better readability.\n" +
		"   - Include all relevant images provided in the input, interspersed naturally with text responses.\n" +
		"   - If multiple images are provided, create separate `image` entries for each.\n\n" +
		"3. **Dynamic Field for [onemessage]**:\n" +
		"   - If the input specifies \"I want this section in add response format [onemessage]\":\n" +
		"      - Add `\"Jenis\": \"onemessage\"` to each `text` type in the `Response` array.\n" +
		"   - If the directive is not present, omit the `Jenis` field entirely.\n" +
		"   - Non-text types like `image` never include the `Jenis` field.\n\n"

	return content
}

// getLastAIResponse gets the last AI response from conversation history
func (s *aiWhatsappService) getLastAIResponse(convHistory []models.ConversationLog) string {
	for _, conv := range convHistory {
		if conv.Sender == "bot" {
			return conv.Message
		}
	}
	return ""
}

// getAPIURL determines the API URL based on device ID
func (s *aiWhatsappService) getAPIURL(idDevice string) string {
	// Special devices use OpenAI API
	if idDevice == "SCHQ-S94" || idDevice == "SCHQ-S12" {
		return "https://api.openai.com/v1/chat/completions"
	}
	// Default to OpenRouter API
	return "https://openrouter.ai/api/v1/chat/completions"
}

// getAIModel determines the AI model based on device and API key option
func (s *aiWhatsappService) getAIModel(idDevice, apiKeyOption string) string {
	// Special devices use GPT-4
	if idDevice == "SCHQ-S94" || idDevice == "SCHQ-S12" {
		return "gpt-4"
	}
	// Use API key option for other devices
	return apiKeyOption
}

// callAIAPI calls the AI API with the given payload
func (s *aiWhatsappService) callAIAPI(apiURL, apiKey string, payload AIWhatsappPayload) (string, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse AIWhatsappAPIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices in API response")
	}

	return apiResponse.Choices[0].Message.Content, nil
}

// parseAIResponse parses the AI response JSON
func (s *aiWhatsappService) parseAIResponse(responseText string) (*AIWhatsappResponse, error) {
	// Clean the response text
	responseText = strings.TrimSpace(responseText)
	
	// Try to extract JSON from markdown code blocks if present
	if strings.Contains(responseText, "```json") {
		start := strings.Index(responseText, "```json") + 7
		end := strings.Index(responseText[start:], "```")
		if end != -1 {
			responseText = responseText[start : start+end]
		}
	} else if strings.Contains(responseText, "```") {
		start := strings.Index(responseText, "```") + 3
		end := strings.Index(responseText[start:], "```")
		if end != -1 {
			responseText = responseText[start : start+end]
		}
	}

	var aiResponse AIWhatsappResponse
	err := json.Unmarshal([]byte(responseText), &aiResponse)
	if err != nil {
		logrus.WithError(err).WithField("response_text", responseText).Error("Failed to parse AI response JSON")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Validate response structure
	if aiResponse.Stage == "" {
		aiResponse.Stage = "default"
	}

	if len(aiResponse.Response) == 0 {
		return nil, fmt.Errorf("empty response from AI")
	}

	return &aiResponse, nil
}

// formatResponseForLogging formats the response items for logging
func (s *aiWhatsappService) formatResponseForLogging(responses []AIWhatsappResponseItem) string {
	var parts []string
	for _, resp := range responses {
		if resp.Type == "text" {
			parts = append(parts, resp.Content)
		} else if resp.Type == "image" {
			parts = append(parts, "[Image: "+resp.Content+"]")
		}
	}
	return strings.Join(parts, " ")
}