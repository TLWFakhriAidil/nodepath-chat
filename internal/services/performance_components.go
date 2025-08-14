package services

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu           sync.Mutex
	tokens       int
	maxTokens    int
	refillRate   int // tokens per second
	lastRefill   time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(tokensPerSecond int) *RateLimiter {
	return &RateLimiter{
		tokens:     tokensPerSecond,
		maxTokens:  tokensPerSecond,
		refillRate: tokensPerSecond,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request can proceed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	
	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed.Seconds()) * rl.refillRate
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}
	
	// Check if we have tokens available
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	
	return false
}

// GetStatus returns the current status of the rate limiter
func (rl *RateLimiter) GetStatus() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	return map[string]interface{}{
		"current_tokens": rl.tokens,
		"max_tokens":     rl.maxTokens,
		"refill_rate":    rl.refillRate,
		"last_refill":    rl.lastRefill,
	}
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreaker implements a circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	failureThreshold int
	successThreshold int
	timeout         time.Duration
	lastFailureTime time.Time
	nextAttempt     time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitBreakerClosed,
		failureThreshold: 5,  // Open after 5 failures
		successThreshold: 3,  // Close after 3 successes in half-open
		timeout:          30 * time.Second, // Try again after 30 seconds
	}
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		return time.Now().After(cb.nextAttempt)
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	switch cb.state {
	case CircuitBreakerClosed:
		cb.failureCount = 0
	case CircuitBreakerHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = CircuitBreakerClosed
			cb.failureCount = 0
			cb.successCount = 0
			logrus.Info("Circuit breaker closed - service recovered")
		}
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	switch cb.state {
	case CircuitBreakerClosed:
		if cb.failureCount >= cb.failureThreshold {
			cb.state = CircuitBreakerOpen
			cb.nextAttempt = time.Now().Add(cb.timeout)
			logrus.WithField("failure_count", cb.failureCount).Warn("Circuit breaker opened - service failing")
		}
	case CircuitBreakerHalfOpen:
		cb.state = CircuitBreakerOpen
		cb.nextAttempt = time.Now().Add(cb.timeout)
		cb.successCount = 0
		logrus.Info("Circuit breaker reopened - service still failing")
	}
}

// GetStatus returns the current status of the circuit breaker
func (cb *CircuitBreaker) GetStatus() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	var stateStr string
	switch cb.state {
	case CircuitBreakerClosed:
		stateStr = "closed"
	case CircuitBreakerOpen:
		stateStr = "open"
	case CircuitBreakerHalfOpen:
		stateStr = "half-open"
	}
	
	return map[string]interface{}{
		"state":             stateStr,
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"failure_threshold": cb.failureThreshold,
		"success_threshold": cb.successThreshold,
		"last_failure_time": cb.lastFailureTime,
		"next_attempt":      cb.nextAttempt,
	}
}

// MessageCacheEntry represents a cached message entry
type MessageCacheEntry struct {
	value     interface{}
	expiredAt time.Time
}

// MessageCache implements a simple in-memory cache with TTL
type MessageCache struct {
	mu       sync.RWMutex
	cache    map[string]*MessageCacheEntry
	maxSize  int
	ttl      time.Duration
	cleanup  *time.Ticker
	stopChan chan bool
}

// NewMessageCache creates a new message cache
func NewMessageCache(maxSize int, ttl time.Duration) *MessageCache {
	cache := &MessageCache{
		cache:    make(map[string]*MessageCacheEntry),
		maxSize:  maxSize,
		ttl:      ttl,
		cleanup:  time.NewTicker(ttl / 2), // Cleanup every half TTL
		stopChan: make(chan bool),
	}
	
	// Start cleanup goroutine
	go cache.cleanupExpired()
	
	return cache
}

// Set stores a value in the cache
func (mc *MessageCache) Set(key string, value interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// Check if cache is full
	if len(mc.cache) >= mc.maxSize {
		// Remove oldest entries (simple LRU)
		mc.evictOldest()
	}
	
	mc.cache[key] = &MessageCacheEntry{
		value:     value,
		expiredAt: time.Now().Add(mc.ttl),
	}
}

