package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RateLimiterConfig defines configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerMinute int           // Maximum requests per minute
	BurstSize         int           // Maximum burst size
	TimeWindow        time.Duration // Time window for rate limiting
}

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	capacity   int          // Maximum tokens in bucket
	tokens     int          // Current tokens available
	refillRate int          // Tokens added per minute
	lastRefill time.Time    // Last time bucket was refilled
	mutex      sync.RWMutex // Mutex for thread safety
}

// APIRateLimiter manages rate limiting for different API providers
type APIRateLimiter struct {
	// Rate limiters for different providers
	openRouterLimiter *TokenBucket
	openAILimiter     *TokenBucket
	// Global rate limiter for all APIs
	globalLimiter *TokenBucket
	// Per-device rate limiters
	deviceLimiters map[string]*TokenBucket
	deviceMutex    sync.RWMutex
	// Configuration
	config *RateLimiterConfig
	// Metrics
	totalRequests    int64
	rejectedRequests int64
	metricsMutex     sync.RWMutex
}

// NewTokenBucket creates a new token bucket with specified capacity and refill rate
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity, // Start with full bucket
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// NewAPIRateLimiter creates a new API rate limiter with different limits for different providers
func NewAPIRateLimiter(config *RateLimiterConfig) *APIRateLimiter {
	return &APIRateLimiter{
		// OpenRouter: More generous limits (60 RPM)
		openRouterLimiter: NewTokenBucket(60, 60),
		// OpenAI: More conservative limits (40 RPM)
		openAILimiter: NewTokenBucket(40, 40),
		// Global limiter: Overall system protection (100 RPM)
		globalLimiter:  NewTokenBucket(100, 100),
		deviceLimiters: make(map[string]*TokenBucket),
		config:         config,
	}
}

// TryConsume attempts to consume a token from the bucket
func (tb *TokenBucket) TryConsume() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	if elapsed > time.Minute {
		// Calculate tokens to add (refillRate per minute)
		tokensToAdd := int(elapsed.Minutes()) * tb.refillRate
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	// Try to consume a token
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// GetAvailableTokens returns the current number of available tokens
func (tb *TokenBucket) GetAvailableTokens() int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.tokens
}

// GetTimeUntilNextToken returns the time until the next token becomes available
func (tb *TokenBucket) GetTimeUntilNextToken() time.Duration {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	if tb.tokens > 0 {
		return 0
	}

	// Calculate time until next refill
	now := time.Now()
	timeUntilNextMinute := time.Minute - now.Sub(tb.lastRefill)
	if timeUntilNextMinute <= 0 {
		return 0
	}

	return timeUntilNextMinute
}

// CheckRateLimit checks if a request is allowed for the specified provider and device
func (rl *APIRateLimiter) CheckRateLimit(provider, deviceID string) error {
	rl.metricsMutex.Lock()
	rl.totalRequests++
	rl.metricsMutex.Unlock()

	// Check global rate limit first
	if !rl.globalLimiter.TryConsume() {
		rl.recordRejection()
		return fmt.Errorf("global rate limit exceeded, try again in %v", rl.globalLimiter.GetTimeUntilNextToken())
	}

	// Check provider-specific rate limit
	var providerLimiter *TokenBucket
	switch provider {
	case "openrouter":
		providerLimiter = rl.openRouterLimiter
	case "openai":
		providerLimiter = rl.openAILimiter
	default:
		providerLimiter = rl.openRouterLimiter // Default to OpenRouter limits
	}

	if !providerLimiter.TryConsume() {
		rl.recordRejection()
		return fmt.Errorf("%s rate limit exceeded, try again in %v", provider, providerLimiter.GetTimeUntilNextToken())
	}

	// Check per-device rate limit (10 RPM per device)
	deviceLimiter := rl.getOrCreateDeviceLimiter(deviceID)
	if !deviceLimiter.TryConsume() {
		rl.recordRejection()
		return fmt.Errorf("device rate limit exceeded for %s, try again in %v", deviceID, deviceLimiter.GetTimeUntilNextToken())
	}

	logrus.WithFields(logrus.Fields{
		"provider":          provider,
		"device_id":         deviceID,
		"global_tokens":     rl.globalLimiter.GetAvailableTokens(),
		"provider_tokens":   providerLimiter.GetAvailableTokens(),
		"device_tokens":     deviceLimiter.GetAvailableTokens(),
		"total_requests":    rl.totalRequests,
		"rejected_requests": rl.rejectedRequests,
	}).Debug("Rate limit check passed")

	return nil
}

// getOrCreateDeviceLimiter gets or creates a rate limiter for a specific device
func (rl *APIRateLimiter) getOrCreateDeviceLimiter(deviceID string) *TokenBucket {
	rl.deviceMutex.RLock()
	limiter, exists := rl.deviceLimiters[deviceID]
	rl.deviceMutex.RUnlock()

	if exists {
		return limiter
	}

	// Create new device limiter (10 requests per minute per device)
	rl.deviceMutex.Lock()
	defer rl.deviceMutex.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := rl.deviceLimiters[deviceID]; exists {
		return limiter
	}

	newLimiter := NewTokenBucket(10, 10) // 10 RPM per device
	rl.deviceLimiters[deviceID] = newLimiter

	logrus.WithField("device_id", deviceID).Debug("Created new device rate limiter")
	return newLimiter
}

// recordRejection records a rejected request for metrics
func (rl *APIRateLimiter) recordRejection() {
	rl.metricsMutex.Lock()
	rl.rejectedRequests++
	rl.metricsMutex.Unlock()
}

// GetMetrics returns current rate limiting metrics
func (rl *APIRateLimiter) GetMetrics() map[string]interface{} {
	rl.metricsMutex.RLock()
	defer rl.metricsMutex.RUnlock()

	rejectionRate := float64(0)
	if rl.totalRequests > 0 {
		rejectionRate = float64(rl.rejectedRequests) / float64(rl.totalRequests) * 100
	}

	return map[string]interface{}{
		"total_requests":    rl.totalRequests,
		"rejected_requests": rl.rejectedRequests,
		"rejection_rate":    rejectionRate,
		"global_tokens":     rl.globalLimiter.GetAvailableTokens(),
		"openrouter_tokens": rl.openRouterLimiter.GetAvailableTokens(),
		"openai_tokens":     rl.openAILimiter.GetAvailableTokens(),
		"active_devices":    len(rl.deviceLimiters),
	}
}

// CleanupInactiveDevices removes device limiters that haven't been used recently
func (rl *APIRateLimiter) CleanupInactiveDevices() {
	rl.deviceMutex.Lock()
	defer rl.deviceMutex.Unlock()

	now := time.Now()
	cleanedCount := 0

	for deviceID, limiter := range rl.deviceLimiters {
		// Remove device limiters that haven't been used in the last hour
		if now.Sub(limiter.lastRefill) > time.Hour {
			delete(rl.deviceLimiters, deviceID)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		logrus.WithFields(logrus.Fields{
			"cleaned_devices":   cleanedCount,
			"remaining_devices": len(rl.deviceLimiters),
		}).Info("Cleaned up inactive device rate limiters")
	}
}

// StartCleanupRoutine starts a background routine to clean up inactive device limiters
func (rl *APIRateLimiter) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(30 * time.Minute) // Cleanup every 30 minutes
		defer ticker.Stop()

		for range ticker.C {
			rl.CleanupInactiveDevices()
		}
	}()

	logrus.Info("Rate limiter cleanup routine started")
}

// Helper function for min
func rateLimiterMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
