package services

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"

	"github.com/sirupsen/logrus"
)

const (
	openRouterBaseURL       = "https://openrouter.ai/api/v1"
	openAIBaseURL           = "https://api.openai.com/v1"
	defaultModel            = "openai/gpt-4o"
	maxRetries              = 3
	retryDelay              = time.Second * 1  // Reduced from 2s for faster retries
	circuitBreakerThreshold = 5                // Number of consecutive failures before circuit opens
	circuitBreakerTimeout   = 30 * time.Second // Time to wait before trying again
)

// CachedResponse represents a cached AI response
type CachedResponse struct {
	Response  string
	Timestamp time.Time
}

// CircuitBreaker represents the state of a circuit breaker
type CircuitBreaker struct {
	failureCount    int
	lastFailureTime time.Time
	isOpen          bool
	mutex           sync.RWMutex
}

// AIService handles AI/OpenRouter integration with caching and concurrency optimization
type AIService struct {
	cfg        *config.Config
	deviceRepo repository.DeviceSettingsRepository
	httpClient *http.Client
	// Response cache for frequently asked questions
	cache    map[string]*CachedResponse
	cacheMux sync.RWMutex
	cacheTTL time.Duration
	// Rate limiting for concurrent requests
	semaphore chan struct{}
	// Circuit breaker for API failure handling
	circuitBreaker *CircuitBreaker
	// Advanced rate limiter for API calls
	rateLimiter *APIRateLimiter
}

// NewAIService creates a new AI service with performance optimizations
func NewAIService(cfg *config.Config, deviceRepo repository.DeviceSettingsRepository) *AIService {
	// Initialize rate limiter configuration
	rateLimiterConfig := &RateLimiterConfig{
		RequestsPerMinute: 100,
		BurstSize:         20,
		TimeWindow:        time.Minute,
	}

	rateLimiter := NewAPIRateLimiter(rateLimiterConfig)
	// Start cleanup routine for inactive device limiters
	rateLimiter.StartCleanupRoutine()

	return &AIService{
		cfg:        cfg,
		deviceRepo: deviceRepo,
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // Reduced from 30s for better real-time performance
		},
		cache:          make(map[string]*CachedResponse),
		cacheTTL:       5 * time.Minute,          // Cache responses for 5 minutes
		semaphore:      make(chan struct{}, 100), // Limit concurrent AI requests
		circuitBreaker: &CircuitBreaker{},        // Initialize circuit breaker
		rateLimiter:    rateLimiter,              // Initialize rate limiter
	}
}

// maskAPIKey masks API key for logging purposes
func maskAPIKey(apiKey string) string {
	// Return full API key for debugging - remove masking
	return apiKey
}

// GenerateResponse generates an AI response using OpenRouter with caching and concurrency control
func (s *AIService) GenerateResponse(systemPrompt, userInput, apiKey, deviceID string, conversationHistory []models.ConversationMessage) (string, error) {
	// Use device-specific API key logic
	apiKey = s.getAPIKey(apiKey, deviceID)

	if apiKey == "" {
		return "", fmt.Errorf("no API key provided")
	}

	// 🔍 DEBUG TRACE: Log final API key state
	logrus.WithFields(logrus.Fields{
		"device_id":                  deviceID,
		"api_key_final_preview":      maskAPIKey(apiKey),
		"system_prompt_length":       len(systemPrompt),
		"user_input":                 userInput,
		"conversation_history_count": len(conversationHistory),
	}).Info("🔍 AI_SERVICE_DEBUG: Final parameters for AI API call")

	// Check cache first
	cacheKey := s.generateCacheKey(systemPrompt, userInput, conversationHistory)
	if cachedResponse := s.getCachedResponse(cacheKey); cachedResponse != "" {
		logrus.Debug("Returning cached AI response")
		return cachedResponse, nil
	}

	// Acquire semaphore for rate limiting
	select {
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
	case <-time.After(10 * time.Second):
		return "", fmt.Errorf("request timeout: too many concurrent AI requests")
	}

	// Build messages for OpenRouter
	messages := s.buildMessages(systemPrompt, userInput, conversationHistory)

	// Create request with PHP payload structure parameters and device-specific model
	request := models.OpenRouterRequest{
		Model:             s.getAIModel(deviceID), // Use device-specific model selection
		Messages:          messages,
		Stream:            false,
		Temperature:       0.67, // Recommended setting from PHP code
		TopP:              1.0,  // Keep responses within natural probability range
		RepetitionPenalty: 1.0,  // Avoid repetitive responses
	}

	// Make API call with retries
	var response *models.OpenRouterResponse
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err = s.makeOpenRouterRequest(request, apiKey, deviceID)
		if err == nil {
			break
		}

		logrus.WithFields(logrus.Fields{
			"attempt": attempt,
			"error":   err.Error(),
		}).Warn("OpenRouter API call failed, retrying")

		if attempt < maxRetries {
			time.Sleep(retryDelay * time.Duration(attempt))
		}
	}

	if err != nil {
		logrus.WithError(err).Error("All OpenRouter API attempts failed")
		return s.getFallbackResponse(userInput), nil
	}

	// Extract response content
	if len(response.Choices) == 0 {
		return s.getFallbackResponse(userInput), nil
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		return s.getFallbackResponse(userInput), nil
	}

	// Cache the response
	s.setCachedResponse(cacheKey, content)

	logrus.WithFields(logrus.Fields{
		"model":         response.Model,
		"prompt_tokens": response.Usage.PromptTokens,
		"total_tokens":  response.Usage.TotalTokens,
	}).Info("OpenRouter API call successful")

	return content, nil
}

