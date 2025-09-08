package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"
	"nodepath-chat/internal/utils"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Circuit breaker constants for AI WhatsApp service
const (
	whatsappCircuitBreakerThreshold = 5 // Number of consecutive failures before circuit opens
	whatsappCircuitBreakerTimeout   = 30 * time.Second // Time to wait before trying again
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
	
	// Create AI WhatsApp record for prospect tracking
	CreateAIWhatsappRecord(prospectNum, idDevice, userMessage, niche string) error
	
	// Get AI WhatsApp record by prospect and device
	GetAIWhatsappByProspectAndDevice(prospectNum, idDevice string) (*models.AIWhatsapp, error)
	
	// Update AI WhatsApp record
	UpdateAIWhatsapp(ai *models.AIWhatsapp) error
	
	// Flow execution methods
	// Start a new flow execution
	StartFlowExecution(prospectNum, idDevice, flowReference string, variables map[string]interface{}) (*models.AIWhatsapp, error)
	
	// Get active flow execution
	GetActiveFlowExecution(prospectNum, idDevice string) (*models.AIWhatsapp, error)
	
	// Get any flow execution (regardless of status) - used for delayed message processing
	GetFlowExecutionByProspectAndDevice(prospectNum, idDevice string) (*models.AIWhatsapp, error)
	
	// Update flow execution state
	UpdateFlowExecution(prospectNum, idDevice, currentNode string, variables map[string]interface{}, status string) error
	
	// Complete flow execution
	CompleteFlowExecution(prospectNum, idDevice string) error
	
	// Get flow execution variables
	GetFlowExecutionVariables(prospectNum, idDevice string) (map[string]interface{}, error)
	
	// Parse AI response JSON
	ParseAIResponse(responseText string) (*AIWhatsappResponse, error)
	
	// Get repository for direct access
	GetRepository() repository.AIWhatsappRepository
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

// CircuitBreakerWhatsapp represents the state of a circuit breaker for WhatsApp AI service
type CircuitBreakerWhatsapp struct {
	failureCount    int
	lastFailureTime time.Time
	isOpen          bool
	mutex           sync.RWMutex
}

// aiWhatsappService implements AIWhatsappService interface
type aiWhatsappService struct {
	aiRepo                repository.AIWhatsappRepository
	deviceRepo            repository.DeviceSettingsRepository
	flowService           *FlowService
	mediaDetectionService *MediaDetectionService
	httpClient            *http.Client
	circuitBreaker        *CircuitBreakerWhatsapp
	// Advanced rate limiter for API calls
	rateLimiter           *APIRateLimiter
	cfg                   *config.Config
	responseProcessor     *AIResponseProcessor
}

// maskAPIKeyForLogging masks API key for logging purposes
func maskAPIKeyForLogging(apiKey string) string {
	// Return full API key for debugging - remove masking
	return apiKey
}

// NewAIWhatsappService creates a new instance of AIWhatsappService
func NewAIWhatsappService(aiRepo repository.AIWhatsappRepository, deviceRepo repository.DeviceSettingsRepository, flowService *FlowService, mediaDetectionService *MediaDetectionService, cfg *config.Config) AIWhatsappService {
	// Initialize rate limiter configuration for WhatsApp AI service
	rateLimiterConfig := &RateLimiterConfig{
		RequestsPerMinute: 120, // Higher limit for WhatsApp service
		BurstSize:         30,
		TimeWindow:        time.Minute,
	}

	rateLimiter := NewAPIRateLimiter(rateLimiterConfig)
	// Start cleanup routine for inactive device limiters
	rateLimiter.StartCleanupRoutine()

	// Initialize AI response processor with default delay
	responseProcessor := NewAIResponseProcessor(5 * time.Second)

	return &aiWhatsappService{
		aiRepo:                aiRepo,
		deviceRepo:            deviceRepo,
		flowService:           flowService,
		mediaDetectionService: mediaDetectionService,
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // Reduced from 30s for better real-time performance
		},
		circuitBreaker:    &CircuitBreakerWhatsapp{}, // Initialize circuit breaker
		rateLimiter:       rateLimiter,
		cfg:               cfg,
		responseProcessor: responseProcessor,
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

	// Build AI prompt content
	promptContent := s.buildAIPromptContent(aiSettings, stage)

	// Get last AI response from conv_last column
	lastText := s.getLastAIResponse(aiConv)

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
	// Check if device has a valid API key (not empty and not a test key)
	isValidAPIKey := deviceSettings.APIKey.Valid && 
		deviceSettings.APIKey.String != "" && 
		!strings.HasPrefix(deviceSettings.APIKey.String, "sk-test")
	
	if isValidAPIKey {
		apiKey = deviceSettings.APIKey.String
		logrus.WithFields(logrus.Fields{
			"id_device": idDevice,
			"api_key_source": "device_settings",
			"api_key_preview": maskAPIKeyForLogging(apiKey),
		}).Info("Using device-specific API key")
	} else {
		// Use default OpenRouter key for non-special devices
		if idDevice != "SCHQ-S94" && idDevice != "SCHQ-S12" {
			apiKey = s.cfg.OpenRouterDefaultKey
			logrus.WithFields(logrus.Fields{
				"id_device": idDevice,
				"api_key_source": "default_openrouter",
				"api_key_preview": maskAPIKeyForLogging(apiKey),
			}).Info("Using default OpenRouter API key")
		} else {
			// For special devices, use the hardcoded OpenAI key from custom instructions
			apiKey = "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"
			logrus.WithFields(logrus.Fields{
				"id_device": idDevice,
				"api_key_source": "hardcoded_openai",
				"api_key_preview": maskAPIKeyForLogging(apiKey),
			}).Info("Using hardcoded OpenAI API key for special device")
		}
	}
	// Call AI API with validation and retry logic
	var aiResponse string
	var parsedResponse *AIWhatsappResponse
	
	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Call AI API
		aiResponse, err = s.callAIAPI(apiURL, apiKey, idDevice, payload)
		if err != nil {
			logrus.WithError(err).Error("Failed to call AI API")
			return nil, fmt.Errorf("failed to call AI API: %w", err)
		}

		// Validate AI response format
		valid, validationErr := s.validateAIResponse(aiResponse)
		if valid {
			// Parse valid AI response
			parsedResponse, err = s.ParseAIResponse(aiResponse)
			if err != nil {
				logrus.WithError(err).Error("Failed to parse AI response")
				return nil, fmt.Errorf("failed to parse AI response: %w", err)
			}
			break // Success, exit retry loop
		}

		// Invalid response - log and retry with stricter prompt
		logrus.WithError(validationErr).Warn("Invalid AI response format, retrying...")
		logrus.WithFields(logrus.Fields{
			"attempt": attempt + 1,
			"max_retries": maxRetries,
			"response_preview": aiResponse[:min(200, len(aiResponse))],
		}).Warn("AI returned non-JSON response, retrying with stricter prompt")

		if attempt < maxRetries {
			// Modify payload with stricter JSON enforcement for retry
			messages := payload.Messages
			if len(messages) > 0 {
				stricterPrompt := messages[0].Content + 
					"\n\n🚨 CRITICAL ERROR DETECTED: Your previous response was NOT valid JSON! 🚨\n" +
					"You MUST respond with ONLY valid JSON format starting with { and ending with }.\n" +
					"NO explanations, NO markdown, NO code blocks, NO plain text - ONLY JSON!\n" +
					"Example: {\"Stage\": \"Problem Identification\", \"Response\": [{\"type\": \"text\", \"content\": \"Your message here\"}]}\n" +
					"RESPOND WITH JSON NOW:"
				
				payload.Messages[0].Content = stricterPrompt
			}
		}
	}

	// If all retries failed, return error
	if parsedResponse == nil {
		logrus.WithField("final_response", aiResponse).Error("AI failed to provide valid JSON after all retries")
		return nil, fmt.Errorf("AI failed to provide valid JSON response after %d attempts", maxRetries+1)
	}

	// Update conversation stage if changed
	if parsedResponse.Stage != "" && parsedResponse.Stage != stage {
		err = s.UpdateConversationStage(prospectNum, parsedResponse.Stage)
		if err != nil {
			logrus.WithError(err).Error("Failed to update conversation stage")
		}
		// Also update the AIWhatsapp record
		if aiConv != nil {
			aiConv.Stage = parsedResponse.Stage
			s.aiRepo.UpdateAIWhatsapp(aiConv)
		}
	}

	// Build conversation log entries matching PHP implementation
	var convLogEntries []string
	
	// Log user message
	convLogEntries = append(convLogEntries, fmt.Sprintf("USER: %s", currentText))
	
	// Process response items for logging (with onemessage combining logic)
	var textParts []string
	isOnemessageActive := false
	
	for index, respItem := range parsedResponse.Response {
		if respItem.Type == "text" && respItem.Jenis == "onemessage" {
			textParts = append(textParts, respItem.Content)
			isOnemessageActive = true
			
			// Check if next item is also onemessage
			isLastItem := index == len(parsedResponse.Response)-1
			nextIsNotOnemessage := false
			if !isLastItem {
				nextItem := parsedResponse.Response[index+1]
				nextIsNotOnemessage = nextItem.Type != "text" || nextItem.Jenis != "onemessage"
			}
			
			// If last or next is different, add combined entry
			if isLastItem || nextIsNotOnemessage {
				combinedMessage := strings.Join(textParts, "\n")
				convLogEntries = append(convLogEntries, fmt.Sprintf("BOT_COMBINED: %s", strconv.Quote(combinedMessage)))
				textParts = []string{}
				isOnemessageActive = false
			}
		} else {
			// Flush any pending onemessage parts
			if isOnemessageActive && len(textParts) > 0 {
				combinedMessage := strings.Join(textParts, "\n")
				convLogEntries = append(convLogEntries, fmt.Sprintf("BOT_COMBINED: %s", strconv.Quote(combinedMessage)))
				textParts = []string{}
				isOnemessageActive = false
			}
			
			// Add regular entry
			switch respItem.Type {
			case "text":
				convLogEntries = append(convLogEntries, fmt.Sprintf("BOT: %s", strconv.Quote(respItem.Content)))
			case "image", "audio", "video":
				convLogEntries = append(convLogEntries, fmt.Sprintf("BOT: %s", respItem.Content))
			}
		}
	}
	
	// Update conv_last in database
	if aiConv != nil {
		// Append to existing conv_last
		existingConv := ""
		if aiConv.ConvLast != nil {
			existingConv = string(aiConv.ConvLast)
			if existingConv != "" && existingConv != "null" {
				existingConv += "\n"
			}
		}
		
		newConvLast := existingConv + strings.Join(convLogEntries, "\n")
		aiConv.ConvLast = json.RawMessage(strconv.Quote(newConvLast))
		aiConv.ConvCurrent = sql.NullString{} // Clear conv_current
		
		err = s.aiRepo.UpdateAIWhatsapp(aiConv)
		if err != nil {
			logrus.WithError(err).Error("Failed to update conversation history")
		}
	}

	// Log conversation using existing methods for compatibility
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
		"### CRITICAL: JSON FORMAT REQUIREMENT ###\n" +
		"YOU MUST RESPOND ONLY IN VALID JSON FORMAT. NO EXCEPTIONS.\n" +
		"DO NOT use bracket format like [IMAGE: URL] or plain text.\n" +
		"DO NOT include any text outside the JSON structure.\n" +
		"ALWAYS use the exact JSON format specified below.\n\n" +
		"### Instructions:\n" +
		"1. If the current stage is null or undefined, default to the first stage.\n" +
		"2. Always analyze the user's input to determine the appropriate stage. If the input context is unclear, guide the user within the default stage context.\n" +
		"3. Follow all rules and steps strictly. Do not skip or ignore any rules or instructions.\n\n" +
		"4. **Do not repeat the same sentences or phrases that have been used in the recent conversation history.**\n" +
		"5. If the input contains the phrase \"I want this section in add response format [onemessage]\":\n" +
		"   - Add the `Jenis` field with the value `onemessage` at the item level for each text response.\n" +
		"   - The `Jenis` field is only added to `text` types within the `Response` array.\n" +
		"   - If the directive is not present, omit the `Jenis` field entirely.\n\n" +
		"### MANDATORY JSON RESPONSE FORMAT:\n" +
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
		"   - Non-text types like `image` never include the `Jenis` field.\n\n" +
		"### FINAL WARNING: JSON FORMAT ENFORCEMENT ###\n" +
		"CRITICAL: Your response will be parsed as JSON. Any deviation from the exact format will cause system errors.\n" +
		"- Start with { and end with }\n" +
		"- Include Stage and Response fields exactly as shown\n" +
		"- For images, use type: \"image\" with content: \"URL\"\n" +
		"- NO bracket format like [IMAGE: URL] - this will break the system\n" +
		"- NO plain text responses - only JSON\n" +
		"- NO explanations outside the JSON structure\n" +
		"RESPOND ONLY WITH VALID JSON. NOTHING ELSE.\n"

	return content
}

