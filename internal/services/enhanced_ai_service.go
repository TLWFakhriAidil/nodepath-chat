package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// EnhancedAIService provides high-performance AI response generation
type EnhancedAIService struct {
	mu                sync.RWMutex
	baseAIService     *AIService // Original AI service
	responseCache     *MessageCache
	fallbackResponses []string
	config            *AIServiceConfig
	metrics           *AIMetrics
	circuitBreaker    *CircuitBreaker
	connectionPool    *ConnectionPool
	logger            *logrus.Entry
}

// AIServiceConfig holds configuration for the AI service
type AIServiceConfig struct {
	EnableCache          bool          `json:"enable_cache"`
	CacheSize            int           `json:"cache_size"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	RequestTimeout       time.Duration `json:"request_timeout"`
	MaxConcurrentRequests int          `json:"max_concurrent_requests"`
	EnableFallback       bool          `json:"enable_fallback"`
	RetryAttempts        int           `json:"retry_attempts"`
	RetryDelay           time.Duration `json:"retry_delay"`
	RateLimitPerMinute   int           `json:"rate_limit_per_minute"`
}

// AIMetrics tracks AI service performance
type AIMetrics struct {
	mu                    sync.RWMutex
	TotalRequests         int64         `json:"total_requests"`
	SuccessfulRequests    int64         `json:"successful_requests"`
	FailedRequests        int64         `json:"failed_requests"`
	CacheHits             int64         `json:"cache_hits"`
	CacheMisses           int64         `json:"cache_misses"`
	FallbackUsed          int64         `json:"fallback_used"`
	AverageResponseTime   time.Duration `json:"average_response_time"`
	PeakResponseTime      time.Duration `json:"peak_response_time"`
	CurrentConcurrency    int           `json:"current_concurrency"`
	MaxConcurrency        int           `json:"max_concurrency"`
	LastRequestTime       time.Time     `json:"last_request_time"`
}

// AIRequest represents an AI generation request
type AIRequest struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Context     []models.ConversationMessage `json:"context,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	SystemPrompt string                `json:"system_prompt,omitempty"`
}