// GenerateAdvancedResponse generates an AI response with structured JSON output for advanced AI prompt nodes
func (s *AIService) GenerateAdvancedResponse(systemPrompt, userInput, apiKey, deviceID string, conversationHistory []models.ConversationMessage, closingPrompt string) (*models.AIPromptResponse, error) {
	// Use device-specific API key logic
	apiKey = s.getAPIKey(apiKey, deviceID)

	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	// Build enhanced system prompt with structured response format
	enhancedSystemPrompt := s.buildEnhancedSystemPrompt(systemPrompt, closingPrompt)

	// Build messages for OpenRouter
	messages := s.buildMessages(enhancedSystemPrompt, userInput, conversationHistory)

	// Create request with PHP payload structure parameters and device-specific model
	request := models.OpenRouterRequest{
		Model:             s.getAIModel(deviceID), // Use device-specific model selection
		Messages:          messages,
		Stream:            false,
		Temperature:       0.67, // Recommended setting from PHP code
		TopP:              1.0,  // Keep responses within natural probability range
		RepetitionPenalty: 1.0,  // Avoid repetitive responses
	}

	// Make API call with retries
	var response *models.OpenRouterResponse
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err = s.makeOpenRouterRequest(request, apiKey, deviceID)
		if err == nil {
			break
		}

		logrus.WithFields(logrus.Fields{
			"attempt": attempt,
			"error":   err.Error(),
		}).Warn("OpenRouter API call failed, retrying")

		if attempt < maxRetries {
			time.Sleep(retryDelay * time.Duration(attempt))
		}
	}

	if err != nil {
		logrus.WithError(err).Error("All OpenRouter API attempts failed")
		return s.getFallbackAdvancedResponse(userInput), nil
	}

	// Extract and parse response content
	if len(response.Choices) == 0 {
		return s.getFallbackAdvancedResponse(userInput), nil
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		return s.getFallbackAdvancedResponse(userInput), nil
	}

	// Parse the structured response
	parsedResponse, err := s.parseAIResponse(content)
	if err != nil {
		logrus.WithError(err).Warn("Failed to parse AI response, using fallback")
		return s.getFallbackAdvancedResponse(userInput), nil
	}

	logrus.WithFields(logrus.Fields{
		"model":          response.Model,
		"prompt_tokens":  response.Usage.PromptTokens,
		"total_tokens":   response.Usage.TotalTokens,
		"stage":          parsedResponse.Stage,
		"response_parts": len(parsedResponse.Response),
	}).Info("Advanced OpenRouter API call successful")

	return parsedResponse, nil
}