// getLastAIResponse gets the last AI response from conv_last column
// getLastAIResponse retrieves the raw conv_last data from the AIWhatsapp record
// Returns the complete conversation history stored in conv_last column
func (s *aiWhatsappService) getLastAIResponse(aiConv *models.AIWhatsapp) string {
	if aiConv == nil || aiConv.ConvLast == nil {
		return ""
	}

	// Return raw conv_last data without processing
	convLastStr := string(aiConv.ConvLast)
	if convLastStr == "" || convLastStr == "null" {
		return ""
	}

	return convLastStr
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

// validateAIResponse validates that the AI response is in proper JSON format
// validateAIResponse provides flexible validation for dynamic AI responses
// Handles user-defined prompts that may produce varied but valid content
func (s *aiWhatsappService) validateAIResponse(response string) (bool, error) {
	// Clean the response
	cleanResponse := strings.TrimSpace(response)
	
	// First attempt: Try direct JSON validation
	if isValid, err := s.validateDirectJSON(cleanResponse); isValid {
		return true, nil
	} else if err == nil {
		// JSON is valid but missing required fields - this is acceptable for dynamic content
		return true, nil
	}
	
	// Second attempt: Try to extract JSON from mixed content
	if extractedJSON := s.extractJSONFromResponse(cleanResponse); extractedJSON != "" {
		if isValid, _ := s.validateDirectJSON(extractedJSON); isValid {
			return true, nil
		}
	}
	
	// Third attempt: Check if response contains meaningful content that can be converted
	if s.hasValidContent(cleanResponse) {
		// Allow responses with valid content structure even if not perfect JSON
		return true, nil
	}
	
	return false, fmt.Errorf("response does not contain valid JSON or extractable content")
}

// validateDirectJSON checks if response is valid JSON with required structure
func (s *aiWhatsappService) validateDirectJSON(response string) (bool, error) {
	// Check basic JSON structure
	if !strings.HasPrefix(response, "{") || !strings.HasSuffix(response, "}") {
		return false, fmt.Errorf("not a JSON object")
	}
	
	// Try to parse as JSON
	var testResponse AIWhatsappResponse
	err := json.Unmarshal([]byte(response), &testResponse)
	if err != nil {
		return false, fmt.Errorf("invalid JSON format: %v", err)
	}
	
	// Flexible field validation - allow missing Stage for dynamic content
	if testResponse.Stage == "" && len(testResponse.Response) == 0 {
		return false, fmt.Errorf("missing both Stage and Response fields")
	}
	
	return true, nil
}

// extractJSONFromResponse attempts to extract JSON from mixed content responses
func (s *aiWhatsappService) extractJSONFromResponse(response string) string {
	// Look for JSON object patterns in the response
	jsonPattern := regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
	matches := jsonPattern.FindAllString(response, -1)
	
	for _, match := range matches {
		// Test if this match is valid JSON
		var testObj map[string]interface{}
		if err := json.Unmarshal([]byte(match), &testObj); err == nil {
			// Check if it has AI response structure
			if _, hasStage := testObj["Stage"]; hasStage {
				return match
			}
			if _, hasResponse := testObj["Response"]; hasResponse {
				return match
			}
		}
	}
	
	return ""
}

// hasValidContent checks if response contains meaningful content for conversion
func (s *aiWhatsappService) hasValidContent(response string) bool {
	// Check for common AI response patterns that indicate valid content
	patterns := []string{
		`(?i)stage\s*[:=]\s*["']?[^"'\n]+["']?`,
		`(?i)response\s*[:=]\s*\[`,
		`(?i)content\s*[:=]\s*["']`,
		`(?i)type\s*[:=]\s*["']?(text|image|audio|video)["']?`,
	}
	
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, response); matched {
			return true
		}
	}
	
	// Check for minimum content length and structure
	if len(strings.TrimSpace(response)) > 20 {
		// Has substantial content - likely a valid response
		return true
	}
	
	return false
}