// AIResponse represents an AI generation response
type AIResponse struct {
	ID           string        `json:"id"`
	Content      string        `json:"content"`
	TokensUsed   int           `json:"tokens_used,omitempty"`
	Model        string        `json:"model,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	FromCache    bool          `json:"from_cache"`
	FromFallback bool          `json:"from_fallback"`
}

// NewEnhancedAIService creates a new enhanced AI service
func NewEnhancedAIService(baseAIService *AIService) *EnhancedAIService {
	config := &AIServiceConfig{
		EnableCache:           true,
		CacheSize:             10000,
		CacheTTL:              30 * time.Minute,
		RequestTimeout:        30 * time.Second,
		MaxConcurrentRequests: 50,
		EnableFallback:        true,
		RetryAttempts:         3,
		RetryDelay:            1 * time.Second,
		RateLimitPerMinute:    1000,
	}
	
	fallbackResponses := []string{
		"I'm currently experiencing high demand. Please try again in a moment.",
		"Thank you for your message. I'm processing many requests right now, but I'll get back to you soon.",
		"I'm here to help! Due to high volume, there might be a slight delay in my response.",
		"Your message is important to me. I'm currently handling many conversations, please bear with me.",
		"I appreciate your patience. I'm working through a high volume of messages right now.",
	}
	
	service := &EnhancedAIService{
		baseAIService:     baseAIService,
		config:            config,
		fallbackResponses: fallbackResponses,
		metrics: &AIMetrics{
			LastRequestTime: time.Now(),
		},
		logger: logrus.WithFields(logrus.Fields{
			"component": "enhanced_ai_service",
		}),
	}
	
	// Initialize components
	if config.EnableCache {
		service.responseCache = NewMessageCache(config.CacheSize, config.CacheTTL)
	}
	
	service.circuitBreaker = NewCircuitBreaker()
	
	return service
}

// GenerateResponse generates an AI response with high performance optimizations
func (ais *EnhancedAIService) GenerateResponse(ctx context.Context, content string, metadata map[string]interface{}) (string, error) {
	start := time.Now()
	
	// Update metrics
	ais.updateConcurrency(1)
	defer ais.updateConcurrency(-1)
	
	ais.metrics.mu.Lock()
	ais.metrics.TotalRequests++
	ais.metrics.LastRequestTime = time.Now()
	ais.metrics.mu.Unlock()
	
	// Create AI request
	request := &AIRequest{
		ID:       fmt.Sprintf("req_%d", time.Now().UnixNano()),
		Content:  content,
		Metadata: metadata,
	}
	
	// Check cache first
	if ais.config.EnableCache {
		if cachedResponse, found := ais.getCachedResponse(request); found {
			ais.metrics.mu.Lock()
			ais.metrics.CacheHits++
			ais.metrics.SuccessfulRequests++
			ais.metrics.mu.Unlock()
			
			ais.logger.WithField("request_id", request.ID).Debug("Cache hit for AI request")
			return cachedResponse, nil
		}
		
		ais.metrics.mu.Lock()
		ais.metrics.CacheMisses++
		ais.metrics.mu.Unlock()
	}
	
	// Check circuit breaker
	if !ais.circuitBreaker.CanExecute() {
		ais.logger.WithField("request_id", request.ID).Warn("Circuit breaker is open, using fallback")
		return ais.getFallbackResponse(), nil
	}
	
	// Generate response with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, ais.config.RequestTimeout)
	defer cancel()
	
	response, err := ais.generateWithRetry(timeoutCtx, request)
	if err != nil {
		ais.circuitBreaker.RecordFailure()
		ais.metrics.mu.Lock()
		ais.metrics.FailedRequests++
		ais.metrics.mu.Unlock()
		
		// Use fallback if enabled
		if ais.config.EnableFallback {
			ais.logger.WithFields(logrus.Fields{
				"request_id": request.ID,
				"error":      err.Error(),
			}).Warn("AI request failed, using fallback")
			
			ais.metrics.mu.Lock()
			ais.metrics.FallbackUsed++
			ais.metrics.mu.Unlock()
			
			return ais.getFallbackResponse(), nil
		}
		
		return "", err
	}
	
	// Record success
	ais.circuitBreaker.RecordSuccess()
	ais.metrics.mu.Lock()
	ais.metrics.SuccessfulRequests++
	ais.metrics.mu.Unlock()
	
	// Cache the response
	if ais.config.EnableCache {
		ais.cacheResponse(request, response.Content)
	}
	
	// Update response time metrics
	responseTime := time.Since(start)
	ais.updateResponseTimeMetrics(responseTime)
	
	ais.logger.WithFields(logrus.Fields{
		"request_id":    request.ID,
		"response_time": responseTime,
		"tokens_used":   response.TokensUsed,
	}).Debug("AI response generated successfully")
	
	return response.Content, nil
}

// generateWithRetry generates AI response with retry logic
func (ais *EnhancedAIService) generateWithRetry(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt <= ais.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			retryDelay := time.Duration(attempt) * ais.config.RetryDelay
			select {
			case <-time.After(retryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		
		response, err := ais.generateResponse(ctx, request)
		if err == nil {
			return response, nil
		}
		
		lastErr = err
		ais.logger.WithFields(logrus.Fields{
			"request_id": request.ID,
			"attempt":    attempt + 1,
			"error":      err.Error(),
		}).Warn("AI request attempt failed")
		
		// Don't retry on context cancellation
		if ctx.Err() != nil {
			break
		}
	}
	
	return nil, fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
}

// generateResponse generates AI response using the base service
func (ais *EnhancedAIService) generateResponse(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	start := time.Now()
	
	// Use the base AI service to generate response
	// This would integrate with your existing AI service
	responseContent, err := ais.baseAIService.GenerateResponse(ctx, request.Content, request.Metadata)
	if err != nil {
		return nil, err
	}
	
	response := &AIResponse{
		ID:           request.ID,
		Content:      responseContent,
		ResponseTime: time.Since(start),
		FromCache:    false,
		FromFallback: false,
	}
	
	return response, nil
}

// getCachedResponse retrieves a cached response
func (ais *EnhancedAIService) getCachedResponse(request *AIRequest) (string, bool) {
	cacheKey := ais.generateCacheKey(request)
	if cached, found := ais.responseCache.Get(cacheKey); found {
		if response, ok := cached.(string); ok {
			return response, true
		}
	}
	return "", false
}

// cacheResponse caches an AI response
func (ais *EnhancedAIService) cacheResponse(request *AIRequest, response string) {
	cacheKey := ais.generateCacheKey(request)
	ais.responseCache.Set(cacheKey, response)
}

// generateCacheKey generates a cache key for the request
func (ais *EnhancedAIService) generateCacheKey(request *AIRequest) string {
	// Create a simple hash of the content for caching
	// In production, you might want to use a more sophisticated hashing
	content := strings.ToLower(strings.TrimSpace(request.Content))
	return fmt.Sprintf("ai_cache_%x", []byte(content))
}

// getFallbackResponse returns a random fallback response
func (ais *EnhancedAIService) getFallbackResponse() string {
	if len(ais.fallbackResponses) == 0 {
		return "I'm currently unavailable. Please try again later."
	}
	
	// Simple random selection based on current time
	index := int(time.Now().UnixNano()) % len(ais.fallbackResponses)
	return ais.fallbackResponses[index]
}

// updateConcurrency updates the current concurrency metrics
func (ais *EnhancedAIService) updateConcurrency(delta int) {
	ais.metrics.mu.Lock()
	defer ais.metrics.mu.Unlock()
	
	ais.metrics.CurrentConcurrency += delta
	if ais.metrics.CurrentConcurrency > ais.metrics.MaxConcurrency {
		ais.metrics.MaxConcurrency = ais.metrics.CurrentConcurrency
	}
}

// updateResponseTimeMetrics updates response time metrics
func (ais *EnhancedAIService) updateResponseTimeMetrics(responseTime time.Duration) {
	ais.metrics.mu.Lock()
	defer ais.metrics.mu.Unlock()
	
	if responseTime > ais.metrics.PeakResponseTime {
		ais.metrics.PeakResponseTime = responseTime
	}
	
	// Calculate average response time
	totalRequests := ais.metrics.SuccessfulRequests
	if totalRequests > 0 {
		currentAvg := ais.metrics.AverageResponseTime
		ais.metrics.AverageResponseTime = time.Duration(
			(int64(currentAvg)*(totalRequests-1) + int64(responseTime)) / totalRequests,
		)
	}
}

// GetMetrics returns current AI service metrics
func (ais *EnhancedAIService) GetMetrics() *AIMetrics {
	ais.metrics.mu.RLock()
	defer ais.metrics.mu.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := *ais.metrics
	return &metrics
}

// GetStatus returns the current status of the AI service
func (ais *EnhancedAIService) GetStatus() map[string]interface{} {
	metrics := ais.GetMetrics()
	
	successRate := float64(0)
	if metrics.TotalRequests > 0 {
		successRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100
	}
	
	cacheHitRate := float64(0)
	if metrics.CacheHits+metrics.CacheMisses > 0 {
		cacheHitRate = float64(metrics.CacheHits) / float64(metrics.CacheHits+metrics.CacheMisses) * 100
	}
	
	return map[string]interface{}{
		"total_requests":       metrics.TotalRequests,
		"successful_requests":  metrics.SuccessfulRequests,
		"failed_requests":      metrics.FailedRequests,
		"success_rate":         successRate,
		"cache_hits":           metrics.CacheHits,
		"cache_misses":         metrics.CacheMisses,
		"cache_hit_rate":       cacheHitRate,
		"fallback_used":        metrics.FallbackUsed,
		"current_concurrency":  metrics.CurrentConcurrency,
		"max_concurrency":      metrics.MaxConcurrency,
		"average_response_time": metrics.AverageResponseTime.String(),
		"peak_response_time":   metrics.PeakResponseTime.String(),
		"last_request_time":    metrics.LastRequestTime,
		"circuit_breaker":      ais.circuitBreaker.GetStatus(),
		"cache_size":           ais.responseCache.Size(),
	}
}

// ClearCache clears the response cache
func (ais *EnhancedAIService) ClearCache() {
	if ais.config.EnableCache && ais.responseCache != nil {
		ais.responseCache.Clear()
		ais.logger.Info("AI response cache cleared")
	}
}

// UpdateConfig updates the AI service configuration
func (ais *EnhancedAIService) UpdateConfig(config *AIServiceConfig) {
	ais.mu.Lock()
	defer ais.mu.Unlock()
	
	ais.config = config
	ais.logger.WithField("config", config).Info("AI service configuration updated")
}

// Stop stops the enhanced AI service
func (ais *EnhancedAIService) Stop() {
	if ais.responseCache != nil {
		ais.responseCache.Stop()
	}
	ais.logger.Info("Enhanced AI service stopped")
}