// buildMessages constructs the message array for OpenRouter API
func (s *AIService) buildMessages(systemPrompt, userInput string, conversationHistory []models.ConversationMessage) []models.OpenRouterMessage {
	var messages []models.OpenRouterMessage

	// Add system prompt
	if systemPrompt != "" {
		messages = append(messages, models.OpenRouterMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Add conversation history (limit to last 10 messages to avoid token limits)
	historyLimit := 10
	startIndex := 0
	if len(conversationHistory) > historyLimit {
		startIndex = len(conversationHistory) - historyLimit
	}

	for i := startIndex; i < len(conversationHistory); i++ {
		msg := conversationHistory[i]
		role := "user"
		if msg.Role == "BOT" {
			role = "assistant"
		}

		messages = append(messages, models.OpenRouterMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Add current user input
	messages = append(messages, models.OpenRouterMessage{
		Role:    "user",
		Content: userInput,
	})

	return messages
}

// getAPIURL determines the API URL based on device ID
// Uses OpenAI for SCHQ-S94 and SCHQ-S12, OpenRouter for all other devices
func (s *AIService) getAPIURL(deviceID string) string {
	// Use OpenAI API for specific devices as per PHP code requirements
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		return openAIBaseURL
	}
	// Use OpenRouter API for all other devices
	return openRouterBaseURL
}

// getAIModel determines the AI model based on device ID
// Uses gpt-4.1 for SCHQ-S94 and SCHQ-S12, api_key_option from database for all other devices
func (s *AIService) getAIModel(deviceID string) string {
	// Use gpt-4.1 for specific devices as per PHP code requirements
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		return "gpt-4.1"
	}

	// Fetch device settings from database to get api_key_option
	deviceSettings, err := s.deviceRepo.GetDeviceSettingsByDevice(deviceID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"error":     err.Error(),
		}).Warn("Failed to fetch device settings for AI model, using default")
		return defaultModel
	}

	// Use API key option from device settings as the model
	if deviceSettings.APIKeyOption != "" {
		return deviceSettings.APIKeyOption
	}

	// Fallback to default model
	return defaultModel
}

// getAPIKey determines the API key based on device ID
// Uses specific OpenAI key for SCHQ-S94 and SCHQ-S12, provided key for all other devices
func (s *AIService) getAPIKey(providedKey, deviceID string) string {
	// Use specific OpenAI API key for SCHQ-S94 and SCHQ-S12 as per PHP code requirements
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		return "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"
	}
	// Use provided API key for all other devices
	if providedKey != "" {
		return providedKey
	}
	// Fallback to default OpenRouter key
	return s.cfg.OpenRouterDefaultKey
}

// makeOpenRouterRequest makes the actual HTTP request to AI API with circuit breaker and rate limiting protection
func (s *AIService) makeOpenRouterRequest(request models.OpenRouterRequest, apiKey, deviceID string) (*models.OpenRouterResponse, error) {
	// Check circuit breaker before making request
	if s.isCircuitBreakerOpen() {
		return nil, fmt.Errorf("circuit breaker is open, API temporarily unavailable")
	}

	// Determine provider and API URL based on device ID
	apiURL := s.getAPIURL(deviceID)
	provider := "openrouter"
	if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
		provider = "openai"
	}

	// Check rate limits before making request
	if err := s.rateLimiter.CheckRateLimit(provider, deviceID); err != nil {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"device_id": deviceID,
			"error":     err.Error(),
		}).Warn("Rate limit exceeded for API request")
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		s.recordAPIFailure()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with device-specific API URL
	req, err := http.NewRequest("POST", apiURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		s.recordAPIFailure()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://nodepath-chat.railway.app")
	req.Header.Set("X-Title", "NodePath Chat")

	// Make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.recordAPIFailure()
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.recordAPIFailure()
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		s.recordAPIFailure()
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(responseBody),
		}).Error("OpenRouter API returned error")
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var response models.OpenRouterResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		s.recordAPIFailure()
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Record successful API call
	s.recordAPISuccess()
	return &response, nil
}

