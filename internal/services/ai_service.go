package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

const (
	openRouterBaseURL = "https://openrouter.ai/api/v1"
	defaultModel      = "openai/gpt-4.1"
	maxRetries        = 3
	retryDelay       = time.Second * 2
)

// AIService handles AI/OpenRouter integration
type AIService struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewAIService creates a new AI service
func NewAIService(cfg *config.Config) *AIService {
	return &AIService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateResponse generates an AI response using OpenRouter
func (s *AIService) GenerateResponse(systemPrompt, userInput, apiKey string, conversationHistory []models.ConversationMessage) (string, error) {
	if apiKey == "" {
		apiKey = s.cfg.OpenRouterDefaultKey
	}

	if apiKey == "" {
		return "", fmt.Errorf("no API key provided")
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
		response, err = s.makeOpenRouterRequest(request, apiKey)
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

	logrus.WithFields(logrus.Fields{
		"model":         response.Model,
		"prompt_tokens": response.Usage.PromptTokens,
		"total_tokens":  response.Usage.TotalTokens,
	}).Info("OpenRouter API call successful")

	return content, nil
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

// makeOpenRouterRequest makes the actual HTTP request to OpenRouter
func (s *AIService) makeOpenRouterRequest(request models.OpenRouterRequest, apiKey string) (*models.OpenRouterResponse, error) {
	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", openRouterBaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
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
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
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
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
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

	_, err := s.makeOpenRouterRequest(testRequest, apiKey)
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