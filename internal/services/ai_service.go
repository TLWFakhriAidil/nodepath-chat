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

	"github.com/sirupsen/logrus"
)

const (
	openRouterBaseURL = "https://openrouter.ai/api/v1"
	defaultModel      = "openai/gpt-4o"
	maxRetries        = 3
	retryDelay       = time.Second * 1 // Reduced from 2s for faster retries
	circuitBreakerThreshold = 5 // Number of consecutive failures before circuit opens
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
	httpClient *http.Client
	// Response cache for frequently asked questions
	cache     map[string]*CachedResponse
	cacheMux  sync.RWMutex
	cacheTTL  time.Duration
	// Rate limiting for concurrent requests
	semaphore chan struct{}
	// Circuit breaker for API failure handling
	circuitBreaker *CircuitBreaker
	// Advanced rate limiter for API calls
	rateLimiter *APIRateLimiter
}

// NewAIService creates a new AI service with performance optimizations
func NewAIService(cfg *config.Config) *AIService {
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
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // Reduced from 30s for better real-time performance
		},
		cache:     make(map[string]*CachedResponse),
		cacheTTL:  5 * time.Minute, // Cache responses for 5 minutes
		semaphore: make(chan struct{}, 100), // Limit concurrent AI requests
		circuitBreaker: &CircuitBreaker{}, // Initialize circuit breaker
		rateLimiter:    rateLimiter,       // Initialize rate limiter
	}
}

// maskAPIKey masks an API key for safe logging
// maskAPIKey returns the full API key for debugging purposes
// WARNING: This exposes the full API key in logs - use only for debugging
func maskAPIKey(apiKey string) string {
	// Return full API key for debugging
	return apiKey
}

// GenerateResponse generates an AI response using OpenRouter with caching and concurrency control
func (s *AIService) GenerateResponse(systemPrompt, userInput, apiKey, deviceID string, conversationHistory []models.ConversationMessage) (string, error) {
	// 🔍 DEBUG TRACE: Log initial API key state
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"api_key_provided": apiKey != "",
		"api_key_source": func() string {
			if apiKey != "" {
				return "parameter"
			}
			return "none"
		}(),
		"api_key_preview": func() string {
			if apiKey != "" {
				return maskAPIKey(apiKey)
			}
			return "none"
		}(),
	}).Info("🔍 AI_SERVICE_DEBUG: Initial API key state")

	if apiKey == "" {
		apiKey = s.cfg.OpenRouterDefaultKey
		// 🔍 DEBUG TRACE: Log fallback to default key
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"api_key_source": "config_default",
			"api_key_preview": func() string {
				if apiKey != "" {
					return maskAPIKey(apiKey)
				}
				return "none"
			}(),
		}).Info("🔍 AI_SERVICE_DEBUG: Using default API key from config")
	}

	if apiKey == "" {
		logrus.WithField("device_id", deviceID).Error("🔍 AI_SERVICE_DEBUG: No API key available after all fallbacks")
		return "", fmt.Errorf("no API key provided")
	}

	// 🔍 DEBUG TRACE: Log final API key state
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"api_key_final_preview": maskAPIKey(apiKey),
		"system_prompt_length": len(systemPrompt),
		"user_input": userInput,
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

	// Create request
	request := models.OpenRouterRequest{
		Model:    defaultModel,
		Messages: messages,
		Stream:   false,
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
	if apiKey == "" {
		apiKey = s.cfg.OpenRouterDefaultKey
	}

	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	// Build enhanced system prompt with structured response format
	enhancedSystemPrompt := s.buildEnhancedSystemPrompt(systemPrompt, closingPrompt)

	// Build messages for OpenRouter
	messages := s.buildMessages(enhancedSystemPrompt, userInput, conversationHistory)

	// Create request
	request := models.OpenRouterRequest{
		Model:    defaultModel,
		Messages: messages,
		Stream:   false,
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
		"model":         response.Model,
		"prompt_tokens": response.Usage.PromptTokens,
		"total_tokens":  response.Usage.TotalTokens,
		"stage":         parsedResponse.Stage,
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

// makeOpenRouterRequest makes the actual HTTP request to OpenRouter with circuit breaker and rate limiting protection
func (s *AIService) makeOpenRouterRequest(request models.OpenRouterRequest, apiKey, deviceID string) (*models.OpenRouterResponse, error) {
	// Check circuit breaker before making request
	if s.isCircuitBreakerOpen() {
		return nil, fmt.Errorf("circuit breaker is open, API temporarily unavailable")
	}

	// Determine provider based on API key or device ID
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

	// Create HTTP request
	req, err := http.NewRequest("POST", openRouterBaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
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

// parseAIResponse parses the AI response JSON into structured format
func (s *AIService) parseAIResponse(content string) (*models.AIPromptResponse, error) {
	// Clean the content - remove code block markers if present
	sanitizedContent := content
	if strings.HasPrefix(content, "```json") {
		sanitizedContent = strings.TrimPrefix(content, "```json")
	}
	if strings.HasSuffix(sanitizedContent, "```") {
		sanitizedContent = strings.TrimSuffix(sanitizedContent, "```")
	}
	sanitizedContent = strings.TrimSpace(sanitizedContent)

	// Try to parse as JSON first
	var response models.AIPromptResponse
	err := json.Unmarshal([]byte(sanitizedContent), &response)
	if err == nil && response.Stage != "" && len(response.Response) > 0 {
		return &response, nil
	}

	// Fallback: try to extract using regex patterns (similar to PHP implementation)
	if stage, responseParts, ok := s.extractWithRegex(content); ok {
		return &models.AIPromptResponse{
			Stage:    stage,
			Response: responseParts,
		}, nil
	}

	// Final fallback: treat as plain text
	return &models.AIPromptResponse{
		Stage: "conversation",
		Response: []models.AIResponsePart{
			{
				Type:    "text",
				Content: content,
				Jenis:   "onemessage",
			},
		},
	}, nil
}

// extractWithRegex attempts to extract stage and response using regex patterns
func (s *AIService) extractWithRegex(content string) (string, []models.AIResponsePart, bool) {
	// Pattern 1: Stage: ... Response: [...]
	pattern1 := `Stage:\s*(.+?)\s*Response:\s*(\[.*?\])$`
	re1 := regexp.MustCompile(pattern1)
	matches1 := re1.FindStringSubmatch(content)
	if len(matches1) == 3 {
		stage := strings.TrimSpace(matches1[1])
		responseJSON := matches1[2]
		
		var responseParts []models.AIResponsePart
		err := json.Unmarshal([]byte(responseJSON), &responseParts)
		if err == nil {
			return stage, responseParts, true
		}
	}

	// Pattern 2: JSON-like structure detection
	pattern2 := `^\s*{\s*"Stage":\s*".+?",\s*"Response":\s*\[.*\]\s*}\s*$`
	re2 := regexp.MustCompile(pattern2)
	if re2.MatchString(content) {
		var response models.AIPromptResponse
		err := json.Unmarshal([]byte(content), &response)
		if err == nil {
			return response.Stage, response.Response, true
		}
	}

	return "", nil, false
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