// generateCacheKey creates a unique cache key for the request
func (s *AIService) generateCacheKey(systemPrompt, userInput string, conversationHistory []models.ConversationMessage) string {
	// Create a hash of the input parameters
	hasher := md5.New()
	hasher.Write([]byte(systemPrompt))
	hasher.Write([]byte(userInput))

	// Include last few messages from conversation history
	for i, msg := range conversationHistory {
		if i >= len(conversationHistory)-3 { // Only last 3 messages for cache key
			hasher.Write([]byte(msg.Content))
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// getCachedResponse retrieves a cached response if it exists and is still valid
func (s *AIService) getCachedResponse(cacheKey string) string {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	cached, exists := s.cache[cacheKey]
	if !exists {
		return ""
	}

	// Check if cache entry is still valid
	if time.Since(cached.Timestamp) > s.cacheTTL {
		// Cache expired, remove it
		go s.removeCachedResponse(cacheKey)
		return ""
	}

	return cached.Response
}

// setCachedResponse stores a response in the cache
func (s *AIService) setCachedResponse(cacheKey, response string) {
	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()

	s.cache[cacheKey] = &CachedResponse{
		Response:  response,
		Timestamp: time.Now(),
	}

	// Clean up old cache entries periodically
	go s.cleanupCache()
}

// removeCachedResponse removes a specific cache entry
func (s *AIService) removeCachedResponse(cacheKey string) {
	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()
	delete(s.cache, cacheKey)
}

// cleanupCache removes expired cache entries
func (s *AIService) cleanupCache() {
	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()

	now := time.Now()
	for key, cached := range s.cache {
		if now.Sub(cached.Timestamp) > s.cacheTTL {
			delete(s.cache, key)
		}
	}
}

// getFallbackResponse returns a fallback response when AI fails
func (s *AIService) getFallbackResponse(userInput string) string {
	fallbackResponses := []string{
		"I'm sorry, I'm having trouble processing your request right now. Please try again later.",
		"I apologize, but I'm experiencing technical difficulties. Can you please rephrase your question?",
		"Sorry, I'm unable to provide a response at the moment. Please contact support if this continues.",
		"I'm currently unable to process your message. Please try again in a few moments.",
	}

	// Simple hash-based selection for consistent fallback
	index := len(userInput) % len(fallbackResponses)
	return fallbackResponses[index]
}

// ValidateAPIKey validates an OpenRouter API key
func (s *AIService) ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Make a simple test request
	testRequest := models.OpenRouterRequest{
		Model: defaultModel,
		Messages: []models.OpenRouterMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Stream: false,
	}

	_, err := s.makeOpenRouterRequest(testRequest, apiKey, "validation")
	if err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}

	return nil
}

// GetSupportedModels returns a list of supported models
func (s *AIService) GetSupportedModels() []string {
	return []string{
		"openai/gpt-4.1",
		"openai/gpt-4",
		"openai/gpt-3.5-turbo",
		"anthropic/claude-3-opus",
		"anthropic/claude-3-sonnet",
		"anthropic/claude-3-haiku",
	}
}

// EstimateTokens provides a rough estimate of token count
func (s *AIService) EstimateTokens(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

// isCircuitBreakerOpen checks if the circuit breaker is open
func (s *AIService) isCircuitBreakerOpen() bool {
	s.circuitBreaker.mutex.RLock()
	defer s.circuitBreaker.mutex.RUnlock()

	if !s.circuitBreaker.isOpen {
		return false
	}

	// Check if enough time has passed to try again
	if time.Since(s.circuitBreaker.lastFailureTime) > circuitBreakerTimeout {
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

// recordAPISuccess records a successful API call
func (s *AIService) recordAPISuccess() {
	s.circuitBreaker.mutex.Lock()
	defer s.circuitBreaker.mutex.Unlock()

	s.circuitBreaker.failureCount = 0
	s.circuitBreaker.isOpen = false
}

// recordAPIFailure records a failed API call
func (s *AIService) recordAPIFailure() {
	s.circuitBreaker.mutex.Lock()
	defer s.circuitBreaker.mutex.Unlock()

	s.circuitBreaker.failureCount++
	s.circuitBreaker.lastFailureTime = time.Now()

	if s.circuitBreaker.failureCount >= circuitBreakerThreshold {
		s.circuitBreaker.isOpen = true
		logrus.WithField("failure_count", s.circuitBreaker.failureCount).Warn("Circuit breaker opened due to consecutive API failures")
	}
}

// TruncateToTokenLimit truncates text to fit within token limits
func (s *AIService) TruncateToTokenLimit(text string, maxTokens int) string {
	estimatedTokens := s.EstimateTokens(text)
	if estimatedTokens <= maxTokens {
		return text
	}

	// Truncate to approximate character limit
	maxChars := maxTokens * 4
	if len(text) <= maxChars {
		return text
	}

	return text[:maxChars] + "..."
}

// buildEnhancedSystemPrompt creates an enhanced system prompt with structured response format
func (s *AIService) buildEnhancedSystemPrompt(systemPrompt, closingPrompt string) string {
	enhancedPrompt := systemPrompt

	// Add structured response format instructions
	enhancedPrompt += "\n\n=== RESPONSE FORMAT ===\n"
	enhancedPrompt += "You MUST respond in the following JSON format:\n"
	enhancedPrompt += `{
`
	enhancedPrompt += `  "Stage": "current_conversation_stage",
`
	enhancedPrompt += `  "Response": [
`
	enhancedPrompt += `    {
`
	enhancedPrompt += `      "type": "text",
`
	enhancedPrompt += `      "content": "your_text_response",
`
	enhancedPrompt += `      "Jenis": "onemessage"
`
	enhancedPrompt += `    },
`
	enhancedPrompt += `    {
`
	enhancedPrompt += `      "type": "image",
`
	enhancedPrompt += `      "url": "image_url_if_needed"
`
	enhancedPrompt += `    }
`
	enhancedPrompt += `  ]
`
	enhancedPrompt += `}
`
	enhancedPrompt += "\nIMPORTANT RULES:\n"
	enhancedPrompt += "- Stage: Update based on conversation progress\n"
	enhancedPrompt += "- Response: Array of response parts (text/image)\n"
	enhancedPrompt += "- For text responses, use 'Jenis: onemessage' to combine multiple text parts\n"
	enhancedPrompt += "- Only include image responses when specifically needed\n"
	enhancedPrompt += "- Always provide valid JSON format\n"

	// Add closing prompt if provided
	if closingPrompt != "" {
		enhancedPrompt += "\n\n=== CLOSING INSTRUCTIONS ===\n"
		enhancedPrompt += closingPrompt
	}

	return enhancedPrompt
}

// parseAIResponse parses the AI response JSON into structured format with comprehensive PHP-based parsing logic
func (s *AIService) parseAIResponse(content string) (*models.AIPromptResponse, error) {
	// Detect and log response format characteristics
	formatInfo := s.detectResponseFormat(content)
	logrus.WithFields(logrus.Fields{
		"content_length":  len(content),
		"format_info":     formatInfo,
		"content_preview": s.getContentPreview(content, 100),
	}).Debug("Starting AI response parsing")

	// Step 1: Try to parse as structured JSON (direct format)
	if response, ok := s.parseStructuredJSON(content); ok {
		logrus.WithFields(logrus.Fields{
			"method":         "structured_json",
			"stage":          response.Stage,
			"response_count": len(response.Response),
		}).Info("Successfully parsed AI response")
		return response, nil
	}

	// Step 2: Try to parse older format with Stage and Response fields
	if response, ok := s.parseOlderFormat(content); ok {
		logrus.WithFields(logrus.Fields{
			"method":         "older_format",
			"stage":          response.Stage,
			"response_count": len(response.Response),
		}).Info("Successfully parsed AI response")
		return response, nil
	}

	// Step 3: Try to extract encapsulated JSON (JSON within text)
	if response, ok := s.parseEncapsulatedJSON(content); ok {
		logrus.WithFields(logrus.Fields{
			"method":         "encapsulated_json",
			"stage":          response.Stage,
			"response_count": len(response.Response),
		}).Info("Successfully parsed AI response")
		return response, nil
	}

	// Step 4: Try advanced regex patterns for various formats
	if response, ok := s.parseWithAdvancedRegex(content); ok {
		logrus.WithFields(logrus.Fields{
			"method":         "advanced_regex",
			"stage":          response.Stage,
			"response_count": len(response.Response),
		}).Info("Successfully parsed AI response")
		return response, nil
	}

	// Step 5: Final fallback - treat as plain text
	logrus.WithFields(logrus.Fields{
		"method":         "plain_text_fallback",
		"content_length": len(content),
		"format_info":    formatInfo,
		"content_sample": s.getContentPreview(content, 200),
	}).Warn("All parsing methods failed, using plain text fallback")
	return s.getPlainTextFallback(content), nil
}

// parseStructuredJSON attempts to parse content as direct structured JSON
func (s *AIService) parseStructuredJSON(content string) (*models.AIPromptResponse, bool) {
	// Clean the content - remove code block markers if present
	sanitizedContent := s.sanitizeContent(content)

	// Try to parse as JSON directly
	var response models.AIPromptResponse
	err := json.Unmarshal([]byte(sanitizedContent), &response)
	if err == nil && response.Stage != "" && len(response.Response) > 0 {
		return &response, true
	}

	return nil, false
}

// parseOlderFormat attempts to parse older format with Stage and Response fields
func (s *AIService) parseOlderFormat(content string) (*models.AIPromptResponse, bool) {
	// Pattern for older format: Stage: ... Response: [...]
	pattern := `(?s)Stage:\s*(.+?)\s*Response:\s*(\[.*\])\s*$`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) == 3 {
		stage := strings.TrimSpace(matches[1])
		responseJSON := matches[2]

		var responseParts []models.AIResponsePart
		err := json.Unmarshal([]byte(responseJSON), &responseParts)
		if err == nil && len(responseParts) > 0 {
			return &models.AIPromptResponse{
				Stage:    stage,
				Response: responseParts,
			}, true
		}
	}

	return nil, false
}

// parseEncapsulatedJSON attempts to extract JSON from within text content with comprehensive patterns
func (s *AIService) parseEncapsulatedJSON(content string) (*models.AIPromptResponse, bool) {
	// Comprehensive patterns to find JSON objects within text (based on PHP implementation)
	patterns := []struct {
		name    string
		pattern string
	}{
		{"standard_json", `(?s)\{\s*"Stage"\s*:\s*"[^"]+"\s*,\s*"Response"\s*:\s*\[.*?\]\s*\}`},
		{"loose_json", `(?s)\{[^{}]*"Stage"[^{}]*"Response"[^{}]*\}`},
		{"nested_json", `(?s)\{.*?"Stage".*?"Response".*?\}`},
		{"multiline_json", `(?s)\{[\s\S]*?"Stage"[\s\S]*?"Response"[\s\S]*?\}`},
		{"escaped_json", `(?s)\\?\{[\s\S]*?"Stage"[\s\S]*?"Response"[\s\S]*?\\?\}`},
		{"quoted_json", `(?s)"\{[\s\S]*?"Stage"[\s\S]*?"Response"[\s\S]*?\}"`},
		{"bracketed_json", `(?s)\[[\s\S]*?\{[\s\S]*?"Stage"[\s\S]*?"Response"[\s\S]*?\}[\s\S]*?\]`},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		matches := re.FindAllString(content, -1)
		for _, match := range matches {
			// Clean the match
			cleanMatch := s.cleanJSONMatch(match)

			var response models.AIPromptResponse
			err := json.Unmarshal([]byte(cleanMatch), &response)
			if err == nil && response.Stage != "" && len(response.Response) > 0 {
				logrus.WithFields(logrus.Fields{
					"pattern":      p.name,
					"match_length": len(cleanMatch),
				}).Debug("Successfully parsed encapsulated JSON")
				return &response, true
			}

			// Log parsing attempt for debugging
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"pattern": p.name,
					"error":   err.Error(),
					"match":   cleanMatch[:minInt(100, len(cleanMatch))],
				}).Debug("Failed to parse encapsulated JSON match")
			}
		}
	}

	return nil, false
}