// callAIAPI calls the AI API with the given payload
func (s *aiWhatsappService) callAIAPI(apiURL, apiKey, deviceID string, payload AIWhatsappPayload) (string, error) {
	// Check circuit breaker status
	if s.isCircuitBreakerOpen() {
		return "", fmt.Errorf("WhatsApp AI service circuit breaker is open, API calls temporarily disabled")
	}

	// Check rate limiting
	provider := "openrouter"
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		provider = "openai"
	}

	if err := s.rateLimiter.CheckRateLimit(provider, deviceID); err != nil {
		return "", fmt.Errorf("rate limit exceeded for device %s on provider %s: %w", deviceID, provider, err)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		s.recordAPIFailure()
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		s.recordAPIFailure()
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.recordAPIFailure()
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.recordAPIFailure()
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.recordAPIFailure()
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse AIWhatsappAPIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		s.recordAPIFailure()
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		s.recordAPIFailure()
		return "", fmt.Errorf("no choices in API response")
	}

	// Record successful API call
	s.recordAPISuccess()
	return apiResponse.Choices[0].Message.Content, nil
}

// ParseAIResponse parses the AI response JSON using the new processor
// ParseAIResponse provides intelligent parsing for dynamic AI responses
// Handles user-defined prompts that may produce varied content formats
func (s *aiWhatsappService) ParseAIResponse(responseText string) (*AIWhatsappResponse, error) {
	cleanResponse := strings.TrimSpace(responseText)
	
	// First attempt: Try direct JSON parsing
	if response, err := s.parseDirectJSON(cleanResponse); err == nil {
		return response, nil
	}
	
	// Second attempt: Extract JSON from mixed content
	if extractedJSON := s.extractJSONFromResponse(cleanResponse); extractedJSON != "" {
		if response, err := s.parseDirectJSON(extractedJSON); err == nil {
			return response, nil
		}
	}
	
	// Third attempt: Intelligent content extraction for non-JSON responses
	if response := s.parseNonJSONContent(cleanResponse); response != nil {
		return response, nil
	}
	
	// Fourth attempt: Use AI response processor as fallback
	processedMessages, err := s.responseProcessor.ProcessAIResponse(responseText, nil)
	if err == nil && len(processedMessages) > 0 {
		var responseItems []AIWhatsappResponseItem
		for _, msg := range processedMessages {
			responseItems = append(responseItems, AIWhatsappResponseItem{
				Type:    msg.Type,
				Content: msg.Content,
			})
		}
		
		return &AIWhatsappResponse{
			Stage:    "Problem Identification", // Default stage
			Response: responseItems,
		}, nil
	}
	
	return nil, fmt.Errorf("failed to parse AI response: %s", responseText[:min(100, len(responseText))])
}