// Get retrieves a value from the cache
func (mc *MessageCache) Get(key string) (interface{}, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	entry, exists := mc.cache[key]
	if !exists {
		return nil, false
	}
	
	// Check if expired
	if time.Now().After(entry.expiredAt) {
		delete(mc.cache, key)
		return nil, false
	}
	
	return entry.value, true
}

// Exists checks if a key exists in the cache
func (mc *MessageCache) Exists(key string) bool {
	_, exists := mc.Get(key)
	return exists
}

// Delete removes a key from the cache
func (mc *MessageCache) Delete(key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	delete(mc.cache, key)
}

// Clear removes all entries from the cache
func (mc *MessageCache) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.cache = make(map[string]*MessageCacheEntry)
}

// Size returns the current size of the cache
func (mc *MessageCache) Size() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	return len(mc.cache)
}

// Stop stops the cache cleanup goroutine
func (mc *MessageCache) Stop() {
	mc.cleanup.Stop()
	close(mc.stopChan)
}

// evictOldest removes the oldest entries from the cache
func (mc *MessageCache) evictOldest() {
	// Simple eviction: remove 10% of entries
	toRemove := mc.maxSize / 10
	if toRemove < 1 {
		toRemove = 1
	}
	
	count := 0
	for key := range mc.cache {
		if count >= toRemove {
			break
		}
		delete(mc.cache, key)
		count++
	}
}

// cleanupExpired removes expired entries from the cache
func (mc *MessageCache) cleanupExpired() {
	for {
		select {
		case <-mc.cleanup.C:
			mc.mu.Lock()
			now := time.Now()
			expiredKeys := make([]string, 0)
			
			for key, entry := range mc.cache {
				if now.After(entry.expiredAt) {
					expiredKeys = append(expiredKeys, key)
				}
			}
			
			for _, key := range expiredKeys {
				delete(mc.cache, key)
			}
			
			if len(expiredKeys) > 0 {
				logrus.WithField("expired_count", len(expiredKeys)).Debug("Cleaned up expired cache entries")
			}
			
			mc.mu.Unlock()
			
		case <-mc.stopChan:
			return
		}
	}
}

// ConnectionPool manages database connections for high performance
type ConnectionPool struct {
	mu          sync.RWMutex
	connections chan interface{} // Generic connection interface
	maxSize     int
	currentSize int
	createFunc  func() (interface{}, error)
	closeFunc   func(interface{}) error
	validateFunc func(interface{}) bool
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(
	maxSize int,
	createFunc func() (interface{}, error),
	closeFunc func(interface{}) error,
	validateFunc func(interface{}) bool,
) *ConnectionPool {
	return &ConnectionPool{
		connections:  make(chan interface{}, maxSize),
		maxSize:      maxSize,
		createFunc:   createFunc,
		closeFunc:    closeFunc,
		validateFunc: validateFunc,
	}
}

// Get retrieves a connection from the pool
func (cp *ConnectionPool) Get() (interface{}, error) {
	select {
	case conn := <-cp.connections:
		// Validate connection
		if cp.validateFunc != nil && !cp.validateFunc(conn) {
			// Connection is invalid, create a new one
			if cp.closeFunc != nil {
				cp.closeFunc(conn)
			}
			return cp.createFunc()
		}
		return conn, nil
	default:
		// No available connections, create a new one
		return cp.createFunc()
	}
}

// Put returns a connection to the pool
func (cp *ConnectionPool) Put(conn interface{}) {
	if conn == nil {
		return
	}
	
	// Validate connection before putting back
	if cp.validateFunc != nil && !cp.validateFunc(conn) {
		if cp.closeFunc != nil {
			cp.closeFunc(conn)
		}
		return
	}
	
	select {
	case cp.connections <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close the connection
		if cp.closeFunc != nil {
			cp.closeFunc(conn)
		}
	}
}

// Close closes all connections in the pool
func (cp *ConnectionPool) Close() {
	close(cp.connections)
	
	if cp.closeFunc != nil {
		for conn := range cp.connections {
			cp.closeFunc(conn)
		}
	}
}

// GetStats returns pool statistics
func (cp *ConnectionPool) GetStats() map[string]interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	
	return map[string]interface{}{
		"max_size":     cp.maxSize,
		"current_size": cp.currentSize,
		"available":    len(cp.connections),
		"in_use":       cp.currentSize - len(cp.connections),
	}
}