// cleanJSONMatch cleans and prepares JSON match for parsing
func (s *AIService) cleanJSONMatch(match string) string {
	// Remove leading/trailing quotes if present
	clean := strings.TrimSpace(match)
	if strings.HasPrefix(clean, `"`) && strings.HasSuffix(clean, `"`) {
		clean = clean[1 : len(clean)-1]
	}

	// Remove escape characters
	clean = strings.ReplaceAll(clean, `\"`, `"`)
	clean = strings.ReplaceAll(clean, `\\`, `\`)

	// Remove array brackets if they wrap the entire JSON
	if strings.HasPrefix(clean, "[") && strings.HasSuffix(clean, "]") {
		// Try to extract the JSON object from within the array
		if idx := strings.Index(clean, "{"); idx != -1 {
			if lastIdx := strings.LastIndex(clean, "}"); lastIdx != -1 && lastIdx > idx {
				clean = clean[idx : lastIdx+1]
			}
		}
	}

	return clean
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseWithAdvancedRegex attempts to parse using advanced regex patterns (comprehensive PHP-based implementation)
func (s *AIService) parseWithAdvancedRegex(content string) (*models.AIPromptResponse, bool) {
	// Comprehensive advanced patterns for various response formats (based on PHP implementation)
	patterns := []struct {
		name    string
		pattern string
	}{
		// Standard patterns
		{"stage_response_multiline", `(?s)Stage\s*:\s*([^\n]+)\s*Response\s*:\s*(\[.*?\])\s*$`},
		{"json_like_structure", `(?s)^\s*\{\s*"Stage"\s*:\s*"([^"]+)"\s*,\s*"Response"\s*:\s*(\[.*?\])\s*\}\s*$`},
		{"loose_json_structure", `(?s)Stage[\s:]*([^\n,}]+)[\s,]*Response[\s:]*\[(.*?)\]`},
		{"quoted_stage_response", `(?s)"Stage"\s*:\s*"([^"]+)".*?"Response"\s*:\s*(\[.*?\])`},

		// Extended patterns for edge cases
		{"stage_colon_response", `(?s)Stage\s*:\s*([^\n\r]+)[\s\n\r]*Response\s*:\s*(\[.*?\])(?:\s*$|\s*[}\]])`},
		{"stage_equals_response", `(?s)Stage\s*=\s*([^\n\r]+)[\s\n\r]*Response\s*=\s*(\[.*?\])`},
		{"stage_arrow_response", `(?s)Stage\s*=>\s*([^\n\r]+)[\s\n\r]*Response\s*=>\s*(\[.*?\])`},
		{"stage_dash_response", `(?s)Stage\s*-\s*([^\n\r]+)[\s\n\r]*Response\s*-\s*(\[.*?\])`},

		// Flexible patterns for malformed JSON
		{"flexible_stage_response", `(?s)(?:Stage|stage|STAGE)[\s:=\-]*([^\n\r,}]+)[\s\n\r,]*(?:Response|response|RESPONSE)[\s:=\-]*(\[.*?\])`},
		{"case_insensitive_json", `(?si)\{[\s\S]*?"?stage"?\s*:\s*"?([^"\n,}]+)"?[\s\S]*?"?response"?\s*:\s*(\[.*?\])[\s\S]*?\}`},
		{"partial_json_structure", `(?s)"?Stage"?\s*:\s*"?([^"\n,}]+)"?[\s\S]*?"?Response"?\s*:\s*(\[.*?\])`},

		// Patterns for text with embedded JSON
		{"text_with_json", `(?s).*?\{[\s\S]*?"Stage"\s*:\s*"([^"]+)"[\s\S]*?"Response"\s*:\s*(\[.*?\])[\s\S]*?\}.*?`},
		{"markdown_json", `(?s)` + "`" + `(?:json)?[\s\S]*?\{[\s\S]*?"Stage"\s*:\s*"([^"]+)"[\s\S]*?"Response"\s*:\s*(\[.*?\])[\s\S]*?\}[\s\S]*?` + "`" + `?`},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) >= 3 {
			stage := s.cleanStageValue(matches[1])
			responseJSON := s.cleanResponseJSON(matches[2])

			// Validate stage is not empty
			if stage == "" {
				logrus.WithField("pattern", p.name).Debug("Empty stage found, skipping")
				continue
			}

			// Try to parse the response array
			var responseParts []models.AIResponsePart
			err := json.Unmarshal([]byte(responseJSON), &responseParts)
			if err == nil && len(responseParts) > 0 {
				logrus.WithFields(logrus.Fields{
					"pattern":        p.name,
					"stage":          stage,
					"response_count": len(responseParts),
				}).Debug("Successfully parsed with advanced regex")
				return &models.AIPromptResponse{
					Stage:    stage,
					Response: responseParts,
				}, true
			}

			// Log parsing failure for debugging
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"pattern":       p.name,
					"stage":         stage,
					"error":         err.Error(),
					"response_json": responseJSON[:minInt(200, len(responseJSON))],
				}).Debug("Failed to parse response JSON in advanced regex")
			}
		}
	}

	return nil, false
}