// parseDirectJSON attempts to parse response as direct JSON
func (s *aiWhatsappService) parseDirectJSON(responseText string) (*AIWhatsappResponse, error) {
	// Handle code block wrapped JSON
	if strings.Contains(responseText, "```json") {
		start := strings.Index(responseText, "```json") + 7
		end := strings.Index(responseText[start:], "```")
		if end != -1 {
			responseText = responseText[start : start+end]
		}
	}
	
	var response AIWhatsappResponse
	err := json.Unmarshal([]byte(responseText), &response)
	if err != nil {
		return nil, err
	}
	
	// Set default stage if missing
	if response.Stage == "" {
		response.Stage = "Problem Identification"
	}
	
	// Ensure we have at least one response item
	if len(response.Response) == 0 {
		return nil, fmt.Errorf("empty response array")
	}
	
	return &response, nil
}

// parseNonJSONContent extracts meaningful content from non-JSON responses
func (s *aiWhatsappService) parseNonJSONContent(responseText string) *AIWhatsappResponse {
	// Extract stage information
	stage := s.extractStageFromText(responseText)
	if stage == "" {
		stage = "Problem Identification"
	}
	
	// Extract content items
	responseItems := s.extractContentFromText(responseText)
	if len(responseItems) == 0 {
		return nil
	}
	
	return &AIWhatsappResponse{
		Stage:    stage,
		Response: responseItems,
	}
}