// cleanStageValue cleans and validates the stage value
func (s *AIService) cleanStageValue(stage string) string {
	clean := strings.TrimSpace(stage)
	clean = strings.Trim(clean, `"'`)
	clean = strings.TrimSpace(clean)

	// Remove common prefixes/suffixes
	clean = strings.TrimSuffix(clean, ",")
	clean = strings.TrimSuffix(clean, ";")
	clean = strings.TrimSpace(clean)

	return clean
}

// cleanResponseJSON cleans and validates the response JSON
func (s *AIService) cleanResponseJSON(responseJSON string) string {
	clean := strings.TrimSpace(responseJSON)

	// Ensure it starts and ends with brackets
	if !strings.HasPrefix(clean, "[") {
		clean = "[" + clean
	}
	if !strings.HasSuffix(clean, "]") {
		clean = clean + "]"
	}

	return clean
}

// sanitizeContent cleans and sanitizes the input content
func (s *AIService) sanitizeContent(content string) string {
	// Remove code block markers
	sanitized := content
	if strings.HasPrefix(content, "```json") {
		sanitized = strings.TrimPrefix(content, "```json")
	}
	if strings.HasPrefix(sanitized, "```") {
		sanitized = strings.TrimPrefix(sanitized, "```")
	}
	if strings.HasSuffix(sanitized, "```") {
		sanitized = strings.TrimSuffix(sanitized, "```")
	}

	// Remove common prefixes and suffixes
	sanitized = strings.TrimSpace(sanitized)
	sanitized = strings.Trim(sanitized, "\n\r\t ")

	// Remove any leading/trailing quotes if they wrap the entire content
	if (strings.HasPrefix(sanitized, `"`) && strings.HasSuffix(sanitized, `"`)) ||
		(strings.HasPrefix(sanitized, "'") && strings.HasSuffix(sanitized, "'")) {
		sanitized = sanitized[1 : len(sanitized)-1]
	}

	return sanitized
}

// getPlainTextFallback creates a fallback response for plain text content
func (s *AIService) getPlainTextFallback(content string) *models.AIPromptResponse {
	// Clean the content
	cleanContent := s.sanitizeContent(content)

	// If content is empty, provide a default response
	if strings.TrimSpace(cleanContent) == "" {
		cleanContent = "I apologize, but I'm having trouble processing your request. Please try again."
	}

	return &models.AIPromptResponse{
		Stage: "conversation",
		Response: []models.AIResponsePart{
			{
				Type:    "text",
				Content: cleanContent,
				Jenis:   "onemessage",
			},
		},
	}
}

// detectResponseFormat analyzes the content to determine its format characteristics
func (s *AIService) detectResponseFormat(content string) map[string]interface{} {
	formatInfo := make(map[string]interface{})

	// Basic content analysis
	formatInfo["has_json_markers"] = strings.Contains(content, "```json") || strings.Contains(content, "```")
	formatInfo["has_stage_field"] = strings.Contains(content, "Stage") || strings.Contains(content, "stage")
	formatInfo["has_response_field"] = strings.Contains(content, "Response") || strings.Contains(content, "response")
	formatInfo["has_curly_braces"] = strings.Contains(content, "{") && strings.Contains(content, "}")
	formatInfo["has_square_brackets"] = strings.Contains(content, "[") && strings.Contains(content, "]")

	// JSON structure detection
	formatInfo["appears_json"] = s.looksLikeJSON(content)
	formatInfo["has_quoted_fields"] = strings.Contains(content, `"Stage"`) && strings.Contains(content, `"Response"`)

	// Content characteristics
	lines := strings.Split(content, "\n")
	formatInfo["line_count"] = len(lines)
	formatInfo["starts_with_brace"] = strings.HasPrefix(strings.TrimSpace(content), "{")
	formatInfo["ends_with_brace"] = strings.HasSuffix(strings.TrimSpace(content), "}")

	// Pattern detection
	formatInfo["has_colon_separator"] = strings.Contains(content, "Stage:") && strings.Contains(content, "Response:")
	formatInfo["has_equals_separator"] = strings.Contains(content, "Stage=") && strings.Contains(content, "Response=")

	return formatInfo
}