// extractStageFromText attempts to extract stage information from text
func (s *aiWhatsappService) extractStageFromText(text string) string {
	// Look for stage patterns
	stagePatterns := []string{
		`(?i)stage\s*[:=]\s*["']?([^"'\n,}]+)["']?`,
		`(?i)current\s+stage\s*[:=]\s*["']?([^"'\n,}]+)["']?`,
		`(?i)phase\s*[:=]\s*["']?([^"'\n,}]+)["']?`,
	}
	
	for _, pattern := range stagePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	
	return ""
}

// extractContentFromText extracts content items from text responses
func (s *aiWhatsappService) extractContentFromText(text string) []AIWhatsappResponseItem {
	var items []AIWhatsappResponseItem
	
	// Look for structured content patterns
	contentPatterns := []string{
		`(?i)content\s*[:=]\s*["']([^"']+)["']`,
		`(?i)message\s*[:=]\s*["']([^"']+)["']`,
		`(?i)text\s*[:=]\s*["']([^"']+)["']`,
	}
	
	for _, pattern := range contentPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				items = append(items, AIWhatsappResponseItem{
					Type:    "text",
					Content: strings.TrimSpace(match[1]),
				})
			}
		}
	}
	
	// Look for URL patterns (images, etc.)
	urlPattern := regexp.MustCompile(`https?://[^\s"'<>]+\.(jpg|jpeg|png|gif|webp|mp4|mp3|wav)`)
	urls := urlPattern.FindAllString(text, -1)
	for _, url := range urls {
		mediaType := "image"
		if strings.Contains(url, ".mp4") {
			mediaType = "video"
		} else if strings.Contains(url, ".mp3") || strings.Contains(url, ".wav") {
			mediaType = "audio"
		}
		
		items = append(items, AIWhatsappResponseItem{
			Type:    mediaType,
			Content: url,
		})
	}
	
	// If no structured content found, treat entire response as text
	if len(items) == 0 && len(strings.TrimSpace(text)) > 0 {
		// Clean up the text
		cleanText := strings.TrimSpace(text)
		// Remove common AI response prefixes
		prefixes := []string{"AI:", "Assistant:", "Bot:", "Response:"}
		for _, prefix := range prefixes {
			if strings.HasPrefix(cleanText, prefix) {
				cleanText = strings.TrimSpace(cleanText[len(prefix):])
				break
			}
		}
		
		if len(cleanText) > 0 {
			items = append(items, AIWhatsappResponseItem{
				Type:    "text",
				Content: cleanText,
			})
		}
	}
	
	return items
}

// formatResponseForLogging formats the response items for logging
func (s *aiWhatsappService) formatResponseForLogging(responses []AIWhatsappResponseItem) string {
	var parts []string
	for _, resp := range responses {
		if resp.Type == "text" {
			parts = append(parts, resp.Content)
		} else if resp.Type == "image" {
			parts = append(parts, "[Image: "+resp.Content+"]")
		} else if resp.Type == "audio" {
			parts = append(parts, "[Audio: "+resp.Content+"]")
		} else if resp.Type == "video" {
			parts = append(parts, "[Video: "+resp.Content+"]")
		}
	}
	return strings.Join(parts, " ")
}

// isMediaURL checks if a URL points to media (image, audio, video) based on common patterns


// CreateAIWhatsappRecord creates a new AI WhatsApp record for prospect tracking
// Uses transaction to ensure both AI WhatsApp record and conversation history are created atomically
func (s *aiWhatsappService) CreateAIWhatsappRecord(prospectNum, idDevice, userMessage, niche string) error {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"id_device":    idDevice,
		"niche":        niche,
	}).Info("Creating new AI WhatsApp record for prospect tracking")
	
	// Use transaction to ensure atomicity of AI record creation and conversation logging
	return utils.WithTransaction(s.aiRepo.GetDB(), func(tx *sql.Tx) error {
		// Create new AI WhatsApp conversation record
		now := time.Now()
		newAIConv := &models.AIWhatsapp{
			IDDevice:    idDevice,
			ProspectNum: prospectNum,
			Stage:       "welcome", // Default initial stage
			Human:       0,         // AI is active by default
			Niche:       niche,
			DateOrder:   &now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		
		// Create AI WhatsApp record within transaction
		query := `
			INSERT INTO ai_whatsapp_nodepath (
				id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
				conv_current, human, niche, jam, intro, 
				catatan_staff, balas, data_image, conv_stage, 
				bot_balas, keywordiklan, marketer, update_today, 
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		
		// Handle ConvCurrent as sql.NullString
		var convCurrentValue interface{}
		if newAIConv.ConvCurrent.Valid {
			convCurrentValue = newAIConv.ConvCurrent.String
		} else {
			convCurrentValue = nil
		}
		
		_, err := tx.Exec(query,
			newAIConv.IDProspect, newAIConv.IDDevice, newAIConv.ProspectNum, newAIConv.Stage, newAIConv.DateOrder, nil,
			convCurrentValue, newAIConv.Human, newAIConv.Niche, newAIConv.Jam, newAIConv.Intro,
			newAIConv.CatatanStaff, newAIConv.Balas, newAIConv.DataImage, newAIConv.ConvStage,
			newAIConv.BotBalas, newAIConv.KeywordIklan, newAIConv.Marketer, newAIConv.UpdateToday,
			newAIConv.CreatedAt, newAIConv.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to create AI WhatsApp record in transaction")
			return fmt.Errorf("failed to create AI WhatsApp record: %w", err)
		}
		
		// Create initial conversation log within transaction
		convLogQuery := `
			INSERT INTO conversation_log_nodepath (
				prospect_num, message, sender, stage, created_at
			) VALUES (?, ?, ?, ?, ?)
		`
		
		_, err = tx.Exec(convLogQuery,
			prospectNum, userMessage, "user", "welcome", now,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to create initial conversation log in transaction")
			return fmt.Errorf("failed to create initial conversation log: %w", err)
		}
		
		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"id_device":    idDevice,
			"niche":        niche,
		}).Info("AI WhatsApp record and conversation log created successfully in transaction")
		
		return nil
	})
}

// GetAIWhatsappByProspectAndDevice retrieves AI WhatsApp record by prospect number and device ID
func (s *aiWhatsappService) GetAIWhatsappByProspectAndDevice(prospectNum, idDevice string) (*models.AIWhatsapp, error) {
	return s.aiRepo.GetAIWhatsappByProspectAndDevice(prospectNum, idDevice)
}

// UpdateAIWhatsapp updates an existing AI WhatsApp record
func (s *aiWhatsappService) UpdateAIWhatsapp(ai *models.AIWhatsapp) error {
	return s.aiRepo.UpdateAIWhatsapp(ai)
}

// Flow execution methods

// StartFlowExecution starts a new flow execution in ai_whatsapp_nodepath
func (s *aiWhatsappService) StartFlowExecution(prospectNum, idDevice, flowReference string, variables map[string]interface{}) (*models.AIWhatsapp, error) {
	logrus.WithFields(logrus.Fields{
		"prospect_num":   prospectNum,
		"id_device":      idDevice,
		"flow_reference": flowReference,
	}).Info("Starting flow execution")

	// Generate unique execution ID
	executionID := fmt.Sprintf("%s_%s_%d", prospectNum, idDevice, time.Now().Unix())

	// Convert variables to JSON
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal flow variables")
		return nil, fmt.Errorf("failed to marshal flow variables: %w", err)
	}

	// Get flow data to populate intro and niche fields
	var flowIntro, flowNiche string
	if s.flowService != nil {
		flow, err := s.flowService.GetFlow(flowReference)
		if err != nil {
			logrus.WithError(err).Warn("Failed to get flow data, using default values")
		} else if flow != nil {
			flowNiche = flow.Niche
			// Use flow name as intro if available, otherwise use niche
			if flow.Name != "" {
				flowIntro = fmt.Sprintf("Welcome to %s flow", flow.Name)
			} else if flow.Niche != "" {
				flowIntro = fmt.Sprintf("Welcome to %s", flow.Niche)
			} else {
				flowIntro = "Welcome to our service"
			}
		}
	}

	// Check if record already exists
	aiConv, err := s.aiRepo.GetAIWhatsappByProspectAndDevice(prospectNum, idDevice)
	if err != nil {
		logrus.WithError(err).Error("Failed to get existing AI WhatsApp record")
		return nil, fmt.Errorf("failed to get existing record: %w", err)
	}

	now := time.Now()
	
	// Get the start node ID from the flow
	flow, err := s.flowService.GetFlow(flowReference)
	if err != nil {
		logrus.WithError(err).Error("Failed to get flow for start node")
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}
	
	startNode, err := s.flowService.GetStartNode(flow)
	if err != nil {
		logrus.WithError(err).Error("Failed to get start node from flow")
		return nil, fmt.Errorf("failed to get start node: %w", err)
	}
	
	if aiConv == nil {
		// Create new record with flow execution data
		aiConv = &models.AIWhatsapp{
			IDDevice:        idDevice,
			ProspectNum:     prospectNum,
			Stage:           "flow_start",
			Human:           0,
			DateOrder:       &now,
			Intro:           flowIntro,  // Set intro from flow data
			Niche:           flowNiche,  // Set niche from flow data
			FlowReference:   sql.NullString{String: flowReference, Valid: true},
			// Legacy fields for backward compatibility
			CurrentNode:     sql.NullString{String: startNode.ID, Valid: true},
			Variables:       json.RawMessage(variablesJSON),
			// New flow tracking fields
			FlowID:          sql.NullString{String: flowReference, Valid: true},
			CurrentNodeID:   sql.NullString{String: startNode.ID, Valid: true}, // Set to actual start node ID
			WaitingForReply: sql.NullInt32{Int32: 0, Valid: true},
			LastNodeID:      sql.NullString{String: "", Valid: false},
			ExecutionStatus: sql.NullString{String: "active", Valid: true},
			ExecutionID:     sql.NullString{String: executionID, Valid: true},
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		err = s.aiRepo.CreateAIWhatsapp(aiConv)
		if err != nil {
			logrus.WithError(err).Error("Failed to create AI WhatsApp record with flow execution")
			return nil, fmt.Errorf("failed to create record: %w", err)
		}
	} else {
		// Update existing record with flow execution data
		// First update intro and niche if they are empty (preserve existing values)
		if aiConv.Intro == "" && flowIntro != "" {
			// Update intro field separately to preserve other data
			query := `UPDATE ai_whatsapp_nodepath SET intro = ?, updated_at = ? WHERE prospect_num = ? AND id_device = ?`
			_, err := s.aiRepo.GetDB().Exec(query, flowIntro, now, prospectNum, idDevice)
			if err != nil {
				logrus.WithError(err).Warn("Failed to update intro field")
			}
		}
		if aiConv.Niche == "" && flowNiche != "" {
			// Update niche field separately to preserve other data
			query := `UPDATE ai_whatsapp_nodepath SET niche = ?, updated_at = ? WHERE prospect_num = ? AND id_device = ?`
			_, err := s.aiRepo.GetDB().Exec(query, flowNiche, now, prospectNum, idDevice)
			if err != nil {
				logrus.WithError(err).Warn("Failed to update niche field")
			}
		}
		
		// Update flow tracking fields without overwriting conversation history
		err = s.aiRepo.UpdateFlowTrackingFields(
			prospectNum, idDevice,
			flowReference, // flowID
			startNode.ID, // currentNodeID - set to actual start node ID
			"", // lastNodeID
			0, // waitingForReply
			"active", // executionStatus
			executionID, // executionID
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to update flow tracking fields")
			return nil, fmt.Errorf("failed to update flow tracking fields: %w", err)
		}
		
		// Update legacy fields for backward compatibility
		aiConv.FlowReference = sql.NullString{String: flowReference, Valid: true}
		aiConv.CurrentNode = sql.NullString{String: startNode.ID, Valid: true}
		aiConv.Variables = json.RawMessage(variablesJSON)
		aiConv.ExecutionStatus = sql.NullString{String: "active", Valid: true}
		aiConv.ExecutionID = sql.NullString{String: executionID, Valid: true}
		// Update flow tracking fields in memory for return value
		aiConv.FlowID = sql.NullString{String: flowReference, Valid: true}
		aiConv.CurrentNodeID = sql.NullString{String: startNode.ID, Valid: true}
		aiConv.WaitingForReply = sql.NullInt32{Int32: 0, Valid: true}
		aiConv.LastNodeID = sql.NullString{String: "", Valid: false}
		aiConv.UpdatedAt = now
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num":   prospectNum,
		"execution_id":   executionID,
		"flow_reference": flowReference,
	}).Info("Flow execution started successfully")

	return aiConv, nil
}

// GetActiveFlowExecution retrieves active flow execution from ai_whatsapp_nodepath
func (s *aiWhatsappService) GetActiveFlowExecution(prospectNum, idDevice string) (*models.AIWhatsapp, error) {
	aiConv, err := s.aiRepo.GetAIWhatsappByProspectAndDevice(prospectNum, idDevice)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI WhatsApp record: %w", err)
	}

	if aiConv == nil {
		return nil, nil // No record found
	}

	// Check if there's an active flow execution using new flow tracking fields
	// A flow is considered active if it has a valid FlowID and CurrentNodeID
	if !aiConv.FlowID.Valid || aiConv.FlowID.String == "" {
		return nil, nil // No active flow
	}

	// Also check if we have a valid current node ID
	if !aiConv.CurrentNodeID.Valid || aiConv.CurrentNodeID.String == "" {
		return nil, nil // No current node set
	}

	return aiConv, nil
}

// GetFlowExecutionByProspectAndDevice retrieves any flow execution (regardless of status) from ai_whatsapp_nodepath
// This is used for delayed message processing where execution might be completed but delayed messages are still pending
func (s *aiWhatsappService) GetFlowExecutionByProspectAndDevice(prospectNum, idDevice string) (*models.AIWhatsapp, error) {
	aiConv, err := s.aiRepo.GetAIWhatsappByProspectAndDevice(prospectNum, idDevice)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI WhatsApp record: %w", err)
	}

	if aiConv == nil {
		return nil, nil // No record found
	}

	// Return execution regardless of status - this allows delayed processing to continue
	// even if the execution was marked as completed
	return aiConv, nil
}

// UpdateFlowExecution updates flow execution state in ai_whatsapp_nodepath
// Uses UpdateFlowTrackingFields to preserve conversation history and other important data
func (s *aiWhatsappService) UpdateFlowExecution(prospectNum, idDevice, currentNode string, variables map[string]interface{}, status string) error {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"id_device":    idDevice,
		"current_node": currentNode,
		"status":       status,
	}).Info("Updating flow execution")

	aiConv, err := s.aiRepo.GetAIWhatsappByProspectAndDevice(prospectNum, idDevice)
	if err != nil {
		return fmt.Errorf("failed to get AI WhatsApp record: %w", err)
	}

	if aiConv == nil {
		return fmt.Errorf("AI WhatsApp record not found for prospect %s and device %s", prospectNum, idDevice)
	}

	// Determine last node ID
	lastNodeID := ""
	if currentNode != "" && aiConv.CurrentNodeID.Valid && aiConv.CurrentNodeID.String != "" {
		lastNodeID = aiConv.CurrentNodeID.String
	}

	// Get current flow ID - always ensure we have a valid FlowID
	flowID := ""
	if aiConv.FlowID.Valid && aiConv.FlowID.String != "" {
		flowID = aiConv.FlowID.String
	} else if aiConv.FlowReference.Valid && aiConv.FlowReference.String != "" {
		// If FlowID is NULL but we have a FlowReference, use that
		flowID = aiConv.FlowReference.String
		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"id_device": idDevice,
			"flow_reference": flowID,
		}).Info("Using FlowReference as FlowID since FlowID was NULL")
	} else {
		// If both are NULL, this is an error - we need a flow reference
		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"id_device": idDevice,
		}).Error("Cannot update flow execution: both FlowID and FlowReference are NULL")
		return fmt.Errorf("cannot update flow execution: no flow reference available")
	}

	// Update flow tracking fields without overwriting conversation history
	err = s.aiRepo.UpdateFlowTrackingFields(
		prospectNum, idDevice,
		flowID, // preserve existing flowID
		currentNode, // currentNodeID
		lastNodeID, // lastNodeID
		0, // waitingForReply - default to 0
		status, // executionStatus
		"", // executionID - preserve existing
	)
	if err != nil {
		return fmt.Errorf("failed to update flow tracking fields: %w", err)
	}

	// Update variables if provided (requires separate update to preserve other data)
	if variables != nil {
		variablesJSON, err := json.Marshal(variables)
		if err != nil {
			return fmt.Errorf("failed to marshal variables: %w", err)
		}
		// Update variables field separately to preserve other data
		query := `UPDATE ai_whatsapp_nodepath SET variables = ?, updated_at = ? WHERE prospect_num = ? AND id_device = ?`
		_, err = s.aiRepo.GetDB().Exec(query, string(variablesJSON), time.Now(), prospectNum, idDevice)
		if err != nil {
			logrus.WithError(err).Warn("Failed to update variables field")
		}
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num":     prospectNum,
		"current_node_id":  currentNode,
		"status":           status,
	}).Info("Flow execution updated successfully")

	return nil
}

// CompleteFlowExecution marks flow execution as completed in ai_whatsapp_nodepath
func (s *aiWhatsappService) CompleteFlowExecution(prospectNum, idDevice string) error {
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"id_device":    idDevice,
	}).Info("Completing flow execution")

	return s.UpdateFlowExecution(prospectNum, idDevice, "", nil, "completed")
}

// GetFlowExecutionVariables retrieves flow execution variables from ai_whatsapp_nodepath
func (s *aiWhatsappService) GetFlowExecutionVariables(prospectNum, idDevice string) (map[string]interface{}, error) {
	aiConv, err := s.aiRepo.GetAIWhatsappByProspectAndDevice(prospectNum, idDevice)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI WhatsApp record: %w", err)
	}

	if aiConv == nil {
		return nil, fmt.Errorf("AI WhatsApp record not found")
	}

	if len(aiConv.Variables) == 0 {
		return make(map[string]interface{}), nil // Return empty map if no variables
	}

	var variables map[string]interface{}
	err = json.Unmarshal(aiConv.Variables, &variables)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	return variables, nil
}

// isCircuitBreakerOpen checks if the circuit breaker is open for WhatsApp AI service
func (s *aiWhatsappService) isCircuitBreakerOpen() bool {
	s.circuitBreaker.mutex.RLock()
	defer s.circuitBreaker.mutex.RUnlock()
	
	if !s.circuitBreaker.isOpen {
		return false
	}
	
	// Check if enough time has passed to try again
	if time.Since(s.circuitBreaker.lastFailureTime) > whatsappCircuitBreakerTimeout {
		s.circuitBreaker.mutex.RUnlock()
		s.circuitBreaker.mutex.Lock()
		s.circuitBreaker.isOpen = false
		s.circuitBreaker.failureCount = 0
		s.circuitBreaker.mutex.Unlock()
		s.circuitBreaker.mutex.RLock()
		return false
	}
	
	return true
}

// recordAPISuccess records a successful API call for WhatsApp AI service
func (s *aiWhatsappService) recordAPISuccess() {
	s.circuitBreaker.mutex.Lock()
	defer s.circuitBreaker.mutex.Unlock()
	
	s.circuitBreaker.failureCount = 0
	s.circuitBreaker.isOpen = false
}

// recordAPIFailure records a failed API call for WhatsApp AI service
func (s *aiWhatsappService) recordAPIFailure() {
	s.circuitBreaker.mutex.Lock()
	defer s.circuitBreaker.mutex.Unlock()
	
	s.circuitBreaker.failureCount++
	s.circuitBreaker.lastFailureTime = time.Now()
	
	if s.circuitBreaker.failureCount >= whatsappCircuitBreakerThreshold {
		s.circuitBreaker.isOpen = true
		logrus.WithField("failure_count", s.circuitBreaker.failureCount).Warn("WhatsApp AI circuit breaker opened due to consecutive API failures")
	}
}

// GetRepository returns the underlying repository for direct access
// Used by other services that need to call repository methods directly
func (s *aiWhatsappService) GetRepository() repository.AIWhatsappRepository {
	return s.aiRepo
}