// looksLikeJSON performs a basic check to see if content resembles JSON
func (s *AIService) looksLikeJSON(content string) bool {
	trimmed := strings.TrimSpace(content)

	// Remove code block markers for analysis
	if strings.HasPrefix(trimmed, "```json") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
	}
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
	}
	if strings.HasSuffix(trimmed, "```") {
		trimmed = strings.TrimSuffix(trimmed, "```")
	}
	trimmed = strings.TrimSpace(trimmed)

	// Basic JSON structure check
	return (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
}

// getContentPreview returns a safe preview of the content for logging
func (s *AIService) getContentPreview(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	// Try to break at a reasonable point
	preview := content[:maxLength]

	// Try to break at word boundary
	if lastSpace := strings.LastIndex(preview, " "); lastSpace > maxLength/2 {
		preview = preview[:lastSpace]
	}

	// Try to break at line boundary
	if lastNewline := strings.LastIndex(preview, "\n"); lastNewline > maxLength/2 {
		preview = preview[:lastNewline]
	}

	return preview + "..."
}

// getFallbackAdvancedResponse returns a fallback response for advanced AI prompts
func (s *AIService) getFallbackAdvancedResponse(userInput string) *models.AIPromptResponse {
	fallbackResponses := []string{
		"I'm sorry, I'm having trouble processing your request right now. Please try again later.",
		"I apologize, but I'm experiencing technical difficulties. Can you please rephrase your question?",
		"Sorry, I'm unable to provide a response at the moment. Please contact support if this continues.",
		"I'm currently unable to process your message. Please try again in a few moments.",
	}

	// Simple hash-based selection for consistent fallback
	index := len(userInput) % len(fallbackResponses)

	return &models.AIPromptResponse{
		Stage: "error",
		Response: []models.AIResponsePart{
			{
				Type:    "text",
				Content: fallbackResponses[index],
				Jenis:   "onemessage",
			},
		},
	}
}
